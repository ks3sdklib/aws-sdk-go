package lib

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"github.com/ks3sdklib/aws-sdk-go/aws"
	"github.com/ks3sdklib/aws-sdk-go/aws/awserr"
	"github.com/ks3sdklib/aws-sdk-go/aws/credentials"
	"github.com/ks3sdklib/aws-sdk-go/service/s3"
	"github.com/ks3sdklib/aws-sdk-go/service/s3/s3manager"
	. "gopkg.in/check.v1"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
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

// TestListObjectsV2 列举bucket下对象(V2版本)
func (s *Ks3utilCommandSuite) TestListObjectsV2(c *C) {
	// 生成测试文件前缀
	testPrefix := "test_listv2_" + randLowStr(8) + "/"

	// 上传15个测试文件
	testKeys := make([]string, 15)
	for i := 0; i < 15; i++ {
		key := fmt.Sprintf("%sfile_%02d.txt", testPrefix, i)
		testKeys[i] = key
		_, err := client.PutObject(&s3.PutObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
			Body:   strings.NewReader(fmt.Sprintf("content %d", i)),
		})
		c.Assert(err, IsNil)
	}

	// 上传带子目录的文件，用于测试Delimiter
	subDirKey := testPrefix + "subdir/file.txt"
	_, err := client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(subDirKey),
		Body:   strings.NewReader("subdir content"),
	})
	c.Assert(err, IsNil)

	// 测试完成后删除所有文件
	defer func() {
		for _, key := range testKeys {
			s.DeleteObject(key, c)
		}
		s.DeleteObject(subDirKey, c)
	}()

	// 1. 基本列举，验证响应基本字段
	resp, err := client.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(testPrefix),
	})
	c.Assert(err, IsNil)
	c.Assert(resp, NotNil)
	c.Assert(*resp.KeyCount, Equals, int64(16))
	c.Assert(resp.IsTruncated, NotNil)
	c.Assert(*resp.Name, Equals, bucket)
	c.Assert(*resp.Prefix, Equals, testPrefix)
	c.Assert(resp.MaxKeys, NotNil)

	// 验证Contents中对象的元数据
	obj := resp.Contents[0]
	c.Assert(obj.Key, NotNil)
	c.Assert(obj.ETag, NotNil)
	c.Assert(obj.LastModified, NotNil)
	c.Assert(obj.Size, NotNil)
	c.Assert(obj.StorageClass, NotNil)

	// 2. 测试MaxKeys限制
	resp, err = client.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket:  aws.String(bucket),
		Prefix:  aws.String(testPrefix),
		MaxKeys: aws.Long(int64(5)),
	})
	c.Assert(err, IsNil)
	c.Assert(*resp.KeyCount, Equals, int64(5))
	c.Assert(*resp.IsTruncated, Equals, true)
	c.Assert(resp.NextContinuationToken, NotNil)
	c.Assert(*resp.MaxKeys, Equals, int64(5))

	// 3. 测试ContinuationToken分页
	continuationToken := *resp.NextContinuationToken
	var totalKeys int64 = 5
	for continuationToken != "" {
		resp, err = client.ListObjectsV2(&s3.ListObjectsV2Input{
			Bucket:            aws.String(bucket),
			Prefix:            aws.String(testPrefix),
			MaxKeys:           aws.Long(int64(5)),
			ContinuationToken: aws.String(continuationToken),
		})
		c.Assert(err, IsNil)
		c.Assert(resp.ContinuationToken, NotNil)
		c.Assert(*resp.ContinuationToken, Equals, continuationToken)
		totalKeys += *resp.KeyCount

		if resp.IsTruncated != nil && *resp.IsTruncated {
			c.Assert(resp.NextContinuationToken, NotNil)
			continuationToken = *resp.NextContinuationToken
		} else {
			continuationToken = ""
		}
	}
	c.Assert(totalKeys, Equals, int64(16))

	// 4. 测试StartAfter参数
	resp, err = client.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket:     aws.String(bucket),
		Prefix:     aws.String(testPrefix),
		MaxKeys:    aws.Long(int64(5)),
		StartAfter: aws.String(testKeys[4]),
	})
	c.Assert(err, IsNil)
	c.Assert(*resp.KeyCount, Equals, int64(5))
	c.Assert(*resp.Contents[0].Key, Equals, testKeys[5])
	c.Assert(*resp.StartAfter, Equals, testKeys[4])

	// 5. 测试Delimiter分隔符，验证CommonPrefixes
	resp, err = client.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket:    aws.String(bucket),
		Prefix:    aws.String(testPrefix),
		Delimiter: aws.String("/"),
		MaxKeys:   aws.Long(int64(100)),
	})
	c.Assert(err, IsNil)
	c.Assert(*resp.Delimiter, Equals, "/")
	// 验证CommonPrefixes包含子目录前缀
	c.Assert(len(resp.CommonPrefixes) > 0, Equals, true)
	c.Assert(*resp.CommonPrefixes[0].Prefix, Equals, testPrefix+"subdir/")

	// 6. 测试FetchOwner参数
	resp, err = client.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket:     aws.String(bucket),
		Prefix:     aws.String(testPrefix),
		MaxKeys:    aws.Long(int64(5)),
		FetchOwner: aws.Boolean(true),
	})
	c.Assert(err, IsNil)
	c.Assert(*resp.KeyCount, Equals, int64(5))
	// FetchOwner为true时，Contents中的Owner字段应该有值
	c.Assert(resp.Contents[0].Owner, NotNil)
	c.Assert(resp.Contents[0].Owner.ID, NotNil)
	c.Assert(resp.Contents[0].Owner.DisplayName, NotNil)

	// 7. 测试EncodingType参数
	resp, err = client.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket:       aws.String(bucket),
		Prefix:       aws.String(testPrefix),
		MaxKeys:      aws.Long(int64(5)),
		EncodingType: aws.String("url"),
	})
	c.Assert(err, IsNil)
	c.Assert(*resp.EncodingType, Equals, "url")
	c.Assert(*resp.KeyCount, Equals, int64(5))

	// 8. 验证HTTP响应字段
	c.Assert(resp.StatusCode, NotNil)
	c.Assert(*resp.StatusCode, Equals, int64(200))
	c.Assert(resp.Metadata, NotNil)
}

