
KS3 Go SDK 使用指南
---------------

### 目录

*   [1、概述](#1)

*   [2、环境准备](#2)

*   [3、初始化](#3)
    *   [3.1 下载安装 SDK](#3.1)
    *   [3.2 获取密钥](#3.2)
    *   [3.3 初始化](#3.3)
    *   [3.4 常见术语介绍](#3.4)
*   [4、空间相关](#4)
     *   [4.1 创建|删除 bucket](#4.1)
     *   [4.2 设置镜像回源规则](#4.2)
     *   [4.3 获取镜像回源规则](#4.3)
     *   [4.4 删除镜像回源规则](#4.4)   
*   [5、对象相关](#5)
    *   [5.1 文件是否存在](#5.1)
    *   [5.2 上传文件](#5.2)
    *   [5.3 下载文件](#5.3)
    *   [5.4 生成文件下载外链](#5.4)
    *   [5.5 生成文件上传外链](#5.5)
    *   [5.6 生成设置文件ACL的外链](#5.6)
    *   [5.7 修改元数据](#5.7)
    *   [5.8 批量删除对象](#5.8)
    *   [5.9 目录删除](#5.9)
    *   [5.10 列举文件](#5.10)
    *   [5.11 抓取远程数据到KS3](#5.11)
    *   [5.12 复制对象](#5.12) 
    *   [5.13 对象标签](#5.1) 
         *   [1 获取对象标签](#5.13.1) 
         *   [2 设置对象标签](#5.13.2) 
         *   [3 删除对象标签](#5.13.3) 
*   [6、分块相关](#6)   
    *   [6.1 初始化分块上传](#6.1)
    *   [6.2 分块上传 - 上传块](#6.2)
    *   [6.3 完成分块上传并合并块](#6.3)
    *   [6.4 取消分块上传](#6.4)
    *   [6.5 罗列分块上传已经上传的块](#6.5)

### 1、概述

此 SDK 是基于 aws-sdk-go 改造，适用于 golang 1.8 开发环境。  

### 2、环境准备

配置 Go 开发环境  

### 3、初始化

### 3.1 下载安装 SDK

*   安装方式：

```
go get github.com/ks3sdklib/aws-sdk-go


```

*   SDK 下载参见 [Github/SDK](https://github.com/ks3sdklib/aws-sdk-go)。

*   使用 demo 参见 [Github/test](https://github.com/ks3sdklib/aws-sdk-go/blob/master/test/ks3_test.go)。  

### 3.2 获取密钥

1.  [开通 KS3 服务, 注册账号](http://www.ksyun.com/user/register)

2.  [进入控制台，获取 AccessKeyID 、AccessKeySecret](https://iam.console.ksyun.com/#!/account)  

### 3.3 初始化

1.  引用相关包

```
     import(

     "github.com/ks3sdklib/aws-sdk-go/aws"

     "github.com/ks3sdklib/aws-sdk-go/aws/credentials"

     "github.com/ks3sdklib/aws-sdk-go/service/s3"

     )



```

2.  初始化客户端

```
     credentials := credentials.NewStaticCredentials("<AccessKeyID>","<AccessKeySecret>","")

     client := s3.New(&aws.Config{

     Region: "BEIJING",

     Credentials: credentials,

     Endpoint:"ks3-cn-beijingcs.ksyun.com",//ks3地址

     DisableSSL:true,//是否禁用https

     LogLevel:1,//是否开启日志,0为关闭日志，1为开启日志

     S3ForcePathStyle:false,//是否强制使用path style方式访问,默认不使用，true开启

     LogHTTPBody:true,//是否把HTTP请求body打入日志

     Logger:os.Stdout,//打日志的位置

     })


```

> 注意：
>
> *   [endpoint 与 Region 对应关系](https://docs.ksyun.com/documents/6761)

### 3.4 常见术语介绍

#### Object（对象，文件）

在 KS3 中，用户操作的基本数据单元是 Object。单个 Object 允许存储 0~48.8TB 的数据。 Object 包含 key 和 data。其中，key 是 Object 的名字；data 是 Object 的数据。key 为 UTF-8 编码，且编码后的长度不得超过 1024 个字符。

#### Key（文件名）

即 Object 的名字，key 为 UTF-8 编码，且编码后的长度不得超过 1024 个字符。Key 中可以带有斜杠，当 Key 中带有斜杠的时候，将会自动在控制台里组织成目录结构。

**其他术语请参考[概念与术语](https://docs.ksyun.com/documents/2286)**



###  [4、空间相关](#4)

### 4.1 创建 bucket

```
func CreateBucket(svc *s3.S3) {
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
}
```


### 4.2 设置镜像回源规则

```
func PutBucketMirrorRules() {

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

}	


```

### 4.3 获取镜像回源规则

```
	//获取镜像回源规则
func GetBucketMirrorRules() {

	params := &s3.GetBucketMirrorInput{
		Bucket: aws.String(BucketName),
	}
	resp, _ := client.GetBucketMirror(params)
	fmt.Println("resp.code is:", resp.HttpCode)
	fmt.Println("resp.Header is:", resp.Header)
	// Pretty-print the response data.
	var bodyStr = string(resp.Body[:])
	fmt.Println("resp.Body is:", bodyStr)

}	


```

### 4.4 删除镜像回源规则

```
	//删除镜像回源规则
    func DeleteBucketMirrorRules() {

	params := &s3.DeleteBucketMirrorInput{
		Bucket: aws.String(BucketName),
	}
	resp, _ := client.DeleteBucketMirror(params)
	fmt.Println("resp.code is:", resp.HttpCode)
	fmt.Println("resp.Header is:", resp.Header)
	// Pretty-print the response data.
	var bodyStr = string(resp.Body[:])
	fmt.Println("resp.Body is:", bodyStr)

}	

```


###  [5、对象相关](#5)


### 5.1  获取元数据
    func headObj(svc *s3.S3) {
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
}




### 5.2 上传文件

```
func putFile(filename string) {

	if len(filename) == 0 {
		filename = "/Users/user/Downloads/aaa.docx"
	}
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
	//获取新的文件名
	fmt.Println(*resp)
}



```

### 5.3 下载文件

```
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

```
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

```
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

```
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

```
func TestModifyObjectMeta(t *testing.T) {
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
}


```

### 5.8 批量删除对象

```
	func DeleteObjects() {

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
    }


```

### 5.9 目录删除

```
	func DeleteBucketPrefix(prefix string) {

	params := &s3.DeleteBucketPrefixInput{
		Bucket: aws.String(""),  //桶名称
		Prefix: aws.String(prefix),    //前缀（目录）名称
	}
	resp, _ := svc.DeleteBucketPrefix(params)    //执行并接受响应结果
	fmt.Println("error keys:",resp.Errors)
	fmt.Println("deleted keys:",resp.Deleted)
    }


```

### 5.10 列举文件

```
	/**
	列举文件
	*/
	func ListObjects() {

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
	}
		
```


### 5.11 抓取远程数据到KS3

```
 func  FetchObcjet() {
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
    }   
}	

```

### 5.12  复制对象

```
 func  CopyObject() {
 
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
    } 

```

###   [5.13 对象标签](###5.13) 

### 5.13.1 设置对象标签

```
func PutObjectTag() {

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
		Bucket:  aws.String(bucketname), // Required
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
}
		
```

### 5.18.2 获取对象标签

```
	/**
	获取对象标签
	*/
	func GetObjectTag(svc *s3.S3) {

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
}
		


```

### 5.18.3 删除对象标签

```
	func DeleteObjectTag() {

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
}
		
```
*   [6、分块相关](#6)

### 6.1 初始化分块上传


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


### 6.2 分块上传 - 上传块

```
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

```
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

```
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

```
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




### 5.7 计算 token（移动端相关）

```
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
