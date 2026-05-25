# System Environment Variables Setting For Alibaba Secrets Manager Client V2

Use Alibaba Secrets Manager client v2 by system environment variables with the below ways:

* Use access key to access aliyun kms, you must set the following system environment variables (for linux):

	- export credentials_type=ak
	- export credentials_access_key_id=\<your access key id>
	- export credentials_access_secret=\<your access key secret>
	- export cache_client_region_id=[{"regionId":"\<your region id>"}]
```
tips:
 	When accessing KMS instance gateway, use the following configuration
	export cache_client_region_id=[{"regionId":"<your region id>","endpoint":"<your kms instanceId>.cryptoservice.kms.aliyuncs.com"}]
```

* Use ECS RAM role to access aliyun kms, you must set the following system environment variables (for linux):

	- export credentials_type=ecs_ram_role
	- export credentials_role_name=\<role name>
	- export cache_client_region_id=[{"regionId":"\<your region id>"}]
```
tips:
 	When accessing KMS instance gateway, use the following configuration
	export cache_client_region_id=[{"regionId":"<your region id>","endpoint":"<your kms instanceId>.cryptoservice.kms.aliyuncs.com"}]
```

* Use OIDC Role ARN to access aliyun kms, you must set the following system environment variables (for linux):

	- export credentials_type=oidc_role_arn
	- export credentials_role_arn=\<role arn> (optional, if not set, the default Aliyun credential chain will be used)
	- export credentials_oidc_provider_arn=\<oidc provider arn> (optional, if not set, the default Aliyun credential chain will be used)
	- export credentials_oidc_token_file_path=\<oidc token file path> (optional, if not set, the default Aliyun credential chain will be used)
	- export cache_client_region_id=[{"regionId":"\<your region id>"}]
```
tips:
 	When accessing KMS instance gateway, use the following configuration
	export cache_client_region_id=[{"regionId":"<your region id>","endpoint":"<your kms instanceId>.cryptoservice.kms.aliyuncs.com"}]
```

* Use mauth ECS Instance Identity to access aliyun kms, you must set the following system environment variables (for linux):

	- export credentials_type=ecs_instance_identity
	- export aap_arn=\<aap arn>
	- export cache_client_region_id=[{"regionId":"\<your region id>"}]
```
tips:
 	When accessing KMS instance gateway, use the following configuration
	export cache_client_region_id=[{"regionId":"<your region id>","endpoint":"<your kms instanceId>.cryptoservice.kms.aliyuncs.com"}]
```

* Use mauth ACK OIDC JWT to access aliyun kms, you must set the following system environment variables (for linux):

	- export credentials_type=ack_oidc_jwt
	- export aap_arn=\<aap arn>
	- export token_path=\<oidc token file path>
	- export cache_client_region_id=[{"regionId":"\<your region id>"}]
```
tips:
 	When accessing KMS instance gateway, use the following configuration
	export cache_client_region_id=[{"regionId":"<your region id>","endpoint":"<your kms instanceId>.cryptoservice.kms.aliyuncs.com"}]
```

* Use mauth ClientKey to access aliyun kms, you must set the following system environment variables (for linux):

	- export credentials_type=client_key
	- export client_key_private_key_path=\<client key config path>
	- export client_key_password_from_file_path=\<password file path> (choose one with client_key_password_from_env_variable)
  # - export client_key_password_from_env_variable=\<env name> (choose one with client_key_password_from_file_path)
	- export cache_client_region_id=[{"regionId":"\<your region id>"}]
```
tips:
 	When accessing KMS instance gateway, use the following configuration
	export cache_client_region_id=[{"regionId":"<your region id>","endpoint":"<your kms instanceId>.cryptoservice.kms.aliyuncs.com"}]
```