// TestPutObject 上传示例 -可设置标签  acl
func (s *Ks3utilCommandSuite) TestPutObject(c *C) {
	//指定目标Object对象标签，可同时设置多个标签，如：TagA=A&TagB=B。
	//说明 Key和Value需要先进行URL编码，如果某项没有“=”，则看作Value为空字符串。详情请见对象标签（https://docs.ksyun.com/documents/39576）。
	v := url.Values{}
	v.Add("name", "yz")
	v.Add("age", "11")
	Tagging := v.Encode()

	object := randLowStr(10)
	createFile(object, 1024*1024*1)
	fd, _ := os.Open(object)
	md5, _ := s3.GetBase64FileMD5Str(object)
	_, err := client.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(object),
		ACL:         aws.String("private"),
		Body:        fd,
		ContentType: aws.String("application/octet-stream"),
		Tagging:     aws.String(Tagging),
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
	md5 := s3.GetBase64MD5Str(text)
	url, err := client.GeneratePresignedUrl(&s3.GeneratePresignedUrlInput{
		HTTPMethod:  s3.PUT,                    // 请求方法，可选值有 PUT, GET, DELETE, HEAD，必填
		Bucket:      aws.String(bucket),        // 存储空间名称，必填
		Key:         aws.String(key),           // 对象的key，必填
		Expires:     3600,                      // 过期时间，例如，3600（表示1小时），必填
		ACL:         aws.String("public-read"), // 对象访问权限，非必填
		ContentType: aws.String("text/plain"),  // 文件类型，非必填
		ContentMd5:  aws.String(md5),           // 文件的MD5
	})
	c.Assert(err, IsNil)
	// 通过外链上传，此处以Golang代码为例，也可以通过其他方式上传
	httpReq, err := http.NewRequest("PUT", url, strings.NewReader(text))
	c.Assert(err, IsNil)
	// 实际上传时的请求头必须与生成链接时的请求头一致，否则会有签名不一致的问题
	httpReq.Header.Add("x-amz-acl", "public-read")
	httpReq.Header.Add("Content-Type", "text/plain")
	httpReq.Header.Add("Content-MD5", md5)
	resp, err := http.DefaultClient.Do(httpReq)
	c.Assert(err, IsNil)
	c.Assert(resp.StatusCode, Equals, http.StatusOK)

	s.DeleteObject(key, c)
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
	c.Assert(s3.GetCannedACL(resp.Grants), Equals, s3.ACLPublicRead)
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
	Tagging := v.Encode()

	//设置对象元素头
	metadata := make(map[string]*string)
	metadata["yourmetakey1"] = aws.String("yourmetavalue1")
	metadata["yourmetakey2"] = aws.String("yourmetavalue2")

	_, err := client.CopyObject(&s3.CopyObjectInput{
		Bucket:            aws.String(bucket),
		Key:               aws.String("copy_" + key),
		CopySource:        aws.String("/" + bucket + "/" + key),
		MetadataDirective: aws.String("REPLACE"),
		Metadata:          metadata,
		Tagging:           aws.String(Tagging),
		TaggingDirective:  aws.String("REPLACE"),
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

	f, err := os.Open(object)
	c.Assert(err, IsNil)

	defer f.Close()
	var partNum int64 = 1
	// 待合并分块
	var compParts []*s3.CompletedPart
	// 缓冲区，分块大小为5MB
	buffer := make([]byte, 5*1024*1024)
	for {
		nr, err := f.Read(buffer)
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
			md5 := s3.GetBase64MD5Str(string(buffer[0:nr]))
			resp, err := client.UploadPart(&s3.UploadPartInput{
				Bucket:        aws.String(bucket),
				Key:           aws.String(key),
				PartNumber:    aws.Long(partNum),
				UploadID:      aws.String(uploadId),
				Body:          bytes.NewReader(buffer[0:nr]),
				ContentLength: aws.Long(int64(len(buffer[0:nr]))),
				//TrafficLimit:  aws.Long(int64(MIN_BANDWIDTH)),
				ContentMD5: aws.String(md5),
			})
			c.Assert(err, IsNil)
			compParts = append(compParts, &s3.CompletedPart{PartNumber: aws.Long(partNum), ETag: resp.ETag})
			partNum++
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
		SSECustomerAlgorithm: aws.String("AES256"),                           //加密类型
		SSECustomerKey:       aws.String(s3.GetBase64Str(SSECustomerKey)),    // 客户端提供的加密密钥
		SSECustomerKeyMD5:    aws.String(s3.GetBase64MD5Str(SSECustomerKey)), // 客户端提供的通过BASE64编码的通过128位MD5加密的密钥的MD5值
	})
	c.Assert(err, IsNil)
}

// TestHeadObject 判断文件是否存在
func (s *Ks3utilCommandSuite) TestHeadObject(c *C) {
	v := url.Values{}
	v.Add("name", "yz")
	v.Add("age", "11")
	Tagging := v.Encode()

	object := randLowStr(10)
	createFile(object, 1024*1024*1)
	fd, _ := os.Open(content)
	resp, err := client.PutObject(&s3.PutObjectInput{
		Bucket:  aws.String(bucket),
		Key:     aws.String(key),
		ACL:     aws.String("public-read"),
		Body:    fd,
		Tagging: aws.String(Tagging),
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

	// 带Days和JobParameters参数的解冻请求
	_, err = client.RestoreObject(&s3.RestoreObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		RestoreRequest: &s3.RestoreRequest{
			Days: aws.Long(int64(7)), // 解冻持续时间
			//JobParameters: &s3.JobParameters{
			//	Tier: aws.String(s3.RestoreTierStandard), // 解冻优先级: Expedited/Standard/Bulk
			//},
		},
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
	// DirPath 要上传的目录
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
	createFile(object, 1024*1024*10)
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

// TestPresignedMultipartUpload 通过外链分块上传
func (s *Ks3utilCommandSuite) TestPresignedMultipartUpload(c *C) {
	object := randLowStr(10)
	createFile(object, 1024*1024*1)
	fd, _ := os.Open(object)
	// 生成init外链
	initUrl, err := client.GeneratePresignedUrl(&s3.GeneratePresignedUrlInput{
		HTTPMethod: s3.POST,
		Bucket:     aws.String(bucket),
		Key:        aws.String(object),
		Expires:    3600,
		ExtendQueryParams: map[string]*string{
			"uploads": nil,
		},
	})

	initRequest, err := http.NewRequest("POST", initUrl, nil)
	c.Assert(err, IsNil)

	initResp, err := http.DefaultClient.Do(initRequest)
	c.Assert(err, IsNil)

	body, err := io.ReadAll(initResp.Body)
	c.Assert(err, IsNil)

	initXml := struct {
		UploadId string `xml:"UploadId"`
	}{}
	err = xml.Unmarshal(body, &initXml)
	c.Assert(err, IsNil)

	// 生成upload part外链
	uploadPartUrl, err := client.GeneratePresignedUrl(&s3.GeneratePresignedUrlInput{
		HTTPMethod: s3.PUT,
		Bucket:     aws.String(bucket),
		Key:        aws.String(object),
		Expires:    3600,
		ExtendQueryParams: map[string]*string{
			"partNumber": aws.String("1"),
			"uploadId":   aws.String(initXml.UploadId),
		},
	})
	c.Assert(err, IsNil)

	uploadPartRequest, err := http.NewRequest("PUT", uploadPartUrl, fd)
	c.Assert(err, IsNil)

	uploadPartResp, err := http.DefaultClient.Do(uploadPartRequest)
	c.Assert(err, IsNil)

	etag := uploadPartResp.Header.Get("ETag")

	// 生成complete外链
	completeUrl, err := client.GeneratePresignedUrl(&s3.GeneratePresignedUrlInput{
		HTTPMethod: s3.POST,
		Bucket:     aws.String(bucket),
		Key:        aws.String(object),
		Expires:    3600,
		ExtendQueryParams: map[string]*string{
			"uploadId": aws.String(initXml.UploadId),
		},
	})
	c.Assert(err, IsNil)

	completeParts := `<CompleteMultipartUpload>
		<Part>
			<PartNumber>1</PartNumber>
			<ETag>` + etag + `</ETag>
		</Part>	
	</CompleteMultipartUpload>`

	completeRequest, err := http.NewRequest("POST", completeUrl, strings.NewReader(completeParts))
	c.Assert(err, IsNil)

	completeResp, err := http.DefaultClient.Do(completeRequest)
	c.Assert(err, IsNil)

	body, err = io.ReadAll(completeResp.Body)
	c.Assert(err, IsNil)

	// 获取head外链
	headUrl, err := client.GeneratePresignedUrl(&s3.GeneratePresignedUrlInput{
		HTTPMethod: s3.HEAD,
		Bucket:     aws.String(bucket),
		Key:        aws.String(object),
		Expires:    3600,
	})
	c.Assert(err, IsNil)

	headRequest, err := http.NewRequest("HEAD", headUrl, nil)
	c.Assert(err, IsNil)

	headResp, err := http.DefaultClient.Do(headRequest)
	c.Assert(err, IsNil)
	c.Assert(headResp.StatusCode, Equals, 200)

	os.Remove(object)
}

// TestPresignedMultipartCopy 通过外链分块复制
func (s *Ks3utilCommandSuite) TestPresignedMultipartCopy(c *C) {
	object := randLowStr(10)
	createFile(object, 1024*1024*1)
	s.PutObject(object, c)

	// 生成init外链
	initUrl, err := client.GeneratePresignedUrl(&s3.GeneratePresignedUrlInput{
		HTTPMethod: s3.POST,
		Bucket:     aws.String(bucket),
		Key:        aws.String(object + "copy"),
		Expires:    3600,
		ExtendQueryParams: map[string]*string{
			"uploads": nil,
		},
	})
	c.Assert(err, IsNil)

	initRequest, err := http.NewRequest("POST", initUrl, nil)
	c.Assert(err, IsNil)

	initResp, err := http.DefaultClient.Do(initRequest)
	c.Assert(err, IsNil)

	body, err := io.ReadAll(initResp.Body)
	c.Assert(err, IsNil)

	initXml := struct {
		UploadId string `xml:"UploadId"`
	}{}

	err = xml.Unmarshal(body, &initXml)
	c.Assert(err, IsNil)

	// 生成upload part外链
	uploadPartUrl, err := client.GeneratePresignedUrl(&s3.GeneratePresignedUrlInput{
		HTTPMethod: s3.PUT,
		Bucket:     aws.String(bucket),
		Key:        aws.String(object + "copy"),
		Expires:    3600,
		ExtendQueryParams: map[string]*string{
			"partNumber": aws.String("1"),
			"uploadId":   aws.String(initXml.UploadId),
		},
		ExtendHeaders: map[string]*string{
			"X-Amz-Copy-Source": aws.String("/" + bucket + "/" + object),
		},
	})
	c.Assert(err, IsNil)

	uploadPartRequest, err := http.NewRequest("PUT", uploadPartUrl, nil)
	c.Assert(err, IsNil)

	// 设置header
	uploadPartRequest.Header.Set("X-Amz-Copy-Source", "/"+bucket+"/"+object)

	uploadPartResp, err := http.DefaultClient.Do(uploadPartRequest)
	c.Assert(err, IsNil)

	body, err = io.ReadAll(uploadPartResp.Body)
	c.Assert(err, IsNil)

	uploadPartXml := struct {
		ETag string `xml:"ETag"`
	}{}

	err = xml.Unmarshal(body, &uploadPartXml)
	c.Assert(err, IsNil)

	etag := uploadPartXml.ETag

	// 生成complete外链
	completeUrl, err := client.GeneratePresignedUrl(&s3.GeneratePresignedUrlInput{
		HTTPMethod: s3.POST,
		Bucket:     aws.String(bucket),
		Key:        aws.String(object + "copy"),
		Expires:    3600,
		ExtendQueryParams: map[string]*string{
			"uploadId": aws.String(initXml.UploadId),
		},
	})
	c.Assert(err, IsNil)

	completeParts := `<CompleteMultipartUpload>
		<Part>
			<PartNumber>1</PartNumber>
			<ETag>` + etag + `</ETag>
		</Part>
	</CompleteMultipartUpload>`

	completeRequest, err := http.NewRequest("POST", completeUrl, strings.NewReader(completeParts))
	c.Assert(err, IsNil)

	completeResp, err := http.DefaultClient.Do(completeRequest)
	c.Assert(err, IsNil)

	body, err = io.ReadAll(completeResp.Body)
	c.Assert(err, IsNil)

	// 获取head外链
	headUrl, err := client.GeneratePresignedUrl(&s3.GeneratePresignedUrlInput{
		HTTPMethod: s3.HEAD,
		Bucket:     aws.String(bucket),
		Key:        aws.String(object + "copy"),
		Expires:    3600,
	})
	c.Assert(err, IsNil)

	headRequest, err := http.NewRequest("HEAD", headUrl, nil)
	c.Assert(err, IsNil)

	headResp, err := http.DefaultClient.Do(headRequest)
	c.Assert(err, IsNil)
	c.Assert(headResp.StatusCode, Equals, 200)

	os.Remove(object)
}

func (s *Ks3utilCommandSuite) TestUploadFile(c *C) {
	object := randLowStr(10)
	createFile(object, 1024*1024*10)
	defer os.Remove(object)
	defer s.DeleteObject(object, c)

	// 高级上传
	_, err := client.UploadFile(&s3.UploadFileInput{
		Bucket:     aws.String(bucket),
		Key:        aws.String(object),
		UploadFile: aws.String(object),
	})
	c.Assert(err, IsNil)

	// 高级上传，设置块大小
	_, err = client.UploadFile(&s3.UploadFileInput{
		Bucket:     aws.String(bucket),
		Key:        aws.String(object),
		UploadFile: aws.String(object),
		PartSize:   aws.Long(20 * 1024 * 1024),
	})
	c.Assert(err, IsNil)

	// 高级上传，开启断点续传
	_, err = client.UploadFile(&s3.UploadFileInput{
		Bucket:           aws.String(bucket),
		Key:              aws.String(object),
		UploadFile:       aws.String(object),
		EnableCheckpoint: aws.Boolean(true),
		CheckpointDir:    aws.String("./checkpoint/"),
	})
	c.Assert(err, IsNil)

	// 高级上传，设置进度回调
	_, err = client.UploadFile(&s3.UploadFileInput{
		Bucket:           aws.String(bucket),
		Key:              aws.String(object),
		UploadFile:       aws.String(object),
		EnableCheckpoint: aws.Boolean(true),
		CheckpointFile:   aws.String(object + s3.CheckpointFileSuffixUploader),
		PartSize:         aws.Long(1024 * 1024),
		ProgressFn: func(increment, completed, total int64) {
		},
	})
	c.Assert(err, IsNil)

	// 高级上传，设置加密
	_, err = client.UploadFile(&s3.UploadFileInput{
		Bucket:               aws.String(bucket),
		Key:                  aws.String(object),
		UploadFile:           aws.String(object),
		SSECustomerAlgorithm: aws.String("AES256"),
		SSECustomerKey:       aws.String(s3.GetBase64Str(customerKey)),
		SSECustomerKeyMD5:    aws.String(s3.GetBase64MD5Str(customerKey)),
	})
	c.Assert(err, IsNil)
}

func (s *Ks3utilCommandSuite) TestDownloadFile(c *C) {
	object := randLowStr(10)
	createFile(object, 1024*1024*10)
	// 高级上传
	_, err := client.UploadFile(&s3.UploadFileInput{
		Bucket:     aws.String(bucket),
		Key:        aws.String(object),
		UploadFile: aws.String(object),
	})
	c.Assert(err, IsNil)

	// 高级下载
	_, err = client.DownloadFile(&s3.DownloadFileInput{
		Bucket:       aws.String(bucket),
		Key:          aws.String(object),
		DownloadFile: aws.String("./" + object),
	})
	c.Assert(err, IsNil)

	// 高级下载，开启断点续传
	_, err = client.DownloadFile(&s3.DownloadFileInput{
		Bucket:           aws.String(bucket),
		Key:              aws.String(object),
		DownloadFile:     aws.String("./" + object),
		EnableCheckpoint: aws.Boolean(true),
		CheckpointDir:    aws.String("./checkpoint/"),
	})
	c.Assert(err, IsNil)

	// 高级下载，设置进度回调
	_, err = client.DownloadFile(&s3.DownloadFileInput{
		Bucket:           aws.String(bucket),
		Key:              aws.String(object),
		DownloadFile:     aws.String("./" + object),
		EnableCheckpoint: aws.Boolean(true),
		CheckpointFile:   aws.String(object + s3.CheckpointFileSuffixDownloader),
		ProgressFn: func(increment, completed, total int64) {
		},
	})
	c.Assert(err, IsNil)

	s.DeleteObject(object, c)

	// 高级上传，设置加密
	_, err = client.UploadFile(&s3.UploadFileInput{
		Bucket:               aws.String(bucket),
		Key:                  aws.String(object),
		UploadFile:           aws.String(object),
		SSECustomerAlgorithm: aws.String("AES256"),
		SSECustomerKey:       aws.String(s3.GetBase64Str(customerKey)),
		SSECustomerKeyMD5:    aws.String(s3.GetBase64MD5Str(customerKey)),
	})
	c.Assert(err, IsNil)

	// 高级下载，下载加密文件
	_, err = client.DownloadFile(&s3.DownloadFileInput{
		Bucket:               aws.String(bucket),
		Key:                  aws.String(object),
		DownloadFile:         aws.String(object),
		SSECustomerAlgorithm: aws.String("AES256"),
		SSECustomerKey:       aws.String(s3.GetBase64Str(customerKey)),
		SSECustomerKeyMD5:    aws.String(s3.GetBase64MD5Str(customerKey)),
	})
	c.Assert(err, IsNil)
	s.DeleteObject(object, c)
	os.Remove(object)

	// 高级下载，Range下载
	uploadRangeFile := randLowStr(10)
	downloadRangeFile := randLowStr(10)
	createFileWithContent(uploadRangeFile, "123456789")

	// 高级上传
	_, err = client.UploadFile(&s3.UploadFileInput{
		Bucket:     aws.String(bucket),
		Key:        aws.String(uploadRangeFile),
		UploadFile: aws.String(uploadRangeFile),
	})
	c.Assert(err, IsNil)
	os.Remove(uploadRangeFile)

	// 高级下载，Range=[0, 1]，下载前两个字节
	_, err = client.DownloadFile(&s3.DownloadFileInput{
		Bucket:       aws.String(bucket),
		Key:          aws.String(uploadRangeFile),
		DownloadFile: aws.String(downloadRangeFile),
		Range:        []int64{0, 1},
	})
	c.Assert(err, IsNil)

	rangeContent, err := os.ReadFile(downloadRangeFile)
	c.Assert(err, IsNil)
	c.Assert(string(rangeContent), Equals, "12")
	os.Remove(downloadRangeFile)

	// 高级下载，Range=[2, 6]，下载第3到7个字节
	_, err = client.DownloadFile(&s3.DownloadFileInput{
		Bucket:       aws.String(bucket),
		Key:          aws.String(uploadRangeFile),
		DownloadFile: aws.String(downloadRangeFile),
		Range:        []int64{2, 6},
	})
	c.Assert(err, IsNil)

	rangeContent, err = os.ReadFile(downloadRangeFile)
	c.Assert(err, IsNil)
	c.Assert(string(rangeContent), Equals, "34567")
	os.Remove(downloadRangeFile)

	// 高级下载，Range=[6, -1]，下载第7个字节至文件末尾
	_, err = client.DownloadFile(&s3.DownloadFileInput{
		Bucket:       aws.String(bucket),
		Key:          aws.String(uploadRangeFile),
		DownloadFile: aws.String(downloadRangeFile),
		Range:        []int64{6, -1},
	})
	c.Assert(err, IsNil)

	rangeContent, err = os.ReadFile(downloadRangeFile)
	c.Assert(err, IsNil)
	c.Assert(string(rangeContent), Equals, "789")
	os.Remove(downloadRangeFile)

	// 高级下载，Range=[-1, 2]，下载最后2个字节
	_, err = client.DownloadFile(&s3.DownloadFileInput{
		Bucket:       aws.String(bucket),
		Key:          aws.String(uploadRangeFile),
		DownloadFile: aws.String(downloadRangeFile),
		Range:        []int64{-1, 2},
	})
	c.Assert(err, IsNil)

	rangeContent, err = os.ReadFile(downloadRangeFile)
	c.Assert(err, IsNil)
	c.Assert(string(rangeContent), Equals, "89")
	os.Remove(downloadRangeFile)

	// 高级下载，Range=[-1, -1]，下载整个文件
	_, err = client.DownloadFile(&s3.DownloadFileInput{
		Bucket:       aws.String(bucket),
		Key:          aws.String(uploadRangeFile),
		DownloadFile: aws.String(downloadRangeFile),
		Range:        []int64{-1, -1},
	})
	c.Assert(err, IsNil)

	rangeContent, err = os.ReadFile(downloadRangeFile)
	c.Assert(err, IsNil)
	c.Assert(string(rangeContent), Equals, "123456789")
	os.Remove(downloadRangeFile)

	// 高级下载，Range=[2, 1]，下载整个文件
	_, err = client.DownloadFile(&s3.DownloadFileInput{
		Bucket:       aws.String(bucket),
		Key:          aws.String(uploadRangeFile),
		DownloadFile: aws.String(downloadRangeFile),
		Range:        []int64{2, 1},
	})
	c.Assert(err, IsNil)

	rangeContent, err = os.ReadFile(downloadRangeFile)
	c.Assert(err, IsNil)
	c.Assert(string(rangeContent), Equals, "123456789")
	os.Remove(downloadRangeFile)

	s.DeleteObject(uploadRangeFile, c)
}

func (s *Ks3utilCommandSuite) TestUploadReader(c *C) {
	// 事先上传一个 12MB 的对象
	object := randLowStr(10)
	createFile(object, 1024*1024*12)
	_, err := client.UploadFile(&s3.UploadFileInput{
		Bucket:     aws.String(bucket),
		Key:        aws.String(object),
		UploadFile: aws.String(object),
	})
	c.Assert(err, IsNil)
	defer s.DeleteObject(object, c)
	os.Remove(object)

	// 先 GetObject 拿到网络流
	getResp, err := client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(err, IsNil)
	defer getResp.Body.Close()

	// 单块上传：PartSize 大于对象大小
	smallKey := randLowStr(10)
	_, err = client.UploadReader(&s3.UploadReaderInput{
		Bucket:   aws.String(bucket),
		Key:      aws.String(smallKey),
		Body:     getResp.Body,
		PartSize: aws.Long(20 * 1024 * 1024),
	})
	c.Assert(err, IsNil)
	defer s.DeleteObject(smallKey, c)

	// 验证上传内容
	smallResp, err := client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(smallKey),
	})
	c.Assert(err, IsNil)
	smallBody, _ := io.ReadAll(smallResp.Body)
	smallResp.Body.Close()
	c.Assert(len(smallBody), Equals, 1024*1024*12)

	// 多块上传：重新 GetObject 拿网络流
	getResp2, err := client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(err, IsNil)
	defer getResp2.Body.Close()

	largeKey := randLowStr(10)
	_, err = client.UploadReader(&s3.UploadReaderInput{
		Bucket:   aws.String(bucket),
		Key:      aws.String(largeKey),
		Body:     getResp2.Body,
		PartSize: aws.Long(5 * 1024 * 1024),
		TaskNum:  aws.Long(3),
	})
	c.Assert(err, IsNil)
	defer s.DeleteObject(largeKey, c)

	// 验证上传内容
	largeResp, err := client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(largeKey),
	})
	c.Assert(err, IsNil)
	largeBody, _ := io.ReadAll(largeResp.Body)
	largeResp.Body.Close()
	c.Assert(len(largeBody), Equals, 1024*1024*12)
}

func (s *Ks3utilCommandSuite) TestCopyFile(c *C) {
	object := randLowStr(10)
	dstObject := object + "_copy"
	createFile(object, 1024*1024*10)
	// 高级上传
	_, err := client.UploadFile(&s3.UploadFileInput{
		Bucket:     aws.String(bucket),
		Key:        aws.String(object),
		UploadFile: aws.String(object),
	})
	c.Assert(err, IsNil)

	// 高级复制
	_, err = client.CopyFile(&s3.CopyFileInput{
		Bucket:       aws.String(bucket),
		Key:          aws.String(dstObject),
		SourceBucket: aws.String(bucket),
		SourceKey:    aws.String(object),
	})
	c.Assert(err, IsNil)
	s.DeleteObject(dstObject, c)

	// 高级复制，设置分块大小
	_, err = client.CopyFile(&s3.CopyFileInput{
		Bucket:       aws.String(bucket),
		Key:          aws.String(dstObject),
		SourceBucket: aws.String(bucket),
		SourceKey:    aws.String(object),
		PartSize:     aws.Long(20 * 1024 * 1024),
	})
	c.Assert(err, IsNil)
	s.DeleteObject(dstObject, c)

	// 高级复制，开启断点续传
	_, err = client.CopyFile(&s3.CopyFileInput{
		Bucket:           aws.String(bucket),
		Key:              aws.String(dstObject),
		SourceBucket:     aws.String(bucket),
		SourceKey:        aws.String(object),
		EnableCheckpoint: aws.Boolean(true),
		CheckpointDir:    aws.String("./checkpoint/"),
	})
	c.Assert(err, IsNil)
	s.DeleteObject(dstObject, c)

	// 高级复制，设置进度回调
	_, err = client.CopyFile(&s3.CopyFileInput{
		Bucket:           aws.String(bucket),
		Key:              aws.String(dstObject),
		SourceBucket:     aws.String(bucket),
		SourceKey:        aws.String(object),
		EnableCheckpoint: aws.Boolean(true),
		CheckpointFile:   aws.String(object + s3.CheckpointFileSuffixCopier),
		ProgressFn: func(increment, completed, total int64) {
		},
	})
	c.Assert(err, IsNil)
	s.DeleteObject(dstObject, c)

	// 高级复制，设置加密
	_, err = client.CopyFile(&s3.CopyFileInput{
		Bucket:               aws.String(bucket),
		Key:                  aws.String(dstObject),
		SourceBucket:         aws.String(bucket),
		SourceKey:            aws.String(object),
		SSECustomerAlgorithm: aws.String("AES256"),
		SSECustomerKey:       aws.String(s3.GetBase64Str(customerKey)),
		SSECustomerKeyMD5:    aws.String(s3.GetBase64MD5Str(customerKey)),
	})
	c.Assert(err, IsNil)
	s.DeleteObject(dstObject, c)

	s.DeleteObject(object, c)
	// 高级上传，设置加密
	_, err = client.UploadFile(&s3.UploadFileInput{
		Bucket:               aws.String(bucket),
		Key:                  aws.String(object),
		UploadFile:           aws.String(object),
		SSECustomerAlgorithm: aws.String("AES256"),
		SSECustomerKey:       aws.String(s3.GetBase64Str(customerKey)),
		SSECustomerKeyMD5:    aws.String(s3.GetBase64MD5Str(customerKey)),
	})
	c.Assert(err, IsNil)

	// 高级复制，复制加密文件
	_, err = client.CopyFile(&s3.CopyFileInput{
		Bucket:                         aws.String(bucket),
		Key:                            aws.String(dstObject),
		SourceBucket:                   aws.String(bucket),
		SourceKey:                      aws.String(object),
		CopySourceSSECustomerAlgorithm: aws.String("AES256"),
		CopySourceSSECustomerKey:       aws.String(s3.GetBase64Str(customerKey)),
		CopySourceSSECustomerKeyMD5:    aws.String(s3.GetBase64MD5Str(customerKey)),
	})
	c.Assert(err, IsNil)
	s.DeleteObject(dstObject, c)
	s.DeleteObject(object, c)

	os.Remove(object)
}

func (s *Ks3utilCommandSuite) TestCopyFileAcrossRegion(c *C) {
	// 目标桶client
	var cre = credentials.NewStaticCredentials(accessKeyID, accessKeySecret, "")
	dstClient := s3.New(&aws.Config{
		Credentials: cre,                           // 访问凭证
		Region:      "SHANGHAI",                    // 填写您的Region
		Endpoint:    "ks3-cn-shanghai.ksyuncs.com", // 填写您的Endpoint
	})

	// 创建上海的桶
	dstBucket := commonNamePrefix + randLowStr(10)
	_, err := dstClient.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(dstBucket),
	})

	object := randLowStr(10)
	dstObject := object + "_copy"
	createFile(object, 1024*1024*10)
	// 高级上传
	_, err = client.UploadFile(&s3.UploadFileInput{
		Bucket:     aws.String(bucket),
		Key:        aws.String(object),
		UploadFile: aws.String(object),
	})
	c.Assert(err, IsNil)

	// 高级复制（跨region）
	_, err = client.CopyFileAcrossRegion(&s3.CopyFileInput{
		Bucket:       aws.String(dstBucket),
		Key:          aws.String(dstObject),
		SourceBucket: aws.String(bucket),
		SourceKey:    aws.String(object),
	}, dstClient)
	c.Assert(err, IsNil)
	s.DeleteObjectWithClient(dstClient, dstBucket, dstObject, c)

	// 高级复制（跨region），设置分块大小
	_, err = client.CopyFileAcrossRegion(&s3.CopyFileInput{
		Bucket:       aws.String(dstBucket),
		Key:          aws.String(dstObject),
		SourceBucket: aws.String(bucket),
		SourceKey:    aws.String(object),
		PartSize:     aws.Long(20 * 1024 * 1024),
	}, dstClient)
	c.Assert(err, IsNil)
	s.DeleteObjectWithClient(dstClient, dstBucket, dstObject, c)

	// 高级复制（跨region），开启断点续传
	_, err = client.CopyFileAcrossRegion(&s3.CopyFileInput{
		Bucket:           aws.String(dstBucket),
		Key:              aws.String(dstObject),
		SourceBucket:     aws.String(bucket),
		SourceKey:        aws.String(object),
		EnableCheckpoint: aws.Boolean(true),
		CheckpointDir:    aws.String("./checkpoint/"),
	}, dstClient)
	c.Assert(err, IsNil)
	s.DeleteObjectWithClient(dstClient, dstBucket, dstObject, c)

	// 高级复制（跨region），设置进度回调
	_, err = client.CopyFileAcrossRegion(&s3.CopyFileInput{
		Bucket:           aws.String(dstBucket),
		Key:              aws.String(dstObject),
		SourceBucket:     aws.String(bucket),
		SourceKey:        aws.String(object),
		EnableCheckpoint: aws.Boolean(true),
		CheckpointFile:   aws.String(object + s3.CheckpointFileSuffixCopier),
		ProgressFn: func(increment, completed, total int64) {
		},
	}, dstClient)
	c.Assert(err, IsNil)
	s.DeleteObjectWithClient(dstClient, dstBucket, dstObject, c)

	// 高级复制（跨region），设置加密
	_, err = client.CopyFileAcrossRegion(&s3.CopyFileInput{
		Bucket:               aws.String(dstBucket),
		Key:                  aws.String(dstObject),
		SourceBucket:         aws.String(bucket),
		SourceKey:            aws.String(object),
		SSECustomerAlgorithm: aws.String("AES256"),
		SSECustomerKey:       aws.String(s3.GetBase64Str(customerKey)),
		SSECustomerKeyMD5:    aws.String(s3.GetBase64MD5Str(customerKey)),
	}, dstClient)
	c.Assert(err, IsNil)
	s.DeleteObjectWithClient(dstClient, dstBucket, dstObject, c)

	s.DeleteObject(object, c)
	// 高级上传，设置加密
	_, err = client.UploadFile(&s3.UploadFileInput{
		Bucket:               aws.String(bucket),
		Key:                  aws.String(object),
		UploadFile:           aws.String(object),
		SSECustomerAlgorithm: aws.String("AES256"),
		SSECustomerKey:       aws.String(s3.GetBase64Str(customerKey)),
		SSECustomerKeyMD5:    aws.String(s3.GetBase64MD5Str(customerKey)),
	})
	c.Assert(err, IsNil)

	// 高级复制（跨region），复制加密文件
	_, err = client.CopyFileAcrossRegion(&s3.CopyFileInput{
		Bucket:                         aws.String(dstBucket),
		Key:                            aws.String(dstObject),
		SourceBucket:                   aws.String(bucket),
		SourceKey:                      aws.String(object),
		CopySourceSSECustomerAlgorithm: aws.String("AES256"),
		CopySourceSSECustomerKey:       aws.String(s3.GetBase64Str(customerKey)),
		CopySourceSSECustomerKeyMD5:    aws.String(s3.GetBase64MD5Str(customerKey)),
	}, dstClient)
	c.Assert(err, IsNil)

	s.DeleteObjectWithClient(dstClient, dstBucket, dstObject, c)
	s.DeleteObject(object, c)

	os.Remove(object)

	// 删除上海的桶
	_, err = dstClient.DeleteBucket(&s3.DeleteBucketInput{
		Bucket: aws.String(dstBucket),
	})
	c.Assert(err, IsNil)
}

