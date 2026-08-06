package authenticator

import (
	"fmt"

	mauthconfig "github.com/aliyun/alibabacloud-secretsmanager-client-go-v2/sdk/mauth/config"
	"github.com/aliyun/alibabacloud-secretsmanager-client-go-v2/sdk/mauth/utils"
	idaasconfig "github.com/cloud-idaas/idaas-go-core-sdk/config"
)

const (
	IdentityTypePKCS7 = "PKCS7"
	IdentityTypeJWT   = "JWT"

	DefaultACKTokenPath = "/var/run/secrets/kubernetes.io/serviceaccount/token"
)

type Config struct {
	// Must
	AuthMethod string
	// ACKOidcJwt config
	TokenPath string
	// IDaaS config
	IDaaSConfigPath   string
	IDaaSClientConfig *idaasconfig.IDaaSClientConfig
}

// ValidateConfig 验证配置参数是否合法
func (c *Config) ValidateConfig() error {
	switch c.AuthMethod {
	case mauthconfig.ECSInstanceIdentity:
	case mauthconfig.ACKOidcJwt:
		// // 检查 JWT 文件是否存在且格式合法
		if c.TokenPath != "" {
			if !utils.IsValidJWTFile(c.TokenPath) {
				return fmt.Errorf("token file does not exist or has invalid JWT format: %s", c.TokenPath)
			}
		}
	case mauthconfig.AwsEc2PKCS7, mauthconfig.AwsEksOIDC, mauthconfig.GcpVmOIDC, mauthconfig.GcpGkeOIDC, mauthconfig.AzureVmOIDC, mauthconfig.AzureAksOIDC, mauthconfig.GenericKubernetesOIDC:
		if c.IDaaSClientConfig != nil {
			// 直接传了配置对象，跳过文件校验
			break
		}
		if c.IDaaSConfigPath != "" {
			if !utils.IsJSONFile(c.IDaaSConfigPath) {
				return fmt.Errorf("idaas config file does not exist or has invalid JSON format: %s", c.IDaaSConfigPath)
			}
		}
	default:
		return fmt.Errorf("unsupported authentication method: %s", c.AuthMethod)
	}
	return nil
}
