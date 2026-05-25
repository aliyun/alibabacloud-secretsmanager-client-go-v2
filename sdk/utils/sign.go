package utils

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/aliyun/alibabacloud-secretsmanager-client-go-v2/sdk/logger"
)

const (
	HeaderDate            = "Date"
	HeaderApiVersion      = "x-kms-apiversion"
	HeaderApiName         = "x-kms-apiname"
	HeaderSignatureMethod = "x-kms-signaturemethod"
	HeaderAccessKeyId     = "x-kms-acccesskeyid"
	AlgoRsaPkcs1Sha256    = "RSA_PKCS1_SHA_256"
	ActionGetSecretValue  = "GetSecretValue"
	ApiVersion20160102    = "2016-01-02"
)

func ExtractSignParams(param interface{}, tokenId string, action string) (queryParams, header map[string]string) {
	// 参数
	queryParams = map[string]string{}
	data, err := json.Marshal(param)
	if err != nil {
		logger.GetCommonLogger(ModeName).Errorf("Error marshaling JSON: %v", err)
		return
	}
	var m map[string]interface{}
	err = json.Unmarshal(data, &m)
	if err != nil {
		logger.GetCommonLogger(ModeName).Errorf("Error unmarshalling JSON: %v", err)
		return
	}
	for k, v := range m {
		if v != nil && fmt.Sprintf("%v", v) != "" {
			queryParams[k] = fmt.Sprintf("%v", v)
		}
	}

	// header
	header = map[string]string{}
	header[HeaderAccessKeyId] = tokenId
	header[HeaderSignatureMethod] = AlgoRsaPkcs1Sha256
	header[HeaderDate] = time.Now().UTC().Format(time.RFC3339)
	header[HeaderApiName] = action
	header[HeaderApiVersion] = ApiVersion20160102
	return
}

func Sign(queryParams, header map[string]string, privateKey string) (string, error) {
	// args 是 args 和 header 里合集
	args := make(map[string]string)
	for k, v := range queryParams {
		args[k] = v
	}
	for k, v := range header {
		args[k] = v
	}
	toSignString := buildStringToSign(args)
	// 签名
	signature, err := Sha256WithRsa(toSignString, privateKey)
	if err != nil {
		return "", err
	}
	return signature, nil
}

func buildStringToSign(args map[string]string) string {
	keys := make([]string, 0, len(args))
	for k := range args {
		keys = append(keys, k)
	}
	//排序
	sort.Strings(keys)
	var buf bytes.Buffer
	for _, key := range keys {
		value := args[key]
		prefix := PercentEncode(key) + "="
		if buf.Len() > 0 {
			buf.WriteByte('&')
		}
		buf.WriteString(prefix)
		buf.WriteString(PercentEncode(value))
	}
	return buf.String()
}

func PercentEncode(s string) string {
	spec := map[string]string{"+": "%20", "*": "%2A", "%7E": "~"}
	s = url.QueryEscape(s)
	for key, value := range spec {
		s = strings.Replace(s, key, value, -1)
	}
	return s
}

func Sha256WithRsa(source, secret string) (signature string, err error) {
	// 1. 解码 PEM 格式的数据
	block, _ := pem.Decode([]byte(secret))
	if block == nil {
		return "", fmt.Errorf("failed to decode PEM block")
	}

	// 解析私钥
	var privateKey interface{}
	privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		privateKey, err = x509.ParsePKCS8PrivateKey(block.Bytes)
	}
	if privateKey == nil {
		return "", fmt.Errorf("failed to parse private key, neither PKCS1 nor PKCS8 format")
	}
	// 3. 类型转换为 RSA 私钥
	rsaPrivateKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return "", fmt.Errorf("not an RSA private key")
	}

	h := crypto.Hash.New(crypto.SHA256)
	h.Write([]byte(source))
	hashed := h.Sum(nil)
	signatureBytes, err := rsa.SignPKCS1v15(rand.Reader, rsaPrivateKey, crypto.SHA256, hashed)
	if err != nil {
		return "", fmt.Errorf("failed to sign data: %v", err)
	}
	return base64.StdEncoding.EncodeToString(signatureBytes), nil
}