func (s *Ks3utilCommandSuite) TestBucketNameEmpty(c *C) {
	// PutObject
	_, err := client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(""),
		Key:    aws.String(key),
		Body:   strings.NewReader("test"),
	})
	c.Assert(err, NotNil)
	c.Assert(strings.Contains(err.Error(), "input member Bucket must not be empty"), Equals, true)

	// GetObject
	_, err = client.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(""),
		Key:    aws.String(key),
	})
	c.Assert(err, NotNil)
	c.Assert(strings.Contains(err.Error(), "input member Bucket must not be empty"), Equals, true)

	// DeleteObject
	_, err = client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(""),
		Key:    aws.String(key),
	})
	c.Assert(err, NotNil)
	c.Assert(strings.Contains(err.Error(), "input member Bucket must not be empty"), Equals, true)
}

func (s *Ks3utilCommandSuite) TestObjectKeyEmpty(c *C) {
	// PutObject
	_, err := client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(""),
		Body:   strings.NewReader("test"),
	})
	c.Assert(err, NotNil)
	c.Assert(strings.Contains(err.Error(), "input member Key must not be empty"), Equals, true)

	// GetObject
	_, err = client.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(""),
	})
	c.Assert(err, NotNil)
	c.Assert(strings.Contains(err.Error(), "input member Key must not be empty"), Equals, true)

	// DeleteObject
	_, err = client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(""),
	})
	c.Assert(err, NotNil)
	c.Assert(strings.Contains(err.Error(), "input member Key must not be empty"), Equals, true)
}

func (s *Ks3utilCommandSuite) TestPutObjectWithExtendHeaders(c *C) {
	object := randLowStr(10)
	createFile(object, 1024*1024*10)
	fd, _ := os.Open(object)

	// 设置扩展头部
	_, err := client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
		Body:   fd,
		ExtendHeaders: map[string]*string{
			"X-Amz-Storage-Class": aws.String(s3.StorageClassIA),
		},
	})
	c.Assert(err, IsNil)

	// 验证扩展头部是否生效
	headResp, err := client.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(err, IsNil)
	c.Assert(*headResp.Metadata["X-Amz-Storage-Class"], Equals, s3.StorageClassIA)

	// 设置值为空的扩展头部
	_, err = client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
		Body:   fd,
		ExtendHeaders: map[string]*string{
			"header1": nil,
			"header2": aws.String(""),
		},
	})
	c.Assert(err, IsNil)

	// 设置非法扩展头部
	_, err = client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
		Body:   fd,
		ExtendHeaders: map[string]*string{
			"": aws.String("value1"),
		},
	})
	c.Assert(err, NotNil)
	c.Assert(strings.Contains(err.Error(), "invalid extend header field name"), Equals, true)

	// 设置非法扩展头部
	_, err = client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
		Body:   fd,
		ExtendHeaders: map[string]*string{
			"\n": aws.String("value1"),
		},
	})
	c.Assert(err, NotNil)
	c.Assert(strings.Contains(err.Error(), "invalid extend header field name"), Equals, true)

	// 删除对象
	s.DeleteObject(object, c)
	os.Remove(object)
}

func (s *Ks3utilCommandSuite) TestPutObjectWithExtendQueryParams(c *C) {
	// 设置扩展查询参数
	resp, err := client.ListObjects(&s3.ListObjectsInput{
		Bucket: aws.String(bucket),
		ExtendQueryParams: map[string]*string{
			"max-keys": aws.String("10"),
		},
	})
	c.Assert(err, IsNil)
	c.Assert(*resp.MaxKeys, Equals, int64(10))

	// 测试扩展查询参数包含多个键值对
	req, _ := client.ListObjectsRequest(&s3.ListObjectsInput{
		Bucket: aws.String(bucket),
		ExtendQueryParams: map[string]*string{
			"param1": aws.String("value1"),
			"param2": aws.String("value2"),
		},
	})
	err = req.Send()
	c.Assert(err, IsNil)
	c.Assert(req.HTTPRequest.URL.Query().Encode(), Equals, "param1=value1&param2=value2")

	// 测试扩展查询参数的值为空或nil的情况
	req, _ = client.ListObjectsRequest(&s3.ListObjectsInput{
		Bucket: aws.String(bucket),
		ExtendQueryParams: map[string]*string{
			"param1": nil,
			"param2": aws.String(""),
		},
	})
	err = req.Send()
	c.Assert(err, IsNil)
	c.Assert(req.HTTPRequest.URL.Query().Encode(), Equals, "param1=&param2=")

	// 测试扩展查询参数的键为空字符串或特殊字符的情况
	req, _ = client.ListObjectsRequest(&s3.ListObjectsInput{
		Bucket: aws.String(bucket),
		ExtendQueryParams: map[string]*string{
			"":   aws.String("value1"),
			"\n": aws.String("value2"),
		},
	})
	err = req.Send()
	c.Assert(err, IsNil)
	c.Assert(req.HTTPRequest.URL.Query().Encode(), Equals, "%0A=value2")

	// 测试扩展查询参数包含子资源
	object := randLowStr(10)
	s.PutObject(object, c)
	getResp, err := client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
		ExtendQueryParams: map[string]*string{
			"acl": aws.String(""),
		},
	})
	c.Assert(err, IsNil)
	body, err := io.ReadAll(getResp.Body)
	c.Assert(err, IsNil)
	c.Assert(strings.Contains(string(body), "AccessControlPolicy"), Equals, true)
	s.DeleteObject(object, c)
	os.Remove(object)
}

func (s *Ks3utilCommandSuite) TestObjectMigration(c *C) {
	c.Skip("Skip TestObjectMigration")
	srcBucketName := "test-bucket1"
	srcObjectKey := "test-file1"
	dstBucketName := "test-bucket2"
	dstObjectKey := "test-file2"

	// 上传源对象
	_, err := client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(srcBucketName),
		Key:    aws.String(srcObjectKey),
		Body:   strings.NewReader(content),
	})
	c.Assert(err, IsNil)

	// 创建迁移任务
	_, err = client.PutObjectMigration(&s3.PutObjectMigrationInput{
		SourceBucket: aws.String(srcBucketName),
		SourceKey:    aws.String(srcObjectKey),
		Bucket:       aws.String(dstBucketName),
		Key:          aws.String(dstObjectKey),
		StorageClass: aws.String(s3.StorageClassIA),
	})
	c.Assert(err, IsNil)

	// 等待迁移任务完成
	time.Sleep(time.Second * 1)

	// 查看迁移任务状态
	resp, err := client.GetObjectMigration(&s3.GetObjectMigrationInput{
		Bucket: aws.String(dstBucketName),
		Key:    aws.String(dstObjectKey),
	})
	c.Assert(err, IsNil)
	c.Assert(*resp.MigrationConfiguration.Status, Equals, "Succeed")

	// 查看目标对象是否存在
	headResp, err := client.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(dstBucketName),
		Key:    aws.String(dstObjectKey),
	})
	c.Assert(err, IsNil)
	c.Assert(*headResp.Metadata[s3.HTTPHeaderAmzStorageClass], Equals, s3.StorageClassIA)

	// 删除源对象
	_, err = client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(srcBucketName),
		Key:    aws.String(srcObjectKey),
	})
	c.Assert(err, IsNil)

	// 删除目标对象
	_, err = client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(dstBucketName),
		Key:    aws.String(dstObjectKey),
	})
	c.Assert(err, IsNil)

	// 上传源对象，设置服务端加密
	putResp, err := client.PutObject(&s3.PutObjectInput{
		Bucket:               aws.String(srcBucketName),
		Key:                  aws.String(srcObjectKey),
		Body:                 strings.NewReader(content),
		ServerSideEncryption: aws.String(s3.AlgorithmAES256),
	})
	c.Assert(err, IsNil)
	c.Assert(*putResp.ServerSideEncryption, Equals, s3.AlgorithmAES256)

	// 创建迁移任务
	_, err = client.PutObjectMigration(&s3.PutObjectMigrationInput{
		SourceBucket:         aws.String(srcBucketName),
		SourceKey:            aws.String(srcObjectKey),
		Bucket:               aws.String(dstBucketName),
		Key:                  aws.String(dstObjectKey),
		ServerSideEncryption: aws.String(s3.AlgorithmAES256),
		StorageClass:         aws.String(s3.StorageClassIA),
	})
	c.Assert(err, IsNil)

	// 等待迁移任务完成
	time.Sleep(time.Second * 1)

	// 查看迁移任务状态
	resp, err = client.GetObjectMigration(&s3.GetObjectMigrationInput{
		Bucket: aws.String(dstBucketName),
		Key:    aws.String(dstObjectKey),
	})
	c.Assert(err, IsNil)
	c.Assert(*resp.MigrationConfiguration.Status, Equals, "Succeed")

	// 查看目标对象是否存在
	headResp, err = client.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(dstBucketName),
		Key:    aws.String(dstObjectKey),
	})
	c.Assert(err, IsNil)
	c.Assert(*headResp.Metadata[s3.HTTPHeaderAmzStorageClass], Equals, s3.StorageClassIA)
	c.Assert(*headResp.ServerSideEncryption, Equals, s3.AlgorithmAES256)

	// 删除源对象
	_, err = client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(srcBucketName),
		Key:    aws.String(srcObjectKey),
	})
	c.Assert(err, IsNil)

	// 	// 删除目标对象
	_, err = client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(dstBucketName),
		Key:    aws.String(dstObjectKey),
	})
	c.Assert(err, IsNil)

	// 上传源对象，设置客户端加密
	putResp, err = client.PutObject(&s3.PutObjectInput{
		Bucket:               aws.String(srcBucketName),
		Key:                  aws.String(srcObjectKey),
		Body:                 strings.NewReader(content),
		SSECustomerAlgorithm: aws.String(s3.AlgorithmAES256),
		SSECustomerKey:       aws.String(s3.GetBase64Str(customerKey)),
		SSECustomerKeyMD5:    aws.String(s3.GetBase64MD5Str(customerKey)),
	})
	c.Assert(err, IsNil)
	c.Assert(*putResp.SSECustomerAlgorithm, Equals, s3.AlgorithmAES256)
	c.Assert(*putResp.SSECustomerKeyMD5, Equals, s3.GetBase64MD5Str(customerKey))

	// 创建迁移任务
	_, err = client.PutObjectMigration(&s3.PutObjectMigrationInput{
		SourceBucket:               aws.String(srcBucketName),
		SourceKey:                  aws.String(srcObjectKey),
		Bucket:                     aws.String(dstBucketName),
		Key:                        aws.String(dstObjectKey),
		SourceSSECustomerAlgorithm: aws.String(s3.AlgorithmAES256),
		SourceSSECustomerKey:       aws.String(s3.GetBase64Str(customerKey)),
		SourceSSECustomerKeyMD5:    aws.String(s3.GetBase64MD5Str(customerKey)),
		SSECustomerAlgorithm:       aws.String(s3.AlgorithmAES256),
		SSECustomerKey:             aws.String(s3.GetBase64Str(customerKey)),
		SSECustomerKeyMD5:          aws.String(s3.GetBase64MD5Str(customerKey)),
		StorageClass:               aws.String(s3.StorageClassIA),
	})
	c.Assert(err, IsNil)

	// 等待迁移任务完成
	time.Sleep(time.Second * 1)

	// 查看迁移任务状态
	resp, err = client.GetObjectMigration(&s3.GetObjectMigrationInput{
		Bucket: aws.String(dstBucketName),
		Key:    aws.String(dstObjectKey),
	})
	c.Assert(err, IsNil)
	c.Assert(*resp.MigrationConfiguration.Status, Equals, "Succeed")

	// 查看目标对象是否存在
	headResp, err = client.HeadObject(&s3.HeadObjectInput{
		Bucket:               aws.String(dstBucketName),
		Key:                  aws.String(dstObjectKey),
		SSECustomerAlgorithm: aws.String(s3.AlgorithmAES256),
		SSECustomerKey:       aws.String(s3.GetBase64Str(customerKey)),
		SSECustomerKeyMD5:    aws.String(s3.GetBase64MD5Str(customerKey)),
	})
	c.Assert(err, IsNil)
	c.Assert(*headResp.Metadata[s3.HTTPHeaderAmzStorageClass], Equals, s3.StorageClassIA)
	c.Assert(*headResp.SSECustomerAlgorithm, Equals, s3.AlgorithmAES256)
	c.Assert(*headResp.SSECustomerKeyMD5, Equals, s3.GetBase64MD5Str(customerKey))

	// 删除源对象
	_, err = client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(srcBucketName),
		Key:    aws.String(srcObjectKey),
	})
	c.Assert(err, IsNil)

	// 	// 删除目标对象
	_, err = client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(dstBucketName),
		Key:    aws.String(dstObjectKey),
	})
	c.Assert(err, IsNil)
}

