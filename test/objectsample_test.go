package lib

import (
	"bytes"
	"fmt"
	"github.com/ks3sdklib/aws-sdk-go/aws"
	"github.com/ks3sdklib/aws-sdk-go/aws/awserr"
	"github.com/ks3sdklib/aws-sdk-go/service/s3"
	"github.com/ks3sdklib/aws-sdk-go/service/s3/s3manager"
	"github.com/ks3sdklib/aws-sdk-go/service/s3/s3util"
	. "gopkg.in/check.v1"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// TestListObjects 列举bucket下对象
func (s *Ks3utilCommandSuite) TestListObjects(c *C) {
	_, err := client.ListObjects(&s3.ListObjectsInput{
		Bucket: aws.String(bucket),
		//Delimiter: aws.String("/"),       //分隔符，用于对一组参数进行分割的字符
		MaxKeys: aws.Long(int64(1000)), //设置响应体中返回的最大记录数（最后实际返回可能小于该值）。默认为1000。如果你想要的result在1000条以后，你可以设定 marker 的值来调整起始位置。
		Prefix:  aws.String(""),        //限定响应result列表使用的前缀，正如你在电脑中使用的文件夹一样。
		Marker:  aws.String(""),        //指定列举指定空间中对象的起始位置。KS3按照字母排序方式返回result，将从给定的 marker 开始返回列表。
	})
	c.Assert(err, IsNil)
}

// TestPutObject 上传示例 -可设置标签  acl
func (s *Ks3utilCommandSuite) TestPutObject(c *C) {
	//指定目标Object对象标签，可同时设置多个标签，如：TagA=A&TagB=B。
	//说明 Key和Value需要先进行URL编码，如果某项没有“=”，则看作Value为空字符串。详情请见对象标签（https://docs.ksyun.com/documents/39576）。
	v := url.Values{}
	v.Add("name", "yz")
	v.Add("age", "11")
	XAmzTagging := v.Encode()

	object := randLowStr(10)
	createFile(object, 1024*1024*1)
	fd, _ := os.Open(object)
	md5, _ := s3util.GetBase64FileMD5Str(object)
	_, err := client.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(object),
		ACL:         aws.String("private"),
		Body:        fd,
		ContentType: aws.String("application/octet-stream"),
		XAmzTagging: aws.String(XAmzTagging),
		ContentMD5:  aws.String(md5),
	})
	c.Assert(err, IsNil)
	os.Remove(object)
}

// TestPutObjectByLimit 上传示例 -限速
func (s *Ks3utilCommandSuite) TestPutObjectByLimit(c *C) {
	minBandwidth := 1024 * 100 * 8 // 100KB/s
	object := randLowStr(10)
	createFile(object, 1024*1024*1) // 1MB大小的文件
	fd, _ := os.Open(object)
	// 记录开始时间
	startTime := time.Now()
	_, err := client.PutObject(&s3.PutObjectInput{
		Bucket:       aws.String(bucket),
		Key:          aws.String(object),
		Body:         fd,
		TrafficLimit: aws.Long(int64(minBandwidth)), //限制上传速度
	})
	c.Assert(err, IsNil)
	// 计算上传耗时
	elapsed := time.Since(startTime)
	fmt.Println("Upload completed successfully.")
	fmt.Println("Elapsed time:", elapsed)
	os.Remove(object)
}

// TestGetObjectByLimit 下载限速示例
func (s *Ks3utilCommandSuite) TestGetObjectByLimit(c *C) {
	minBandwidth := 1024 * 100 * 8 // 100KB/s
	_, err := client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		ACL:    aws.String("public-read"),
		Body:   strings.NewReader(content),
	})
	c.Assert(err, IsNil)

	//下载
	_, err = client.GetObject(&s3.GetObjectInput{
		Bucket:       aws.String(bucket),
		Key:          aws.String(key),
		TrafficLimit: aws.Long(int64(minBandwidth)), //限制下载速度
	})
	c.Assert(err, IsNil)
}

