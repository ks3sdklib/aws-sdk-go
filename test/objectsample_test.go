package lib

import (
	"bytes"
	"fmt"
	"github.com/ks3sdklib/aws-sdk-go/aws"
	"github.com/ks3sdklib/aws-sdk-go/aws/awserr"
	"github.com/ks3sdklib/aws-sdk-go/aws/awsutil"
	"github.com/ks3sdklib/aws-sdk-go/service/s3"
	"github.com/ks3sdklib/aws-sdk-go/service/s3/s3manager"
	. "gopkg.in/check.v1"
	"net/url"
	"os"
	"time"
)

var (
	randKey  = randLowStr(10)
	key      = "123.txt"
	key_copy = randLowStr(10)
)

//列举bucket下对象
func (s *Ks3utilCommandSuite) TestListObjects(c *C) {

	resp, err := client.ListObjects(&s3.ListObjectsInput{
		Bucket: aws.String(bucket),
		//Delimiter: aws.String("/"),       //分隔符，用于对一组参数进行分割的字符
		MaxKeys: aws.Long(int64(1000)), //设置响应体中返回的最大记录数（最后实际返回可能小于该值）。默认为1000。如果你想要的result在1000条以后，你可以设定 marker 的值来调整起始位置。
		Prefix:  aws.String(""),        //限定响应result列表使用的前缀，正如你在电脑中使用的文件夹一样。
		Marker:  aws.String(""),        //指定列举指定空间中对象的起始位置。KS3按照字母排序方式返回result，将从给定的 marker 开始返回列表。
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("result：\n", awsutil.StringValue(resp))
}

/**
上传示例 -可设置标签  acl
*/
func (s *Ks3utilCommandSuite) TestPutObject(c *C) {

	//指定目标Object对象标签，可同时设置多个标签，如：TagA=A&TagB=B。
	//说明 Key和Value需要先进行URL编码，如果某项没有“=”，则看作Value为空字符串。详情请见对象标签（https://docs.ksyun.com/documents/39576）。
	//v := url.Values{}
	//v.Add("name", "yz")
	//v.Add("age", "11")
	//XAmzTagging := v.Encode()

	object := randLowStr(10)
	createFile(object, 1024*1024*1)
	fd, _ := os.Open(object)
	//md5, _ := utilfile.GetFileMD5(content)
	input := s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(object),
		ACL:         aws.String("private"),
		Body:        fd,
		ContentType: aws.String("application/octet-stream"),
		//XAmzTagging: aws.String(XAmzTagging),
		//ContentMD5:  aws.String(md5),
	}

	resp, err := client.PutObject(&input)
	if err != nil {
		panic(err)
	}
	fmt.Println("result：\n", awsutil.StringValue(resp))
	os.Remove(object)

}

/**
上传示例 -限速
*/
func (s *Ks3utilCommandSuite) TestPutObjectByLimit(c *C) {

	MIN_BANDWIDTH := 1024 * 100 * 8  // 100KB/s
	createFile(content, 1024*1024*1) // 1MB大小的文件
	object := randLowStr(10)
	fd, _ := os.Open(content)
	input := s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
		Body:   fd,
		//设置上传速度
		TrafficLimit: aws.Long(int64(MIN_BANDWIDTH)),
	}
	// 记录开始时间
	startTime := time.Now()
	resp, err := client.PutObject(&input)
	if err != nil {
		panic(err)
	}
	// 计算上传耗时
	elapsed := time.Since(startTime)

	fmt.Println("Upload completed successfully.")
	fmt.Println("Elapsed time:", elapsed)
	fmt.Println("result：\n", awsutil.StringValue(resp))
}