func (s *Ks3utilCommandSuite) TestCreateJob(c *C) {
	c.Skip("Skip TestCreateJob")
	// 新建设置对象ACL操作，设置预定义ACL
	createJobInput := &s3.CreateJobInput{
		CreateJobRequest: &s3.CreateJobRequest{
			Description: aws.String("this-is-a-test-job"),
			Priority:    aws.Long(100),
			Operation: &s3.JobOperation{
				KS3PutObjectAcl: &s3.KS3PutObjectAcl{
					CannedAccessControlList: aws.String(s3.ACLPublicRead),
				},
			},
			Manifest: &s3.JobManifest{
				Location: &s3.ManifestLocation{
					Filters: []*s3.LocationFilter{
						{
							Bucket:   aws.String("krn:ksc:ks3:::test-bucket"),
							Prefixes: []string{"prefix1/"},
						},
					},
				},
				Spec: &s3.ManifestSpec{
					Format: aws.String("KS3BatchOperations_Bucket_V1"),
				},
			},
			Report: &s3.JobReport{
				Bucket:      aws.String("krn:ksc:ks3:::test-bucket"),
				Prefix:      aws.String("result/"),
				Enabled:     aws.Boolean(true),
				ReportScope: aws.String("FailedTasksOnly"),
			},
		},
	}
	createResp, err := client.CreateJob(createJobInput)
	c.Assert(err, IsNil)
	jobId1 := *createResp.CreateJobResult.JobId

	resp, err := client.DescribeJob(&s3.DescribeJobInput{
		JobId: aws.String(jobId1),
	})
	c.Assert(err, IsNil)
	c.Assert(*resp.DescribeJobResult.JobId, Equals, jobId1)
	c.Assert(*resp.DescribeJobResult.Description, Equals, "this-is-a-test-job")
	c.Assert(*resp.DescribeJobResult.Priority, Equals, int64(100))
	c.Assert(len(resp.DescribeJobResult.Manifest.Location.Filters), Equals, 1)
	c.Assert(*resp.DescribeJobResult.Manifest.Location.Filters[0].Bucket, Equals, "krn:ksc:ks3:::test-bucket")
	c.Assert(len(resp.DescribeJobResult.Manifest.Location.Filters[0].Prefixes), Equals, 1)
	c.Assert(resp.DescribeJobResult.Manifest.Location.Filters[0].Prefixes[0], Equals, "prefix1/")
	c.Assert(*resp.DescribeJobResult.Manifest.Spec.Format, Equals, "KS3BatchOperations_Bucket_V1")
	c.Assert(*resp.DescribeJobResult.Operation.KS3PutObjectAcl.CannedAccessControlList, Equals, s3.ACLPublicRead)
	c.Assert(*resp.DescribeJobResult.Report.Bucket, Equals, "krn:ksc:ks3:::test-bucket")
	c.Assert(*resp.DescribeJobResult.Report.Prefix, Equals, "result/")
	c.Assert(*resp.DescribeJobResult.Report.Enabled, Equals, true)
	c.Assert(*resp.DescribeJobResult.Report.ReportScope, Equals, "FailedTasksOnly")

	// 新建设置对象ACL操作，设置自定义ACL
	createJobInput.CreateJobRequest.ClientRequestToken = nil
	createJobInput.CreateJobRequest.Operation = &s3.JobOperation{
		KS3PutObjectAcl: &s3.KS3PutObjectAcl{
			AccessControlList: &s3.JobAccessControlList{
				Grants: []*s3.JobGrant{
					{
						Grantee:    aws.String("12345678"),
						Permission: aws.String("READ"),
					},
				},
			},
		},
	}
	createJobInput.CreateJobRequest.Manifest.Location.Filters[0].Prefixes = []string{"prefix2/"}
	createResp, err = client.CreateJob(createJobInput)
	c.Assert(err, IsNil)
	jobId2 := *createResp.CreateJobResult.JobId

	resp, err = client.DescribeJob(&s3.DescribeJobInput{
		JobId: aws.String(jobId2),
	})
	c.Assert(err, IsNil)
	c.Assert(*resp.DescribeJobResult.JobId, Equals, jobId2)
	c.Assert(len(resp.DescribeJobResult.Operation.KS3PutObjectAcl.AccessControlList.Grants), Equals, 1)
	c.Assert(*resp.DescribeJobResult.Operation.KS3PutObjectAcl.AccessControlList.Grants[0].Grantee, Equals, "12345678")
	c.Assert(*resp.DescribeJobResult.Operation.KS3PutObjectAcl.AccessControlList.Grants[0].Permission, Equals, "READ")

	// 新建解冻操作
	createJobInput.CreateJobRequest.ClientRequestToken = nil
	createJobInput.CreateJobRequest.Operation = &s3.JobOperation{
		KS3RestoreObject: &s3.KS3RestoreObject{
			StorageClass: aws.String(s3.StorageClassArchive),
			Days:         aws.Long(7),
		},
	}
	createJobInput.CreateJobRequest.Manifest.Location.Filters[0].Prefixes = []string{"prefix3/"}
	createResp, err = client.CreateJob(createJobInput)
	c.Assert(err, IsNil)
	jobId3 := *createResp.CreateJobResult.JobId

	resp, err = client.DescribeJob(&s3.DescribeJobInput{
		JobId: aws.String(jobId3),
	})
	c.Assert(err, IsNil)
	c.Assert(*resp.DescribeJobResult.JobId, Equals, jobId3)
	c.Assert(*resp.DescribeJobResult.Operation.KS3RestoreObject.StorageClass, Equals, s3.StorageClassArchive)
	c.Assert(*resp.DescribeJobResult.Operation.KS3RestoreObject.Days, Equals, int64(7))

	// 新建删除操作
	createJobInput.CreateJobRequest.ClientRequestToken = nil
	createJobInput.CreateJobRequest.Operation = &s3.JobOperation{
		KS3DeleteObject: &s3.KS3DeleteObject{},
	}
	createJobInput.CreateJobRequest.Manifest.Location.Filters[0].Prefixes = []string{"prefix4/"}
	createResp, err = client.CreateJob(createJobInput)
	c.Assert(err, IsNil)
	jobId4 := *createResp.CreateJobResult.JobId

	resp, err = client.DescribeJob(&s3.DescribeJobInput{
		JobId: aws.String(jobId4),
	})
	c.Assert(err, IsNil)
	c.Assert(*resp.DescribeJobResult.JobId, Equals, jobId4)
	c.Assert(*resp.DescribeJobResult.Operation.KS3DeleteObject, NotNil)

	// 新建同城冗余转换操作
	createJobInput.CreateJobRequest.ClientRequestToken = nil
	createJobInput.CreateJobRequest.Operation = &s3.JobOperation{
		KS3PutObjectDataRedundancyTransition: &s3.KS3PutObjectDataRedundancyTransition{
			DataRedundancyType: aws.String(s3.DataRedundancyTypeZRS),
		},
	}
	createJobInput.CreateJobRequest.Manifest.Location.Filters[0].Prefixes = []string{"prefix5/"}
	createResp, err = client.CreateJob(createJobInput)
	c.Assert(err, IsNil)
	jobId5 := *createResp.CreateJobResult.JobId

	resp, err = client.DescribeJob(&s3.DescribeJobInput{
		JobId: aws.String(jobId5),
	})
	c.Assert(err, IsNil)
	c.Assert(*resp.DescribeJobResult.JobId, Equals, jobId5)
	c.Assert(*resp.DescribeJobResult.Operation.KS3PutObjectDataRedundancyTransition.DataRedundancyType, Equals, s3.DataRedundancyTypeZRS)

	// 新建打包压缩操作
	createJobInput.CreateJobRequest.ClientRequestToken = nil
	createJobInput.CreateJobRequest.Operation = &s3.JobOperation{
		KS3CompressObject: &s3.KS3CompressObject{
			Format:      aws.String("zip"),
			IgnoreError: aws.Boolean(true),
			Output: &s3.CompressOutput{
				Bucket:     aws.String("krn:ksc:ks3:::test-bucket"),
				Prefix:     aws.String("output/"),
				ObjectName: aws.String("archive.zip"),
			},
		},
	}
	createJobInput.CreateJobRequest.Manifest.Location.Filters[0].Prefixes = []string{"prefix6/"}
	createResp, err = client.CreateJob(createJobInput)
	c.Assert(err, IsNil)
	jobId6 := *createResp.CreateJobResult.JobId

	resp, err = client.DescribeJob(&s3.DescribeJobInput{
		JobId: aws.String(jobId6),
	})
	c.Assert(err, IsNil)
	c.Assert(*resp.DescribeJobResult.JobId, Equals, jobId6)
	c.Assert(*resp.DescribeJobResult.Operation.KS3CompressObject.Format, Equals, "zip")
	c.Assert(*resp.DescribeJobResult.Operation.KS3CompressObject.IgnoreError, Equals, true)
	c.Assert(*resp.DescribeJobResult.Operation.KS3CompressObject.Output.Bucket, Equals, "krn:ksc:ks3:::test-bucket")
	c.Assert(*resp.DescribeJobResult.Operation.KS3CompressObject.Output.Prefix, Equals, "output/")
	c.Assert(*resp.DescribeJobResult.Operation.KS3CompressObject.Output.ObjectName, Equals, "archive.zip")

	listResp, err := client.ListJobs(&s3.ListJobsInput{
		MaxResults: aws.Long(100),
	})
	c.Assert(err, IsNil)
	c.Assert(len(listResp.ListJobsResult.Jobs.Members) >= 6, Equals, true)

	_, err = client.DeleteJob(&s3.DeleteJobInput{
		JobId: aws.String(jobId1),
	})
	c.Assert(err, IsNil)

	_, err = client.DeleteJob(&s3.DeleteJobInput{
		JobId: aws.String(jobId2),
	})
	c.Assert(err, IsNil)

	_, err = client.DeleteJob(&s3.DeleteJobInput{
		JobId: aws.String(jobId3),
	})
	c.Assert(err, IsNil)

	_, err = client.DeleteJob(&s3.DeleteJobInput{
		JobId: aws.String(jobId4),
	})
	c.Assert(err, IsNil)

	_, err = client.DeleteJob(&s3.DeleteJobInput{
		JobId: aws.String(jobId5),
	})
	c.Assert(err, IsNil)

	_, err = client.DeleteJob(&s3.DeleteJobInput{
		JobId: aws.String(jobId6),
	})
	c.Assert(err, IsNil)
}

func (s *Ks3utilCommandSuite) TestUpdateJob(c *C) {
	c.Skip("Skip TestUpdateJob")
	// 新建设置对象ACL操作，设置对象ACL
	createJobInput := &s3.CreateJobInput{
		CreateJobRequest: &s3.CreateJobRequest{
			Description: aws.String("this-is-a-test-job"),
			Priority:    aws.Long(100),
			Operation: &s3.JobOperation{
				KS3RestoreObject: &s3.KS3RestoreObject{
					StorageClass: aws.String(s3.StorageClassArchive),
					Days:         aws.Long(7),
				},
			},
			Manifest: &s3.JobManifest{
				Location: &s3.ManifestLocation{
					Filters: []*s3.LocationFilter{
						{
							Bucket:   aws.String("krn:ksc:ks3:::test-bucket"),
							Prefixes: []string{"prefix6/"},
						},
					},
				},
				Spec: &s3.ManifestSpec{
					Format: aws.String("KS3BatchOperations_Bucket_V1"),
				},
			},
			Report: &s3.JobReport{
				Bucket:      aws.String("krn:ksc:ks3:::test-bucket"),
				Prefix:      aws.String("result/"),
				Enabled:     aws.Boolean(true),
				ReportScope: aws.String("FailedTasksOnly"),
			},
		},
	}
	// 创建任务
	createResp, err := client.CreateJob(createJobInput)
	c.Assert(err, IsNil)
	jobId1 := *createResp.CreateJobResult.JobId

	//查看任务详情
	resp, err := client.DescribeJob(&s3.DescribeJobInput{
		JobId: aws.String(jobId1),
	})
	c.Assert(err, IsNil)
	c.Assert(*resp.DescribeJobResult.JobId, Equals, jobId1)
	c.Assert(*resp.DescribeJobResult.Priority, Equals, int64(100))

	// 更新任务优先级
	updateResp, err := client.UpdateJobPriority(&s3.UpdateJobPriorityInput{
		JobId:    aws.String(jobId1),
		Priority: aws.Long(50),
	})
	c.Assert(err, IsNil)
	c.Assert(*updateResp.UpdateJobPriorityResult.JobId, Equals, jobId1)
	c.Assert(*updateResp.UpdateJobPriorityResult.Priority, Equals, int64(50))

	// 删除任务
	_, err = client.DeleteJob(&s3.DeleteJobInput{
		JobId: aws.String(jobId1),
	})
	c.Assert(err, IsNil)
}

func (s *Ks3utilCommandSuite) TestGenerateShareUrl(c *C) {
	var signerVersions = []string{"V2", "V4", "V4_UNSIGNED_PAYLOAD_SIGNER"}
	var cre = credentials.NewStaticCredentials(accessKeyID, accessKeySecret, "")
	for _, sigVer := range signerVersions {
		client := s3.New(&aws.Config{
			Credentials:   cre,
			Region:        region,
			Endpoint:      endpoint,
			SignerVersion: sigVer,
		})

		objectName := "test/" + sigVer
		s.PutObject(objectName, c)

		urlStr, err := client.GenerateShareUrl(&s3.GenerateShareUrlInput{})
		c.Assert(err, NotNil)
		c.Assert(err.Error(), Equals, "bucket is required")

		urlStr, err = client.GenerateShareUrl(&s3.GenerateShareUrlInput{
			Bucket: aws.String(bucket),
			Prefix: aws.String("test/"),
		})
		c.Assert(err, IsNil)

		resp, err := sendRequestByShareUrl("GET", urlStr)
		c.Assert(err, IsNil)
		body, err := io.ReadAll(resp.Body)
		c.Assert(err, IsNil)
		out := &s3.ListObjectsOutput{}
		err = xml.Unmarshal(body, out)
		resp.Body.Close()
		c.Assert(err, IsNil)
		c.Assert(len(out.Contents), Equals, 1)

		url, err := url.Parse(urlStr)
		c.Assert(err, IsNil)

		url.Path = objectName
		resp, err = sendRequestByShareUrl("GET", url.String())
		c.Assert(err, IsNil)
		text, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		c.Assert(err, IsNil)
		c.Assert(string(text), Equals, content)

		resp, err = sendRequestByShareUrl("HEAD", url.String())
		resp.Body.Close()
		c.Assert(err, IsNil)
		c.Assert(resp.Header.Get(s3.HTTPHeaderContentLength), Equals, "3")

		token, err := s3.EncryptUrlToToken(urlStr, "")
		c.Assert(err, NotNil)
		c.Assert(err.Error(), Equals, "accessCode is required")

		token, err = s3.EncryptUrlToToken(urlStr, "123")
		c.Assert(err, NotNil)
		c.Assert(err.Error(), Equals, "accessCode must be 6 characters long and contain letters and numbers only")

		token, err = s3.EncryptUrlToToken(urlStr, "123456")
		c.Assert(err, IsNil)

		urlStr2, err := s3.DecryptTokenToUrl(token, "111111")
		c.Assert(err, NotNil)

		urlStr2, err = s3.DecryptTokenToUrl(token, "123456")
		c.Assert(err, IsNil)
		c.Assert(urlStr, Equals, urlStr2)

		s.DeleteObject(objectName, c)
	}
}

func (s *Ks3utilCommandSuite) TestGenerateShareUrlWithAccessCode(c *C) {
	var signerVersions = []string{"V2", "V4", "V4_UNSIGNED_PAYLOAD_SIGNER"}
	var cre = credentials.NewStaticCredentials(accessKeyID, accessKeySecret, "")
	for _, sigVer := range signerVersions {
		client := s3.New(&aws.Config{
			Credentials:   cre,
			Region:        region,
			Endpoint:      endpoint,
			SignerVersion: sigVer,
		})

		objectName := "test/" + sigVer
		s.PutObject(objectName, c)

		htmlShareUrl, err := client.GenerateShareUrl(&s3.GenerateShareUrlInput{
			Bucket:     aws.String(bucket),
			Prefix:     aws.String("test/"),
			AccessCode: aws.String("123456"),
		})
		c.Assert(err, IsNil)

		token := htmlShareUrl[strings.Index(htmlShareUrl, "token=")+6:]
		urlStr, err := s3.DecryptTokenToUrl(token, "123456")
		c.Assert(err, IsNil)

		resp, err := sendRequestByShareUrl("GET", urlStr)
		c.Assert(err, IsNil)
		body, err := io.ReadAll(resp.Body)
		c.Assert(err, IsNil)
		out := &s3.ListObjectsOutput{}
		err = xml.Unmarshal(body, out)
		resp.Body.Close()
		c.Assert(err, IsNil)
		c.Assert(len(out.Contents), Equals, 1)

		url, err := url.Parse(urlStr)
		c.Assert(err, IsNil)

		url.Path = objectName
		resp, err = sendRequestByShareUrl("GET", url.String())
		c.Assert(err, IsNil)
		text, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		c.Assert(err, IsNil)
		c.Assert(string(text), Equals, content)

		resp, err = sendRequestByShareUrl("HEAD", url.String())
		resp.Body.Close()
		c.Assert(err, IsNil)
		c.Assert(resp.Header.Get(s3.HTTPHeaderContentLength), Equals, "3")

		s.DeleteObject(objectName, c)
	}
}

func (s *Ks3utilCommandSuite) TestGenerateShareUrlByPolicy(c *C) {
	var signerVersions = []string{"V2", "V4", "V4_UNSIGNED_PAYLOAD_SIGNER"}
	var cre = credentials.NewStaticCredentials(accessKeyID, accessKeySecret, "")
	for _, sigVer := range signerVersions {
		client := s3.New(&aws.Config{
			Credentials:   cre,
			Region:        region,
			Endpoint:      endpoint,
			SignerVersion: sigVer,
		})

		objectName := "test/" + sigVer
		s.PutObject(objectName, c)

		policy, err := s3.BuildPolicy("", []string{}, []string{})
		c.Assert(err, NotNil)
		c.Assert(err.Error(), Equals, "bucketName is required")

		policy, err = s3.BuildPolicy(bucket, []string{}, []string{})
		c.Assert(err, NotNil)
		c.Assert(err.Error(), Equals, "prefixes or keys must be provided")

		policy, err = s3.BuildPolicy(bucket, []string{"test/"}, []string{})
		c.Assert(err, IsNil)

		urlStr, err := client.GenerateShareUrl(&s3.GenerateShareUrlInput{
			Bucket: aws.String(bucket),
			Policy: aws.String(policy),
		})
		c.Assert(err, IsNil)
		urlStr += "&prefix=test%2F"

		resp, err := sendRequestByShareUrl("GET", urlStr)
		c.Assert(err, IsNil)
		c.Assert(resp.StatusCode, Equals, 200)
		body, err := io.ReadAll(resp.Body)
		c.Assert(err, IsNil)
		out := &s3.ListObjectsOutput{}
		err = xml.Unmarshal(body, out)
		resp.Body.Close()
		c.Assert(err, IsNil)
		c.Assert(len(out.Contents), Equals, 1)

		url, err := url.Parse(urlStr)
		c.Assert(err, IsNil)

		url.Path = objectName
		resp, err = sendRequestByShareUrl("GET", url.String())
		c.Assert(err, IsNil)
		text, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		c.Assert(err, IsNil)
		c.Assert(string(text), Equals, content)

		resp, err = sendRequestByShareUrl("HEAD", url.String())
		resp.Body.Close()
		c.Assert(err, IsNil)
		c.Assert(resp.Header.Get(s3.HTTPHeaderContentLength), Equals, "3")

		policy, err = s3.BuildPolicy(bucket, []string{}, []string{objectName})
		c.Assert(err, IsNil)

		urlStr, err = client.GenerateShareUrl(&s3.GenerateShareUrlInput{
			Bucket: aws.String(bucket),
			Policy: aws.String(policy),
		})
		c.Assert(err, IsNil)
		urlStr += "&prefix=test%2F"

		resp, err = sendRequestByShareUrl("GET", urlStr)
		c.Assert(err, IsNil)
		c.Assert(resp.StatusCode, Equals, 400)

		url, err = url.Parse(urlStr)
		c.Assert(err, IsNil)

		url.Path = objectName
		resp, err = sendRequestByShareUrl("GET", url.String())
		c.Assert(err, IsNil)
		text, err = io.ReadAll(resp.Body)
		resp.Body.Close()
		c.Assert(err, IsNil)
		c.Assert(string(text), Equals, content)

		resp, err = sendRequestByShareUrl("HEAD", url.String())
		resp.Body.Close()
		c.Assert(err, IsNil)
		c.Assert(resp.Header.Get(s3.HTTPHeaderContentLength), Equals, "3")

		s.DeleteObject(objectName, c)
	}
}

