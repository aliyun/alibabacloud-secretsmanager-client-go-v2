package exchanger

import (
	"context"
	"encoding/json"
	"fmt"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	openapiutil "github.com/alibabacloud-go/openapi-util/service"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/aliyun/credentials-go/credentials"
)

type KMSExchanger struct {
	kmsEndpoint string
	aapArn      string
	ca          string
}

type TemporaryCredential struct {
	CredentialType string
	Credential     string
	ExpiredAt      string
	Region         string
	RequestId      string
}

const (
	ActionIssueTemporaryCredential = "IssueTemporaryCredential"
	argAapArn                      = "AapArn"
)

func (kms *KMSExchanger) ExchangeCredential(ctx context.Context, bearerToken string) (string, string, error) {
	credentialsConfig := new(credentials.Config).
		SetType("bearer").
		SetBearerToken(bearerToken)
	bearerCredential, err := credentials.NewCredential(credentialsConfig)
	if err != nil {
		return "", "", fmt.Errorf("failed to create aliyun sdk credential: %v", err)
	}
	config := &openapi.Config{
		Endpoint:       tea.String(kms.kmsEndpoint),
		ReadTimeout:    tea.Int(3 * 1000),
		ConnectTimeout: tea.Int(3 * 1000),
		Protocol:       tea.String("https"),
	}
	if kms.ca != "" {
		config.Ca = tea.String(kms.ca)
	}
	config.Credential = bearerCredential
	client, err := openapi.NewClient(config)
	if err != nil {
		return "", "", fmt.Errorf("failed to create client: %v", err)
	}
	params := &openapi.Params{
		// 设置API的行动、版本和其他必要参数
		Action:      tea.String(ActionIssueTemporaryCredential), // API名称
		Version:     tea.String("2016-01-20"),                   // API版本号
		Protocol:    tea.String("HTTPS"),                        // 请求协议：HTTPS或HTTP，建议使用HTTPS。
		Method:      tea.String("POST"),                         // 请求方法
		AuthType:    tea.String("bearer"),                       // 认证类型，默认即可。当OpenAPI支持匿名请求时，您可以传入 Anonymous 发起匿名请求。
		Style:       tea.String("RPC"),                          // API风格：RPC、ROA
		Pathname:    tea.String("/"),                            // 接口 PATH，RPC接口默认"/"
		ReqBodyType: tea.String("json"),                         // 接口请求体内容格式。
		BodyType:    tea.String("json"),                         // 接口响应体内容格式。
	}

	// 设置查询参数
	var aapArn string
	val := ctx.Value("AapArn")
	if strVal, ok := val.(string); ok && strVal != "" {
		aapArn = strVal
	} else {
		aapArn = kms.aapArn
	}

	query := map[string]interface{}{
		argAapArn: tea.String(aapArn),
	}
	// 设置运行时选项
	runtime := &util.RuntimeOptions{}
	// 创建API请求并设置参数
	request := &openapi.OpenApiRequest{
		Query: openapiutil.Query(query),
	}

	//添加请求头
	request.Headers = map[string]*string{
		"X-Acs-Bearer-Token": tea.String(bearerToken),
	}

	// 调用API并处理返回结果
	response, err := client.CallApi(params, request, runtime)
	if err != nil {
		return "", "", fmt.Errorf("failed to call api: %v", err)
	}
	// 返回值为Map类型，可从Map中获得三类数据：body、headers、HTTP返回的状态码 statusCode。
	body := response["body"]
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal body: %v", err)
	}

	var ret TemporaryCredential
	err = json.Unmarshal(bodyBytes, &ret)
	if err != nil {
		return "", "", fmt.Errorf("failed to unmarshal body: %v", err)
	}

	return ret.CredentialType, ret.Credential, nil
}