// TestGetObject 下载示例
func (s *Ks3utilCommandSuite) TestGetObject(c *C) {
	s.PutObject(key, c)
	_, err := client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	c.Assert(err, IsNil)
}

// TestDeleteObject 删除对象
func (s *Ks3utilCommandSuite) TestDeleteObject(c *C) {
	s.PutObject(key, c)
	_, err := client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	c.Assert(err, IsNil)
}

// TestGeneratePresignedUrl 根据方法生成外链
func (s *Ks3utilCommandSuite) TestGeneratePresignedUrl(c *C) {
	_, err := client.GeneratePresignedUrl(&s3.GeneratePresignedUrlInput{
		Bucket: aws.String(bucket), // 设置 bucket 名称
		Key:    aws.String(key),    // 设置 object key
		//TrafficLimit: aws.Long(1000),            // 设置速度限制
		//ContentType:  aws.String("image/jpeg"),  //如果是PUT方法，需要设置content-type
		Expires:    3600,   // 过期时间
		HTTPMethod: s3.GET, //可选值有 PUT, GET, DELETE, HEAD
	})
	c.Assert(err, IsNil)
}

// TestGeneratePUTPresignedUrl 根据外链PUT上传
func (s *Ks3utilCommandSuite) TestGeneratePUTPresignedUrl(c *C) {
	text := "test content"
	md5 := s3util.GetBase64MD5Str(text)
	url, err := client.GeneratePresignedUrl(&s3.GeneratePresignedUrlInput{
		Bucket:      aws.String(bucket),       // 设置 bucket 名称
		Key:         aws.String(key),          // 设置 object key
		ContentType: aws.String("text/plain"), //如果是PUT方法，需要设置content-type
		ContentMd5:  aws.String(md5),          // 文件的MD5
		Expires:     3600,                     // 过期时间
		HTTPMethod:  s3.PUT,                   //可选值有 PUT, GET, DELETE, HEAD
	})
	c.Assert(err, IsNil)
	// 通过外链上传
	httpReq, err := http.NewRequest("PUT", url, strings.NewReader(text))
	if err != nil {
		panic(err)
	}
	// 生成外链时传入的请求头参数需要与此处保持一致
	httpReq.Header.Add("Content-Type", "text/plain")
	httpReq.Header.Add("Content-MD5", md5)
	_, err = http.DefaultClient.Do(httpReq)
	c.Assert(err, IsNil)
	_, err = client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	c.Assert(err, IsNil)
}

// TestGetObjectAcl 获取对象Acl
func (s *Ks3utilCommandSuite) TestGetObjectAcl(c *C) {
	_, err := client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		ACL:    aws.String("public-read"),
		Body:   strings.NewReader(content),
	})
	c.Assert(err, IsNil)

	resp, err := client.GetObjectACL(&s3.GetObjectACLInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	c.Assert(err, IsNil)
	c.Assert(s3.GetAcl(*resp), Equals, s3.PublicRead)
}

// TestPutObjectAcl 设置对象Acl
func (s *Ks3utilCommandSuite) TestPutObjectAcl(c *C) {
	s.PutObject(key, c)
	_, err := client.PutObjectACL(&s3.PutObjectACLInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		ACL:    aws.String(s3.ACLPublicRead),
	})
	c.Assert(err, IsNil)
}

// TestCopyObject 复制对象
func (s *Ks3utilCommandSuite) TestCopyObject(c *C) {
	s.PutObject(key, c)
	//设置对象Tag
	v := url.Values{}
	v.Add("school", "yz")
	v.Add("class", "11")
	XAmzTagging := v.Encode()

	//设置对象元素头
	metadata := make(map[string]*string)
	metadata["yourmetakey1"] = aws.String("yourmetavalue1")
	metadata["yourmetakey2"] = aws.String("yourmetavalue2")

	_, err := client.CopyObject(&s3.CopyObjectInput{
		Bucket:               aws.String(bucket),
		Key:                  aws.String("copy_" + key),
		CopySource:           aws.String("/" + bucket + "/" + key),
		MetadataDirective:    aws.String("REPLACE"),
		Metadata:             metadata,
		XAmzTagging:          aws.String(XAmzTagging),
		XAmzTaggingDirective: aws.String("REPLACE"),
	})
	c.Assert(err, IsNil)
}