func (s *Ks3utilCommandSuite) TestPutObjectDataRedundancyTransition(c *C) {
	object := randLowStr(10)
	// 上传对象
	s.PutObject(object, c)

	// 获取对象冗余类型，默认是LRS
	headResp, err := client.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(err, IsNil)
	c.Assert(*headResp.Metadata[s3.HTTPHeaderAmzDataRedundancyType], Equals, s3.DataRedundancyTypeLRS)

	// 转换为ZRS
	_, err = client.PutObjectDataRedundancyTransition(&s3.PutObjectDataRedundancyTransitionInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(err, IsNil)

	// 获取对象冗余类型，验证是否转换成功
	headResp, err = client.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(err, IsNil)
	c.Assert(*headResp.Metadata[s3.HTTPHeaderAmzDataRedundancyType], Equals, s3.DataRedundancyTypeZRS)

	// 删除对象
	s.DeleteObject(object, c)
}

func (s *Ks3utilCommandSuite) TestPutObjectWithIfMatch(c *C) {
	object := randLowStr(10)
	// 上传对象
	s.PutObject(object, c)

	// 获取对象的ETag
	headResp, err := client.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(err, IsNil)
	eTag := *headResp.ETag

	// 使用正确的If-Match头部上传对象
	putResp, err := client.PutObject(&s3.PutObjectInput{
		Bucket:  aws.String(bucket),
		Key:     aws.String(object),
		Body:    strings.NewReader("content1"),
		IfMatch: aws.String(eTag),
	})
	c.Assert(err, IsNil)
	c.Assert(*putResp.StatusCode, Equals, int64(200))
	eTag = *putResp.ETag

	// 使用多个ETag值的If-Match头部上传对象，其中一个匹配
	putResp, err = client.PutObject(&s3.PutObjectInput{
		Bucket:  aws.String(bucket),
		Key:     aws.String(object),
		Body:    strings.NewReader("content2"),
		IfMatch: aws.String(eTag + ",non-matching-etag"),
	})
	c.Assert(err, IsNil)
	c.Assert(*putResp.StatusCode, Equals, int64(200))
	eTag = *putResp.ETag

	// 使用错误的If-Match头部上传对象
	putResp, err = client.PutObject(&s3.PutObjectInput{
		Bucket:  aws.String(bucket),
		Key:     aws.String(object),
		Body:    strings.NewReader("content3"),
		IfMatch: aws.String("non-matching-etag"),
	})
	c.Assert(err, NotNil)
	c.Assert(*putResp.StatusCode, Equals, int64(412))

	// 删除对象
	s.DeleteObject(object, c)
}

func (s *Ks3utilCommandSuite) TestPutObjectWithIfNoneMatch(c *C) {
	object := randLowStr(10)
	// 上传对象
	s.PutObject(object, c)

	// 获取对象的ETag
	headResp, err := client.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(err, IsNil)
	eTag := *headResp.ETag

	// 使用不匹配的If-None-Match头部上传对象
	putResp, err := client.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(object),
		Body:        strings.NewReader("content1"),
		IfNoneMatch: aws.String("non-matching-etag"),
	})
	c.Assert(err, IsNil)
	c.Assert(*putResp.StatusCode, Equals, int64(200))
	eTag = *putResp.ETag

	// 使用多个ETag值的If-None-Match头部上传对象，全都不匹配
	putResp, err = client.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(object),
		Body:        strings.NewReader("content2"),
		IfNoneMatch: aws.String("non-matching-etag,non-matching-etag2"),
	})
	c.Assert(err, IsNil)
	c.Assert(*putResp.StatusCode, Equals, int64(200))
	eTag = *putResp.ETag

	// 使用匹配的If-None-Match头部上传对象
	putResp, err = client.PutObject(&s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(object),
		Body:        strings.NewReader("content3"),
		IfNoneMatch: aws.String(eTag),
	})
	c.Assert(err, NotNil)
	c.Assert(*putResp.StatusCode, Equals, int64(412))

	// 删除对象
	s.DeleteObject(object, c)
}

func (s *Ks3utilCommandSuite) TestMultipartUploadWithIfMatch(c *C) {
	object := randLowStr(10)
	// 上传对象
	s.PutObject(object, c)

	// 获取对象的ETag
	headResp, err := client.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(err, IsNil)
	eTag := *headResp.ETag

	// 使用正确的If-Match头部上传对象
	putResp, err := multipartUpload(bucket, object, "content1", &s3.CompleteMultipartUploadInput{
		IfMatch: aws.String(eTag),
	})
	c.Assert(err, IsNil)
	c.Assert(*putResp.StatusCode, Equals, int64(200))
	eTag = *putResp.ETag

	// 使用多个ETag值的If-Match头部上传对象，其中一个匹配
	putResp, err = multipartUpload(bucket, object, "content2", &s3.CompleteMultipartUploadInput{
		IfMatch: aws.String(eTag + ",non-matching-etag"),
	})
	c.Assert(err, IsNil)
	c.Assert(*putResp.StatusCode, Equals, int64(200))
	eTag = *putResp.ETag

	// 使用错误的If-Match头部上传对象
	putResp, err = multipartUpload(bucket, object, "content3", &s3.CompleteMultipartUploadInput{
		IfMatch: aws.String("non-matching-etag"),
	})
	c.Assert(err, NotNil)
	c.Assert(*putResp.StatusCode, Equals, int64(412))

	// 删除对象
	s.DeleteObject(object, c)
}

func (s *Ks3utilCommandSuite) TestMultipartUploadWithIfNoneMatch(c *C) {
	object := randLowStr(10)
	// 上传对象
	s.PutObject(object, c)

	// 获取对象的ETag
	headResp, err := client.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(err, IsNil)
	eTag := *headResp.ETag

	// 使用不匹配的If-None-Match头部上传对象
	putResp, err := multipartUpload(bucket, object, "content1", &s3.CompleteMultipartUploadInput{
		IfNoneMatch: aws.String("non-matching-etag"),
	})
	c.Assert(err, IsNil)
	c.Assert(*putResp.StatusCode, Equals, int64(200))
	eTag = *putResp.ETag

	// 使用多个ETag值的If-None-Match头部上传对象，全都不匹配
	putResp, err = multipartUpload(bucket, object, "content2", &s3.CompleteMultipartUploadInput{
		IfNoneMatch: aws.String("non-matching-etag,non-matching-etag2"),
	})
	c.Assert(err, IsNil)
	c.Assert(*putResp.StatusCode, Equals, int64(200))
	eTag = *putResp.ETag

	// 使用匹配的If-None-Match头部上传对象
	putResp, err = multipartUpload(bucket, object, "content3", &s3.CompleteMultipartUploadInput{
		IfNoneMatch: aws.String(eTag),
	})
	c.Assert(err, NotNil)
	c.Assert(*putResp.StatusCode, Equals, int64(412))

	// 删除对象
	s.DeleteObject(object, c)
}

func (s *Ks3utilCommandSuite) TestUploadFileWithIfMatch(c *C) {
	object := randLowStr(10)
	// 上传对象
	s.PutObject(object, c)

	// 获取对象的ETag
	headResp, err := client.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(err, IsNil)
	eTag := *headResp.ETag

	// 使用正确的If-Match头部上传对象
	createFileWithContent(object, "content1")
	uploadResp, err := client.UploadFile(&s3.UploadFileInput{
		Bucket:     aws.String(bucket),
		Key:        aws.String(object),
		UploadFile: aws.String(object),
		IfMatch:    aws.String(eTag),
	})
	c.Assert(err, IsNil)
	eTag = *uploadResp.ETag

	// 使用多个ETag值的If-Match头部上传对象，其中一个匹配
	createFileWithContent(object, "content2")
	uploadResp, err = client.UploadFile(&s3.UploadFileInput{
		Bucket:     aws.String(bucket),
		Key:        aws.String(object),
		UploadFile: aws.String(object),
		IfMatch:    aws.String(eTag + ",non-matching-etag"),
	})
	c.Assert(err, IsNil)
	eTag = *uploadResp.ETag

	// 使用错误的If-Match头部上传对象
	createFileWithContent(object, "content3")
	uploadResp, err = client.UploadFile(&s3.UploadFileInput{
		Bucket:     aws.String(bucket),
		Key:        aws.String(object),
		UploadFile: aws.String(object),
		IfMatch:    aws.String("non-matching-etag"),
	})
	c.Assert(err, NotNil)
	c.Assert(strings.Contains(err.Error(), "412"), Equals, true)

	// 删除本地文件
	os.Remove(object)

	// 删除对象
	s.DeleteObject(object, c)
}

func (s *Ks3utilCommandSuite) TestUploadFileIfNoneMatch(c *C) {
	object := randLowStr(10)
	// 上传对象
	s.PutObject(object, c)

	// 获取对象的ETag
	headResp, err := client.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(err, IsNil)
	eTag := *headResp.ETag

	// 使用不匹配的If-None-Match头部上传对象
	createFileWithContent(object, "content1")
	uploadResp, err := client.UploadFile(&s3.UploadFileInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(object),
		UploadFile:  aws.String(object),
		IfNoneMatch: aws.String("non-matching-etag"),
	})
	c.Assert(err, IsNil)
	eTag = *uploadResp.ETag

	// 使用多个ETag值的If-None-Match头部上传对象，全都不匹配
	createFileWithContent(object, "content2")
	uploadResp, err = client.UploadFile(&s3.UploadFileInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(object),
		UploadFile:  aws.String(object),
		IfNoneMatch: aws.String("non-matching-etag,non-matching-etag2"),
	})
	c.Assert(err, IsNil)
	eTag = *uploadResp.ETag

	// 使用匹配的If-None-Match头部上传对象
	createFileWithContent(object, "content3")
	uploadResp, err = client.UploadFile(&s3.UploadFileInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(object),
		UploadFile:  aws.String(object),
		IfNoneMatch: aws.String(eTag),
	})
	c.Assert(err, NotNil)
	c.Assert(strings.Contains(err.Error(), "412"), Equals, true)

	// 删除本地文件
	os.Remove(object)

	// 删除对象
	s.DeleteObject(object, c)
}

func (s *Ks3utilCommandSuite) TestDeleteObjectWithIfMatch(c *C) {
	object := randLowStr(10)
	// 上传对象
	s.PutObject(object, c)

	// 获取对象的ETag
	headResp, err := client.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(err, IsNil)
	eTag := *headResp.ETag

	// 使用正确的If-Match头部删除对象
	_, err = client.DeleteObject(&s3.DeleteObjectInput{
		Bucket:  aws.String(bucket),
		Key:     aws.String(object),
		IfMatch: aws.String(eTag),
	})
	c.Assert(err, IsNil)

	// 上传对象
	s.PutObject(object, c)

	// 获取对象的ETag
	headResp, err = client.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(err, IsNil)
	eTag = *headResp.ETag

	// 使用多个ETag值的If-Match头部删除对象，其中一个匹配
	_, err = client.DeleteObject(&s3.DeleteObjectInput{
		Bucket:  aws.String(bucket),
		Key:     aws.String(object),
		IfMatch: aws.String(eTag + ",non-matching-etag"),
	})
	c.Assert(err, IsNil)

	// 上传对象
	s.PutObject(object, c)

	// 获取对象的ETag
	headResp, err = client.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(err, IsNil)
	eTag = *headResp.ETag

	// 使用错误的If-Match头部删除对象
	_, err = client.DeleteObject(&s3.DeleteObjectInput{
		Bucket:  aws.String(bucket),
		Key:     aws.String(object),
		IfMatch: aws.String("non-matching-etag"),
	})
	c.Assert(err, NotNil)
	c.Assert(strings.Contains(err.Error(), "412"), Equals, true)

	// 删除对象
	s.DeleteObject(object, c)
}

// TestListObjectsPaginator 测试ListObjects分页器
func (s *Ks3utilCommandSuite) TestListObjectsPaginator(c *C) {
	testPrefix := "test_paginator_v1_" + randLowStr(8) + "/"

	testKeys := make([]string, 10)
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("%sfile_%02d.txt", testPrefix, i)
		testKeys[i] = key
		_, err := client.PutObject(&s3.PutObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
			Body:   strings.NewReader(fmt.Sprintf("content %d", i)),
		})
		c.Assert(err, IsNil)
	}

	defer func() {
		for _, key := range testKeys {
			s.DeleteObject(key, c)
		}
	}()

	paginator := client.NewListObjectsPaginator(&s3.ListObjectsInput{
		Bucket:  aws.String(bucket),
		Prefix:  aws.String(testPrefix),
		MaxKeys: aws.Long(3),
	})

	var totalObjects int
	pageCount := 0
	for paginator.HasNext() {
		resp, err := paginator.NextPage()
		c.Assert(err, IsNil)
		totalObjects += len(resp.Contents)
		pageCount++
	}

	c.Assert(totalObjects, Equals, 10)
	c.Assert(pageCount > 1, Equals, true)

	_, err := paginator.NextPage()
	c.Assert(err, NotNil)
}

// TestListMultipartUploadsPaginator 测试ListMultipartUploads分页器
func (s *Ks3utilCommandSuite) TestListMultipartUploadsPaginator(c *C) {
	testPrefix := "test_paginator_mpu_" + randLowStr(8) + "/"

	// 创建3个分块上传任务
	uploadIDs := make([]string, 3)
	for i := 0; i < 3; i++ {
		key := fmt.Sprintf("%sfile_%02d.txt", testPrefix, i)
		resp, err := client.CreateMultipartUpload(&s3.CreateMultipartUploadInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		})
		c.Assert(err, IsNil)
		uploadIDs[i] = *resp.UploadID
	}

	defer func() {
		for i, uploadID := range uploadIDs {
			_, _ = client.AbortMultipartUpload(&s3.AbortMultipartUploadInput{
				Bucket:   aws.String(bucket),
				Key:      aws.String(fmt.Sprintf("%sfile_%02d.txt", testPrefix, i)),
				UploadID: aws.String(uploadID),
			})
		}
	}()

	paginator := client.NewListMultipartUploadsPaginator(&s3.ListMultipartUploadsInput{
		Bucket:     aws.String(bucket),
		Prefix:     aws.String(testPrefix),
		MaxUploads: aws.Long(2),
	})

	var totalUploads int
	for paginator.HasNext() {
		resp, err := paginator.NextPage()
		c.Assert(err, IsNil)
		totalUploads += len(resp.Uploads)
	}

	c.Assert(totalUploads, Equals, 3)

	_, err := paginator.NextPage()
	c.Assert(err, NotNil)
}

// TestListPartsPaginator 测试ListParts分页器
func (s *Ks3utilCommandSuite) TestListPartsPaginator(c *C) {
	object := randLowStr(10)

	// 创建分块上传
	createResp, err := client.CreateMultipartUpload(&s3.CreateMultipartUploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(err, IsNil)
	uploadID := *createResp.UploadID

	defer func() {
		_, _ = client.AbortMultipartUpload(&s3.AbortMultipartUploadInput{
			Bucket:   aws.String(bucket),
			Key:      aws.String(object),
			UploadID: aws.String(uploadID),
		})
	}()

	// 上传5个分块
	for i := 1; i <= 5; i++ {
		_, err := client.UploadPart(&s3.UploadPartInput{
			Bucket:     aws.String(bucket),
			Key:        aws.String(object),
			PartNumber: aws.Long(int64(i)),
			UploadID:   aws.String(uploadID),
			Body:       strings.NewReader(fmt.Sprintf("part %d", i)),
		})
		c.Assert(err, IsNil)
	}

	paginator := client.NewListPartsPaginator(&s3.ListPartsInput{
		Bucket:   aws.String(bucket),
		Key:      aws.String(object),
		UploadID: aws.String(uploadID),
		MaxParts: aws.Long(2),
	})

	var totalParts int
	for paginator.HasNext() {
		resp, err := paginator.NextPage()
		c.Assert(err, IsNil)
		totalParts += len(resp.Parts)
	}

	c.Assert(totalParts, Equals, 5)

	_, err = paginator.NextPage()
	c.Assert(err, NotNil)
}

// TestListObjectsV2Paginator 测试ListObjectsV2分页器
func (s *Ks3utilCommandSuite) TestListObjectsV2Paginator(c *C) {
	testPrefix := "test_paginator_v2_" + randLowStr(8) + "/"

	testKeys := make([]string, 10)
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("%sfile_%02d.txt", testPrefix, i)
		testKeys[i] = key
		_, err := client.PutObject(&s3.PutObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
			Body:   strings.NewReader(fmt.Sprintf("content %d", i)),
		})
		c.Assert(err, IsNil)
	}

	defer func() {
		for _, key := range testKeys {
			s.DeleteObject(key, c)
		}
	}()

	paginator := client.NewListObjectsV2Paginator(&s3.ListObjectsV2Input{
		Bucket:  aws.String(bucket),
		Prefix:  aws.String(testPrefix),
		MaxKeys: aws.Long(3),
	})

	var totalObjects int
	pageCount := 0
	for paginator.HasNext() {
		resp, err := paginator.NextPage()
		c.Assert(err, IsNil)
		totalObjects += len(resp.Contents)
		pageCount++
	}

	c.Assert(totalObjects, Equals, 10)
	c.Assert(pageCount > 1, Equals, true)

	_, err := paginator.NextPage()
	c.Assert(err, NotNil)
}

