package authenticator

import (
	"context"
	"fmt"

	"github.com/aliyun/alibabacloud-secretsmanager-client-go-v2/sdk/mauth/utils"
	idaasconfig "github.com/cloud-idaas/idaas-go-core-sdk/config"
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