// TestUploadPartCopy 分块拷贝用例
func (s *Ks3utilCommandSuite) TestUploadPartCopy(c *C) {
	s.PutObject(key, c)
	dstKey := "xxx/copy/" + key
	//初始化分块
	initResp, err := client.CreateMultipartUpload(&s3.CreateMultipartUploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(dstKey),
	})
	c.Assert(err, IsNil)

	uploadPartCopyResp, err := client.UploadPartCopy(&s3.UploadPartCopyInput{
		Bucket:          aws.String(bucket),
		Key:             aws.String(dstKey),
		CopySource:      aws.String("/" + bucket + "/" + key),
		UploadID:        initResp.UploadID,
		PartNumber:      aws.Long(1),
		CopySourceRange: aws.String("bytes=0-1024"),
	})
	c.Assert(err, IsNil)

	//合并分块
	_, err = client.CompleteMultipartUpload(&s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(bucket),
		Key:      aws.String(dstKey),
		UploadID: initResp.UploadID,
		MultipartUpload: &s3.CompletedMultipartUpload{
			Parts: []*s3.CompletedPart{
				{
					PartNumber: aws.Long(1),
					ETag:       uploadPartCopyResp.CopyPartResult.ETag,
				},
			},
		},
	})
	c.Assert(err, IsNil)
}

// TestFetchObject 抓取第三方URL上传到KS3
func (s *Ks3utilCommandSuite) TestFetchObject(c *C) {
	s.PutObject(key, c)
	// 填写源站对象的url
	sourceUrl := fmt.Sprintf("https://%s.%s/%s", bucket, endpoint, key)
	// 通过第三方URL拉取文件上传
	_, err := client.FetchObject(&s3.FetchObjectInput{
		Bucket:    aws.String(bucket),        // 存储空间名称，必填
		Key:       aws.String(key),           // 对象的key，必填
		SourceUrl: aws.String(sourceUrl),     // 编码后的源站url，必填
		ACL:       aws.String("public-read"), // 对象访问权限，非必填
	})
	c.Assert(err, IsNil)
}

// TestModifyObjectMeta 修改元数据信息
func (s *Ks3utilCommandSuite) TestModifyObjectMeta(c *C) {
	s.PutObject(key, c)

	metadata := make(map[string]*string)
	metadata["yourmetakey1"] = aws.String("yourmetavalue1")
	metadata["yourmetakey2"] = aws.String("yourmetavalue2")

	_, err := client.CopyObject(&s3.CopyObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		////空间名称与对象的object key名称的组合，通过斜杠分隔(’/’)。
		CopySource: aws.String("/" + bucket + "/" + key),
		//指定如何设置目标Object的对象标签。
		//默认值：COPY
		//有效值：
		//1. COPY（默认值）：复制源Object的对象标签到目标 Object。
		//2. REPLACE：忽略源Object的对象标签，直接采用请求中指定的对象标签。
		MetadataDirective: aws.String("REPLACE"),
		Metadata:          metadata,
	})
	c.Assert(err, IsNil)
}

