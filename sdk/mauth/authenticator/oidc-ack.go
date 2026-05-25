package authenticator

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aliyun/alibabacloud-secretsmanager-client-go-v2/sdk/mauth/utils"
)

type AckOIDCAuthenticator struct {
	TokenPath string
}

// GetIdentity 获取身份信息详情，包含过期时间
func (a *AckOIDCAuthenticator) GetIdentity(ctx context.Context) (*Identity, error) {

	// 缓存未命中或已过期，从磁盘读取新JWT
	tokenBytes, err := os.ReadFile(a.TokenPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read token file: %v", err)
	}
	token := strings.TrimSpace(string(tokenBytes))

	// 校验 JWT 格式
	if !utils.IsValidJWT(token) {
		return nil, fmt.Errorf("invalid JWT format")
	}

	// 构造身份信息
	identity := &Identity{
		IdentityType:  IdentityTypeJWT,
		IdentityValue: token,
	}
	return identity, nil
}
