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