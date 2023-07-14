## 1、概述

金山云对象存储（Kingsoft Standard Storage Service，简称KS3），是金山云提供的无限制、多备份、分布式的低成本存储空间解决方案。目前提供多种语言SDK，替开发者解决存储扩容、数据可靠安全以及分布式访问等相关复杂问题，开发者可以快速的开发出涉及存储业务的程序或服务。

## 2、环境准备

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

## 3、初始化

### 3.1 下载安装 SDK

- 安装方式：

```shell
go get github.com/ks3sdklib/aws-sdk-go
```

- 使用方法 参见 [Demo](https://github.com/ks3sdklib/aws-sdk-go/tree/master/test)。

### 3.2 获取密钥

1. [开通 KS3 服务, 注册账号](http://www.ksyun.com/user/register)
2. [进入控制台，获取 AccessKeyID 、AccessKeySecret](https://iam.console.ksyun.com/#!/account)

### 3.3 初始化

1. 初始化客户端

```go
  credentials := credentials.NewStaticCredentials("<AccessKeyID>","<AccessKeySecret>","")
	client = s3.New(&aws.Config{
		//Region 可参考 https://docs.ksyun.com/documents/6761
		Region:      region,
		Credentials: cre,
		//Endpoint 可参考 https://docs.ksyun.com/documents/6761
		Endpoint:         endpoint,
		DisableSSL:       true,  //是否禁用https
		LogLevel:         0,     //是否开启日志,0为关闭日志，1为开启日志
		LogHTTPBody:      false, //是否把HTTP请求body打入日志
		S3ForcePathStyle: false,
		Logger:           nil,   //打日志的位置
		DomainMode:       false, //是否开启自定义bucket绑定域名，当开启时 S3ForcePathStyle 参数不生效。
		//可选值有 ： V2 OR V4 OR V4_UNSIGNED_PAYLOAD_SIGNER
		SignerVersion: "V4",
		MaxRetries:    1,
	})
```

> 注意：
>
> - [endpoint 与 Region 对应关系](https://docs.ksyun.com/documents/6761)


### 3.4 常见术语介绍

#### Object（对象，文件）

在 KS3 中，用户操作的基本数据单元是 Object。单个 Object 允许存储 0~48.8TB 的数据。 Object 包含 key 和 data。其中，key 是 Object 的名字；data 是 Object 的数据。key 为 UTF-8 编码，且编码后的长度不得超过 1024 个字符。

#### Key（文件名）

即 Object 的名字，key 为 UTF-8 编码，且编码后的长度不得超过 1024 个字符。Key 中可以带有斜杠，当 Key 中带有斜杠的时候，将会自动在控制台里组织成目录结构。

**其他术语请参考**[**概念与术语**](https://docs.ksyun.com/documents/2286)

## 4、空间相关

### 4.1 创建bucket

```go
resp, err := svc.CreateBucket(&s3.CreateBucketInput{
    ACL:    aws.String("public-read"),//权限
    Bucket: aws.String("bucket"),
    ProjectId: aws.String("123123"),//项目制id
})
if err != nil {
    if awsErr, ok := err.(awserr.Error); ok {
        fmt.Println(awsErr.Code(), awsErr.Message(), awsErr.OrigErr())
        if reqErr, ok := err.(awserr.RequestFailure); ok {
            fmt.Println(reqErr.Code(), reqErr.Message(), reqErr.StatusCode(), reqErr.RequestID())
        }
    } else {
        fmt.Println(err.Error())
    }
}
fmt.Println("结果：\n", awsutil.StringValue(resp))
```

### 4.2 判断bucket是否存在

```python
//判断bucket桶是否存在
func (s *Ks3utilCommandSuite) TestBucketExist(c *C) {

	exist := client.HeadBucketExist(bucket)
	if exist {
		fmt.Println("bucket exist")
	} else {
		fmt.Println("bucket not exist")
	}
}

```

### 镜像回源规则

#### 1. 设置镜像回源规则

```go
params := &s3.PutBucketMirrorInput{
    Bucket: aws.String(BucketName), // 桶名，必填字段
    BucketMirror: &s3.BucketMirror{
        Version:          aws.String("V3"),//回源规则版本
        UseDefaultRobots: aws.Boolean(false),//是否使用默认的robots.txt
        //异步回源规则，设置源站url、ACL权限
        AsyncMirrorRule: &s3.AsyncMirrorRule{
            MirrorUrls: []*string{
                aws.String("http://abc.om"),
                aws.String("http://wps.om"),
            },
            //SavingSetting: &s3.SavingSetting{
            //	ACL: "private",
            //},
        },
        //同步回源规则，设置触发条件（http_codes和文件前缀）、源站url、query string是否透传、是否follow 302/301、header配置、ACL权限
        SyncMirrorRules: []*s3.SyncMirrorRules{
            {
                MatchCondition: s3.MatchCondition{
                    HTTPCodes: []*string{
                        aws.String("404"),
                    },
                    KeyPrefixes: []*string{
                        aws.String("abc"),
                    },
                },
                MirrorURL: aws.String("http://v-ks-a-i.originalvod.com"),
                MirrorRequestSetting: &s3.MirrorRequestSetting{
                    PassQueryString: aws.Boolean(true),
                    Follow3Xx:       aws.Boolean(true),
                    HeaderSetting: &s3.HeaderSetting{
                        SetHeaders: []*s3.SetHeaders{
                            {
                                aws.String("a"),
                                aws.String("b"),
                            },
                        },
                        RemoveHeaders: []*s3.RemoveHeaders{
                            {
                                aws.String("daaaaa"),
                            },
                        },
                        PassAll: aws.Boolean(true),
                        //PassHeaders: []*s3.PassHeaders{
                        //	{
                        //		aws.String("asdb"),
                        //	},
                        //},
                    },
                },
                SavingSetting: &s3.SavingSetting{
                    ACL: aws.String("private"),
                },
            },
        },
    },
}
resp, err := client.PutBucketMirror(params)
if err != nil {
    if awsErr, ok := err.(awserr.Error); ok {
        // Generic AWS Error with Code, Message, and original error (if any)
        fmt.Println(awsErr.Code(), awsErr.Message(), awsErr.OrigErr())
        if reqErr, ok := err.(awserr.RequestFailure); ok {
            // A service error occurred
            fmt.Println(reqErr.Code(), reqErr.Message(), reqErr.StatusCode(), reqErr.RequestID())
        }
    } else {
        // This case should never be hit, The SDK should alwsy return an
        // error which satisfies the awserr.Error interface.
        fmt.Println(err.Error())
    }
}

fmt.Println("resp.code is:", resp.HttpCode)
fmt.Println("resp.Header is:", resp.Header)
// Pretty-print the response data.
var bodyStr = string(resp.Body[:])
fmt.Println("resp.Body is:", bodyStr)
```

#### 2.获取镜像回源规则

```go
 	params := &s3.GetBucketMirrorInput{
	  	Bucket: aws.String(BucketName),
	}
	resp, _ := client.GetBucketMirror(params)
	fmt.Println("resp.code is:", resp.HttpCode)
	fmt.Println("resp.Header is:", resp.Header)
	// Pretty-print the response data.
	var bodyStr = string(resp.Body[:])
	fmt.Println("resp.Body is:", bodyStr)
```

#### 3. 删除镜像回源规则

```python
params := &s3.DeleteBucketMirrorInput{
    Bucket: aws.String(BucketName),
}
resp, _ := client.DeleteBucketMirror(params)
fmt.Println("resp.code is:", resp.HttpCode)
fmt.Println("resp.Header is:", resp.Header)
// Pretty-print the response data.
var bodyStr = string(resp.Body[:])
fmt.Println("resp.Body is:", bodyStr)
```

#### 

### 生命周期规则

#### 1.配置生命周期规则

```go
lifecycleConfiguration := &s3.LifecycleConfiguration{
    Rules: []*s3.LifecycleRule{
        {
            ID:     aws.String("rule1"),
            Status: aws.String("Enabled"),
            Expiration: &s3.LifecycleExpiration{
                Days: aws.Long(30),
            },
        },
    },
}
resp, err := client.PutBucketLifecycle(&s3.PutBucketLifecycleInput{
    Bucket:                 aws.String(bucket),
    LifecycleConfiguration: lifecycleConfiguration,
})
fmt.Println("结果：\n", awsutil.StringValue(resp), err)
```

#### 2.获取生命周期规则

```go
resp, err := client.GetBucketLifecycle(&s3.GetBucketLifecycleInput{
		Bucket: aws.String(bucket),
	})
fmt.Println("结果：\n", awsutil.StringValue(resp), err)
```

#### 3.删除生命周期规则

```go
resp, err := client.DeleteBucketLifecycle(&s3.DeleteBucketLifecycleInput{})
fmt.Println("结果：\n", awsutil.StringValue(resp), err)
```

### 跨域规则

#### 1.配置跨域规则

```go
// 配置CORS规则
corsConfiguration := &s3.CORSConfiguration{
    Rules: []*s3.CORSRule{
        {
            AllowedHeader: []string{
                "*",
            },
            AllowedMethod: []string{
                "PUT",
            },
            AllowedOrigin: []string{
                "*",
            },
            MaxAgeSeconds: 100,
        },
    },
}
// 设置桶的CORS配置
resp, err := client.PutBucketCORS(&s3.PutBucketCORSInput{
    Bucket:            aws.String(bucket),
    CORSConfiguration: corsConfiguration,
})
fmt.Println("结果：\n", awsutil.StringValue(resp), err)
```

#### 2.获取跨域规则

```go
resp, err := client.GetBucketCORS(&s3.GetBucketCORSInput{
    Bucket: aws.String(bucket),
})
fmt.Println("结果：\n", awsutil.StringValue(resp), err)
```

#### 3.删除跨域规则

```go
	resp, err := client.DeleteBucketCORS(&s3.DeleteBucketCORSInput{
		Bucket: aws.String(bucket),
	})
	fmt.Println("结果：\n", awsutil.StringValue(resp), err)
```

### 桶日志

#### 1.配置日志

```go
logStatus := &s3.BucketLoggingStatus{
    LoggingEnabled: &s3.LoggingEnabled{
        TargetBucket: aws.String(bucket),
        TargetPrefix: aws.String(bucket),
    },
}
resp, err := client.PutBucketLogging(&s3.PutBucketLoggingInput{
    Bucket:              aws.String(bucket),
    BucketLoggingStatus: logStatus,
    ContentType:         aws.String("application/xml"),
})
fmt.Println("结果：\n", awsutil.StringValue(resp), err)
```

#### 2.获取日志配置

```go
resp, err := client.GetBucketLogging(&s3.GetBucketLoggingInput{
    Bucket: aws.String(bucket),
})
fmt.Println("结果：\n", awsutil.StringValue(resp), err)
```

## 5、对象相关

### 5.1  获取元数据

```go
input := s3.HeadObjectInput{
	Bucket: aws.String(bucketname),
	Key:    aws.String(key),
}
resp, err := client.HeadObject(&input)

if err != nil {
	if awsErr, ok := err.(awserr.Error); ok {
		fmt.Println(awsErr.Code(), awsErr.Message(), awsErr.OrigErr())
		if reqErr, ok := err.(awserr.RequestFailure); ok {
			fmt.Println(reqErr.Code(), reqErr.Message(), reqErr.StatusCode(), reqErr.RequestID())
		}
	} else {
		fmt.Println(err.Error())
	}
}
fmt.Println("结果：\n", awsutil.StringValue(resp))
```

### 5.2 上传本地文件

```go
filename = "/Users/user/Downloads/aaa.docx"
file, err := ioutil.ReadFile(filename)
if err != nil {
    fmt.Println("Failed to open file", err)
    os.Exit(1)
}
params := &s3.PutObjectInput{
    Bucket:      aws.String(BucketName),                // bucket名称
    Key:         aws.String("go-demo/test"),            // object key
    ACL:         aws.String("public-read"),             //权限，支持private(私有)，public-read(公开读)
    Body:        bytes.NewReader(file),                 //要上传的内容
    ContentType: aws.String("application/octet-stream"), //设置content-type
    Metadata: map[string]*string{
        //"Key": aws.String("MetadataValue"), // 设置用户元数据
        // More values...
    },
}
resp, err := client.PutObject(params)
if err != nil {
    panic(err)
}
fmt.Println(resp)
```

### 5.3 下载文件

```go
 params := &s3.GetObjectInput{
 			Bucket: aws.String("BucketName"), // bucket名称
	    Key: aws.String("ObjectKey"), // object key
 }
 resp, err := client.GetObject(params)
 if err != nil{
     panic(err)
 }
 //读取返回结果中body的前20个字节
 b := make([]byte, 20)
 n, err := resp.Body.Read(b)
 fmt.Printf("%-20s %-2v %v\n", b[:n], n, err)
```

### 5.4 生成文件下载外链

```go
 params := &s3.GetObjectInput{
     Bucket: aws.String("BucketName"), // bucket名称
     Key: aws.String("ObjectKey"), // object key
     ResponseCacheControl: aws.String("ResponseCacheControl"),
    //控制返回的Cache-Control header
     ResponseContentDisposition: aws.String("ResponseContentDisposition"),
    //控制返回的Content-Disposition header
     ResponseContentEncoding: aws.String("ResponseContentEncoding"),
    //控制返回的Content-Encoding header
     ResponseContentLanguage: aws.String("ResponseContentLanguage"),
    //控制返回的Content-Language header
     ResponseContentType: aws.String("ResponseContentType"),
    //控制返回的Content-Type header
 }

 resp, err := client.GetObjectPresignedUrl(params,1444370289000000000)
//第二个参数为外链过期时间，为纳秒级的时间戳
 if err!=nil {
 			panic(err)
 }
 fmt.Println(resp)//resp即生成的外链地址,类型为url.URL
```

### 5.5 生成文件上传外链

```go
 params := &s3.PutObjectInput{
     Bucket: aws.String(bucket), // bucket名称
     Key: aws.String(key), // object key
     ACL: aws.String("public-read"),//设置ACL
     ContentType: aws.String("application/octet-stream"),//设置文件的content-type
     ContentMaxLength: aws.Long(20),//设置允许的最大长度，对应的header：x-amz-content-maxlength
 }

 resp, err := client.PutObjectPresignedUrl(params,1444370289000000000)
//第二个参数为外链过期时间，为纳秒级的时间戳
 if err!=nil {
     panic(err)
 }

 httpReq, _ := http.NewRequest("PUT", "", strings.NewReader("123"))
 httpReq.URL = resp
 httpReq.Header["x-amz-acl"] = []string{"public-read"}
 httpReq.Header["x-amz-content-maxlength"] = []string{"20"}
 httpReq.Header.Add("Content-Type","application/octet-stream")
 fmt.Println(httpReq)
 httpRep,err := http.DefaultClient.Do(httpReq)
 fmt.Println(httpRep)

 if err != nil{
 		panic(err)
 }
```

### 5.6 生成设置文件 ACL 的外链

```go
 params := &s3.PutObjectACLInput{
     Bucket: aws.String(bucket), // bucket名称
     Key: aws.String(key), // object key
     ACL: aws.String("private"),//设置ACL
     ContentType: aws.String("text/plain"),
 }

 resp, err := client.PutObjectACLPresignedUrl(params,1444370289000000000)
//第二个参数为外链过期时间，为纳秒级的时间戳
 if err!=nil {
     panic(err)
 }
 fmt.Println(resp)//resp即生成的外链地址,类型为url.URL

 httpReq, _ := http.NewRequest("PUT", "", nil)
 httpReq.URL = resp
 httpReq.Header["x-amz-acl"] = []string{"private"}
 httpReq.Header.Add("Content-Type","text/plain")
 fmt.Println(httpReq)

 httpRep,err := http.DefaultClient.Do(httpReq)
 if err != nil{
		 panic(err)
 }
 fmt.Println(httpRep)
```

### 5.7 修改元数据

```go
  key_modify_meta := string("yourkey")
  metadata := make(map[string]*string)
  metadata["yourmetakey1"] = aws.String("yourmetavalue1")
  metadata["yourmetakey2"] = aws.String("yourmetavalue2")

  _,err := svc.CopyObject(&s3.CopyObjectInput{
    Bucket:aws.String(bucket),
    Key:aws.String(key_modify_meta),
    CopySource:aws.String("/" + bucket+"/" + key_modify_meta),
    MetadataDirective:aws.String("REPLACE"),
    Metadata:metadata,
  })
  assert.NoError(t,err)
  assert.True(t,objectExists(bucket,key))
```

### 5.8 批量删除对象

```go
 params := &s3.DeleteObjectsInput{
    Bucket: aws.String(""), // 桶名称
    Delete: &s3.Delete{     // Delete Required
      Objects: []*s3.ObjectIdentifier{
        {
          Key:       aws.String("1"), // 目标对象1的key
        },
        {
          Key:       aws.String("2"), // 目标对象2的key
        },
        // More values...
      },
    },
}
resp := svc.DeleteObjects(params)    //执行并返回响应结果
fmt.Println("error keys:",resp.Errors)
fmt.Println("deleted keys:",resp.Deleted)
```

### 5.9 目录删除

```go
params := &s3.DeleteBucketPrefixInput{
  Bucket: aws.String(""),  //桶名称
  Prefix: aws.String(prefix),    //前缀（目录）名称
}
resp, _ := svc.DeleteBucketPrefix(params)    //执行并接受响应结果
fmt.Println("error keys:",resp.Errors)
fmt.Println("deleted keys:",resp.Deleted)
```

### 5.10 列举文件

```go
resp, err := client.ListObjects(&s3.ListObjectsInput{
  Bucket: aws.String(BucketName),
  Prefix: aws.String("onlineTask/"), //文件前缀
  //	Marker:  aws.String("Marker"), //从按个key开始获取
  MaxKeys: aws.Long(1000), //最大数量
})
if err != nil {
  fmt.Println(resp)
}
fmt.Println(len(resp.Contents))
```

### 5.11 抓取远程数据到KS3

```go
sourceUrl := "https://img0.pconline.com.cn/pconline/1111/04/2483449_20061139501.jpg"
    input := s3.FetchObjectInput{
    Bucket:      aws.String(bucketname),
    Key:         aws.String("dst/testa"),
    SourceUrl:   aws.String(sourceUrl),
    ACL:         aws.String("public-read"),
}
resp, err := svc.FetchObject(&input)
if err != nil {
    if awsErr, ok := err.(awserr.Error); ok {
      fmt.Println(awsErr.Code(), awsErr.Message(), awsErr.OrigErr())
      if reqErr, ok := err.(awserr.RequestFailure); ok {
        fmt.Println(reqErr.Code(), reqErr.Message(), reqErr.StatusCode(), reqErr.RequestID())
      }
    } else {
      fmt.Println(err.Error())
    }
}
fmt.Println("结果：\n", awsutil.StringValue(resp))
```

### 5.12  复制对象

```go
input := s3.CopyObjectInput{
    Bucket:               aws.String(bucketname),
    Key:                  aws.String(key),
    CopySource:           aws.String("/cqc-test-b/yztestfile1"),
}
resp, err := svc.CopyObject(&input)
if err != nil {
    if awsErr, ok := err.(awserr.Error); ok {
      fmt.Println(awsErr.Code(), awsErr.Message(), awsErr.OrigErr())
      if reqErr, ok := err.(awserr.RequestFailure); ok {
        fmt.Println(reqErr.Code(), reqErr.Message(), reqErr.StatusCode(), reqErr.RequestID())
      }
    } else {
      fmt.Println(err.Error())
    }
}

fmt.Println("结果：\n", awsutil.StringValue(resp))
```

### 5.13 对象标签

### 5.13.1 设置对象标签

```go
tagkey := "name"
tagval := "yz"
tagkey2 := "sex"
tagval2 := "female"
objTagging := s3.Tagging{
    TagSet: []*s3.Tag{&s3.Tag{
        Key:   aws.String(tagkey),
        Value: aws.String(tagval),
    }, &s3.Tag{
        Key:   aws.String(tagkey2),
        Value: aws.String(tagval2),
    },
    },
}

params := &s3.PutObjectTaggingInput{
    Bucket:  aws.String(bucketname),
    Key:     aws.String(key),
    Tagging: &objTagging,
}
resp, err := client.PutObjectTagging(params)

if err != nil {
    if awsErr, ok := err.(awserr.Error); ok {
        fmt.Println(awsErr.Code(), awsErr.Message(), awsErr.OrigErr())
        if reqErr, ok := err.(awserr.RequestFailure); ok {
            fmt.Println(reqErr.Code(), reqErr.Message(), reqErr.StatusCode(), reqErr.RequestID())
        }
    } else {
        fmt.Println(err.Error())
    }
}
fmt.Println("结果：\n", awsutil.StringValue(resp))
```

### 5.18.2 获取对象标签

```go
params := &s3.GetObjectTaggingInput{
    Bucket: aws.String(bucketname), // Required
    Key:    aws.String(key),
}
resp, err := client.GetObjectTagging(params)

if err != nil {
    if awsErr, ok := err.(awserr.Error); ok {
        fmt.Println(awsErr.Code(), awsErr.Message(), awsErr.OrigErr())
        if reqErr, ok := err.(awserr.RequestFailure); ok {
            fmt.Println(reqErr.Code(), reqErr.Message(), reqErr.StatusCode(), reqErr.RequestID())
        }
    } else {
        fmt.Println(err.Error())
    }
}
fmt.Println("结果：\n", awsutil.StringValue(resp))
```

### 5.18.3 删除对象标签

```go
 params := &s3.DeleteObjectTaggingInput{
		Bucket: aws.String(bucketname), // Required
		Key:    aws.String(key),
	}
	resp, err := client.DeleteObjectTagging(params)

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			fmt.Println(awsErr.Code(), awsErr.Message(), awsErr.OrigErr())
			if reqErr, ok := err.(awserr.RequestFailure); ok {
				fmt.Println(reqErr.Code(), reqErr.Message(), reqErr.StatusCode(), reqErr.RequestID())
			}
		} else {
			fmt.Println(err.Error())
		}
	}

	fmt.Println("结果：\n", awsutil.StringValue(resp))
```

## 6、分块相关

### 6.1 初始化分块上传

```go
params := &s3.CreateMultipartUploadInput{
     Bucket: aws.String("BucketName"), // bucket名称
     Key: aws.String("ObjectKey"), // object key
     ACL: aws.String("ObjectCannedACL"),//权限，支持private(私有)，public-read(公开读)
     ContentType: aws.String("application/octet-stream"),//设置content-type
     Metadata: map[string]*string{
     //"Key": aws.String("MetadataValue"), // 设置用户元数据
     // More values...
    },

}

resp, err := client.CreateMultipartUpload(params)
if err != nil{
    panic(err)
}
//获取这次初始化的uploadid
fmt.Println(*resp.UploadID)
```

### 6.2 分块上传 - 上传块

```go
params := &s3.UploadPartInput{
 Bucket:aws.String(bucket),//bucket名称
 Key:aws.String(key),//文件名
 PartNumber:aws.Long(1),//当前块的序号
 UploadID:aws.String(uploadId),//由初始化获取到得uploadid
 Body:strings.NewReader(content),//当前块的内容
 ContentLength:aws.Long(int64(len(content))),//内容长度
}

resp,err := client.UploadPart(params)
if err != nil{
     panic(err)
}
fmt.Println(resp)
```

### 6.3 完成分块上传并合并块

```go
params := &s3.CompleteMultipartUploadInput{
 Bucket:aws.String(bucket),//bucket名称
 Key:aws.String(key),//文件名
 UploadID:aws.String(uploadId),//由初始化获取到得uploadid
 MultipartUpload:&s3.CompletedMultipartUpload{
 Parts:<已经完成的块列表>,//类型为*s3.CompletedPart数组
},
}

resp,err := client.CompleteMultipartUpload(params)
if err != nil{
   panic(err)
}
fmt.Println(resp)
```

### 6.4 取消分块上传

```go
 params := &s3.AbortMultipartUploadInput{
     Bucket:aws.String(bucket),//bucket名称
     Key:aws.String(key),//文件名
     UploadID:aws.String(uploadId),//由初始化获取到得uploadid
 }

 resp,err := client.AbortMultipartUpload(params)
 if err != nil{
		 panic(err)
 }
 fmt.Println(resp)
```

### 6.5 罗列分块上传已经上传的块

```go
 params := &s3.ListPartsInput{
     Bucket:aws.String(bucket),//bucket名称
     Key:aws.String(key),//文件名
     UploadID:aws.String(uploadId),//由初始化获取到得uploadid
 }

 resp,err := client.ListParts(params)
 if err != nil{
 		panic(err)
 }
 fmt.Println(resp)
```

## 7.部分示例

### 计算签名

```go
 package main

 import(

 "crypto/hmac"
 "crypto/sha1"
 "encoding/base64"
 "fmt"
 )

 func main(){

   AccessKeyId := "AccessKeyId"
   AccessKeySecret:= "AccessKeySecret"
   stringToSign := "stringToSign"
   signature := string(base64Encode(makeHmac([]byte(AccessKeySecret), []byte(stringToSign))))
   token := "KSS "+AccessKeyId+":"+signature
   fmt.Println(token)
 }

 func makeHmac(key []byte, data []byte) []byte {
     hash := hmac.New(sha1.New, key)
     hash.Write(data)
     return hash.Sum(nil)
 }

 func base64Encode(src []byte) []byte {
	   	return []byte(base64.StdEncoding.EncodeToString(src))
 }
```

### 上传文件夹

```go
package main

import (
	"github.com/ks3sdklib/aws-sdk-go/aws"
	"github.com/ks3sdklib/aws-sdk-go/aws/credentials"
	"github.com/ks3sdklib/aws-sdk-go/service/s3"
	"github.com/ks3sdklib/aws-sdk-go/service/s3/s3manager"
	"os"
)

var (
	endpoint               = os.Getenv("KS3_TEST_ENDPOINT")
	accessKeyID            = os.Getenv("KS3_TEST_ACCESS_KEY_ID")
	accessKeySecret        = os.Getenv("KS3_TEST_ACCESS_KEY_SECRET")
	bucket                 = os.Getenv("KS3_TEST_BUCKET")
	region                 = os.Getenv("KS3_TEST_REGION")
	bucketEndpoint         = os.Getenv("KS3_TEST_BUCKET_ENDPOINT")
	client          *s3.S3 = nil
)

func main() {

	var cre = credentials.NewStaticCredentials(accessKeyID, accessKeySecret, "") //online
	client = s3.New(&aws.Config{
		//Region 可参考 https://docs.ksyun.com/documents/6761
		Region:      region,
		Credentials: cre,
		//Endpoint 可参考 https://docs.ksyun.com/documents/6761
		Endpoint:         endpoint,
		DisableSSL:       true,  //是否禁用https
		LogLevel:         0,     //是否开启日志,0为关闭日志，1为开启日志
		LogHTTPBody:      false, //是否把HTTP请求body打入日志
		S3ForcePathStyle: false,
		Logger:           nil,   //打日志的位置
		DomainMode:       false, //是否开启自定义bucket绑定域名，当开启时 S3ForcePathStyle 参数不生效。
	})

	dir := "/Users/cqc/Desktop/terraFormTest"
	uploader := s3manager.NewUploader(&s3manager.UploadOptions{
		//分块大小 5MB
		PartSize: 0,
		//单文件内部操作的并发任务数
		Parallel: 2,
		//多文件操作时的并发任务数
		Jobs: 2,
		S3:   client,
	})
	//dir 要上传的目录
	//bucket 上传的目标桶
	//prefix 桶下的路径
	uploader.UploadDir(dir, bucket, "aaa/")
}
```