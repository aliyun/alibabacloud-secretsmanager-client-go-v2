package authenticator

import (
	"fmt"

	"github.com/aliyun/alibabacloud-secretsmanager-client-go-v2/sdk/mauth/utils"
)

const (
	ECSInstanceIdentity = "ECSInstanceIdentity"
	ACKOidcJwt          = "ACKOidcJwt"

	IdentityTypePKCS7 = "PKCS7"
	IdentityTypeJWT   = "JWT"

	DefaultACKTokenPath = "/var/run/secrets/kubernetes.io/serviceaccount/token"
)

type Config struct {
	// Must
	AuthMethod string
	// ACKOidcJwt config
	TokenPath string
}

// ValidateConfig 验证配置参数是否合法
func (c *Config) ValidateConfig() error {
	switch c.AuthMethod {
	case ECSInstanceIdentity:
	case ACKOidcJwt:
		// // 检查 JWT 文件是否存在且格式合法
		if c.TokenPath != "" {
			if !utils.IsValidJWTFile(c.TokenPath) {
				return fmt.Errorf("token file does not exist or has invalid JWT format: %s", c.TokenPath)
			}
		}
	default:
		return fmt.Errorf("unsupported authentication method: %s", c.AuthMethod)
	}
	return nil
}