// TestMultipartUpload 分块上传
// 此操作将启动一个分块上传任务并返回 upload ID。在一个确定的分块上传任务中，upload ID用于关联所有分块。连续分块上传请求中的 upload ID由用户指定。在Complete Multipart Upload 和 Abort Multipart Upload请求中同样包含 upload ID。
// 关于请求签名的问题，分块上传为一系列的请求（初始化分块上传，上传块，完成分块上传，终止分块上传），用户启动任务，发送一个或多个分块，最终完成任务。用户需要对每一个请求单独签名。
// 注意: 当你启动分块上传后，并开始上传分块，你必须完成或者放弃上传任务，才能终止因为存储造成的收费。
func (s *Ks3utilCommandSuite) TestMultipartUpload(c *C) {
	//MIN_BANDWIDTH := 1024 * 100 * 8 //100K bits/s
	object := randLowStr(10)
	createFile(object, 1024*1024*10)
	initRet, err := client.CreateMultipartUpload(&s3.CreateMultipartUploadInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		ACL:         aws.String("public-read"),
		ContentType: aws.String("application/octet-stream"),
	})
	c.Assert(err, IsNil)
	//获取分块Id
	uploadId := *initRet.UploadID
	fmt.Printf("%s%s\n", "uploadId=", uploadId)

	f, err := os.Open(object)
	c.Assert(err, IsNil)

	defer f.Close()
	var i int64 = 1
	//组装分块参数
	var compParts []*s3.CompletedPart
	partsNum := []int64{0}
	sc := make([]byte, 52428800)

	for {
		nr, err := f.Read(sc[:])
		if nr < 0 {
			fmt.Fprintf(os.Stderr, "cat: error reading: %s\n", err.Error())
			os.Exit(1)
		} else if nr == 0 {
			break
		} else {
			//上传分块
			//此操作将在分块上传任务中上传一个块。
			//在你上传任一块之前你必须先要启动一个分块上传任务。在你发送一个启动请求后，KS3会给你一个唯一的 upload ID。每次上传块时，都需要将上传ID包含在请求中。
			//块的数量可以是1到10,000中的任意一个（包含1和10,000）。块序号用于标识一个块以及其在对象创建时的位置。如果你上传一个新的块，使用之前已经使用的序列号，那么之前的那个块将会被覆盖。当所有块总大小大于5M时，除了最后一个块没有大小限制外，其余的块的大小均要求在5MB以上。当所有块总大小小于5M时，除了最后一个块没有大小限制外，其余的块的大小均要求在100K以上。如果不符合上述要求，会返回413状态码。
			//为了保证数据在传输过程中没有损坏，请使用 Content-MD5 头部。当使用此头部时，KS3会自动计算出MD5，并根据用户提供的MD5进行校验，如果不匹配，将会返回错误信息。
			//计算sc[:nr]的md5值
			md5 := s3util.GetBase64MD5Str(string(sc[0:nr]))
			resp, err := client.UploadPart(&s3.UploadPartInput{
				Bucket:        aws.String(bucket),
				Key:           aws.String(key),
				PartNumber:    aws.Long(i),
				UploadID:      aws.String(uploadId),
				Body:          bytes.NewReader(sc[0:nr]),
				ContentLength: aws.Long(int64(len(sc[0:nr]))),
				//TrafficLimit:  aws.Long(int64(MIN_BANDWIDTH)),
				ContentMD5: aws.String(md5),
			})
			c.Assert(err, IsNil)
			partsNum = append(partsNum, i)
			compParts = append(compParts, &s3.CompletedPart{PartNumber: &partsNum[i], ETag: resp.ETag})
			i++
		}
	}

	//此操作将完成对象装配之前的块上传任务 。
	//用户启动一个分块上传任务后，会使用 Upload Parts 接口上传所有的块。成功上传所有相关块之后，用户需要调用此接口来完成分块上传。收到完成请求后，KS3将会根据块序号将所有的块组装起来创建一个新的对象。在用户的完成任务请求中需要用户提供分块列表，由于KS3将会按照列表将所有块连接起来，所以要求用户保证所有的块已经完成上传。对于分块列表中的每一个块，用户需要在上传块时添加块序号以及对象的 ETag 头部，KS3则会在块完成上传后回复完成响应。
	//请注意，如果 Complete Multipart Upload 请求失败了，用户应用应当能够进行重试操作。
	_, err = client.CompleteMultipartUpload(&s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(bucket),
		Key:      aws.String(key),
		UploadID: aws.String(uploadId),
		MultipartUpload: &s3.CompletedMultipartUpload{
			Parts: compParts,
		},
	})
	c.Assert(err, IsNil)
	os.Remove(object)
}