// TestUploadDir 上传目录完整测试
func (s *Ks3utilCommandSuite) TestUploadDir(c *C) {
	prefix := randLowStr(6) + "/"
	dir := randLowStr(8) + "/"
	os.RemoveAll(dir)
	os.MkdirAll(dir+"subdir/", 0755)
	createFileWithContent(dir+"a.txt", "hello")
	createFileWithContent(dir+"subdir/b.txt", "world")
	createFile(dir+"c.bin", 1024*5)

	// 1. 基本上传：3个文件（含子目录、大文件分块）
	output, err := client.UploadDir(&s3.UploadDirInput{
		DirPath: aws.String(dir),
		Bucket:  aws.String(bucket),
		Prefix:  aws.String(prefix),
	})
	c.Assert(err, IsNil)
	c.Assert(output.TotalNum, Equals, int64(3))
	c.Assert(output.SuccessNum, Equals, int64(3))
	c.Assert(output.FailNum, Equals, int64(0))
	c.Assert(output.SkipNum, Equals, int64(0))
	c.Assert(len(output.SuccessFiles), Equals, 3)
	c.Assert(len(output.ErrorFiles), Equals, 0)
	// 验证文件内容
	getResp, _ := client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket), Key: aws.String(prefix + "a.txt"),
	})
	body, _ := io.ReadAll(getResp.Body)
	getResp.Body.Close()
	c.Assert(string(body), Equals, "hello")

	// 2. SkipRule=IfExists：对象已存在，全部跳过
	output2, err := client.UploadDir(&s3.UploadDirInput{
		DirPath:  aws.String(dir),
		Bucket:   aws.String(bucket),
		Prefix:   aws.String(prefix),
		SkipRule: aws.String(s3.SkipIfExists),
	})
	c.Assert(err, IsNil)
	c.Assert(output2.SkipNum, Equals, int64(3))
	c.Assert(output2.SuccessNum, Equals, int64(0))
	c.Assert(output.TotalNum, Equals, int64(3))

	// 3. SkipRule=IfSizeEquals：a.txt改大 → 不跳过；b.txt/c.bin大小不变 → 跳过
	createFileWithContent(dir+"a.txt", "hello world! long content now")
	output3, err := client.UploadDir(&s3.UploadDirInput{
		DirPath:  aws.String(dir),
		Bucket:   aws.String(bucket),
		Prefix:   aws.String(prefix),
		SkipRule: aws.String(s3.SkipIfSizeEquals),
	})
	c.Assert(err, IsNil)
	c.Assert(output3.SkipNum, Equals, int64(2))
	c.Assert(output3.SuccessNum, Equals, int64(1))
	// 验证a.txt内容已被覆盖
	getResp2, _ := client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket), Key: aws.String(prefix + "a.txt"),
	})
	body2, _ := io.ReadAll(getResp2.Body)
	getResp2.Body.Close()
	c.Assert(string(body2), Equals, "hello world! long content now")

	// 4. SkipRule=IfNewer：本地文件ModTime设为过去 → 远程更新 → 跳过
	pastTime := time.Now().Add(-48 * time.Hour)
	os.Chtimes(dir+"a.txt", pastTime, pastTime)
	os.Chtimes(dir+"subdir/b.txt", pastTime, pastTime)
	os.Chtimes(dir+"c.bin", pastTime, pastTime)
	output4, err := client.UploadDir(&s3.UploadDirInput{
		DirPath:  aws.String(dir),
		Bucket:   aws.String(bucket),
		Prefix:   aws.String(prefix),
		SkipRule: aws.String(s3.SkipIfNewer),
	})
	c.Assert(err, IsNil)
	c.Assert(output4.SkipNum, Equals, int64(3)) // 全部跳过，因为本地比远程旧
	c.Assert(output4.SuccessNum, Equals, int64(0))

	// 本地设为未来 → 本地比远程新 → 不跳过
	futureTime := time.Now().Add(24 * time.Hour)
	os.Chtimes(dir+"a.txt", futureTime, futureTime)
	output5, err := client.UploadDir(&s3.UploadDirInput{
		DirPath:  aws.String(dir),
		Bucket:   aws.String(bucket),
		Prefix:   aws.String(prefix),
		SkipRule: aws.String(s3.SkipIfNewer),
	})
	c.Assert(err, IsNil)
	c.Assert(output5.SkipNum, Equals, int64(2)) // b.txt/c.bin远程更新，跳过
	c.Assert(output5.SuccessNum, Equals, int64(1)) // a.txt本地更新，重新上传

	// 5. SkipRule=IfNewerAndSizeEquals：b.txt恢复原内容+设为未来时间 → 本地更新且大小相等 → 跳过
	createFileWithContent(dir+"subdir/b.txt", "world")
	os.Chtimes(dir+"subdir/b.txt", futureTime, futureTime)
	output6, err := client.UploadDir(&s3.UploadDirInput{
		DirPath:  aws.String(dir),
		Bucket:   aws.String(bucket),
		Prefix:   aws.String(prefix),
		SkipRule: aws.String(s3.SkipIfNewerAndSizeEquals),
	})
	c.Assert(err, IsNil)
	c.Assert(output6.SkipNum, Equals, int64(1))    // b.txt本地更新且大小相等，跳过
	c.Assert(output6.SuccessNum, Equals, int64(2)) // a.txt大小不等，c.bin本地不更新

	// 6. SkipRule=IfCrc64Equals：改a.txt内容 → CRC不等 → 不跳过
	createFileWithContent(dir+"a.txt", "different content for crc")
	output7, err := client.UploadDir(&s3.UploadDirInput{
		DirPath:  aws.String(dir),
		Bucket:   aws.String(bucket),
		Prefix:   aws.String(prefix + "crc/"),
		SkipRule: aws.String(s3.SkipIfCrc64Equals),
	})
	c.Assert(err, IsNil)
	c.Assert(output7.SuccessNum, Equals, int64(3))
	c.Assert(output7.SkipNum, Equals, int64(0))
	// 再上传同内容 → CRC匹配 → 全部跳过
	output8, err := client.UploadDir(&s3.UploadDirInput{
		DirPath:  aws.String(dir),
		Bucket:   aws.String(bucket),
		Prefix:   aws.String(prefix + "crc/"),
		SkipRule: aws.String(s3.SkipIfCrc64Equals),
	})
	c.Assert(err, IsNil)
	c.Assert(output8.SkipNum, Equals, int64(3))
	c.Assert(output8.SuccessNum, Equals, int64(0))

	// 7. SkipRule=Never（默认值）：不跳过任何文件
	createFileWithContent(dir+"a.txt", "never skip test")
	output8b, err := client.UploadDir(&s3.UploadDirInput{
		DirPath:  aws.String(dir),
		Bucket:   aws.String(bucket),
		Prefix:   aws.String(prefix + "never/"),
		SkipRule: aws.String(s3.SkipNever),
	})
	c.Assert(err, IsNil)
	c.Assert(output8b.SkipNum, Equals, int64(0))
	c.Assert(output8b.SuccessNum, Equals, int64(3))

	// 8. IgnoreSuccessFiles=true
	output9, err := client.UploadDir(&s3.UploadDirInput{
		DirPath:            aws.String(dir),
		Bucket:             aws.String(bucket),
		Prefix:             aws.String(prefix + "ignore/"),
		IgnoreSuccessFiles: aws.Boolean(true),
	})
	c.Assert(err, IsNil)
	c.Assert(output9.SuccessNum, Equals, int64(3))
	c.Assert(len(output9.SuccessFiles), Equals, 0)

	// 9. SkipSymlinks=true跳过，false跟随
	os.Remove(dir + "link.txt")
	os.Symlink("a.txt", dir+"link.txt")
	output10, err := client.UploadDir(&s3.UploadDirInput{
		DirPath:      aws.String(dir),
		Bucket:       aws.String(bucket),
		Prefix:       aws.String(prefix + "symlink-skip/"),
		SkipSymlinks: aws.Boolean(true),
	})
	c.Assert(err, IsNil)
	c.Assert(output10.SuccessNum, Equals, int64(3))

	output11, err := client.UploadDir(&s3.UploadDirInput{
		DirPath: aws.String(dir),
		Bucket:  aws.String(bucket),
		Prefix:  aws.String(prefix + "symlink-follow/"),
	})
	c.Assert(err, IsNil)
	c.Assert(output11.SuccessNum, Equals, int64(4))

	// 10. 并发参数：Jobs=1, Parallel=1（串行）
	output12, err := client.UploadDir(&s3.UploadDirInput{
		DirPath:  aws.String(dir),
		Bucket:   aws.String(bucket),
		Prefix:   aws.String(prefix + "serial/"),
		Jobs:     aws.Long(1),
		Parallel: aws.Long(1),
	})
	c.Assert(err, IsNil)
	c.Assert(output12.SuccessNum, Equals, int64(4))

	// 11. 自定义PartSize
	output13, err := client.UploadDir(&s3.UploadDirInput{
		DirPath:  aws.String(dir),
		Bucket:   aws.String(bucket),
		Prefix:   aws.String(prefix + "partsize/"),
		PartSize: aws.Long(1024 * 1024),
	})
	c.Assert(err, IsNil)
	c.Assert(output13.SuccessNum, Equals, int64(4))

	// 12. 带ACL、StorageClass、Metadata、Tagging上传，验证生效
	output14, err := client.UploadDir(&s3.UploadDirInput{
		DirPath:      aws.String(dir),
		Bucket:       aws.String(bucket),
		Prefix:       aws.String(prefix + "meta/"),
		ACL:          aws.String("private"),
		StorageClass: aws.String("STANDARD_IA"),
		Metadata:     map[string]*string{"x-amz-meta-foo": aws.String("bar")},
		Tagging:      aws.String("key1=val1"),
	})
	c.Assert(err, IsNil)
	c.Assert(output14.SuccessNum, Equals, int64(4))
	// 验证metadata生效：key可能被HTTP/2转为小写
	headResp, _ := client.HeadObjectWithContext(context.Background(), &s3.HeadObjectInput{
		Bucket: aws.String(bucket), Key: aws.String(prefix + "meta/a.txt"),
	})
	var metaVal string
	for k, v := range headResp.Metadata {
		if strings.HasSuffix(strings.ToLower(k), "meta-foo") {
			metaVal = aws.ToString(v)
			break
		}
	}
	c.Assert(metaVal, Equals, "bar")
	// 验证tagging生效
	tagResp, _ := client.GetObjectTaggingWithContext(context.Background(), &s3.GetObjectTaggingInput{
		Bucket: aws.String(bucket), Key: aws.String(prefix + "meta/a.txt"),
	})
	c.Assert(len(tagResp.Tagging.TagSet) > 0, Equals, true)
	c.Assert(aws.ToString(tagResp.Tagging.TagSet[0].Key), Equals, "key1")

	// 13. ProgressFn回调
	var progressCalls int64
	output15, err := client.UploadDir(&s3.UploadDirInput{
		DirPath: aws.String(dir),
		Bucket:  aws.String(bucket),
		Prefix:  aws.String(prefix + "progress/"),
		ProgressFn: func(stat s3.DirResult) {
			atomic.AddInt64(&progressCalls, 1)
		},
	})
	c.Assert(err, IsNil)
	c.Assert(output15.SuccessNum, Equals, int64(4))
	c.Assert(atomic.LoadInt64(&progressCalls) > 0, Equals, true)

	// 14. EnableCheckpoint
	cpDir := randLowStr(8)
	os.MkdirAll(cpDir, 0755)
	output16, err := client.UploadDir(&s3.UploadDirInput{
		DirPath:          aws.String(dir),
		Bucket:           aws.String(bucket),
		Prefix:           aws.String(prefix + "checkpoint/"),
		EnableCheckpoint: aws.Boolean(true),
		CheckpointDir:    aws.String(cpDir),
	})
	c.Assert(err, IsNil)
	c.Assert(output16.SuccessNum, Equals, int64(4))

	// 15. 空目录：无常规文件可上传
	emptyDir := filepath.Join(os.TempDir(), randLowStr(8))
	os.MkdirAll(emptyDir, 0755)
	output17, err := client.UploadDir(&s3.UploadDirInput{
		DirPath: aws.String(emptyDir),
		Bucket:  aws.String(bucket),
		Prefix:  aws.String(prefix + "empty/"),
	})
	c.Assert(err, IsNil)
	c.Assert(output17.TotalNum, Equals, int64(0))
	c.Assert(output17.SuccessNum, Equals, int64(0))

	// 16. 参数校验：DirPath不存在
	_, err = client.UploadDir(&s3.UploadDirInput{
		DirPath: aws.String("/nonexistent_dir_" + randLowStr(8)),
		Bucket:  aws.String(bucket),
		Prefix:  aws.String(prefix),
	})
	c.Assert(err, NotNil)

	// 17. 参数校验：Bucket为空
	_, err = client.UploadDir(&s3.UploadDirInput{
		DirPath: aws.String(dir),
		Bucket:  aws.String(""),
		Prefix:  aws.String(prefix),
	})
	c.Assert(err, NotNil)

	// 清理
	resp, _ := client.ListObjects(&s3.ListObjectsInput{
		Bucket: aws.String(bucket), Prefix: aws.String(prefix),
	})
	for _, obj := range resp.Contents {
		client.DeleteObject(&s3.DeleteObjectInput{Bucket: aws.String(bucket), Key: obj.Key})
	}
	os.RemoveAll(dir)
	os.RemoveAll(cpDir)
	os.RemoveAll(emptyDir)
}

// TestDownloadDir 下载目录完整测试
func (s *Ks3utilCommandSuite) TestDownloadDir(c *C) {
	prefix := randLowStr(6) + "/"
	key1 := prefix + randLowStr(10)
	key2 := prefix + "subdir/" + randLowStr(10)
	content1 := randLowStr(100)
	content2 := randLowStr(200)

	client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucket), Key: aws.String(key1),
		Body: bytes.NewReader([]byte(content1)),
	})
	client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucket), Key: aws.String(key2),
		Body: bytes.NewReader([]byte(content2)),
	})

	// 1. 基本下载：验证文件内容和子目录创建
	localDir := randLowStr(8)
	output, err := client.DownloadDir(&s3.DownloadDirInput{
		Bucket:      aws.String(bucket),
		Prefix:      aws.String(prefix),
		DownloadDir: aws.String(localDir),
	})
	c.Assert(err, IsNil)
	c.Assert(output.TotalNum, Equals, int64(2))
	c.Assert(output.SuccessNum, Equals, int64(2))
	c.Assert(output.FailNum, Equals, int64(0))
	c.Assert(output.SkipNum, Equals, int64(0))
	c.Assert(len(output.SuccessFiles), Equals, 2)
	c.Assert(len(output.ErrorFiles), Equals, 0)
	// 验证文件存在
	_, err1 := os.Stat(filepath.Join(localDir, key1[len(prefix):]))
	_, err2 := os.Stat(filepath.Join(localDir, key2[len(prefix):]))
	c.Assert(err1, IsNil)
	c.Assert(err2, IsNil)
	// 验证文件内容
	data, _ := os.ReadFile(filepath.Join(localDir, key1[len(prefix):]))
	c.Assert(string(data), Equals, content1)

	// 2. SkipRule=IfExists：本地文件已存在，全部跳过
	output2, err := client.DownloadDir(&s3.DownloadDirInput{
		Bucket:      aws.String(bucket),
		Prefix:      aws.String(prefix),
		DownloadDir: aws.String(localDir),
		SkipRule:    aws.String(s3.SkipIfExists),
	})
	c.Assert(err, IsNil)
	c.Assert(output2.SkipNum, Equals, int64(2))
	c.Assert(output2.SuccessNum, Equals, int64(0))

	// 3. SkipRule=IfSizeEquals：改小本地key1文件 → 大小不等 → 不跳过
	createFileWithContent(filepath.Join(localDir, key1[len(prefix):]), "short")
	output3, err := client.DownloadDir(&s3.DownloadDirInput{
		Bucket:      aws.String(bucket),
		Prefix:      aws.String(prefix),
		DownloadDir: aws.String(localDir),
		SkipRule:    aws.String(s3.SkipIfSizeEquals),
	})
	c.Assert(err, IsNil)
	c.Assert(output3.SkipNum, Equals, int64(1))  // key2大小相等，跳过
	c.Assert(output3.SuccessNum, Equals, int64(1)) // key1大小不等，重新下载
	// 验证key1内容已被覆盖
	data3, _ := os.ReadFile(filepath.Join(localDir, key1[len(prefix):]))
	c.Assert(string(data3), Equals, content1)

	// 4. SkipRule=IfNewer：把本地key1时间改旧 → 本地比远程旧 → 不跳过
	pastTime := time.Now().Add(-24 * time.Hour)
	os.Chtimes(filepath.Join(localDir, key1[len(prefix):]), pastTime, pastTime)
	os.Chtimes(filepath.Join(localDir, key2[len(prefix):]), pastTime, pastTime)
	output4, err := client.DownloadDir(&s3.DownloadDirInput{
		Bucket:      aws.String(bucket),
		Prefix:      aws.String(prefix),
		DownloadDir: aws.String(localDir),
		SkipRule:    aws.String(s3.SkipIfNewer),
	})
	c.Assert(err, IsNil)
	c.Assert(output4.SkipNum, Equals, int64(0))  // 两个本地都更旧，都不跳过
	c.Assert(output4.SuccessNum, Equals, int64(2))

	// 把本地时间改为未来 → 本地比远程新 → 全部跳过
	futureTime := time.Now().Add(24 * time.Hour)
	os.Chtimes(filepath.Join(localDir, key1[len(prefix):]), futureTime, futureTime)
	os.Chtimes(filepath.Join(localDir, key2[len(prefix):]), futureTime, futureTime)
	output4b, err := client.DownloadDir(&s3.DownloadDirInput{
		Bucket:      aws.String(bucket),
		Prefix:      aws.String(prefix),
		DownloadDir: aws.String(localDir),
		SkipRule:    aws.String(s3.SkipIfNewer),
	})
	c.Assert(err, IsNil)
	c.Assert(output4b.SkipNum, Equals, int64(2)) // 全部本地更新，跳过
	c.Assert(output4b.SuccessNum, Equals, int64(0))

	// 5. SkipRule=IfNewerAndSizeEquals：本地更新且大小相等才跳过
	createFileWithContent(filepath.Join(localDir, key1[len(prefix):]), content1)
	os.Chtimes(filepath.Join(localDir, key1[len(prefix):]), futureTime, futureTime)
	os.Chtimes(filepath.Join(localDir, key2[len(prefix):]), futureTime, futureTime)
	output6, err := client.DownloadDir(&s3.DownloadDirInput{
		Bucket:      aws.String(bucket),
		Prefix:      aws.String(prefix),
		DownloadDir: aws.String(localDir),
		SkipRule:    aws.String(s3.SkipIfNewerAndSizeEquals),
	})
	c.Assert(err, IsNil)
	c.Assert(output6.SkipNum, Equals, int64(2)) // 两个本地都更新且大小相等，跳过
	c.Assert(output6.SuccessNum, Equals, int64(0))

	// 6. SkipRule=IfCrc64Equals：改key1本地内容 → CRC不等 → 不跳过
	createFileWithContent(filepath.Join(localDir, key1[len(prefix):]), "wrong content")
	output7, err := client.DownloadDir(&s3.DownloadDirInput{
		Bucket:      aws.String(bucket),
		Prefix:      aws.String(prefix),
		DownloadDir: aws.String(localDir),
		SkipRule:    aws.String(s3.SkipIfCrc64Equals),
	})
	c.Assert(err, IsNil)
	c.Assert(output7.SkipNum, Equals, int64(1))  // key2 CRC相等，跳过
	c.Assert(output7.SuccessNum, Equals, int64(1)) // key1 CRC不等，重新下载
	// 再下载同内容 → CRC匹配 → 全部跳过
	output7b, err := client.DownloadDir(&s3.DownloadDirInput{
		Bucket:      aws.String(bucket),
		Prefix:      aws.String(prefix),
		DownloadDir: aws.String(localDir),
		SkipRule:    aws.String(s3.SkipIfCrc64Equals),
	})
	c.Assert(err, IsNil)
	c.Assert(output7b.SkipNum, Equals, int64(2))
	c.Assert(output7b.SuccessNum, Equals, int64(0))

	// 7. SkipRule=Never：不跳过任何文件
	output7c, err := client.DownloadDir(&s3.DownloadDirInput{
		Bucket:      aws.String(bucket),
		Prefix:      aws.String(prefix),
		DownloadDir: aws.String(localDir),
		SkipRule:    aws.String(s3.SkipNever),
	})
	c.Assert(err, IsNil)
	c.Assert(output7c.SkipNum, Equals, int64(0))
	c.Assert(output7c.SuccessNum, Equals, int64(2))

	// 8. 并发参数：Jobs=1, Parallel=1
	localDir2 := randLowStr(8)
	output8, err := client.DownloadDir(&s3.DownloadDirInput{
		Bucket:      aws.String(bucket),
		Prefix:      aws.String(prefix),
		DownloadDir: aws.String(localDir2),
		Jobs:        aws.Long(1),
		Parallel:    aws.Long(1),
	})
	c.Assert(err, IsNil)
	c.Assert(output8.SuccessNum, Equals, int64(2))

	// 9. 自定义PartSize
	localDir3 := randLowStr(8)
	output8b, err := client.DownloadDir(&s3.DownloadDirInput{
		Bucket:      aws.String(bucket),
		Prefix:      aws.String(prefix),
		DownloadDir: aws.String(localDir3),
		PartSize:    aws.Long(1024 * 1024),
	})
	c.Assert(err, IsNil)
	c.Assert(output8b.SuccessNum, Equals, int64(2))

	// 10. ProgressFn回调
	var progressCalls int64
	localDir4 := randLowStr(8)
	output9, err := client.DownloadDir(&s3.DownloadDirInput{
		Bucket:      aws.String(bucket),
		Prefix:      aws.String(prefix),
		DownloadDir: aws.String(localDir4),
		ProgressFn: func(stat s3.DirResult) {
			atomic.AddInt64(&progressCalls, 1)
		},
	})
	c.Assert(err, IsNil)
	c.Assert(output9.SuccessNum, Equals, int64(2))
	c.Assert(atomic.LoadInt64(&progressCalls) > 0, Equals, true)

	// 11. IgnoreSuccessFiles
	localDir5 := randLowStr(8)
	output10, err := client.DownloadDir(&s3.DownloadDirInput{
		Bucket:             aws.String(bucket),
		Prefix:             aws.String(prefix),
		DownloadDir:        aws.String(localDir5),
		IgnoreSuccessFiles: aws.Boolean(true),
	})
	c.Assert(err, IsNil)
	c.Assert(output10.SuccessNum, Equals, int64(2))
	c.Assert(len(output10.SuccessFiles), Equals, 0)

	// 12. EnableCheckpoint
	cpDir := randLowStr(8)
	os.MkdirAll(cpDir, 0755)
	localDir6 := randLowStr(8)
	output11, err := client.DownloadDir(&s3.DownloadDirInput{
		Bucket:           aws.String(bucket),
		Prefix:           aws.String(prefix),
		DownloadDir:      aws.String(localDir6),
		EnableCheckpoint: aws.Boolean(true),
		CheckpointDir:    aws.String(cpDir),
	})
	c.Assert(err, IsNil)
	c.Assert(output11.SuccessNum, Equals, int64(2))

	// 13. 空前缀：桶中无匹配对象
	localDir7 := randLowStr(8)
	output12, err := client.DownloadDir(&s3.DownloadDirInput{
		Bucket:      aws.String(bucket),
		Prefix:      aws.String(prefix + "nonexistent/"),
		DownloadDir: aws.String(localDir7),
	})
	c.Assert(err, IsNil)
	c.Assert(output12.TotalNum, Equals, int64(0))
	c.Assert(output12.SuccessNum, Equals, int64(0))

	// 14. 参数校验：Bucket为空
	_, err = client.DownloadDir(&s3.DownloadDirInput{
		Bucket:      aws.String(""),
		Prefix:      aws.String(prefix),
		DownloadDir: aws.String(localDir),
	})
	c.Assert(err, NotNil)

	// 清理
	for _, key := range []string{key1, key2} {
		client.DeleteObject(&s3.DeleteObjectInput{Bucket: aws.String(bucket), Key: aws.String(key)})
	}
	for _, d := range []string{localDir, localDir2, localDir3, localDir4, localDir5, localDir6, localDir7, cpDir} {
		os.RemoveAll(d)
	}
}

