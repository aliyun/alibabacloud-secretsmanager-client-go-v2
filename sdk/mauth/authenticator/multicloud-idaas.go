package authenticator

import (
	"context"
	"fmt"
	"strings"

	"github.com/aliyun/alibabacloud-secretsmanager-client-go-v2/sdk/mauth/utils"
	idaasconfig "github.com/cloud-idaas/idaas-go-core-sdk/config"
	idaasenums "github.com/cloud-idaas/idaas-go-core-sdk/enums"
	"github.com/cloud-idaas/idaas-go-core-sdk/factory"
)

type IDaaSAuthenticator struct {
	IDaaSConfigPath string
	getTokenFunc    func() (string, error)
}

// getOauthTokenFunc 用于支持单元测试 mock
var getOauthTokenFunc = GetOauthToken

func (a *IDaaSAuthenticator) GetIdentity(ctx context.Context) (*Identity, error) {
	var token string
	var err error
	if a.getTokenFunc != nil {
		token, err = a.getTokenFunc()
	} else {
		token, err = getOauthTokenFunc(a.IDaaSConfigPath)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get OAuth token: %v", err)
	}

	if !utils.IsValidJWT(token) {
		return nil, fmt.Errorf("invalid JWT format")
	}

	return &Identity{
		IdentityType:  IdentityTypeJWT,
		IdentityValue: token,
	}, nil
}

func GetOauthToken(idaasConfigPath string) (token string, err error) {
	reader := idaasconfig.NewConfigReader()
	cfg, err := reader.LoadWithPriority(idaasConfigPath)
	if err != nil {
		return "", fmt.Errorf("failed to load config: %v", err)
	}
	return getOauthTokenFromConfig(cfg)
}

func getOauthTokenFromConfig(cfg *idaasconfig.IDaaSClientConfig) (token string, err error) {
	normalizeIDaaSConfig(cfg)

	factoryInstance := factory.GetInstance()
	if !factoryInstance.IsInitialized() {
		err = factoryInstance.Initialize(cfg)
		if err != nil {
			return "", fmt.Errorf("failed to initialize factory: %v", err)
		}
	}

	provider, err := factoryInstance.CreateCredentialProvider()
	if err != nil {
		return "", fmt.Errorf("failed to create credential provider: %v", err)
	}

	cred, err := provider.GetCredential()
	if err != nil {
		return "", fmt.Errorf("failed to get credential: %v", err)
	}

	return cred.GetAccessToken(), nil
}

// normalizeIDaaSConfig 对 IDaaS 配置对象进行大小写归一化。
// 由于 idaas-go-core-sdk 仅在 JSON 反序列化时通过 UnmarshalJSON 做归一化，
// 对象直接赋值方式会跳过该逻辑，导致枚举值大小写不匹配。
// 此函数补齐对象方式的归一化，使行为与文件加载方式保持一致。
func normalizeIDaaSConfig(cfg *idaasconfig.IDaaSClientConfig) {
	if cfg == nil || cfg.AuthnConfiguration == nil {
		return
	}
	authn := cfg.AuthnConfiguration
	if authn.AuthnMethod != "" {
		authn.AuthnMethod = idaasenums.TokenAuthnMethod(strings.ToLower(string(authn.AuthnMethod)))
	}
	if authn.IdentityType != "" {
		authn.IdentityType = idaasenums.AuthenticationIdentityEnum(strings.ToUpper(string(authn.IdentityType)))
	}
	if authn.ClientDeployEnvironment != "" {
		authn.ClientDeployEnvironment = idaasenums.ClientDeployEnvironmentEnum(strings.ToUpper(string(authn.ClientDeployEnvironment)))
	}
}
