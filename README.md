## 1、概述

金山云对象存储（Kingsoft Standard Storage Service，简称KS3），是金山云提供的无限制、多备份、分布式的低成本存储空间解决方案。目前提供多种语言SDK，替开发者解决存储扩容、数据可靠安全以及分布式访问等相关复杂问题，开发者可以快速的开发出涉及存储业务的程序或服务。

## 2、完整文档
该文档仅介绍了SDK的基本用法，如果您想了解更多用法，请查阅[官网文档](https://docs.ksyun.com/documents/40487)。

## 3、环境准备

- 环境要求
  使用Golang 1.6及以上版本。

请参考[Golang](https://golang.org/doc/install/source?spm=a2c4g.11186623.0.0.105764a8Y1NXVs)安装下载和安装Go编译运行环境。Go安装完毕后请新建系统变量GOPATH，并将其指向您的代码目录。要了解更多GOPATH相关信息，请执行以下命令。

```shell
go help gopath
```

- 查看语言版本
  要查看Go语言版本，请执行以下命令。

```shell
go version
```

## 4、初始化

### 4.1 下载安装 SDK

- 安装方式：

```shell
go get github.com/ks3sdklib/aws-sdk-go
```

- 使用方法 参见 [Demo](https://github.com/ks3sdklib/aws-sdk-go/tree/master/test)。

### 4.2 获取密钥

1. [开通 KS3 服务, 注册账号](http://www.ksyun.com/user/register)
2. [进入控制台，获取 AccessKeyID 、AccessKeySecret](https://iam.console.ksyun.com/#!/account)

### 4.3 初始化

1. 初始化客户端

```go
// 金山云主账号AccessKey拥有所有API的访问权限，风险很高。
// 强烈建议您创建并使用子账号进行API访问或日常运维，请登录https://uc.console.ksyun.com/pro/iam/#/user/list创建子账号。
// 通过指定Host(Endpoint)，您可以在指定的地域创建新的存储空间。

// 创建访问凭证，请将<AccessKeyID>与<SecretAccessKey>替换成真正的值
cre := credentials.NewStaticCredentials("<AccessKeyID>", "<SecretAccessKey>", "")
// 创建Ks3Client
client := s3.New(&aws.Config{
    Credentials:      cre,                          // 访问凭证，必填
    Region:           "BEIJING",                    // 访问的地域，必填
    Endpoint:         "ks3-cn-beijing.ksyuncs.com", // 访问的域名，必填
    DisableSSL:       false,                        // 禁用HTTPS，默认值为false
    LogLevel:         aws.Off,                      // 日志等级，默认关闭日志，可选值：Off, Error, Warn, Info, Debug
    LogHTTPBody:      false,                        // 把HTTP请求body打入日志，默认值为false
    Logger:           nil,                          // 日志输出位置，可设置指定文件
    S3ForcePathStyle: false,                        // 使用二级域名，默认值为false
    DomainMode:       false,                        // 开启自定义Bucket绑定域名，当开启时S3ForcePathStyle参数不生效，默认值为false
    SignerVersion:    "V2",                         // 签名方式可选值有：V2 OR V4 OR V4_UNSIGNED_PAYLOAD_SIGNER，默认值为V2
    MaxRetries:       3,                            // 请求失败时最大重试次数，默认值为3，值小于0时不重试，如-1表示不重试
    CrcCheckEnabled:  true,                         // 开启CRC64校验，默认值为false
    HTTPClient:       nil,                          // HTTP请求的Client对象，若为空则使用默认值
    DnsCache:         true,                         // 启用DNS缓存，默认值为true
})
```

> 注意：
>
> - [endpoint 与 Region 对应关系](https://docs.ksyun.com/documents/6761)


### 4.4 常见术语介绍

#### Object（对象，文件）

在 KS3 中，用户操作的基本数据单元是 Object。单个 Object 允许存储 0~48.8TB 的数据。 Object 包含 key 和 data。其中，key 是 Object 的名字；data 是 Object 的数据。key 为 UTF-8 编码，且编码后的长度不得超过 1024 个字符。

#### Key（文件名）

即 Object 的名字，key 为 UTF-8 编码，且编码后的长度不得超过 1024 个字符。Key 中可以带有斜杠，当 Key 中带有斜杠的时候，将会自动在控制台里组织成目录结构。

**其他术语请参考**[**概念与术语**](https://docs.ksyun.com/documents/2286)

## 5、快速使用

### 5.1 创建存储空间

```go
package main

import (
  "fmt"
  "github.com/ks3sdklib/aws-sdk-go/aws"
  "github.com/ks3sdklib/aws-sdk-go/aws/awsutil"
  "github.com/ks3sdklib/aws-sdk-go/aws/credentials"
  "github.com/ks3sdklib/aws-sdk-go/service/s3"
)

func main() {
  // 创建访问凭证，请将<AccessKeyID>与<SecretAccessKey>替换成真正的值
  cre := credentials.NewStaticCredentials("<AccessKeyID>", "<SecretAccessKey>", "")
  // 创建S3Client，更多配置项请查看Go-SDK初始化文档
  client := s3.New(&aws.Config{
    Credentials: cre,                          // 访问凭证
    Region:      "BEIJING",                    // 填写您的Region
    Endpoint:    "ks3-cn-beijing.ksyuncs.com", // 填写您的Endpoint
  })
  // 填写存储空间名称
  bucket := "<bucket_name>"
  // 创建存储空间
  resp, err := client.CreateBucket(&s3.CreateBucketInput{
    Bucket:    aws.String(bucket),        // 存储空间名称，必填
    ACL:       aws.String("public-read"), // 存储空间访问权限，非必填
    ProjectId: aws.String(""),            // 项目制id，非必填
  })
  if err != nil {
    panic(err)
  }
  fmt.Println("结果：\n", awsutil.StringValue(resp))
}
```

### 5.2 上传对象

```go
package main

import (
  "bytes"
  "fmt"
  "github.com/ks3sdklib/aws-sdk-go/aws"
  "github.com/ks3sdklib/aws-sdk-go/aws/awsutil"
  "github.com/ks3sdklib/aws-sdk-go/aws/credentials"
  "github.com/ks3sdklib/aws-sdk-go/service/s3"
  "io/ioutil"
)

func main() {
  // 创建访问凭证，请将<AccessKeyID>与<SecretAccessKey>替换成真正的值
  cre := credentials.NewStaticCredentials("<AccessKeyID>", "<SecretAccessKey>", "")
  // 创建S3Client，更多配置项请查看Go-SDK初始化文档
  client := s3.New(&aws.Config{
    Credentials: cre,                          // 访问凭证
    Region:      "BEIJING",                    // 填写您的Region
    Endpoint:    "ks3-cn-beijing.ksyuncs.com", // 填写您的Endpoint
  })
  // 填写存储空间名称
  bucket := "<bucket_name>"
  // 填写对象的key
  key := "<object_key>"
  // 填写上传文件路径
  filePath := "/Users/test/demo.txt"
  // 读取文件
  file, err := ioutil.ReadFile(filePath)
  if err != nil {
    panic(err)
  }
  // 上传对象
  resp, err := client.PutObject(&s3.PutObjectInput{
    Bucket: aws.String(bucket),        // 存储空间名称，必填
    Key:    aws.String(key),           // 对象的key，必填
    Body:   bytes.NewReader(file),     // 要上传的文件，必填
    ACL:    aws.String("public-read"), // 对象的访问权限，非必填
  })
  if err != nil {
    panic(err)
  }
  fmt.Println("结果：\n", awsutil.StringValue(resp))
}
```

### 5.3 列举对象

```go
package main

import (
  "fmt"
  "github.com/ks3sdklib/aws-sdk-go/aws"
  "github.com/ks3sdklib/aws-sdk-go/aws/awsutil"
  "github.com/ks3sdklib/aws-sdk-go/aws/credentials"
  "github.com/ks3sdklib/aws-sdk-go/service/s3"
)

func main() {
  // 创建访问凭证，请将<AccessKeyID>与<SecretAccessKey>替换成真正的值
  cre := credentials.NewStaticCredentials("<AccessKeyID>", "<SecretAccessKey>", "")
  // 创建S3Client，更多配置项请查看Go-SDK初始化文档
  client := s3.New(&aws.Config{
    Credentials: cre,                          // 访问凭证
    Region:      "BEIJING",                    // 填写您的Region
    Endpoint:    "ks3-cn-beijing.ksyuncs.com", // 填写您的Endpoint
  })
  // 填写存储空间名称
  bucket := "<bucket_name>"
  // 获取存储对象列表
  resp, err := client.ListObjects(&s3.ListObjectsInput{
    Bucket:    aws.String(bucket),    // 存储空间名称，必填
    Delimiter: aws.String("/"),       // 分隔符，用于对一组参数进行分割的字符，非必填
    MaxKeys:   aws.Long(int64(1000)), // 设置响应体中返回的最大记录数，默认为1000，非必填
    Prefix:    aws.String(""),        // 限定响应结果列表使用的前缀，正如您在电脑中使用的文件夹一样，非必填
    Marker:    aws.String(""),        // 指定列举指定空间中对象的起始位置，非必填
  })
  if err != nil {
    panic(err)
  }
  fmt.Println("结果：\n", awsutil.StringValue(resp))
}
```

### 5.4 删除对象

```go
package main

import (
  "fmt"
  "github.com/ks3sdklib/aws-sdk-go/aws"
  "github.com/ks3sdklib/aws-sdk-go/aws/awsutil"
  "github.com/ks3sdklib/aws-sdk-go/aws/credentials"
  "github.com/ks3sdklib/aws-sdk-go/service/s3"
)

func main() {
  // 创建访问凭证，请将<AccessKeyID>与<SecretAccessKey>替换成真正的值
  cre := credentials.NewStaticCredentials("<AccessKeyID>", "<SecretAccessKey>", "")
  // 创建S3Client，更多配置项请查看Go-SDK初始化文档
  client := s3.New(&aws.Config{
    Credentials: cre,                          // 访问凭证
    Region:      "BEIJING",                    // 填写您的Region
    Endpoint:    "ks3-cn-beijing.ksyuncs.com", // 填写您的Endpoint
  })
  // 填写存储空间名称
  bucket := "<bucket_name>"
  // 填写删除对象的key
  key := "<object_key>"
  // 删除对象
  resp, err := client.DeleteObject(&s3.DeleteObjectInput{
    Bucket: aws.String(bucket), // 存储空间名称，必填
    Key:    aws.String(key),    // 对象的key，必填
  })
  if err != nil {
    panic(err)
  }
  fmt.Println("结果：\n", awsutil.StringValue(resp))
}
```