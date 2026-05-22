package mauth

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/aliyun/alibabacloud-secretsmanager-client-go-v2/sdk/mauth/authenticator"
	"github.com/aliyun/alibabacloud-secretsmanager-client-go-v2/sdk/mauth/config"
	"github.com/aliyun/alibabacloud-secretsmanager-client-go-v2/sdk/mauth/exchanger"
	"github.com/aliyun/alibabacloud-secretsmanager-client-go-v2/sdk/mauth/logger"
	"github.com/aliyun/alibabacloud-secretsmanager-client-go-v2/sdk/mauth/utils"
)

/*
 1. 创建一个 mauth 对象，用于认证，通过 Indentity 置换 token
 2. mauth 对象有 authenticator、exchanger、cache、config、decryptor 构成
*/

const (
	AuthTokenTypeClientKey = "ClientKey"

	MaxRemainingTime = 48
)

type MAuth struct {
	authenticator authenticator.Authenticator
	exchanger     exchanger.Exchanger
	config        config.AuthConfig
	mutex         sync.RWMutex
	cacheToken    map[string]*AuthToken
	lw            logger.Wrapper
	cancel        context.CancelFunc
}

type AuthToken struct {
	TokenType  string
	TokenKeyId string
	TokenValue string
	ExpiresAt  time.Time
}

/*
NewMAuth
使用 auth config 文件，初始化 Mauth
*/
func NewMAuth(c config.AuthConfig, lw logger.Wrapper) (*MAuth, error) {
	if err := c.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	mAuth := &MAuth{
		config: c,
		mutex:  sync.RWMutex{},
	}

	if c.AuthMethod == config.ClientKey {
		return mAuth, nil
	}

	// 创建认证器管理器
	authenticatorManager, err := authenticator.NewAuthenticatorManager(
		authenticator.Config{
			AuthMethod: c.AuthMethod,
			TokenPath:  c.TokenPath,
		})
	if err != nil {
		return nil, fmt.Errorf("failed to create authenticator manager: %w", err)
	}
	mAuth.authenticator = authenticatorManager

	// 创建交换器
	exchangerConfig := exchanger.Config{
		ExchangerType: exchanger.ExchangerTypeKms,
		KmsEndpoint:   c.KmsEndpoint,
		AapArn:        c.AapArn,
		Ca:            c.Ca,
	}

	exchangerObj, err := exchanger.NewExchanger(exchangerConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create exchanger: %w", err)
	}
	mAuth.exchanger = exchangerObj

	if lw != nil {
		mAuth.lw = lw
	} else {
		mAuth.lw = logger.NewDefaultLogger()
	}

	ctx, cancel := context.WithCancel(context.Background())
	mAuth.cancel = cancel

	// 新增协程主动更新缓存，要防止 panic
	go func() {
		defer func() {
			if r := recover(); r != nil {
				mAuth.lw.Error("Panic in RefreshToken goroutine: %v", r)
			}
		}()
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(10 * time.Minute):
				mAuth.RefreshToken()
			}
		}
	}()

	return mAuth, nil
}

/*
GetToken 获取 token
整体逻辑:
服务端签发 72 小时有效期的临时client key, 客户端在使用一天后会重新获取，获取失败也能有 48 小时的容错窗口。 服务端对客户端使用 48 小时内凭据访问进行监控和报警。
流程：
1. 如果缓存的 token 过期时间大于 48 小时，则返回缓存的 token
2. 如果缓存的 token 过期时间小于 48 小时，则重新获取 token
3. 如果缓存的 token 不存在，则重新获取 token
4. 如果获取 token 失败，缓存 token 未过期，则返回缓存的 token
5. 如果 auth method 是 ClientKey, 则调用 GetClientKeyByLocalFile获取
6. 如果 auth method 是 ECSInstanceIdentity、ACKOidcJwt 则调用 GetTokenFromRemote获取
7. 获取失败，假如缓存中存在 未过期的token，则返回缓存的 token
8. 获取成功，则返回 token，缓存 token
*/

