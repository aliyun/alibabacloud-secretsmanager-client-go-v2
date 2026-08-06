package authenticator

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIDaaSAuthenticator_GetIdentity_Success(t *testing.T) {
	// JWT.io 示例 token，格式有效（不验证签名）
	validJWT := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"

	originalFunc := getOauthTokenFunc
	getOauthTokenFunc = func(configPath string) (string, error) {
		return validJWT, nil
	}
	defer func() { getOauthTokenFunc = originalFunc }()

	auth := &IDaaSAuthenticator{IDaaSConfigPath: "/tmp/test.json"}
	identity, err := auth.GetIdentity(context.Background())

	assert.Nil(t, err)
	assert.NotNil(t, identity)
	assert.Equal(t, IdentityTypeJWT, identity.IdentityType)
	assert.Equal(t, validJWT, identity.IdentityValue)
}

func TestIDaaSAuthenticator_GetIdentity_GetTokenError(t *testing.T) {
	originalFunc := getOauthTokenFunc
	getOauthTokenFunc = func(configPath string) (string, error) {
		return "", errors.New("factory not initialized")
	}
	defer func() { getOauthTokenFunc = originalFunc }()

	auth := &IDaaSAuthenticator{IDaaSConfigPath: "/tmp/test.json"}
	identity, err := auth.GetIdentity(context.Background())

	assert.NotNil(t, err)
	assert.Nil(t, identity)
	assert.Contains(t, err.Error(), "failed to get OAuth token")
}

func TestIDaaSAuthenticator_GetIdentity_InvalidJWT(t *testing.T) {
	originalFunc := getOauthTokenFunc
	getOauthTokenFunc = func(configPath string) (string, error) {
		return "not-a-valid-jwt-token", nil
	}
	defer func() { getOauthTokenFunc = originalFunc }()

	auth := &IDaaSAuthenticator{IDaaSConfigPath: "/tmp/test.json"}
	identity, err := auth.GetIdentity(context.Background())

	assert.NotNil(t, err)
	assert.Nil(t, identity)
	assert.Contains(t, err.Error(), "invalid JWT format")
}