/**
下载限速示例
*/
func (s *Ks3utilCommandSuite) TestGetObjectByLimit(c *C) {

	MIN_BANDWIDTH := 1024 * 100 * 8 // 100KB/s
	fd, _ := os.Open(content)
	input := s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		ACL:    aws.String("public-read"),
		Body:   fd,
	}
	resp, err := client.PutObject(&input)
	if err != nil {
		panic(err)
	}
	fmt.Println("result：\n", awsutil.StringValue(resp))

	//下载
	getInput := s3.GetObjectInput{
		Bucket:       aws.String(bucket),
		Key:          aws.String(key),
		TrafficLimit: aws.Long(int64(MIN_BANDWIDTH)),
	}
	DownloadResp, err := client.GetObject(&getInput)
	if err != nil {
		panic(err)
	}
	fmt.Println("result：\n", awsutil.StringValue(DownloadResp))

}

/**
下载示例
*/
func (s *Ks3utilCommandSuite) TestGetObject(c *C) {

	fd, _ := os.Open(content)
	input := s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		ACL:    aws.String("public-read"),
		Body:   fd,
	}
	resp, err := client.PutObject(&input)
	if err != nil {
		panic(err)
	}
	fmt.Println("result：\n", awsutil.StringValue(resp))

	//下载
	getInput := s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}
	DownloadResp, err := client.GetObject(&getInput)
	if err != nil {
		panic(err)
	}
	fmt.Println("result：\n", awsutil.StringValue(DownloadResp))

}

//删除对象
func (s *Ks3utilCommandSuite) TestDelObject(c *C) {

	resp, err := client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("result：\n", awsutil.StringValue(resp))
}

//根据方法生成外链
func (s *Ks3utilCommandSuite) TestGeneratePresignedUrl(c *C) {

	params := &s3.GeneratePresignedUrlInput{
		Bucket: aws.String(bucket), // 设置 bucket 名称
		Key:    aws.String(key),    // 设置 object key
		//TrafficLimit: aws.Long(1000),            // 设置速度限制
		//ContentType:  aws.String("image/jpeg"),  //如果是PUT方法，需要设置content-type
		Expires:    3600,   // 过期时间
		HTTPMethod: s3.GET, //可选值有 PUT, GET, DELETE, HEAD
	}
	url, err := client.GeneratePresignedUrl(params)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	fmt.Println("Result:\n", url)
}

