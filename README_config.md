# Alibaba Cloud Secrets Manager V2 Client Profile Settings

Build the client credentials with the configuration file (secretsmanager.properties) in the directory where the program runs:

1. Use Aliyun AK SK to access Aliyun KMS, you must set the following configuration variables

```properties
# the type of access credentials
credentials_type=ak
# AK
credentials_access_key_id=<access key id>
# SK
credentials_access_secret=<access key secret>
# the region information
cache_client_region_id=[{"regionId":"<regionId>"}]
# When accessing KMS instance gateway, use the following configuration
# cache_client_region_id=[{"regionId":"<regionId>","endpoint":"<you kms instanceId>.cryptoservice.kms.aliyuncs.com"}]
```

2. Use ECS RAM role to access Aliyun KMS, you must set the following configuration variables

```properties
# the type of access credentials
credentials_type=ecs_ram_role
# ECS RAM Role name
credentials_role_name=<your role name>
# the region information
cache_client_region_id=[{"regionId":"<regionId>"}]
# When accessing KMS instance gateway, use the following configuration
# cache_client_region_id=[{"regionId":"<regionId>","endpoint":"<you kms instanceId>.cryptoservice.kms.aliyuncs.com"}]
```

3. Use OIDC Role ARN to access Aliyun KMS, you must set the following configuration variables

```properties
# the type of access credentials
credentials_type=oidc_role_arn
# role arn (optional, if not set, the default Aliyun credential chain will be used)
credentials_role_arn=<role arn>
# OIDC provider arn (optional, if not set, the default Aliyun credential chain will be used)
credentials_oidc_provider_arn=<oidc provider arn>
# OIDC token file path (optional, if not set, the default Aliyun credential chain will be used)
credentials_oidc_token_file_path=<oidc token file path>
# the region information
cache_client_region_id=[{"regionId":"<regionId>"}]
# When accessing KMS instance gateway, use the following configuration
# cache_client_region_id=[{"regionId":"<regionId>","endpoint":"<you kms instanceId>.cryptoservice.kms.aliyuncs.com"}]
```