// TestCopyDir 复制目录完整测试
func (s *Ks3utilCommandSuite) TestCopyDir(c *C) {
	srcPrefix := randLowStr(6) + "/"
	key1 := srcPrefix + randLowStr(10)
	key2 := srcPrefix + "subdir/" + randLowStr(10)
	content1 := randLowStr(100)
	content2 := randLowStr(200)

	client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucket), Key: aws.String(key1),
		Body: bytes.NewReader([]byte(content1)),
	})
	client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucket), Key: aws.String(key2),
		Body: bytes.NewReader([]byte(content2)),
	})

	// 创建目标桶，避免同桶复制时的ObjectAlreadyExists问题
	dstBucket := commonNamePrefix + randLowStr(10)
	s.CreateBucket(dstBucket, c)
	defer s.DeleteBucket(dstBucket, c)

	// 1. 基本复制：验证目标对象存在且内容正确
	dstPrefix := randLowStr(6) + "/"
	output, err := client.CopyDir(&s3.CopyDirInput{
		SourceBucket: aws.String(bucket),
		SourcePrefix: aws.String(srcPrefix),
		Bucket:       aws.String(dstBucket),
		Prefix:    aws.String(dstPrefix),
	})
	c.Assert(err, IsNil, Commentf("CopyDir basic failed: %v", err))
	c.Assert(output.TotalNum, Equals, int64(2))
	c.Assert(output.SuccessNum, Equals, int64(2))
	c.Assert(output.FailNum, Equals, int64(0))
	c.Assert(output.SkipNum, Equals, int64(0))
	c.Assert(len(output.SuccessFiles), Equals, 2)
	// 验证目标对象内容
	getResp, _ := client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(dstBucket), Key: aws.String(dstPrefix + key1[len(srcPrefix):]),
	})
	body, _ := io.ReadAll(getResp.Body)
	getResp.Body.Close()
	c.Assert(string(body), Equals, content1)

	// 2. SkipRule=IfExists：目标已存在，全部跳过
	output2, err := client.CopyDir(&s3.CopyDirInput{
		SourceBucket: aws.String(bucket),
		SourcePrefix: aws.String(srcPrefix),
		Bucket:       aws.String(dstBucket),
		Prefix:    aws.String(dstPrefix),
		SkipRule:     aws.String(s3.SkipIfExists),
	})
	c.Assert(err, IsNil)
	c.Assert(output2.SkipNum, Equals, int64(2))
	c.Assert(output2.SuccessNum, Equals, int64(0))

	// 3. SkipRule=IfSizeEquals：改源key1大小 → 不跳过
	client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucket), Key: aws.String(key1),
		Body: bytes.NewReader([]byte(randLowStr(300))),
	})
	time.Sleep(time.Second * 1)
	output3, err := client.CopyDir(&s3.CopyDirInput{
		SourceBucket: aws.String(bucket),
		SourcePrefix: aws.String(srcPrefix),
		Bucket:       aws.String(dstBucket),
		Prefix:    aws.String(dstPrefix),
		SkipRule:     aws.String(s3.SkipIfSizeEquals),
	})
	c.Assert(err, IsNil)
	c.Assert(output3.SkipNum, Equals, int64(1))
	c.Assert(output3.SuccessNum, Equals, int64(1))

	// 4. SkipRule=IfNewer：刚复制完dst比src更新 → 全部跳过
	output4, err := client.CopyDir(&s3.CopyDirInput{
		SourceBucket: aws.String(bucket),
		SourcePrefix: aws.String(srcPrefix),
		Bucket:       aws.String(dstBucket),
		Prefix:    aws.String(dstPrefix),
		SkipRule:     aws.String(s3.SkipIfNewer),
	})
	c.Assert(err, IsNil)
	c.Assert(output4.SkipNum, Equals, int64(2))
	c.Assert(output4.SuccessNum, Equals, int64(0))

	// 重新上传src key2 → src更新 → 不跳过
	client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucket), Key: aws.String(key2),
		Body: bytes.NewReader([]byte(randLowStr(200))),
	})
	time.Sleep(time.Second * 1)
	output5, err := client.CopyDir(&s3.CopyDirInput{
		SourceBucket: aws.String(bucket),
		SourcePrefix: aws.String(srcPrefix),
		Bucket:       aws.String(dstBucket),
		Prefix:    aws.String(dstPrefix),
		SkipRule:     aws.String(s3.SkipIfNewer),
	})
	c.Assert(err, IsNil)
	c.Assert(output5.SkipNum, Equals, int64(1))
	c.Assert(output5.SuccessNum, Equals, int64(1))

	// 5. SkipRule=IfNewerAndSizeEquals：dst更新且大小相等 → 跳过
	output6, err := client.CopyDir(&s3.CopyDirInput{
		SourceBucket: aws.String(bucket),
		SourcePrefix: aws.String(srcPrefix),
		Bucket:       aws.String(dstBucket),
		Prefix:    aws.String(dstPrefix),
		SkipRule:     aws.String(s3.SkipIfNewerAndSizeEquals),
	})
	c.Assert(err, IsNil)
	c.Assert(output6.SkipNum, Equals, int64(2))

	// 改src key1大小 → dst不满足大小相等 → 不跳过
	client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucket), Key: aws.String(key1),
		Body: bytes.NewReader([]byte(randLowStr(500))),
	})
	time.Sleep(time.Second * 1)
	output7, err := client.CopyDir(&s3.CopyDirInput{
		SourceBucket: aws.String(bucket),
		SourcePrefix: aws.String(srcPrefix),
		Bucket:       aws.String(dstBucket),
		Prefix:    aws.String(dstPrefix),
		SkipRule:     aws.String(s3.SkipIfNewerAndSizeEquals),
	})
	c.Assert(err, IsNil)
	c.Assert(output7.SkipNum, Equals, int64(1))
	c.Assert(output7.SuccessNum, Equals, int64(1))

	// 6. SkipRule=IfCrc64Equals：改src key1内容 → CRC不等 → 不跳过
	client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucket), Key: aws.String(key1),
		Body: bytes.NewReader([]byte("totally different content")),
	})
	time.Sleep(time.Second * 1)
	output8, err := client.CopyDir(&s3.CopyDirInput{
		SourceBucket: aws.String(bucket),
		SourcePrefix: aws.String(srcPrefix),
		Bucket:       aws.String(dstBucket),
		Prefix:    aws.String(dstPrefix),
		SkipRule:     aws.String(s3.SkipIfCrc64Equals),
	})
	c.Assert(err, IsNil)
	c.Assert(output8.SkipNum, Equals, int64(1))
	c.Assert(output8.SuccessNum, Equals, int64(1))
	// 再复制同内容 → CRC匹配 → 全部跳过
	output8b, err := client.CopyDir(&s3.CopyDirInput{
		SourceBucket: aws.String(bucket),
		SourcePrefix: aws.String(srcPrefix),
		Bucket:       aws.String(dstBucket),
		Prefix:    aws.String(dstPrefix),
		SkipRule:     aws.String(s3.SkipIfCrc64Equals),
	})
	c.Assert(err, IsNil)
	c.Assert(output8b.SkipNum, Equals, int64(2))
	c.Assert(output8b.SuccessNum, Equals, int64(0))

	// 7. SkipRule=Never：不跳过任何文件
	client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucket), Key: aws.String(key1),
		Body: bytes.NewReader([]byte("never skip test")),
	})
	output8c, err := client.CopyDir(&s3.CopyDirInput{
		SourceBucket: aws.String(bucket),
		SourcePrefix: aws.String(srcPrefix),
		Bucket:       aws.String(dstBucket),
		Prefix:    aws.String(dstPrefix),
		SkipRule:     aws.String(s3.SkipNever),
	})
	c.Assert(err, IsNil)
	c.Assert(output8c.SkipNum, Equals, int64(0))
	c.Assert(output8c.SuccessNum, Equals, int64(2))

	// 8. 带StorageClass复制，验证生效
	storagePrefix := dstPrefix + "archive/"
	output9, err := client.CopyDir(&s3.CopyDirInput{
		SourceBucket: aws.String(bucket),
		SourcePrefix: aws.String(srcPrefix),
		Bucket:       aws.String(dstBucket),
		Prefix:    aws.String(storagePrefix),
		StorageClass: aws.String("ARCHIVE"),
	})
	c.Assert(err, IsNil)
	c.Assert(output9.SuccessNum, Equals, int64(2))
	// 验证StorageClass生效：key可能被HTTP/2转为小写
	headResp, _ := client.HeadObjectWithContext(context.Background(), &s3.HeadObjectInput{
		Bucket: aws.String(dstBucket), Key: aws.String(storagePrefix + key1[len(srcPrefix):]),
	})
	var storageClass string
	for k, v := range headResp.Metadata {
		if strings.EqualFold(k, s3.HTTPHeaderAmzStorageClass) {
			storageClass = aws.ToString(v)
			break
		}
	}
	c.Assert(storageClass, Equals, "ARCHIVE")

	// 9. 带ACL、Metadata、Tagging复制，验证REPLACE生效
	metaPrefix := dstPrefix + "meta/"
	output10, err := client.CopyDir(&s3.CopyDirInput{
		SourceBucket:      aws.String(bucket),
		SourcePrefix:      aws.String(srcPrefix),
		Bucket:            aws.String(dstBucket),
		Prefix:         aws.String(metaPrefix),
		ACL:               aws.String("private"),
		StorageClass:      aws.String("STANDARD"),
		Metadata:          map[string]*string{"x-amz-meta-foo": aws.String("bar")},
		MetadataDirective: aws.String("REPLACE"),
		Tagging:           aws.String("key1=val1"),
		TaggingDirective:  aws.String("REPLACE"),
	})
	c.Assert(err, IsNil)
	c.Assert(output10.SuccessNum, Equals, int64(2))
	// 验证metadata被替换：key可能被HTTP/2转为小写
	metaResp, _ := client.HeadObjectWithContext(context.Background(), &s3.HeadObjectInput{
		Bucket: aws.String(dstBucket), Key: aws.String(metaPrefix + key1[len(srcPrefix):]),
	})
	var metaVal string
	for k, v := range metaResp.Metadata {
		if strings.HasSuffix(strings.ToLower(k), "meta-foo") {
			metaVal = aws.ToString(v)
			break
		}
	}
	c.Assert(metaVal, Equals, "bar")
	// 验证tagging被替换
	tagResp, _ := client.GetObjectTaggingWithContext(context.Background(), &s3.GetObjectTaggingInput{
		Bucket: aws.String(dstBucket), Key: aws.String(metaPrefix + key1[len(srcPrefix):]),
	})
	c.Assert(len(tagResp.Tagging.TagSet) > 0, Equals, true)
	c.Assert(aws.ToString(tagResp.Tagging.TagSet[0].Key), Equals, "key1")
	c.Assert(aws.ToString(tagResp.Tagging.TagSet[0].Value), Equals, "val1")

	// 10. 覆盖复制：同目标复制两次，验证ForbidOverwrite=false允许覆盖
	overwritePrefix := dstPrefix + "overwrite/"
	output11, err := client.CopyDir(&s3.CopyDirInput{
		SourceBucket: aws.String(bucket),
		SourcePrefix: aws.String(srcPrefix),
		Bucket:       aws.String(dstBucket),
		Prefix:    aws.String(overwritePrefix),
	})
	c.Assert(err, IsNil)
	c.Assert(output11.SuccessNum, Equals, int64(2))
	// 再次复制到同一目标，应覆盖成功
	output12, err := client.CopyDir(&s3.CopyDirInput{
		SourceBucket: aws.String(bucket),
		SourcePrefix: aws.String(srcPrefix),
		Bucket:       aws.String(dstBucket),
		Prefix:    aws.String(overwritePrefix),
	})
	c.Assert(err, IsNil)
	c.Assert(output12.SuccessNum, Equals, int64(2))

	// 11. 并发参数
	serialPrefix := dstPrefix + "serial/"
	output13, err := client.CopyDir(&s3.CopyDirInput{
		SourceBucket: aws.String(bucket),
		SourcePrefix: aws.String(srcPrefix),
		Bucket:       aws.String(dstBucket),
		Prefix:    aws.String(serialPrefix),
		Jobs:     aws.Long(1),
		Parallel: aws.Long(1),
	})
	c.Assert(err, IsNil)
	c.Assert(output13.SuccessNum, Equals, int64(2))

	// 12. 自定义PartSize
	partSizePrefix := dstPrefix + "partsize/"
	output13b, err := client.CopyDir(&s3.CopyDirInput{
		SourceBucket: aws.String(bucket),
		SourcePrefix: aws.String(srcPrefix),
		Bucket:       aws.String(dstBucket),
		Prefix:    aws.String(partSizePrefix),
		PartSize:     aws.Long(1024 * 1024),
	})
	c.Assert(err, IsNil)
	c.Assert(output13b.SuccessNum, Equals, int64(2))

	// 13. ProgressFn回调
	var progressCalls int64
	progressPrefix := dstPrefix + "progress/"
	output14, err := client.CopyDir(&s3.CopyDirInput{
		SourceBucket: aws.String(bucket),
		SourcePrefix: aws.String(srcPrefix),
		Bucket:       aws.String(dstBucket),
		Prefix:    aws.String(progressPrefix),
		ProgressFn: func(stat s3.DirResult) {
			atomic.AddInt64(&progressCalls, 1)
		},
	})
	c.Assert(err, IsNil)
	c.Assert(output14.SuccessNum, Equals, int64(2))
	c.Assert(atomic.LoadInt64(&progressCalls) > 0, Equals, true)

	// 14. EnableCheckpoint
	cpDir := randLowStr(8)
	os.MkdirAll(cpDir, 0755)
	cpPrefix := dstPrefix + "checkpoint/"
	output15, err := client.CopyDir(&s3.CopyDirInput{
		SourceBucket:     aws.String(bucket),
		SourcePrefix:     aws.String(srcPrefix),
		Bucket:           aws.String(dstBucket),
		Prefix:        aws.String(cpPrefix),
		EnableCheckpoint: aws.Boolean(true),
		CheckpointDir:    aws.String(cpDir),
	})
	c.Assert(err, IsNil)
	c.Assert(output15.SuccessNum, Equals, int64(2))

	// 15. IgnoreSuccessFiles
	ignorePrefix := dstPrefix + "ignore/"
	output16, err := client.CopyDir(&s3.CopyDirInput{
		SourceBucket:       aws.String(bucket),
		SourcePrefix:       aws.String(srcPrefix),
		Bucket:             aws.String(dstBucket),
		Prefix:          aws.String(ignorePrefix),
		IgnoreSuccessFiles: aws.Boolean(true),
	})
	c.Assert(err, IsNil)
	c.Assert(output16.SuccessNum, Equals, int64(2))
	c.Assert(len(output16.SuccessFiles), Equals, 0)

	// 16. 空源前缀：无匹配对象
	emptyPrefix := dstPrefix + "empty/"
	output17, err := client.CopyDir(&s3.CopyDirInput{
		SourceBucket: aws.String(bucket),
		SourcePrefix: aws.String(srcPrefix + "nonexistent/"),
		Bucket:       aws.String(dstBucket),
		Prefix:    aws.String(emptyPrefix),
	})
	c.Assert(err, IsNil)
	c.Assert(output17.TotalNum, Equals, int64(0))

	// 17. 参数校验：SourceBucket为空
	_, err = client.CopyDir(&s3.CopyDirInput{
		SourceBucket: aws.String(""),
		SourcePrefix: aws.String(srcPrefix),
		Bucket:       aws.String(dstBucket),
		Prefix:    aws.String(dstPrefix),
	})
	c.Assert(err, NotNil)

	// 18. 参数校验：Bucket为空
	_, err = client.CopyDir(&s3.CopyDirInput{
		SourceBucket: aws.String(bucket),
		SourcePrefix: aws.String(srcPrefix),
		Bucket:       aws.String(""),
		Prefix:    aws.String(dstPrefix),
	})
	c.Assert(err, NotNil)

	// 清理源对象
	resp, _ := client.ListObjects(&s3.ListObjectsInput{
		Bucket: aws.String(bucket), Prefix: aws.String(srcPrefix),
	})
	for _, obj := range resp.Contents {
		client.DeleteObject(&s3.DeleteObjectInput{Bucket: aws.String(bucket), Key: obj.Key})
	}
	// 清理目标对象
	resp, _ = client.ListObjects(&s3.ListObjectsInput{
		Bucket: aws.String(dstBucket), Prefix: aws.String(dstPrefix),
	})
	for _, obj := range resp.Contents {
		client.DeleteObject(&s3.DeleteObjectInput{Bucket: aws.String(dstBucket), Key: obj.Key})
	}
	os.RemoveAll(cpDir)
}