// TestPutObjectWithSSEC 上传加密
// 服务器端加密关乎静态数据加密，即 KS3 在将您的数据写入数据中心内的磁盘时会在对象级别上加密这些数据，并在您访问这些数据时为您解密这些数据。
// 只要您验证了您的请求并且拥有访问权限，您访问加密和未加密数据元的方式就没有区别。
// 例如，如果您使用预签名的 URL 来共享您的对象，那么对于加密和解密对象，该 URL 的工作方式是相同的。
func (s *Ks3utilCommandSuite) TestPutObjectWithSSEC(c *C) {
	SSECustomerKey := "0123456789abcdef"
	_, err := client.PutObject(&s3.PutObjectInput{
		Bucket:               aws.String(bucket),
		Key:                  aws.String(key),
		SSECustomerAlgorithm: aws.String("AES256"),                               //加密类型
		SSECustomerKey:       aws.String(s3util.GetBase64Str(SSECustomerKey)),    // 客户端提供的加密密钥
		SSECustomerKeyMD5:    aws.String(s3util.GetBase64MD5Str(SSECustomerKey)), // 客户端提供的通过BASE64编码的通过128位MD5加密的密钥的MD5值
	})
	c.Assert(err, IsNil)
}

// TestHeadObject 判断文件是否存在
func (s *Ks3utilCommandSuite) TestHeadObject(c *C) {
	v := url.Values{}
	v.Add("name", "yz")
	v.Add("age", "11")
	XAmzTagging := v.Encode()

	object := randLowStr(10)
	createFile(object, 1024*1024*1)
	fd, _ := os.Open(content)
	resp, err := client.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		ACL:         aws.String("public-read"),
		Body:        fd,
		XAmzTagging: aws.String(XAmzTagging),
	})
	c.Assert(err, IsNil)
	os.Remove(object)

	_, err = client.HeadObject(&s3.HeadObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		IfNoneMatch: aws.String(*resp.ETag),
	})
	//判断err的状态码是否为304
	if awsErr, ok := err.(awserr.RequestFailure); ok {
		c.Assert(awsErr.StatusCode(), Equals, 304)
	}
}

// TestDeleteObjects 批量删除对象
func (s *Ks3utilCommandSuite) TestDeleteObjects(c *C) {
	s.PutObject("key1", c)
	s.PutObject("key2", c)
	resp, err := client.DeleteObjects(&s3.DeleteObjectsInput{
		Bucket:          aws.String(bucket), // Required
		IsReTurnResults: aws.Boolean(true),
		Delete: &s3.Delete{ // Required
			Objects: []*s3.ObjectIdentifier{
				{
					Key: aws.String("key1"), // Required
				},
				{
					Key: aws.String("key2"), // Required
				},
				// More values...
			},
		},
	})
	c.Assert(err, IsNil)
	c.Assert(len(resp.Errors), Equals, 0)
	c.Assert(len(resp.Deleted), Equals, 2)
}

// TestDeleteBucketPrefix 删除前缀
func (s *Ks3utilCommandSuite) TestDeleteBucketPrefix(c *C) {
	s.PutObject("123/key1", c)
	s.PutObject("123/key2", c)
	resp, err := client.DeleteBucketPrefix(&s3.DeleteBucketPrefixInput{
		Bucket:          aws.String(bucket), // Required
		IsReTurnResults: aws.Boolean(true),
		Prefix:          aws.String("123/"),
	})
	c.Assert(err, IsNil)
	c.Assert(len(resp.Errors), Equals, 0)
	c.Assert(len(resp.Deleted), Equals, 2)
}

// TestTryDeleteBucketPrefix 删除前缀(包含三次重试)
func (s *Ks3utilCommandSuite) TestTryDeleteBucketPrefix(c *C) {
	s.PutObject("123/key1", c)
	s.PutObject("123/key2", c)
	resp, err := client.TryDeleteBucketPrefix(&s3.DeleteBucketPrefixInput{
		Bucket:          aws.String(bucket),
		IsReTurnResults: aws.Boolean(true),
		Prefix:          aws.String("123/"),
	})
	c.Assert(err, IsNil)
	c.Assert(len(resp.Errors), Equals, 0)
	c.Assert(len(resp.Deleted), Equals, 2)
}

