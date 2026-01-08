# 阿里云凭据管家客户端V2系统环境变量设置 

通过以下系统环境变量设置方式使用阿里云凭据管家客户端V2：

* 通过使用AK访问KMS，你必须要设置如下系统环境变量 (linux):

	- export credentials\_type=ak
	- export credentials\_access\_key\_id=\<your access key id>
	- export credentials\_access\_secret=\<your access key secret>
	- export cache\_client\_region\_id=[{"regionId":"\<your region id>"}]
```
提示:
	访问KMS实例网关时，使用如下配置
	export cache_client_region_id=[{"regionId":"<your region id>","endpoint":"<your kms instanceId>.cryptoservice.kms.aliyuncs.com"}]
```
* 通过使用ECS RAM Role访问KMS，你必须要设置如下系统环境变量 (linux):

	- export credentials\_type=ecs\_ram\_role
	- export credentials\_role\_name=\<role name>
	- export cache\_client\_region\_id=[{"regionId":"\<your region id>"}]
```
提示:
	访问KMS实例网关时，使用如下配置
	export cache_client_region_id=[{"regionId":"<your region id>","endpoint":"<your kms instanceId>.cryptoservice.kms.aliyuncs.com"}]
```

* 通过使用OIDC Role ARN访问KMS，你必须要设置如下系统环境变量 (linux):

	- export credentials\_type=oidc\_role\_arn
	- export credentials\_role\_arn=\<role arn> (可选，不填则使用阿里云默认凭据链)
	- export credentials\_oidc\_provider\_arn=\<oidc provider arn> (可选，不填则使用阿里云默认凭据链)
	- export credentials\_oidc\_token\_file\_path=\<oidc token file path> (可选，不填则使用阿里云默认凭据链)
    - export cache\_client\_region\_id=[{"regionId":"\<your region id>"}]
```
提示:
	访问KMS实例网关时，使用如下配置
	export cache_client_region_id=[{"regionId":"<your region id>","endpoint":"<your kms instanceId>.cryptoservice.kms.aliyuncs.com"}]
```