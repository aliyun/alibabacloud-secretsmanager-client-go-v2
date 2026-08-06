# Alibaba Cloud Secrets Manager V2 Client Profile Settings

Build the client credentials with the configuration file (secretsmanager.properties) in the directory where the program runs:

1. Use Alibaba Cloud AK SK to access Alibaba Cloud KMS, you must set the following configuration variables

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

2. Use ECS RAM role to access Alibaba Cloud KMS, you must set the following configuration variables

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

3. Use OIDC Role ARN to access Alibaba Cloud KMS, you must set the following configuration variables

```properties
# the type of access credentials
credentials_type=oidc_role_arn
# role arn (optional, if not set, the default Alibaba Cloud credential chain will be used)
credentials_role_arn=<role arn>
# OIDC provider arn (optional, if not set, the default Alibaba Cloud credential chain will be used)
credentials_oidc_provider_arn=<oidc provider arn>
# OIDC token file path (optional, if not set, the default Alibaba Cloud credential chain will be used)
credentials_oidc_token_file_path=<oidc token file path>
# the region information
cache_client_region_id=[{"regionId":"<regionId>"}]
# When accessing KMS instance gateway, use the following configuration
# cache_client_region_id=[{"regionId":"<regionId>","endpoint":"<you kms instanceId>.cryptoservice.kms.aliyuncs.com"}]
```

4. Use Cloud Native ECS Instance Identity to access Alibaba Cloud KMS, you must set the following configuration variables

```properties
# the type of access credentials
credentials_type=ecs_instance_identity
# AAP ARN
aap_arn=<aap arn>
# the region information
cache_client_region_id=[{"regionId":"<regionId>"}]
# When accessing KMS instance gateway, use the following configuration
# cache_client_region_id=[{"regionId":"<regionId>","endpoint":"<you kms instanceId>.cryptoservice.kms.aliyuncs.com"}]
```

5. Use Cloud Native ACK OIDC JWT to access Alibaba Cloud KMS, you must set the following configuration variables

```properties
# the type of access credentials
credentials_type=ack_oidc_jwt
# AAP ARN
aap_arn=<aap arn>
# OIDC token file path
token_path=<oidc token file path>
# the region information
cache_client_region_id=[{"regionId":"<regionId>"}]
# When accessing KMS instance gateway, use the following configuration
# cache_client_region_id=[{"regionId":"<regionId>","endpoint":"<you kms instanceId>.cryptoservice.kms.aliyuncs.com"}]
```

6. Use Cloud Native ClientKey to access Alibaba Cloud KMS, you must set the following configuration variables

```properties
# the type of access credentials
credentials_type=client_key
# ClientKey config file path
client_key_private_key_path=<client key config path>
# ClientKey password environment variable name (choose one with client_key_password_from_file_path)
client_key_password_from_env_variable=<env name>
# ClientKey password file path (choose one with client_key_password_from_env_variable)
# client_key_password_from_file_path=<password file path>
# the region information
cache_client_region_id=[{"regionId":"<regionId>"}]
# When accessing KMS instance gateway, use the following configuration
# cache_client_region_id=[{"regionId":"<regionId>","endpoint":"<you kms instanceId>.cryptoservice.kms.aliyuncs.com"}]
```

7. Use Multi-Cloud IDaaS to access Alibaba Cloud KMS, you must set the following configuration variables

```properties
# the type of access credentials, supported values: aws_ec2_pkcs7, aws_eks_oidc, gcp_vm_oidc, gcp_gke_oidc, azure_vm_oidc, azure_aks_oidc, generic_kubernetes_oidc
credentials_type=aws_ec2_pkcs7
# IDaaS client-config.json file path
idaas_config_path=<idaas config path>
# the Application Access Point ARN
aap_arn=<aap arn>
# the region information
cache_client_region_id=[{"regionId":"<regionId>"}]
# When accessing KMS instance gateway, use the following configuration
# cache_client_region_id=[{"regionId":"<regionId>","endpoint":"<you kms instanceId>.cryptoservice.kms.aliyuncs.com"}]
```