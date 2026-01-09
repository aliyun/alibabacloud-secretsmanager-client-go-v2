# Alibaba Cloud Secrets Manager Client V2 for Go

The Alibaba Cloud Secrets Manager Client V2 for Go enables Go developers to easily work with Alibaba Cloud KMS Secrets.

*Read this in other languages: [English](README.md), [简体中文](README.zh-cn.md)*
## License

[Apache License 2.0](https://www.apache.org/licenses/LICENSE-2.0.html)

## Advantages

* Support users to quickly integrate to obtain secret information
* Support Alibaba Cloud secrets cache (memory cache or encrypted file cache)
* Support tolerant disaster recovery by secrets with the same secret name and secret data in different regions
* Support default backoff strategy and user-defined backoff strategy

## Requirements

- You must use Go 1.10.x or later.

## Installation

Use `go get` to install SDK：

```sh
$ go get -u github.com/aliyun/alibabacloud-secretsmanager-client-go-v2
```

## Sample Code

### General User Code

* Build Secrets Manager Client by system environment variables or configuration file (secretsmanager.properties) ([system environment variables setting for details](README_environment.md),[configure configuration details](README_config.md))

```go
package main

import "github.com/aliyun/alibabacloud-secretsmanager-client-go-v2/sdk"

func main() {
	client, err := sdk.NewClient()
	if err != nil {
		// Handle exceptions
		panic(err)
	}
	secretInfo, err := client.GetSecretInfo("#secretName#")
	if err != nil {
		// Handle exceptions
		panic(err)
	}
}
```

* Build Secrets Manager Client by custom configuration file (you can customize the file name or file path name) ([configuration file setting details](README_config.md))

```go
package main

import (
	"github.com/aliyun/alibabacloud-secretsmanager-client-go-v2/sdk"
	"github.com/aliyun/alibabacloud-secretsmanager-client-go-v2/sdk/service"
)

func main() {
	client, err := sdk.NewSecretCacheClientBuilder(
		service.NewDefaultSecretManagerClientBuilder().Standard().WithCustomConfigFile("#customConfigFileName#").Build()).Build()
	if err != nil {
		// Handle exceptions
		panic(err)
	}
	secretInfo, err := client.GetSecretInfo("#secretName#")
	if err != nil {
		// Handle exceptions
		panic(err)
	}
}
```

* Build Secrets Manager Client by the given parameters(accessKey, accessSecret, regionId, etc)

```go
package main

import (
	"os"
	"github.com/aliyun/alibabacloud-secretsmanager-client-go-v2/sdk"
	"github.com/aliyun/alibabacloud-secretsmanager-client-go-v2/sdk/service"
)

func main() {
	client, err := sdk.NewSecretCacheClientBuilder(
		service.NewDefaultSecretManagerClientBuilder().Standard().WithAccessKey(os.Getenv("#accessKeyId#"), os.Getenv("#accessKeySecret#")).WithRegion("#regionId#").Build()).Build()
	if err != nil {
		// Handle exceptions
		panic(err)
	}
	secretInfo, err := client.GetSecretInfo("#secretName#")
	if err != nil {
		// Handle exceptions
		panic(err)
	}
}
```

* Build Secrets Manager Client by specifying Alibaba Cloud default credential chain parameters. For more information, please refer to [Alibaba Cloud default credential chain](https://www.alibabacloud.com/help/en/sdk/developer-reference/v2-manage-access-credentials?spm=a3c0i.7911826.9556232360.637.4b1c3870EQFmKo#b031e67396a5e).

```go
package main

import (
	"github.com/aliyun/alibabacloud-secretsmanager-client-go-v2/sdk"
	"github.com/aliyun/alibabacloud-secretsmanager-client-go-v2/sdk/service"
)

func main() {
	client, err := sdk.NewSecretCacheClientBuilder(
		service.NewDefaultSecretManagerClientBuilder().Standard().WithRegion("#regionId#").Build()).Build()
	if err != nil {
		// Handle exceptions
		panic(err)
	}
	secretInfo, err := client.GetSecretInfo("#secretName#")
	if err != nil {
		// Handle exceptions
		panic(err)
	}
}
```

* Build Secrets Manager Client by specifying parameters (roleArn, oidcProviderArn, oidcTokenFilePath, etc.)

```go
package main

import (
	"github.com/aliyun/alibabacloud-secretsmanager-client-go-v2/sdk"
	"github.com/aliyun/alibabacloud-secretsmanager-client-go-v2/sdk/service"
	"github.com/aliyun/alibabacloud-secretsmanager-client-go-v2/sdk/utils"
)

func main() {
	provider, _ := utils.CredentialsWithSimpleOIDCRoleArn("#roleArn#", "#oidcProviderArn#", "#oidcTokenFilePath#")

	client, err := sdk.NewSecretCacheClientBuilder(
		service.NewDefaultSecretManagerClientBuilder().Standard().
			WithCredential(provider).
			WithRegion("#regionId#").
			Build()).Build()
	if err != nil {
		// Handle exceptions
		panic(err)
	}
	secretInfo, err := client.GetSecretInfo("#secretName#")
	if err != nil {
		// Handle exceptions
		panic(err)
	}
}
```

### Customized User Code

* Use custom parameters or customized implementation

```go
package main

import (
	"os"
	"github.com/aliyun/alibabacloud-secretsmanager-client-go-v2/sdk"
	"github.com/aliyun/alibabacloud-secretsmanager-client-go-v2/sdk/cache"
	"github.com/aliyun/alibabacloud-secretsmanager-client-go-v2/sdk/service"
)

func main() {
	client, err := sdk.NewSecretCacheClientBuilder(service.NewDefaultSecretManagerClientBuilder().Standard().
		WithAccessKey(os.Getenv("#accessKeyId#"), os.Getenv("#accessKeySecret#")).
		WithRegion("#regionId#").
		WithBackoffStrategy(&service.FullJitterBackoffStrategy{RetryMaxAttempts: 3, RetryInitialIntervalMills: 2000, Capacity: 10000}).Build()).
		WithCacheSecretStrategy(cache.NewFileCacheSecretStoreStrategy("#cacheSecretPath#", true, "#salt#")).
		WithRefreshSecretStrategy(service.NewDefaultRefreshSecretStrategy("#jsonTTLPropertyName#")).
		WithCacheStage("ACSCurrent").
		WithSecretTTL("#secretName#", 1*60*1000).
		WithSecretTTL("#secretName1#", 2*60*1000).Build()
	if err != nil {
		// Handle exceptions
		panic(err)
	}
	secretInfo, err := client.GetSecretInfo("#secretName#")
	if err != nil {
		// Handle exceptions
		panic(err)
	}
}
```

## Frequently Asked Questions (FAQ)

### 1. What should I do if I encounter the error "cannot find the built-in ca certificate for region[$regionId], please provide the caFilePath parameter."?

**Cause:** The built-in CA certificate for this region does not exist in the SDK.

**Solutions:**
1. Please update the SDK to the latest version.

2. If you still get this error after updating to the latest version, you can download the latest CA certificate (CA certificate can be downloaded on the [Key Management Service](https://yundun.console.aliyun.com/?spm=5176.12818093.ProductAndResource--ali--widget-product-recent.dre3.3be916d0yK6Zzx&p=kms#/keyStore/list/base/) - Instance Management - Instance Details page), and pass in the caFilePath parameter. The specific methods are as follows:

**Method 1: Pass CA certificate path by coding**
```go
import (
	"github.com/aliyun/alibabacloud-secretsmanager-client-go-v2/sdk"
	"github.com/aliyun/alibabacloud-secretsmanager-client-go-v2/sdk/models"
	"github.com/aliyun/alibabacloud-secretsmanager-client-go-v2/sdk/service"
	"github.com/aliyun/alibabacloud-secretsmanager-client-go-v2/sdk/utils"
)

func main() {
	// Create RegionInfo containing CA certificate path
	regionInfo := &models.RegionInfo{
		RegionId:   "#regionId#",
		Endpoint:   "#kmsInstanceEndpoint#", // Specify KMS instance address
		CaFilePath: "#caFilePath#",          // Specify CA certificate file path
	}
	
	client, err := sdk.NewSecretCacheClientBuilder(
		service.NewDefaultSecretManagerClientBuilder().
			Standard().
			WithAccessKey(os.Getenv("#accessKeyId#"), os.Getenv("#accessKeySecret#")).
			AddRegionInfo(regionInfo). // Use RegionInfo with CA certificate path
			Build()).Build()
	if err != nil {
		// Handle exceptions
		panic(err)
	}
	// ... use client
}
```

**Method 2: Pass CA certificate path by configuration file**
Add the `caFilePath` parameter in the `secretsmanager.properties` configuration file:
```properties
# The associated KMS service region, including the CA certificate path and instance address
cache_client_region_id=[{"regionId":"<regionId>","endpoint":"<kmsInstanceId>.cryptoservice.kms.aliyuncs.com","caFilePath":"<ca certificate file path>"}]
```

**Method 3: Pass CA certificate path by environment variables**
Refer to [Environment Variable Configuration Description](README_environment.md), add the CA certificate path parameter in the environment variable configuration:
```
# The associated KMS service region, including the CA certificate path and instance address
cache_client_region_id=[{"regionId":"<regionId>","endpoint":"<kmsInstanceId>.cryptoservice.kms.aliyuncs.com","caFilePath":"<ca certificate file path>"}]
```