package authenticator

import (
	"context"
	"fmt"
)

type Identity struct {
	IdentityType  string
	IdentityValue string
}

type Authenticator interface {
	GetIdentity(ctx context.Context) (*Identity, error)
}

// NewAuthenticatorManager 根据配置创建认证管理器
func NewAuthenticatorManager(config Config) (Authenticator, error) {
	// 首先验证配置
	if err := config.ValidateConfig(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	var authenticator Authenticator

	switch config.AuthMethod {
	case ACKOidcJwt:
		tokenPath := config.TokenPath
		// 默认 ack token path
		if tokenPath == "" {
			tokenPath = DefaultACKTokenPath
		}
		authenticator = &AckOIDCAuthenticator{
			TokenPath: tokenPath,
		}
	case ECSInstanceIdentity:
		authenticator = &ECSAuthenticator{}
	default:
		return nil, fmt.Errorf("unsupported authentication method: %s", config.AuthMethod)
	}

	return authenticator, nil
}