// TestRestoreObject 文件解冻
func (s *Ks3utilCommandSuite) TestRestoreObject(c *C) {
	_, err := client.PutObject(&s3.PutObjectInput{
		Bucket:       aws.String(bucket),
		Key:          aws.String(key),
		Body:         strings.NewReader(content),
		StorageClass: aws.String(s3.StorageClassArchive),
	})
	c.Assert(err, IsNil)
	_, err = client.RestoreObject(&s3.RestoreObjectInput{
		Bucket: aws.String(bucket), // bucket名称
		Key:    aws.String(key),    // object key
	})
	c.Assert(err, IsNil)
	s.DeleteObject(key, c)
}

// TestDeleteObjectTagging 删除对象Tag
func (s *Ks3utilCommandSuite) TestDeleteObjectTagging(c *C) {
	s.PutObject(key, c)
	//指定目标Object对象标签
	objTagging := s3.Tagging{
		TagSet: []*s3.Tag{{
			Key:   aws.String("name"),
			Value: aws.String("yz"),
		}, {
			Key:   aws.String("sex"),
			Value: aws.String("female"),
		},
		},
	}
	_, err := client.PutObjectTagging(&s3.PutObjectTaggingInput{
		Bucket:  aws.String(bucket), // Required
		Key:     aws.String(key),
		Tagging: &objTagging,
	})
	c.Assert(err, IsNil)
	_, err = client.DeleteObjectTagging(&s3.DeleteObjectTaggingInput{
		Bucket: aws.String(bucket), // Required
		Key:    aws.String(key),
	})
	c.Assert(err, IsNil)
}

// TestGetObjectTagging 获取对象Tag
func (s *Ks3utilCommandSuite) TestGetObjectTagging(c *C) {
	s.PutObject(key, c)
	//指定目标Object对象标签
	objTagging := s3.Tagging{
		TagSet: []*s3.Tag{{
			Key:   aws.String("name"),
			Value: aws.String("yz"),
		}, {
			Key:   aws.String("sex"),
			Value: aws.String("female"),
		},
		},
	}
	_, err := client.PutObjectTagging(&s3.PutObjectTaggingInput{
		Bucket:  aws.String(bucket), // Required
		Key:     aws.String(key),
		Tagging: &objTagging,
	})
	c.Assert(err, IsNil)
	_, err = client.GetObjectTagging(&s3.GetObjectTaggingInput{
		Bucket: aws.String(bucket), // Required
		Key:    aws.String(key),
	})
	c.Assert(err, IsNil)
}

// TestPutObjectTagging 设置对象Tag
func (s *Ks3utilCommandSuite) TestPutObjectTagging(c *C) {
	s.PutObject(key, c)
	//指定目标Object对象标签
	objTagging := s3.Tagging{
		TagSet: []*s3.Tag{{
			Key:   aws.String("name"),
			Value: aws.String("yz"),
		}, {
			Key:   aws.String("sex"),
			Value: aws.String("female"),
		},
		},
	}
	_, err := client.PutObjectTagging(&s3.PutObjectTaggingInput{
		Bucket:  aws.String(bucket), // Required
		Key:     aws.String(key),
		Tagging: &objTagging,
	})
	c.Assert(err, IsNil)
}

// TestBatchUploadWithClient 上传文件夹
func (s *Ks3utilCommandSuite) TestBatchUploadWithClient(c *C) {
	os.MkdirAll("temp/", os.ModePerm)
	createFile("temp/1.txt", 1024*1024*1)
	createFile("temp/2.txt", 1024*1024*10)
	uploader := s3manager.NewUploader(&s3manager.UploadOptions{
		//分块大小 5MB
		PartSize: 5 * 1024 * 1024,
		//单文件内部操作的并发任务数
		Parallel: 2,
		//多文件操作时的并发任务数
		Jobs:            10,
		S3:              client,
		UploadHidden:    true,
		SkipAlreadyFile: true,
	})
	// RootDir 要上传的目录
	// Bucket 上传的目标桶
	// Prefix 桶下的路径
	err := uploader.UploadDir(&s3manager.UploadDirInput{
		RootDir: "temp/",
		Bucket:  bucket,
		Prefix:  "test-prefix/",
	})
	c.Assert(err, IsNil)
	resp, err := client.ListObjects(&s3.ListObjectsInput{
		Bucket: aws.String(bucket),
		Prefix: aws.String("test-prefix/"),
	})
	c.Assert(err, IsNil)
	c.Assert(len(resp.Contents), Equals, 2)
	os.RemoveAll("temp/")
}