func (m *MAuth) GetToken(ctx context.Context) (*AuthToken, error) {

	var aapArn string
	val := ctx.Value("AapArn")
	if strVal, ok := val.(string); ok && strVal != "" {
		aapArn = strVal
	} else {
		aapArn = m.config.AapArn
	}

	if m.config.AuthMethod == config.ClientKey {
		aapArn = "ClientKey"
	}

	m.mutex.RLock()

	// 1. 如果缓存的 token 过期时间大于 48 小时，则返回缓存的 token
	if m.cacheToken != nil && m.cacheToken[aapArn] != nil {
		remainingTime := time.Until(m.cacheToken[aapArn].ExpiresAt)
		if remainingTime > MaxRemainingTime*time.Hour {
			token := m.cacheToken[aapArn]
			m.mutex.RUnlock()
			return token, nil
		}
	}
	m.mutex.RUnlock()

	var newToken *AuthToken
	var err error

	// 4. 如果 auth method 是 ClientKey, 则调用 GetClientKeyByLocalFile获取
	if m.config.AuthMethod == config.ClientKey {
		newToken, err = m.getClientKeyByLocalFile(ctx)
	} else {
		// 5. 如果 auth method 是 ECSInstanceIdentity、ACKOidcJwt 则调用 GetTokenFromRemote获取
		newToken, err = m.getTokenFromRemote(ctx)
	}

	// 6. 获取失败，假如缓存中存在 未过期的token，则返回缓存的 token
	if err != nil {
		m.mutex.RLock()
		defer m.mutex.RUnlock()
		if m.cacheToken != nil && m.cacheToken[aapArn] != nil {
			if time.Now().Before(m.cacheToken[aapArn].ExpiresAt) {
				m.lw.Warn("Using cached token for AAP ARN %s, expires at %v", aapArn, m.cacheToken[aapArn].ExpiresAt)
				return m.cacheToken[aapArn], nil
			}
		}
		return nil, fmt.Errorf("failed to get new token and no valid cached token available: %w", err)
	}

	// 7. 获取成功，则返回 token，缓存 token
	m.mutex.Lock()
	if m.cacheToken == nil {
		m.cacheToken = make(map[string]*AuthToken)
	}
	m.cacheToken[aapArn] = newToken
	m.mutex.Unlock()

	return newToken, nil
}

/*
RefreshToken: 定时刷新缓存的 token
1. 遍历缓存的 token
3. 如果 token 过期时间小于 48 小时，则重新获取 token, 如果 token 过期时间大于 48 小时，则不需要更新
4. 打印出刷新中的异常，以及缓存的 token 必要信息，不能打印token 内容。
*/
func (m *MAuth) RefreshToken() {
	if m.cacheToken == nil {
		m.lw.Info("cacheToken is nil")
		return
	}
	if len(m.cacheToken) == 0 {
		m.lw.Info("cacheToken is empty")
		return
	}
	currentTime := time.Now()
	var toRefresh []string

	m.mutex.RLock()
	for aapArn, token := range m.cacheToken {
		if token.ExpiresAt.Sub(currentTime) <= MaxRemainingTime*time.Hour {
			toRefresh = append(toRefresh, aapArn)
		}
	}
	m.mutex.RUnlock() // 筛选完立即释放锁

	// 3. 在锁外执行网络请求（GetToken）
	for _, aapArn := range toRefresh {
		m.lw.Info("Token for AAP ARN %s needs refresh", aapArn)

		ctx := context.WithValue(context.Background(), "AapArn", aapArn)
		newToken, err := m.GetToken(ctx)
		if err != nil {
			m.lw.Error("Failed to refresh token for %s: %v", aapArn, err)
			continue
		}
		// 检查 token 有效期，看是不是拿了缓存的
		if newToken.ExpiresAt.Sub(currentTime) <= MaxRemainingTime*time.Hour {
			m.lw.Warn("Using cached token for AAP ARN %s, expires at %v", aapArn, newToken.ExpiresAt)
		} else {
			m.lw.Info("Successfully refreshed token for AAP ARN %s, expires at %v", aapArn, newToken.ExpiresAt)
		}
	}
}

