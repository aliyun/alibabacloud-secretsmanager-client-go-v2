# 阿里云托管凭据客户端v2配置文件设置

在程序运行目录下，通过配置文件（secretsmanager.properties）构建客户端：

1. 采用阿里云AK SK作为访问鉴权方式

```properties
# 访问凭据类型
credentials_type=ak
# AK
credentials_access_key_id=<access key id>
# SK
credentials_access_secret=<access key secret>
# 关联的KMS服务地域
cache_client_region_id=[{"regionId":"<regionId>"}]
# 访问KMS实例网关时，使用如下配置
# cache_client_region_id=[{"regionId":"<regionId>","endpoint":"<you kms instanceId>.cryptoservice.kms.aliyuncs.com"}]
```

2. 采用阿里云ECS Ram Role作为访问鉴权方式

```properties
# 访问凭据类型
credentials_type=ecs_ram_role
# ECS RAM Role名称
credentials_role_name=<your role name>
# 关联的KMS服务地域
cache_client_region_id=[{"regionId":"<regionId>"}]
# 访问KMS实例网关时，使用如下配置
# cache_client_region_id=[{"regionId":"<regionId>","endpoint":"<you kms instanceId>.cryptoservice.kms.aliyuncs.com"}]
```

3. 采用阿里云OIDC Role ARN作为访问鉴权方式

```properties
# 访问凭据类型
credentials_type=oidc_role_arn
# 角色ARN (可选，不填则使用阿里云默认凭据链)
credentials_role_arn=<role arn>
# OIDC提供者ARN (可选，不填则使用阿里云默认凭据链)
credentials_oidc_provider_arn=<oidc provider arn>
# OIDC令牌文件路径 (可选，不填则使用阿里云默认凭据链)
credentials_oidc_token_file_path=<oidc token file path>
# 关联的KMS服务地域
cache_client_region_id=[{"regionId":"<regionId>"}]
# 访问KMS实例网关时，使用如下配置
# cache_client_region_id=[{"regionId":"<regionId>","endpoint":"<you kms instanceId>.cryptoservice.kms.aliyuncs.com"}]
```

4. 采用云原生 ECS Instance Identity作为访问鉴权方式

```properties
# 访问凭据类型
credentials_type=ecs_instance_identity
# AAP ARN
aap_arn=<aap arn>
# 关联的KMS服务地域
cache_client_region_id=[{"regionId":"<regionId>"}]
# 访问KMS实例网关时，使用如下配置
# cache_client_region_id=[{"regionId":"<regionId>","endpoint":"<you kms instanceId>.cryptoservice.kms.aliyuncs.com"}]
```

5. 采用云原生 ACK OIDC JWT作为访问鉴权方式

```properties
# 访问凭据类型
credentials_type=ack_oidc_jwt
# AAP ARN
aap_arn=<aap arn>
# OIDC令牌文件路径
token_path=<oidc token file path>
# 关联的KMS服务地域
cache_client_region_id=[{"regionId":"<regionId>"}]
# 访问KMS实例网关时，使用如下配置
# cache_client_region_id=[{"regionId":"<regionId>","endpoint":"<you kms instanceId>.cryptoservice.kms.aliyuncs.com"}]
```

6. 采用云原生 ClientKey作为访问鉴权方式

```properties
# 访问凭据类型
credentials_type=client_key
# ClientKey配置文件路径
client_key_private_key_path=<client key config path>
# ClientKey密码环境变量名称（与client_key_password_from_file_path二选一）
client_key_password_from_env_variable=<env name>
# ClientKey密码文件路径（与client_key_password_from_env_variable二选一）
# client_key_password_from_file_path=<password file path>
# 关联的KMS服务地域
cache_client_region_id=[{"regionId":"<regionId>"}]
# 访问KMS实例网关时，使用如下配置
# cache_client_region_id=[{"regionId":"<regionId>","endpoint":"<you kms instanceId>.cryptoservice.kms.aliyuncs.com"}]
```

7. 采用多云 IDaaS作为访问鉴权方式

```properties
# 访问凭据类型，可选值：aws_ec2_pkcs7、aws_eks_oidc、gcp_vm_oidc、gcp_gke_oidc、azure_vm_oidc、azure_aks_oidc、generic_kubernetes_oidc
credentials_type=aws_ec2_pkcs7
# IDaaS client-config.json 配置文件路径
idaas_config_path=<idaas config path>
# 应用访问点 ARN
aap_arn=<aap arn>
# 关联的KMS服务地域
cache_client_region_id=[{"regionId":"<regionId>"}]
# 访问KMS实例网关时，使用如下配置
# cache_client_region_id=[{"regionId":"<regionId>","endpoint":"<you kms instanceId>.cryptoservice.kms.aliyuncs.com"}]
```