// TestPutObjectCharacterSet 上传文件，测试字符集
func (s *Ks3utilCommandSuite) TestPutObjectCharacterSet(c *C) {
	strList := []string{
		`①②③④⑤⑥⑦⑧⑨⑩⑪⑫⑬⑭⑮⑯⑰⑱⑲⑳⓪❶❷❸❹❺❻❼❽❾❿⓫⓬⓭⓮⓯⓰⓱⓲⓳⓴㊀㊁㊂㊃㊄㊅㊆㊇㊈㊉㈠㈡㈢㈣㈤㈥㈦㈧㈨㈩`,
		`⑴⑵⑶⑷⑸⑹⑺⑻⑼⑽⑾⑿⒀⒁⒂⒃⒄⒅⒆⒇⒈⒉⒊⒋⒌⒍⒎⒏⒐⒑⒒⒓⒔⒕⒖⒗⒘⒙⒚⒛ⅠⅡⅢⅣⅤⅥⅦⅧⅨⅩⅪⅫⅰⅱⅲⅳⅴⅵⅶⅷⅸⅹⒶⒷⒸⒹⒺⒻⒼⒽⒾⒿⓀⓁⓂⓃⓄⓅⓆⓇⓈⓉⓊⓋⓌⓍⓎⓏⓐⓑⓒⓓⓔⓕⓖⓗⓘⓙⓚⓛⓜⓝⓞⓟⓠⓡⓢⓣⓤⓥⓦⓧⓨⓩ⒜⒝⒞⒟⒠⒡⒢⒣⒤⒥⒦⒧⒨⒩⒪⒫⒬⒭⒮⒯⒰⒱⒲⒳⒴⒵`,
		`﹢﹣×÷±/=≌∽≦≧≒﹤﹥≈≡≠=≤≥<>≮≯∷∶∫∮∝∞∧∨∑∏∪∩∈∵∴⊥∥∠⌒⊙√∟⊿㏒㏑%`,
		`‰⅟½⅓⅕⅙⅛⅔⅖⅚⅜¾⅗⅝⅞⅘≂≃≄≅≆≇≈≉≊≋≌≍≎≏≐≑≒≓≔≕≖≗≘≙≚≛≜≝≞≟≠≡≢≣≤≥≦≧≨≩⊰⊱⋛⋚∫∬∭∮∯∰∱∲∳%℅‰‱øØπ`,
		`=, +=, -=, *=, /, =, ==, ===, !=, !==, >, <, >=, <=, +, -, *, /, %, &&, ||, !,  &, |, ^, ~, <<, >>, >>>`,
		`(), [], {}, "", ;, ?, :, \, #,  /* */, ￥, $`,
		`测试中文 ** 特殊符号 && @@ ！@#￥%……&*（）——+{}|：“《》？【】、；‘’，。、`,
		"\n\t\\",
		`abc//////////////`,
	}
	for _, str := range strList {
		srcKey := str
		dstKey := str + "copy"
		s.PutObject(srcKey, c)
		s.CopyObject(srcKey, dstKey, c)
		s.HeadObject(srcKey, c)
		s.HeadObject(dstKey, c)
	}
}

// TestCopyObjectSourceUrlEncoded 复制对象，源URL编码
func (s *Ks3utilCommandSuite) TestCopyObjectSourceUrlEncoded(c *C) {
	srcKey := "测试文件///"
	dstKey := "测试文件_copy///"
	s.PutObject(srcKey, c)
	_, err := client.CopyObject(&s3.CopyObjectInput{
		Bucket:       aws.String(bucket),
		Key:          aws.String(dstKey),
		SourceBucket: aws.String(bucket),
		SourceKey:    aws.String(srcKey),
	})
	c.Assert(err, IsNil)
}

