package authenticator

import (
	"context"
	"fmt"
	"time"

	mauthconfig "github.com/aliyun/alibabacloud-secretsmanager-client-go-v2/sdk/mauth/config"
)

type Identity struct {
	IdentityType  string
	IdentityValue string
	ExpiresAt     time.Time
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
	case mauthconfig.ACKOidcJwt:
		tokenPath := config.TokenPath
		// 默认 ack token path
		if tokenPath == "" {
			tokenPath = DefaultACKTokenPath
		}
		authenticator = &AckOIDCAuthenticator{
			TokenPath: tokenPath,
		}
	case mauthconfig.ECSInstanceIdentity:
		authenticator = &ECSAuthenticator{}
	case mauthconfig.AwsEc2PKCS7, mauthconfig.AwsEksOIDC, mauthconfig.GcpVmOIDC, mauthconfig.GcpGkeOIDC, mauthconfig.AzureVmOIDC, mauthconfig.AzureAksOIDC, mauthconfig.GenericKubernetesOIDC:
		auth := &IDaaSAuthenticator{IDaaSConfigPath: config.IDaaSConfigPath}
		if config.IDaaSClientConfig != nil {
			auth.getTokenFunc = func() (string, error) {
				return getOauthTokenFromConfig(config.IDaaSClientConfig)
			}
		}
		authenticator = auth
	default:
		return nil, fmt.Errorf("unsupported authentication method: %s", config.AuthMethod)
	}

	return authenticator, nil
}