//获取对象Acl
func (s *Ks3utilCommandSuite) TestGetObjectAcl(c *C) {

	resp, err := client.GetObjectACL(&s3.GetObjectACLInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("result：\n", awsutil.StringValue(resp))
	fmt.Println("acl type：\n", s3.GetAcl(*resp))
}

//设置对象Acl
func (s *Ks3utilCommandSuite) TestPutObjectAcl(c *C) {

	resp, err := client.PutObjectACL(&s3.PutObjectACLInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		ACL:    aws.String("public-read"),
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("result：\n", awsutil.StringValue(resp))

}

//复制对象
func (s *Ks3utilCommandSuite) TestCopyObject(c *C) {

	//设置对象Tag
	v := url.Values{}
	v.Add("school", "yz")
	v.Add("class", "11")
	XAmzTagging := v.Encode()

	//设置对象元素头
	metadata := make(map[string]*string)
	metadata["yourmetakey1"] = aws.String("yourmetavalue1")
	metadata["yourmetakey2"] = aws.String("yourmetavalue2")

	resp, err := client.CopyObject(&s3.CopyObjectInput{
		Bucket:               aws.String(bucket),
		Key:                  aws.String("copy_" + key),
		CopySource:           aws.String("/" + bucket + "/" + key),
		MetadataDirective:    aws.String("REPLACE"),
		Metadata:             metadata,
		XAmzTagging:          aws.String(XAmzTagging),
		XAmzTaggingDirective: aws.String("REPLACE"),
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("result：\n", awsutil.StringValue(resp))
}

//分块拷贝用例
func (s *Ks3utilCommandSuite) TestUploadPartCopy(c *C) {

	dstKey := "xxx/copy/" + key
	//初始化分块
	initResp, err := client.CreateMultipartUpload(&s3.CreateMultipartUploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(dstKey),
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("result：\n", awsutil.StringValue(initResp))

	uploadPartCopyresp, err := client.UploadPartCopy(&s3.UploadPartCopyInput{
		Bucket:          aws.String(bucket),
		Key:             aws.String(dstKey),
		CopySource:      aws.String("/" + bucket + "/" + key),
		UploadID:        initResp.UploadID,
		PartNumber:      aws.Long(1),
		CopySourceRange: aws.String("bytes=0-1024"),
	})
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	fmt.Println("result：\n", awsutil.StringValue(uploadPartCopyresp))

	//合并分块
	completeMultipartResp, err := client.CompleteMultipartUpload(&s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(bucket),
		Key:      aws.String(dstKey),
		UploadID: initResp.UploadID,
		MultipartUpload: &s3.CompletedMultipartUpload{
			Parts: []*s3.CompletedPart{
				{
					PartNumber: aws.Long(1),
					ETag:       uploadPartCopyresp.CopyPartResult.ETag,
				},
			},
		},
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("result：\n", awsutil.StringValue(completeMultipartResp))
}

//抓取第三方URL上传到KS3
func (s *Ks3utilCommandSuite) TestFetchObject(c *C) {
	// 填写源站对象的url
	sourceUrl := "https://img0.pconline.com.cn/pconline/1111/04/2483449_20061139501.jpg"
	// 对源站对象url进行编码
	encodedUrl := url.QueryEscape(sourceUrl)
	// 通过第三方URL拉取文件上传
	resp, err := client.FetchObject(&s3.FetchObjectInput{
		Bucket:    aws.String(bucket),        // 存储空间名称，必填
		Key:       aws.String(key),           // 对象的key，必填
		SourceUrl: aws.String(encodedUrl),    // 编码后的源站url，必填
		ACL:       aws.String("public-read"), // 对象访问权限，非必填
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("结果：\n", awsutil.StringValue(resp))
}

//修改元数据信息
func (s *Ks3utilCommandSuite) TestModifyObjectMeta(c *C) {
	key_modify_meta := string("yourkey")

	metadata := make(map[string]*string)
	metadata["yourmetakey1"] = aws.String("yourmetavalue1")
	metadata["yourmetakey2"] = aws.String("yourmetavalue2")

	resp, err := client.CopyObject(&s3.CopyObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key_modify_meta),
		////空间名称与对象的object key名称的组合，通过斜杠分隔(’/’)。
		CopySource: aws.String("/" + bucket + "/" + key_modify_meta),
		//指定如何设置目标Object的对象标签。
		//默认值：COPY
		//有效值：
		//1. COPY（默认值）：复制源Object的对象标签到目标 Object。
		//2. REPLACE：忽略源Object的对象标签，直接采用请求中指定的对象标签。
		MetadataDirective: aws.String("REPLACE"),
		Metadata:          metadata,
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("result：\n", awsutil.StringValue(resp))
}

//分块上传
//此操作将启动一个分块上传任务并返回 upload ID。在一个确定的分块上传任务中，upload ID用于关联所有分块。连续分块上传请求中的 upload ID由用户指定。在Complete Multipart Upload 和 Abort Multipart Upload请求中同样包含 upload ID。
//关于请求签名的问题，分块上传为一系列的请求（初始化分块上传，上传块，完成分块上传，终止分块上传），用户启动任务，发送一个或多个分块，最终完成任务。用户需要对每一个请求单独签名。
//
//注意: 当你启动分块上传后，并开始上传分块，你必须完成或者放弃上传任务，才能终止因为存储造成的收费。
func (s *Ks3utilCommandSuite) TestMultipartUpload(c *C) {

	//MIN_BANDWIDTH := 1024 * 100 * 8 //100K bits/s
	createFile(content, 1024*1024*10)
	initRet, err := client.CreateMultipartUpload(&s3.CreateMultipartUploadInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		ACL:         aws.String("public-read"),
		ContentType: aws.String("application/octet-stream"),
	})
	if err != nil {
		panic(err)
	}
	//获取分块Id
	uploadId := *initRet.UploadID
	fmt.Printf("%s %s", "uploadId=", uploadId)

	f, err := os.Open(content)
	if err != nil {
		fmt.Println("can't opened this file")
		return
	}

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
			//
			//在你上传任一块之前你必须先要启动一个分块上传任务。在你发送一个启动请求后，KS3会给你一个唯一的 upload ID。每次上传块时，都需要将上传ID包含在请求中。
			//
			//块的数量可以是1到10,000中的任意一个（包含1和10,000）。块序号用于标识一个块以及其在对象创建时的位置。如果你上传一个新的块，使用之前已经使用的序列号，那么之前的那个块将会被覆盖。当所有块总大小大于5M时，除了最后一个块没有大小限制外，其余的块的大小均要求在5MB以上。当所有块总大小小于5M时，除了最后一个块没有大小限制外，其余的块的大小均要求在100K以上。如果不符合上述要求，会返回413状态码。
			//
			//为了保证数据在传输过程中没有损坏，请使用 Content-MD5 头部。当使用此头部时，KS3会自动计算出MD5，并根据用户提供的MD5进行校验，如果不匹配，将会返回错误信息。
			//计算sc[:nr]的md5值
			//md5Ctx := md5.New()
			//reader := bytes.NewReader(sc[0:nr])
			//io.Copy(md5Ctx, reader)
			//cipherStr := md5Ctx.Sum(nil)
			//md5Str := hex.EncodeToString(cipherStr)
			//fmt.Println(md5Str)
			resp, err := client.UploadPart(&s3.UploadPartInput{
				Bucket:        aws.String(bucket),
				Key:           aws.String(key),
				PartNumber:    aws.Long(i),
				UploadID:      aws.String(uploadId),
				Body:          bytes.NewReader(sc[0:nr]),
				ContentLength: aws.Long(int64(len(sc[0:nr]))),
				//TrafficLimit:  aws.Long(int64(MIN_BANDWIDTH)),
				//ContentMD5: aws.String(md5Str),
			})
			if err != nil {
				panic(err)
			}
			partsNum = append(partsNum, i)
			compParts = append(compParts, &s3.CompletedPart{PartNumber: &partsNum[i], ETag: resp.ETag})
			i++
			fmt.Println("result：\n", awsutil.StringValue(resp))
		}
	}

	//此操作将完成对象装配之前的块上传任务 。
	//
	//用户启动一个分块上传任务后，会使用 Upload Parts 接口上传所有的块。成功上传所有相关块之后，用户需要调用此接口来完成分块上传。收到完成请求后，KS3将会根据块序号将所有的块组装起来创建一个新的对象。在用户的完成任务请求中需要用户提供分块列表，由于KS3将会按照列表将所有块连接起来，所以要求用户保证所有的块已经完成上传。对于分块列表中的每一个块，用户需要在上传块时添加块序号以及对象的 ETag 头部，KS3则会在块完成上传后回复完成响应。
	//
	//请注意，如果 Complete Multipart Upload 请求失败了，用户应用应当能够进行重试操作。
	compRet, err := client.CompleteMultipartUpload(&s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(bucket),
		Key:      aws.String(key),
		UploadID: aws.String(uploadId),
		MultipartUpload: &s3.CompletedMultipartUpload{
			Parts: compParts,
		},
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("result：\n", awsutil.StringValue(compRet))

}

//上传加密
//服务器端加密关乎静态数据加密，即 KS3 在将您的数据写入数据中心内的磁盘时会在对象级别上加密这些数据，并在您访问这些数据时为您解密这些数据。
//只要您验证了您的请求并且拥有访问权限，您访问加密和未加密数据元的方式就没有区别。
//例如，如果您使用预签名的 URL 来共享您的对象，那么对于加密和解密对象，该 URL 的工作方式是相同的。
func (s *Ks3utilCommandSuite) TestPutObjectWithSSEC(c *C) {
	resp, err := client.PutObject(&s3.PutObjectInput{
		Bucket:               aws.String(bucket),
		Key:                  aws.String(key),
		SSECustomerAlgorithm: aws.String("AES256"),                                       //加密类型
		SSECustomerKey:       aws.String("WKEOUdGRnEK7HISNCOjMjM0NTY3ODlBQkNER5AldAUY="), // 客户端提供的加密密钥
		SSECustomerKeyMD5:    aws.String("mrtoxMuTClsH3tw78pFoIA=="),                     // 客户端提供的通过BASE64编码的通过128位MD5加密的密钥的MD5值
	})
	if err != nil {
		panic(err)
	}
	fmt.Println("result：\n", awsutil.StringValue(resp))
}

//判断文件是否存在
func (s *Ks3utilCommandSuite) TestHeaObject(c *C) {

	etag := s.uploadTmpFile(c)
	params := &s3.HeadObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		IfNoneMatch: aws.String(etag),
	}
	resp, err := client.HeadObject(params)
	//判断err的状态码是否为304
	if awsErr, ok := err.(awserr.RequestFailure); ok {
		if awsErr.StatusCode() == 304 {
			fmt.Println("文件未修改")
		} else if awsErr.StatusCode() == 404 {
			fmt.Println("文件不存在")
		}
	}
	fmt.Println("result：\n", awsutil.StringValue(resp))
}

/**
批量删除对象
*/
func (s *Ks3utilCommandSuite) TestDeleteObjects(c *C) {

	params := &s3.DeleteObjectsInput{
		Bucket:          aws.String(bucket), // Required
		IsReTurnResults: aws.Boolean(true),
		Delete: &s3.Delete{ // Required
			Objects: []*s3.ObjectIdentifier{
				{
					Key: aws.String("1"), // Required
				},
				{
					Key: aws.String("2"), // Required
				},
				// More values...
			},
		},
	}
	resp, err := client.DeleteObjects(params)
	if err != nil {
		panic(err)
	}
	fmt.Println("error keys:", awsutil.StringValue(resp.Errors))
	fmt.Println("deleted keys:", awsutil.StringValue(resp.Deleted))
}

/**
删除前缀
*/
func (s *Ks3utilCommandSuite) TestDeleteBucketPrefix(c *C) {

	params := &s3.DeleteBucketPrefixInput{
		Bucket:          aws.String(bucket), // Required
		IsReTurnResults: aws.Boolean(true),
		Prefix:          aws.String("123/"),
	}
	resp, err := client.DeleteBucketPrefix(params)
	if err != nil {
		panic(err)
	}
	fmt.Println("error keys:", awsutil.StringValue(resp.Errors))
	fmt.Println("deleted keys:", awsutil.StringValue(resp.Deleted))
}

/**
删除前缀(包含三次重试)
*/
func (s *Ks3utilCommandSuite) TestTryDeleteBucketPrefix(c *C) {

	params := &s3.DeleteBucketPrefixInput{
		Bucket:          aws.String(bucket),
		IsReTurnResults: aws.Boolean(true),
		Prefix:          aws.String("123/"),
	}
	resp, err := client.TryDeleteBucketPrefix(params)
	if err != nil {
		panic(err)
	}
	fmt.Println("error keys:", awsutil.StringValue(resp.Errors))
	fmt.Println("deleted keys:", awsutil.StringValue(resp.Deleted))
}

//文件解冻
func (s *Ks3utilCommandSuite) TestRestoreObject(c *C) {

	params := &s3.RestoreObjectInput{
		Bucket: aws.String(bucket), // bucket名称
		Key:    aws.String(key),    // object key
	}
	resp, err := client.RestoreObject(params)
	if err != nil {
		panic(err)
	}
	fmt.Println("result：\n", awsutil.StringValue(resp))
}

//delObjectTagging
func (s *Ks3utilCommandSuite) TestDeleteObjectTagging(c *C) {

	params := &s3.DeleteObjectTaggingInput{
		Bucket: aws.String(bucket), // Required
		Key:    aws.String(key),
	}
	resp, err := client.DeleteObjectTagging(params)
	if err != nil {
		panic(err)
	}
	fmt.Println("result：\n", awsutil.StringValue(resp))
}

//getObjectTagging
func (s *Ks3utilCommandSuite) TestGetObjectTagging(c *C) {

	params := &s3.GetObjectTaggingInput{
		Bucket: aws.String(bucket), // Required
		Key:    aws.String(key),
	}
	resp, err := client.GetObjectTagging(params)
	if err != nil {
		panic(err)
	}
	fmt.Println("result：\n", awsutil.StringValue(resp))
}

//设置对象Tag
func (s *Ks3utilCommandSuite) TestPutObjectTagging(c *C) {

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
	params := &s3.PutObjectTaggingInput{
		Bucket:  aws.String(bucket), // Required
		Key:     aws.String(key),
		Tagging: &objTagging,
	}
	resp, err := client.PutObjectTagging(params)
	if err != nil {
		panic(err)
	}
	fmt.Println("result：\n", awsutil.StringValue(resp))
}

//上传文件夹
func (s *Ks3utilCommandSuite) TestBatchUploadWithClient(c *C) {

	dir := "/Users/test/data/未命名文件夹"
	uploader := s3manager.NewUploader(&s3manager.UploadOptions{
		//分块大小 5MB
		PartSize: 0,
		//单文件内部操作的并发任务数
		Parallel: 2,
		//多文件操作时的并发任务数
		Jobs:            10,
		S3:              client,
		UploadHidden:    true,
		SkipAlreadyFile: true,
	})
	//dir 要上传的目录
	//bucket 上传的目标桶
	//prefix 桶下的路径
	uploader.UploadDir(dir, bucket, "sns/")

}

func (s *Ks3utilCommandSuite) uploadTmpFile(c *C) (etag string) {
	v := url.Values{}
	v.Add("name", "yz")
	v.Add("age", "11")
	XAmzTagging := v.Encode()

	object := randLowStr(10)
	s.createFile(object, content, c)
	fd, _ := os.Open(content)
	input := s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		ACL:         aws.String("public-read"),
		Body:        fd,
		XAmzTagging: aws.String(XAmzTagging),
	}
	resp, _ := client.PutObject(&input)
	fmt.Println("result：\n", awsutil.StringValue(resp))
	os.Remove(object)
	return *resp.ETag
}
func createFile(filePath string, size int64) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Set the file size
	err = file.Truncate(size)
	if err != nil {
		return err
	}

	fmt.Printf("File created: %s (size: %d bytes)\n", filePath, size)
	return nil
}

func (s *Ks3utilCommandSuite) TestPG(c *C) {

	for i := 0; i < 2; i++ {
		object := randLowStr(10)
		createFile(object, 1024*1024*1)
		fd, _ := os.Open(object)
		input := s3.PutObjectInput{
			Bucket:      aws.String(bucket),
			Key:         aws.String(object),
			Body:        fd,
			ContentType: aws.String("application/octet-stream"),
		}

		resp, _ := client.PutObject(&input)
		fmt.Println("result：\n", awsutil.StringValue(resp))
		os.Remove(object)

		params := &s3.GeneratePresignedUrlInput{
			Bucket:       aws.String(bucket), // 设置 bucket 名称
			Key:          aws.String(object), // 设置 object key
			TrafficLimit: aws.Long(1000),     // 设置速度限制
			// ContentType:  aws.String("image/jpeg"),  //如果是PUT方法，需要设置content-type
			Expires:    3600,    //多久后过期
			HTTPMethod: s3.HEAD, //可选值有 PUT, GET, DELETE, HEAD
		}
		url, err := client.GeneratePresignedUrl(params)
		if err != nil {
			panic(err)
		}
		fmt.Println("Result:\n", url)
	}

}
