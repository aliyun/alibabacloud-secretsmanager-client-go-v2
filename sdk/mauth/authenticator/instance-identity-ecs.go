package authenticator

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type ECSAuthenticator struct{}

// GetIdentity GetIdentityInfo 包含过期时间
func (a *ECSAuthenticator) GetIdentity(ctx context.Context) (*Identity, error) {
	// 缓存未命中或已过期，获取新的PKCS7文档
	token, err := a.getMetadataToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get metadata token: %v", err)
	}

	// audience
	audience := fmt.Sprintf("%d", time.Now().Unix())
	// 带上 Token 获取 PKCS7
	url := fmt.Sprintf("http://100.100.100.200/latest/dynamic/instance-identity/pkcs7?audience=%s", audience)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-aliyun-ecs-metadata-token", token)

	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get metadata pkcs7: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get metadata pkcs7: %s", resp.Status)
	}

	document, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata pkcs7: %v", err)
	}

	type ECSRecipient struct {
		Document string
	}
	ecsRecipient := &ECSRecipient{Document: strings.TrimSpace(string(document))}
	doc, err := json.Marshal(ecsRecipient)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal ecsRecipient: %v", err)
	}

	identity := &Identity{
		IdentityType:  IdentityTypePKCS7,
		IdentityValue: string(doc),
	}

	return identity, nil
}

// 获取 Metadata Token 的方法
func (a *ECSAuthenticator) getMetadataToken() (string, error) {
	client := &http.Client{Timeout: 2 * time.Second}
	req, err := http.NewRequest("PUT", "http://100.100.100.200/latest/api/token", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("X-aliyun-ecs-metadata-token-ttl-seconds", "3600")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get metadata token: %s", resp.Status)
	}

	token, err := io.ReadAll(resp.Body)
	return string(token), err
}
