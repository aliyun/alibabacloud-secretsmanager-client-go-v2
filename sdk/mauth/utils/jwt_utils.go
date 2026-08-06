package utils

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

// isValidJWT 简单的 JWT 格式校验 (不进行验签，只检查结构)
func IsValidJWT(token string) bool {
	// JWT 必须由两点分隔成三个部分 (Header.Payload.Signature)
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return false
	}

	// 简单的 Base64 解码检查，确保它是合法的 JSON 结构
	for _, part := range parts[:2] {
		// JWT 使用 URL-safe Base64，需要填充 padding
		p := addPadding(part)
		decoded, err := base64.URLEncoding.DecodeString(p)
		if err != nil {
			return false
		}
		// 增加 JSON 结构体检查
		var js map[string]interface{}
		if err := json.Unmarshal(decoded, &js); err != nil {
			return false // 如果不是有效的 JSON，那肯定不是 JWT
		}
	}
	return true
}

func addPadding(s string) string {
	l := len(s) % 4
	if l == 2 {
		s += "=="
	} else if l == 3 {
		s += "="
	}
	return s
}

// ParseJWTClaims 解析 JWT 的 Claims 部分
func ParseJWTClaims(token string) (map[string]interface{}, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid JWT format: not 3 parts")
	}

	payload := parts[1]

	// 添加必要的 padding
	paddedPayload := addPadding(payload)

	// 解码 payload 部分
	decoded, err := base64.URLEncoding.DecodeString(paddedPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to decode JWT payload: %v", err)
	}

	// 解析 JSON
	var claims map[string]interface{}
	err = json.Unmarshal(decoded, &claims)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JWT claims: %v", err)
	}

	return claims, nil
}

// IsValidJWTFile 检查JWT文件是否存在且格式有效
func IsValidJWTFile(filePath string) bool {
	// 检查文件是否存在
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false
	}

	// 读取文件内容
	tokenBytes, err := os.ReadFile(filePath)
	if err != nil {
		return false
	}

	token := strings.TrimSpace(string(tokenBytes))
	return IsValidJWT(token)
}

// 校验是不是 json
func IsJSONFile(filePath string) bool {
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false
	}
	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		return false
	}
	var js map[string]interface{}
	err = json.Unmarshal(fileBytes, &js)
	if err != nil {
		return false
	}
	return true
}

// GetJWTExpiration 获取 JWT 的过期时间
func GetJWTExpiration(token string) (time.Time, error) {
	claims, err := ParseJWTClaims(token)
	if err != nil {
		return time.Time{}, err
	}

	// 检查 exp 字段是否存在
	expInterface, exists := claims["exp"]
	if !exists {
		return time.Time{}, fmt.Errorf("JWT does not contain 'exp' claim")
	}

	// 将 exp 转换为 int64 时间戳
	var expInt int64
	switch v := expInterface.(type) {
	case float64:
		expInt = int64(v)
	case int64:
		expInt = v
	case int:
		expInt = int64(v)
	case json.Number:
		val, err := v.Int64()
		if err != nil {
			return time.Time{}, fmt.Errorf("failed to convert 'exp' claim to int64: %v", err)
		}
		expInt = val
	default:
		return time.Time{}, fmt.Errorf("'exp' claim is not a number: %T", expInterface)
	}

	// 将 Unix 时间戳转换为 time.Time
	expiration := time.Unix(expInt, 0)
	return expiration, nil
}
