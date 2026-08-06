package utils

const (
	// StageAcsCurrent 当前stage
	StageAcsCurrent = "ACSCurrent"

	// IvLength 随机IV字节长度
	IvLength = 16

	// RandomKeyLength 随机密钥字节长度
	RandomKeyLength = 32

	// DefaultRetryMaxAttempts 默认最大重试次数
	DefaultRetryMaxAttempts = 5

	// DefaultRetryInitialIntervalMills 默认重试间隔时间(毫秒)
	DefaultRetryInitialIntervalMills = 2000

	// DefaultCapacity 默认最大等待时间
	DefaultCapacity = 10000

	// RequestWaitingTime 请求等待时间(毫秒)
	RequestWaitingTime = 2 * 60 * 1000

	// DefaultProtocol 默认协议
	DefaultProtocol = "https"

	// SdkReadTimeout KMS服务Socket连接超时错误码
	SdkReadTimeout = "connect timed out"

	// ErrorCodeForbiddenInDebtOverDue TeaException 欠费errorCode
	ErrorCodeForbiddenInDebtOverDue = "Forbidden.InDebtOverdue"

	// ErrorCodeForbiddenInDebt TeaException 欠费errorCode
	ErrorCodeForbiddenInDebt = "Forbidden.InDebt"

	// VariableCacheClientRegionIdKey 地域ID配置键名
	VariableCacheClientRegionIdKey = "cache_client_region_id"

	// VariableCredentialsTypeKey 凭据类型配置键名
	VariableCredentialsTypeKey = "credentials_type"

	// VariableCredentialsTypeAccessKey AccessKey 凭据类型值
	VariableCredentialsTypeAccessKey = "ak"

	// VariableCredentialsTypeEcsRamRole ECS RAM Role 凭据类型值
	VariableCredentialsTypeEcsRamRole = "ecs_ram_role"

	// VariableCredentialsTypeOidcRoleArn OIDC Role ARN 凭据类型值
	VariableCredentialsTypeOidcRoleArn = "oidc_role_arn"

	// VariableCredentialsTypeClientKey mauth ClientKey 凭据类型值
	VariableCredentialsTypeClientKey = "client_key"

	// VariableCredentialsTypeAckOidcJwt mauth ACK OIDC JWT 凭据类型值
	VariableCredentialsTypeAckOidcJwt = "ack_oidc_jwt"

	// VariableCredentialsTypeEcsInstanceIdentity mauth ECS Instance Identity 凭据类型值
	VariableCredentialsTypeEcsInstanceIdentity = "ecs_instance_identity"

	// VariableCredentialsTypeAwsEc2PKCS7 AWS EC2 PKCS7 凭据类型值
	VariableCredentialsTypeAwsEc2PKCS7 = "aws_ec2_pkcs7"
	// VariableCredentialsTypeAwsEksOIDC AWS EKS OIDC 凭据类型值
	VariableCredentialsTypeAwsEksOIDC = "aws_eks_oidc"
	// VariableCredentialsTypeGcpVmOIDC GCP VM OIDC 凭据类型值
	VariableCredentialsTypeGcpVmOIDC = "gcp_vm_oidc"
	// VariableCredentialsTypeGcpGkeOIDC GCP GKE OIDC 凭据类型值
	VariableCredentialsTypeGcpGkeOIDC = "gcp_gke_oidc"
	// VariableCredentialsTypeAzureVmOIDC Azure VM OIDC 凭据类型值
	VariableCredentialsTypeAzureVmOIDC = "azure_vm_oidc"
	// VariableCredentialsTypeAzureAksOIDC Azure AKS OIDC 凭据类型值
	VariableCredentialsTypeAzureAksOIDC = "azure_aks_oidc"
	// VariableCredentialsTypeGenericKubernetesOIDC Generic Kubernetes OIDC 凭据类型值
	VariableCredentialsTypeGenericKubernetesOIDC = "generic_kubernetes_oidc"

	// VariableCredentialsAccessKeyIdKey AccessKey ID配置键名
	VariableCredentialsAccessKeyIdKey = "credentials_access_key_id"

	// VariableCredentialsAccessSecretKey AccessKey Secret配置键名
	VariableCredentialsAccessSecretKey = "credentials_access_secret"

	// VariableCredentialsRoleNameKey 角色名称配置键名
	VariableCredentialsRoleNameKey = "credentials_role_name"

	// VariableCredentialsOidcDurationSecondsKey OIDC角色会话过期时间配置键名
	VariableCredentialsOidcDurationSecondsKey = "credentials_duration_seconds"

	// VariableCredentialsOidcProviderArnKey OIDC凭证ARN配置键名
	VariableCredentialsOidcProviderArnKey = "credentials_oidc_provider_arn"

	// VariableCredentialsOidcTokenFilePathKey OIDC令牌文件路径配置键名
	VariableCredentialsOidcTokenFilePathKey = "credentials_oidc_token_file_path"

	// VariableCredentialsOidcRoleSessionNameKey OIDC角色会话名称配置键名
	VariableCredentialsOidcRoleSessionNameKey = "credentials_role_session_name"

	// VariableCredentialsOidcRoleArnKey OIDC角色ARN配置键名
	VariableCredentialsOidcRoleArnKey = "credentials_role_arn"

	// VariableCredentialsOidcPolicyKey OIDC策略配置键名
	VariableCredentialsOidcPolicyKey = "credentials_policy"

	// VariableCredentialsOidcStsEndpointKey OIDC sts域名配置键名
	VariableCredentialsOidcStsEndpointKey = "credentials_sts_endpoint"

	// VariableRegionEndpointNameKey 地域域名配置键名
	VariableRegionEndpointNameKey = "endpoint"

	// VariableRegionRegionIdNameKey 地域ID配置键名
	VariableRegionRegionIdNameKey = "regionId"

	// VariableRegionVpcNameKey VPC配置键名
	VariableRegionVpcNameKey = "vpc"

	// VariableRegionCaFilePathNameKey CA文件路径配置键名
	VariableRegionCaFilePathNameKey = "caFilePath"

	// VariableCredentialsAapArnKey mauth AAP ARN配置键名
	VariableCredentialsAapArnKey = "aap_arn"

	// VariableCredentialsTokenPathKey mauth令牌路径配置键名
	VariableCredentialsTokenPathKey = "token_path"

	// VariableCredentialsClientKeyConfigPathKey mauth ClientKey配置文件路径配置键名
	VariableCredentialsClientKeyConfigPathKey = "client_key_private_key_path"

	// VariableCredentialsClientKeyPasswordKey mauth ClientKey密码配置键名
	VariableCredentialsClientKeyPasswordKey = "client_key_password"

	// VariableCredentialsClientKeyPasswordEnvNameKey mauth ClientKey密码环境变量名称配置键名
	VariableCredentialsClientKeyPasswordEnvNameKey = "client_key_password_from_env_variable"

	// VariableCredentialsClientKeyPasswordPathKey mauth ClientKey密码文件路径配置键名
	VariableCredentialsClientKeyPasswordPathKey = "client_key_password_from_file_path"

	// VariableCredentialsIDaaSConfigPathKey IDaaS 原生配置文件路径配置键名
	VariableCredentialsIDaaSConfigPathKey = "idaas_config_path"

	// TextDataType 凭据文本数据类型
	TextDataType = "text"

	// BinaryDataType 凭据二进制数据类型
	BinaryDataType = "binary"

	// ModeName 模块名称
	ModeName = "SecretsManagerClientV2"

	// ProjectVersion 项目版本
	ProjectVersion = "2.0.2"

	// CredentialsPropertiesConfigName 凭据配置文件名称
	CredentialsPropertiesConfigName = "secretsmanager.properties"

	// SourceTypeConfig 配置来源类型标识
	SourceTypeConfig = "config"

	// SourceTypeEnv 环境变量来源类型标识
	SourceTypeEnv = "env"

	// ErrorCodeInvalidAccessKeyIdNotFound 无效的AccessKeyId错误码
	ErrorCodeInvalidAccessKeyIdNotFound = "InvalidAccessKeyId.NotFound"

	// CheckParamErrorMessage 凭据参数缺失错误信息模板
	CheckParamErrorMessage = "%s credentials missing required parameters[%s]"

	// UserAgentOfSecretsManagerV2Go UserAgentOfSecretsManagerGo Secrets Manager Client V2 Go的User Agent
	UserAgentOfSecretsManagerV2Go = "alibabacloud-secretsmanager-client-go-v2"

	// InstanceGatewayDomainSuffix 实例网关域名后缀
	InstanceGatewayDomainSuffix = "cryptoservice.kms.aliyuncs.com"
)