func (s *Ks3utilCommandSuite) TestPutObjectProgress(c *C) {
	object := randLowStr(10)
	createFile(object, 1024*1024*10)
	fd, _ := os.Open(object)
	_, err := client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
		Body:   fd,
		ProgressFn: func(increment, completed, total int64) {
			fmt.Printf("percent: %.2f%%\n", float64(completed)/float64(total)*100)
		},
	})
	c.Assert(err, IsNil)
	os.Remove(object)
}

func (s *Ks3utilCommandSuite) TestGetObjectToFileProgress(c *C) {
	object := randLowStr(10)
	filePath := object + "_download"
	createFile(object, 1024*1024*10)
	fd, _ := os.Open(object)
	_, err := client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
		Body:   fd,
	})
	c.Assert(err, IsNil)
	err = client.GetObjectToFile(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
		ProgressFn: func(increment, completed, total int64) {
			fmt.Printf("percent: %.2f%%\n", float64(completed)/float64(total)*100)
		},
	}, filePath)
	c.Assert(err, IsNil)
	os.Remove(object)
	os.Remove(filePath)
}

func (s *Ks3utilCommandSuite) TestAppendObjectProgress(c *C) {
	object := randLowStr(10)
	createFile(object, 1024*1024*10)
	fd, _ := os.Open(object)
	_, err := client.AppendObject(&s3.AppendObjectInput{
		Bucket:   aws.String(bucket),
		Key:      aws.String(object),
		Position: aws.Long(0),
		Body:     fd,
		ProgressFn: func(increment, completed, total int64) {
			fmt.Printf("percent: %.2f%%\n", float64(completed)/float64(total)*100)
		},
	})
	c.Assert(err, IsNil)
	os.Remove(object)
}

func (s *Ks3utilCommandSuite) TestUploadPartProgress(c *C) {
	object := randLowStr(10)
	createFile(object, 1024*1024*10)
	fd, _ := os.Open(object)

	initResp, err := client.CreateMultipartUpload(&s3.CreateMultipartUploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(err, IsNil)

	partResp, err := client.UploadPart(&s3.UploadPartInput{
		Bucket:     aws.String(bucket),
		Key:        aws.String(object),
		UploadID:   initResp.UploadID,
		PartNumber: aws.Long(1),
		Body:       fd,
		ProgressFn: func(increment, completed, total int64) {
			fmt.Printf("percent: %.2f%%\n", float64(completed)/float64(total)*100)
		},
	})
	c.Assert(err, IsNil)

	_, err = client.CompleteMultipartUpload(&s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(bucket),
		Key:      aws.String(object),
		UploadID: initResp.UploadID,
		MultipartUpload: &s3.CompletedMultipartUpload{
			Parts: []*s3.CompletedPart{
				{
					PartNumber: aws.Long(1),
					ETag:       partResp.ETag,
				},
			},
		},
	})
	c.Assert(err, IsNil)
	os.Remove(object)
}

// TestPutObject10GB 上传10GB文件，报413 Request Entity Too Large错误，错误类型为html
func (s *Ks3utilCommandSuite) TestPutObject10GB(c *C) {
	object := randLowStr(10)
	createFile(object, 1024*1024*1)
	fd, _ := os.Open(object)
	_, err := client.PutObject(&s3.PutObjectInput{
		Bucket:        aws.String(bucket),
		Key:           aws.String(object),
		Body:          fd,
		ContentLength: aws.Long(1024 * 1024 * 1024 * 10),
	})
	c.Assert(err, NotNil)
	c.Assert(strings.Contains(err.Error(), "413 Request Entity Too Large"), Equals, true)
	os.Remove(object)
}

// TestHeadNotExistsObject head不存在的对象，报404错误，request id不为空
func (s *Ks3utilCommandSuite) TestHeadNotExistsObject(c *C) {
	object := randLowStr(10)
	_, err := client.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(err, NotNil)
	c.Assert(strings.Index(err.Error(), "[")+1 != strings.Index(err.Error(), "]"), Equals, true)
}