// TestCopyDirAcrossRegion 跨区域复制目录完整测试
func (s *Ks3utilCommandSuite) TestCopyDirAcrossRegion(c *C) {
	// 目标端client
	var cre = credentials.NewStaticCredentials(accessKeyID, accessKeySecret, "")
	dstClient := s3.New(&aws.Config{
		Credentials: cre,
		Region:      "SHANGHAI",
		Endpoint:    "ks3-cn-shanghai.ksyuncs.com",
	})

	// 创建上海的目标桶
	dstBucket := commonNamePrefix + randLowStr(10)
	_, err := dstClient.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(dstBucket),
	})
	c.Assert(err, IsNil)

	srcPrefix := randLowStr(6) + "/"
	key1 := srcPrefix + randLowStr(10)
	key2 := srcPrefix + "subdir/" + randLowStr(10)
	content1 := randLowStr(100)
	content2 := randLowStr(200)

	// 上传源对象到源桶
	client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucket), Key: aws.String(key1),
		Body: bytes.NewReader([]byte(content1)),
	})
	client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucket), Key: aws.String(key2),
		Body: bytes.NewReader([]byte(content2)),
	})

	// 1. 基本跨区域复制：验证目标对象内容
	dstPrefix := randLowStr(6) + "/"
	output, err := client.CopyDirAcrossRegion(&s3.CopyDirInput{
		SourceBucket: aws.String(bucket),
		SourcePrefix: aws.String(srcPrefix),
		Bucket:       aws.String(dstBucket),
		Prefix:    aws.String(dstPrefix),
	}, dstClient)
	c.Assert(err, IsNil)
	c.Assert(output.TotalNum, Equals, int64(2))
	c.Assert(output.SuccessNum, Equals, int64(2))
	c.Assert(output.FailNum, Equals, int64(0))
	c.Assert(output.SkipNum, Equals, int64(0))
	c.Assert(len(output.SuccessFiles), Equals, 2)
	// 验证目标对象内容
	getResp, _ := dstClient.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(dstBucket), Key: aws.String(dstPrefix + key1[len(srcPrefix):]),
	})
	body, _ := io.ReadAll(getResp.Body)
	getResp.Body.Close()
	c.Assert(string(body), Equals, content1)

	// 2. SkipRule=IfExists：目标已存在，全部跳过
	output2, err := client.CopyDirAcrossRegion(&s3.CopyDirInput{
		SourceBucket: aws.String(bucket),
		SourcePrefix: aws.String(srcPrefix),
		Bucket:       aws.String(dstBucket),
		Prefix:    aws.String(dstPrefix),
		SkipRule:     aws.String(s3.SkipIfExists),
	}, dstClient)
	c.Assert(err, IsNil)
	c.Assert(output2.SkipNum, Equals, int64(2))
	c.Assert(output2.SuccessNum, Equals, int64(0))

	// 3. SkipRule=IfSizeEquals：改源key1大小 → 不跳过
	client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucket), Key: aws.String(key1),
		Body: bytes.NewReader([]byte(randLowStr(300))),
	})
	time.Sleep(time.Second * 3)
	output3, err := client.CopyDirAcrossRegion(&s3.CopyDirInput{
		SourceBucket: aws.String(bucket),
		SourcePrefix: aws.String(srcPrefix),
		Bucket:       aws.String(dstBucket),
		Prefix:    aws.String(dstPrefix),
		SkipRule:     aws.String(s3.SkipIfSizeEquals),
	}, dstClient)
	c.Assert(err, IsNil)
	c.Assert(output3.SkipNum, Equals, int64(1))
	c.Assert(output3.SuccessNum, Equals, int64(1))

	// 4. SkipRule=IfNewer：刚复制完dst比src更新 → 全部跳过
	output4, err := client.CopyDirAcrossRegion(&s3.CopyDirInput{
		SourceBucket: aws.String(bucket),
		SourcePrefix: aws.String(srcPrefix),
		Bucket:       aws.String(dstBucket),
		Prefix:       aws.String(dstPrefix),
		SkipRule:     aws.String(s3.SkipIfNewer),
	}, dstClient)
	c.Assert(err, IsNil)
	c.Assert(output4.SkipNum, Equals, int64(2))
	c.Assert(output4.SuccessNum, Equals, int64(0))

	// 重新上传src key2 → src更新 → 不跳过
	client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucket), Key: aws.String(key2),
		Body: bytes.NewReader([]byte(randLowStr(200))),
	})
	time.Sleep(time.Second * 3)
	output5, err := client.CopyDirAcrossRegion(&s3.CopyDirInput{
		SourceBucket: aws.String(bucket),
		SourcePrefix: aws.String(srcPrefix),
		Bucket:       aws.String(dstBucket),
		Prefix:    aws.String(dstPrefix),
		SkipRule:     aws.String(s3.SkipIfNewer),
	}, dstClient)
	c.Assert(err, IsNil)
	c.Assert(output5.SkipNum, Equals, int64(1))
	c.Assert(output5.SuccessNum, Equals, int64(1))

	// 5. SkipRule=IfNewerAndSizeEquals：dst更新且大小相等 → 全部跳过
	output6, err := client.CopyDirAcrossRegion(&s3.CopyDirInput{
		SourceBucket: aws.String(bucket),
		SourcePrefix: aws.String(srcPrefix),
		Bucket:       aws.String(dstBucket),
		Prefix:       aws.String(dstPrefix),
		SkipRule:     aws.String(s3.SkipIfNewerAndSizeEquals),
	}, dstClient)
	c.Assert(err, IsNil)
	c.Assert(output6.SkipNum, Equals, int64(2))
	// 改src key1大小 → dst不满足大小相等 → 不跳过
	client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucket), Key: aws.String(key1),
		Body: bytes.NewReader([]byte(randLowStr(500))),
	})
	time.Sleep(time.Second * 3)
	output6b, err := client.CopyDirAcrossRegion(&s3.CopyDirInput{
		SourceBucket: aws.String(bucket),
		SourcePrefix: aws.String(srcPrefix),
		Bucket:       aws.String(dstBucket),
		Prefix:       aws.String(dstPrefix),
		SkipRule:     aws.String(s3.SkipIfNewerAndSizeEquals),
	}, dstClient)
	c.Assert(err, IsNil)
	c.Assert(output6b.SkipNum, Equals, int64(1))
	c.Assert(output6b.SuccessNum, Equals, int64(1))

	// 6. SkipRule=IfCrc64Equals：改src key1内容 → CRC不等 → 不跳过
	client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucket), Key: aws.String(key1),
		Body: bytes.NewReader([]byte("totally different content")),
	})
	time.Sleep(time.Second * 3)
	output7, err := client.CopyDirAcrossRegion(&s3.CopyDirInput{
		SourceBucket: aws.String(bucket),
		SourcePrefix: aws.String(srcPrefix),
		Bucket:       aws.String(dstBucket),
		Prefix:       aws.String(dstPrefix),
		SkipRule:     aws.String(s3.SkipIfCrc64Equals),
	}, dstClient)
	c.Assert(err, IsNil)
	c.Assert(output7.SkipNum, Equals, int64(1))
	c.Assert(output7.SuccessNum, Equals, int64(1))
	// 再复制同内容 → CRC匹配 → 全部跳过
	output7b, err := client.CopyDirAcrossRegion(&s3.CopyDirInput{
		SourceBucket: aws.String(bucket),
		SourcePrefix: aws.String(srcPrefix),
		Bucket:       aws.String(dstBucket),
		Prefix:    aws.String(dstPrefix),
		SkipRule:     aws.String(s3.SkipIfCrc64Equals),
	}, dstClient)
	c.Assert(err, IsNil)
	c.Assert(output7b.SkipNum, Equals, int64(2))
	c.Assert(output7b.SuccessNum, Equals, int64(0))

	// 7. SkipRule=Never：不跳过任何文件
	client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucket), Key: aws.String(key1),
		Body: bytes.NewReader([]byte("never skip test")),
	})
	output7c, err := client.CopyDirAcrossRegion(&s3.CopyDirInput{
		SourceBucket: aws.String(bucket),
		SourcePrefix: aws.String(srcPrefix),
		Bucket:       aws.String(dstBucket),
		Prefix:    aws.String(dstPrefix),
		SkipRule:     aws.String(s3.SkipNever),
	}, dstClient)
	c.Assert(err, IsNil)
	c.Assert(output7c.SkipNum, Equals, int64(0))
	c.Assert(output7c.SuccessNum, Equals, int64(2))

	// 8. 带StorageClass跨区域复制，验证生效
	storageDstPrefix := randLowStr(6) + "/"
	output8, err := client.CopyDirAcrossRegion(&s3.CopyDirInput{
		SourceBucket: aws.String(bucket),
		SourcePrefix: aws.String(srcPrefix),
		Bucket:       aws.String(dstBucket),
		Prefix:    aws.String(storageDstPrefix),
		StorageClass: aws.String("ARCHIVE"),
	}, dstClient)
	c.Assert(err, IsNil)
	c.Assert(output8.SuccessNum, Equals, int64(2))
	headResp, _ := dstClient.HeadObjectWithContext(context.Background(), &s3.HeadObjectInput{
		Bucket: aws.String(dstBucket), Key: aws.String(storageDstPrefix + key1[len(srcPrefix):]),
	})
	var storageClass string
	for k, v := range headResp.Metadata {
		if strings.EqualFold(k, s3.HTTPHeaderAmzStorageClass) {
			storageClass = aws.ToString(v)
			break
		}
	}
	c.Assert(storageClass, Equals, "ARCHIVE")

	// 9. 带ACL、Metadata、Tagging跨区域复制，验证REPLACE生效
	metaDstPrefix := randLowStr(6) + "/"
	output9, err := client.CopyDirAcrossRegion(&s3.CopyDirInput{
		SourceBucket:      aws.String(bucket),
		SourcePrefix:      aws.String(srcPrefix),
		Bucket:            aws.String(dstBucket),
		Prefix:         aws.String(metaDstPrefix),
		ACL:               aws.String("private"),
		StorageClass:      aws.String("STANDARD"),
		Metadata:          map[string]*string{"x-amz-meta-foo": aws.String("bar")},
		MetadataDirective: aws.String("REPLACE"),
		Tagging:           aws.String("key1=val1"),
		TaggingDirective:  aws.String("REPLACE"),
	}, dstClient)
	c.Assert(err, IsNil)
	c.Assert(output9.SuccessNum, Equals, int64(2))
	// 验证metadata被替换：key可能被HTTP/2转为小写
	metaResp, _ := dstClient.HeadObjectWithContext(context.Background(), &s3.HeadObjectInput{
		Bucket: aws.String(dstBucket), Key: aws.String(metaDstPrefix + key1[len(srcPrefix):]),
	})
	var metaVal string
	for k, v := range metaResp.Metadata {
		if strings.HasSuffix(strings.ToLower(k), "meta-foo") {
			metaVal = aws.ToString(v)
			break
		}
	}
	c.Assert(metaVal, Equals, "bar")
	// 验证tagging被替换
	tagResp, _ := dstClient.GetObjectTaggingWithContext(context.Background(), &s3.GetObjectTaggingInput{
		Bucket: aws.String(dstBucket), Key: aws.String(metaDstPrefix + key1[len(srcPrefix):]),
	})
	c.Assert(len(tagResp.Tagging.TagSet) > 0, Equals, true)
	c.Assert(aws.ToString(tagResp.Tagging.TagSet[0].Key), Equals, "key1")
	c.Assert(aws.ToString(tagResp.Tagging.TagSet[0].Value), Equals, "val1")

	// 10. 覆盖复制：验证ForbidOverwrite=false允许跨区域覆盖
	overwriteDstPrefix := randLowStr(6) + "/"
	output10, err := client.CopyDirAcrossRegion(&s3.CopyDirInput{
		SourceBucket: aws.String(bucket),
		SourcePrefix: aws.String(srcPrefix),
		Bucket:       aws.String(dstBucket),
		Prefix:    aws.String(overwriteDstPrefix),
	}, dstClient)
	c.Assert(err, IsNil)
	c.Assert(output10.SuccessNum, Equals, int64(2))
	output10b, err := client.CopyDirAcrossRegion(&s3.CopyDirInput{
		SourceBucket: aws.String(bucket),
		SourcePrefix: aws.String(srcPrefix),
		Bucket:       aws.String(dstBucket),
		Prefix:    aws.String(overwriteDstPrefix),
	}, dstClient)
	c.Assert(err, IsNil)
	c.Assert(output10b.SuccessNum, Equals, int64(2))

	// 11. 并发参数
	serialDstPrefix := randLowStr(6) + "/"
	output11, err := client.CopyDirAcrossRegion(&s3.CopyDirInput{
		SourceBucket: aws.String(bucket),
		SourcePrefix: aws.String(srcPrefix),
		Bucket:       aws.String(dstBucket),
		Prefix:    aws.String(serialDstPrefix),
		Jobs:         aws.Long(1),
		Parallel:     aws.Long(1),
	}, dstClient)
	c.Assert(err, IsNil)
	c.Assert(output11.SuccessNum, Equals, int64(2))

	// 12. 自定义PartSize
	partSizeDstPrefix := randLowStr(6) + "/"
	output11b, err := client.CopyDirAcrossRegion(&s3.CopyDirInput{
		SourceBucket: aws.String(bucket),
		SourcePrefix: aws.String(srcPrefix),
		Bucket:       aws.String(dstBucket),
		Prefix:    aws.String(partSizeDstPrefix),
		PartSize:     aws.Long(1024 * 1024),
	}, dstClient)
	c.Assert(err, IsNil)
	c.Assert(output11b.SuccessNum, Equals, int64(2))

	// 13. ProgressFn回调
	var progressCalls int64
	progressDstPrefix := randLowStr(6) + "/"
	output12, err := client.CopyDirAcrossRegion(&s3.CopyDirInput{
		SourceBucket: aws.String(bucket),
		SourcePrefix: aws.String(srcPrefix),
		Bucket:       aws.String(dstBucket),
		Prefix:    aws.String(progressDstPrefix),
		ProgressFn: func(stat s3.DirResult) {
			atomic.AddInt64(&progressCalls, 1)
		},
	}, dstClient)
	c.Assert(err, IsNil)
	c.Assert(output12.SuccessNum, Equals, int64(2))
	c.Assert(atomic.LoadInt64(&progressCalls) > 0, Equals, true)

	// 14. EnableCheckpoint
	cpDir := randLowStr(8)
	os.MkdirAll(cpDir, 0755)
	cpDstPrefix := randLowStr(6) + "/"
	output13, err := client.CopyDirAcrossRegion(&s3.CopyDirInput{
		SourceBucket:     aws.String(bucket),
		SourcePrefix:     aws.String(srcPrefix),
		Bucket:           aws.String(dstBucket),
		Prefix:        aws.String(cpDstPrefix),
		EnableCheckpoint: aws.Boolean(true),
		CheckpointDir:    aws.String(cpDir),
	}, dstClient)
	c.Assert(err, IsNil)
	c.Assert(output13.SuccessNum, Equals, int64(2))

	// 15. IgnoreSuccessFiles
	ignoreDstPrefix := randLowStr(6) + "/"
	output14, err := client.CopyDirAcrossRegion(&s3.CopyDirInput{
		SourceBucket:       aws.String(bucket),
		SourcePrefix:       aws.String(srcPrefix),
		Bucket:             aws.String(dstBucket),
		Prefix:          aws.String(ignoreDstPrefix),
		IgnoreSuccessFiles: aws.Boolean(true),
	}, dstClient)
	c.Assert(err, IsNil)
	c.Assert(output14.SuccessNum, Equals, int64(2))
	c.Assert(len(output14.SuccessFiles), Equals, 0)

	// 16. 空源前缀：无匹配对象
	emptyDstPrefix := randLowStr(6) + "/"
	output15, err := client.CopyDirAcrossRegion(&s3.CopyDirInput{
		SourceBucket: aws.String(bucket),
		SourcePrefix: aws.String(srcPrefix + "nonexistent/"),
		Bucket:       aws.String(dstBucket),
		Prefix:    aws.String(emptyDstPrefix),
	}, dstClient)
	c.Assert(err, IsNil)
	c.Assert(output15.TotalNum, Equals, int64(0))

	// 17. 参数校验：SourceBucket为空
	_, err = client.CopyDirAcrossRegion(&s3.CopyDirInput{
		SourceBucket: aws.String(""),
		SourcePrefix: aws.String(srcPrefix),
		Bucket:       aws.String(dstBucket),
		Prefix:    aws.String(dstPrefix),
	}, dstClient)
	c.Assert(err, NotNil)

	// 18. 参数校验：Bucket为空
	_, err = client.CopyDirAcrossRegion(&s3.CopyDirInput{
		SourceBucket: aws.String(bucket),
		SourcePrefix: aws.String(srcPrefix),
		Bucket:       aws.String(""),
		Prefix:    aws.String(dstPrefix),
	}, dstClient)
	c.Assert(err, NotNil)

	// 清理源对象
	for _, key := range []string{key1, key2} {
		client.DeleteObject(&s3.DeleteObjectInput{Bucket: aws.String(bucket), Key: aws.String(key)})
	}
	// 清理目标对象和桶
	for _, p := range []string{dstPrefix, storageDstPrefix, metaDstPrefix, overwriteDstPrefix, serialDstPrefix, partSizeDstPrefix, progressDstPrefix, cpDstPrefix, ignoreDstPrefix, emptyDstPrefix} {
		resp, _ := dstClient.ListObjects(&s3.ListObjectsInput{
			Bucket: aws.String(dstBucket), Prefix: aws.String(p),
		})
		for _, obj := range resp.Contents {
			dstClient.DeleteObject(&s3.DeleteObjectInput{Bucket: aws.String(dstBucket), Key: obj.Key})
		}
	}
	dstClient.DeleteBucket(&s3.DeleteBucketInput{Bucket: aws.String(dstBucket)})
	os.RemoveAll(cpDir)
}
