# 阿里云凭据管家Go V2客户端

阿里云凭据管家Go V2客户端可以使Go开发者快速使用阿里云凭据。

*其他语言版本: [English](README.md), [简体中文](README.zh-cn.md)*

## 许可证

[Apache License 2.0](https://www.apache.org/licenses/LICENSE-2.0.html)

## 优势

* 支持用户快速集成获取凭据信息
* 支持阿里云凭据管家内存和文件两种缓存凭据机制
* 支持凭据名称相同场景下的跨地域容灾
* 支持默认规避策略和用户自定义规避策略

## 软件要求

- Go 1.10 或以上版本

## 安装

使用 go get 下载安装 SDK

```
go get -u github.com/aliyun/alibabacloud-secretsmanager-client-go-v2
```


## 示例代码
### 一般用户代码

* 通过系统环境变量或配置文件(secretsmanager.properties)
  构建客户端([系统环境变量设置详情](README_environment.zh-cn.md)、[配置文件设置详情](README_config.zh-cn.md))

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

* 通过自定义配置文件(可自定义文件名称或文件路径名称)构建客户端([配置文件设置详情](README_config.zh-cn.md))

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

* 通过指定参数(accessKey、accessSecret、regionId等)构建客户端

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

* 通过指定阿里云默认凭据链参数构建客户端。更多信息请参考 [阿里云默认凭据链](https://help.aliyun.com/zh/sdk/developer-reference/v2-manage-go-access-credentials?spm=a2c4g.11186623.help-menu-262060.d_1_9_1_2.7099279fpgGI9b#b0ae259ed4xa0)。

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

* 通过指定参数(roleArn、oidcProviderArn、oidcTokenFilePath等)构建客户端

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

### 定制化用户代码

* 使用自定义参数或用户自己实现

```go
package main

import (
  "github.com/aliyun/alibabacloud-secretsmanager-client-go-v2/sdk"
  "github.com/aliyun/alibabacloud-secretsmanager-client-go-v2/sdk/cache"
  "github.com/aliyun/alibabacloud-secretsmanager-client-go-v2/sdk/service"
  "os"
)

func main() {
  client, err := sdk.NewSecretCacheClientBuilder(
    service.NewDefaultSecretManagerClientBuilder().Standard().WithAccessKey(os.Getenv("#accessKeyId#"), os.Getenv("#accessKeySecret#")).WithRegion("#regionId#").WithBackoffStrategy(&service.FullJitterBackoffStrategy{RetryMaxAttempts: 3, RetryInitialIntervalMills: 2000, Capacity: 10000}).Build()).WithCacheSecretStrategy(cache.NewFileCacheSecretStoreStrategy("#cacheSecretPath#", true, "#salt#")).WithRefreshSecretStrategy(service.NewDefaultRefreshSecretStrategy("#jsonTTLPropertyName#")).WithCacheStage("ACSCurrent").WithSecretTTL("#secretName#", 1*60*1000).Build()
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

## 常见问题 FAQ

### 1. 出现 "cannot find the built-in ca certificate for region[$regionId], please provide the caFilePath parameter." 错误怎么办？

**问题原因：** SDK 中该地域内置的 CA 证书不存在。

**解决方案：**
1. 请更新 SDK 到最新版本。

2. 如果已更新到最新版本仍然报此错误，可以下载最新的CA证书（CA证书可在[密钥管理服务](https://yundun.console.aliyun.com/?spm=5176.12818093.ProductAndResource--ali--widget-product-recent.dre3.3be916d0yK6Zzx&p=kms#/keyStore/list/base/) - 实例管理 - 实例详情 页面下载），并传入CA证书路径参数。具体方式如下：

**方式一：编码方式传递 CA 证书路径**
```go
package main
import (
  "github.com/aliyun/alibabacloud-secretsmanager-client-go-v2/sdk"
  "github.com/aliyun/alibabacloud-secretsmanager-client-go-v2/sdk/models"
  "github.com/aliyun/alibabacloud-secretsmanager-client-go-v2/sdk/service"
  "os"
)

func main() {
  // 创建包含 CA 证书路径的 RegionInfo
  regionInfo := &models.RegionInfo{
    RegionId:   "#regionId#",
    Endpoint:   "#kmsInstanceEndpoint#", // 指定 KMS 实例地址
    CaFilePath: "#caFilePath#",          // 指定 CA 证书文件路径
  }

  client, err := sdk.NewSecretCacheClientBuilder(
    service.NewDefaultSecretManagerClientBuilder().
      Standard().
      WithAccessKey(os.Getenv("#accessKeyId#"), os.Getenv("#accessKeySecret#")).
      AddRegionInfo(regionInfo). // 使用带 CA 证书路径的 RegionInfo
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

**方式二：通过配置文件方式传递 CA 证书路径**
在 `secretsmanager.properties` 配置文件中添加 `caFilePath` 参数：
```properties
# 关联的KMS服务地域，包含CA证书路径和实例地址
cache_client_region_id=[{"regionId":"<regionId>","endpoint":"<kmsInstanceId>.cryptoservice.kms.aliyuncs.com","caFilePath":"<ca证书文件路径>"}]
```

**方式三：通过环境变量方式传递 CA 证书路径**
参考 [环境变量配置说明](README_environment.zh-cn.md)，在环境变量配置中添加 CA 证书路径参数：
```
# 关联的KMS服务地域，包含CA证书路径和实例地址
cache_client_region_id=[{"regionId":"<regionId>","endpoint":"<kmsInstanceId>.cryptoservice.kms.aliyuncs.com","caFilePath":"<ca证书文件路径>"}]
```