/*
关闭MAuth，停止后台goroutine
*/
func (m *MAuth) Close() {
	if m.cancel != nil {
		m.cancel()
	}
}

/*
清除Token 缓存
*/
func (m *MAuth) ClearToken(ctx context.Context) error {
	var aapArn string
	val := ctx.Value("AapArn")
	if strVal, ok := val.(string); ok && strVal != "" {
		aapArn = strVal
	} else {
		aapArn = m.config.AapArn
	}

	if m.config.AuthMethod == config.ClientKey {
		aapArn = "ClientKey"
	}

	m.mutex.Lock()
	if m.cacheToken != nil {
		delete(m.cacheToken, aapArn)
	}
	m.mutex.Unlock()
	return nil
}

/*
GetClientKeyByLocalFile
1. 从本地文件获取客户端密钥和密码
2. 调用 decryptor.DecryptClientKey 解密客户端密钥和密码, 获取私钥
3. 构成 AuthToken，返回
*/
type ClientKeyConfig struct {
	KeyId          string
	PrivateKeyData string
}

func (m *MAuth) getClientKeyByLocalFile(ctx context.Context) (*AuthToken, error) {
	// 1. 从本地文件获取客户端密钥和密码
	if m.config.ClientKeyConfigPath == "" {
		return nil, fmt.Errorf("client key config path is not set")
	}

	// 读取配置文件
	configData, err := os.ReadFile(m.config.ClientKeyConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read client key config file: %w", err)
	}
	var cc ClientKeyConfig
	err = json.Unmarshal(configData, &cc)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal client key config: %w", err)
	}

	// 获取密码：优先使用已配置的密码，否则从密码文件读取
	password := m.config.ClientKeyPassword
	if password == "" {
		if m.config.ClientKeyPasswordPath == "" {
			return nil, fmt.Errorf("client key password or password path is not set")
		}
		passwordData, err := os.ReadFile(m.config.ClientKeyPasswordPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read client key password file: %w", err)
		}
		password = strings.TrimSpace(string(passwordData))
	}

	// 2. 调用 decryptor.DecryptClientKey 解密客户端密钥和密码, 获取私钥
	privateKeyPem, expiresAt, err := utils.ExtractClientKey(password, cc.PrivateKeyData)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt client key: %w", err)
	}

	// 3. 构成 AuthToken，返回
	token := &AuthToken{
		TokenType:  AuthTokenTypeClientKey,
		TokenKeyId: cc.KeyId,
		TokenValue: privateKeyPem,
		ExpiresAt:  expiresAt,
	}

	return token, nil
}

type ClientKeyCredential struct {
	ClientKeyId    string
	Password       string
	PrivateKeyData string
	KeyAlgorithm   string
	KeyOrigin      string
	NotAfter       time.Time
	NotBefore      time.Time
}

func (m *MAuth) getTokenFromRemote(ctx context.Context) (*AuthToken, error) {

	// 通过authenticator获取身份凭证
	identity, err := m.authenticator.GetIdentity(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get identity: %w", err)
	}

	// 通过exchanger交换令牌
	tokenType, tokenValue, err := m.exchanger.ExchangeCredential(ctx, identity.IdentityValue)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange credential: %w", err)
	}

	if tokenType != AuthTokenTypeClientKey {
		return nil, fmt.Errorf("invalid token type: %s", tokenType)
	}

	var cc ClientKeyCredential
	err = json.Unmarshal([]byte(tokenValue), &cc)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal client key credential: %w", err)
	}

	privateKeyPem, expiresAt, err := utils.ExtractClientKey(cc.Password, cc.PrivateKeyData)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt client key: %w", err)
	}

	// 构建并返回AuthToken
	token := &AuthToken{
		TokenType:  tokenType,
		TokenKeyId: cc.ClientKeyId,
		TokenValue: privateKeyPem,
		ExpiresAt:  expiresAt,
	}

	return token, nil
}
