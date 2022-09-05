package lib

import (
	"bytes"
	"fmt"
	"github.com/ks3sdklib/aws-sdk-go/aws"
	"github.com/ks3sdklib/aws-sdk-go/aws/awserr"
	"github.com/ks3sdklib/aws-sdk-go/aws/awsutil"
	"github.com/ks3sdklib/aws-sdk-go/service/s3"
	. "gopkg.in/check.v1"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	key      = randLowStr(10)
	key_copy = randLowStr(10)
)

//列表bucket下对象
func (s *Ks3utilCommandSuite) TestListObjects(c *C) {

	objects1, _ := client.ListObjects(&s3.ListObjectsInput{
		Bucket:    aws.String(bucket),
		Delimiter: aws.String("/"),       //分隔符，用于对一组参数进行分割的字符
		MaxKeys:   aws.Long(int64(1000)), //设置响应体中返回的最大记录数（最后实际返回可能小于该值）。默认为1000。如果你想要的结果在1000条以后，你可以设定 marker 的值来调整起始位置。
		Prefix:    aws.String(""),        //限定响应结果列表使用的前缀，正如你在电脑中使用的文件夹一样。
		Marker:    aws.String(""),        //指定列举指定空间中对象的起始位置。KS3按照字母排序方式返回结果，将从给定的 marker 开始返回列表。
	})
	//获取对象列表
	objectList := objects1.Contents
	fmt.Println(objectList)
}

/**
  上传示例 -可设置标签  acl
*/
func (s *Ks3utilCommandSuite) TestPutObject(c *C) {

	//指定目标Object对象标签，可同时设置多个标签，如：TagA=A&TagB=B。
	//说明 Key和Value需要先进行URL编码，如果某项没有“=”，则看作Value为空字符串。详情请见对象标签（https://docs.ksyun.com/documents/39576）。
	v := url.Values{}
	v.Add("name", "yz")
	v.Add("age", "11")
	XAmzTagging := v.Encode()

	object := randLowStr(10)
	s.createFile(object, content, c)
	fd, _ := os.Open(content)
	input := s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(object),
		ACL:         aws.String("public-read"),
		Body:        fd,
		XAmzTagging: aws.String(XAmzTagging),
	}
	resp, _ := client.PutObject(&input)
	fmt.Println("结果：\n", awsutil.StringValue(resp))
	os.Remove(object)
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
	resp, _ := client.PutObject(&input)
	fmt.Println("结果：\n", awsutil.StringValue(resp))

	//下载
	getInput := s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}
	DownloadResp, _ := client.GetObject(&getInput)
	fmt.Println("结果：\n", awsutil.StringValue(DownloadResp))

}

//删除对象
func (s *Ks3utilCommandSuite) TestDelObject(c *C) {

	resp, _ := client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	fmt.Println("结果：\n", awsutil.StringValue(resp))
}

//生成下载地址
func (s *Ks3utilCommandSuite) TestGetObjectPresignedUrl(c *C) {

	resp, _ := client.GetObjectPresignedUrl(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}, time.Second*time.Duration(time.Now().Add(time.Second*600).Unix())) //在当前时间多久后到期
	fmt.Println("结果：\n", awsutil.StringValue(resp))
}

//获取对象Acl
func (s *Ks3utilCommandSuite) TestGetObjectAcl(c *C) {

	resp, _ := client.GetObjectACL(&s3.GetObjectACLInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	foundFull := false
	foundRead := false
	for i := 0; i < len(resp.Grants); i++ {
		grant := resp.Grants[i]
		if *grant.Permission == "FULL_CONTROL" {
			foundFull = true
		} else if *grant.Permission == "READ" {
			foundRead = true
		}
	}
	fmt.Println(foundFull, foundRead)
	fmt.Println("结果：\n", awsutil.StringValue(resp))

}

//设置对象Acl
func (s *Ks3utilCommandSuite) TestPutObjectAcl(c *C) {

	resp, _ := client.PutObjectACL(&s3.PutObjectACLInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		ACL:    aws.String("private"),
	})
	fmt.Println("结果：\n", awsutil.StringValue(resp))

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

	resp, _ := client.CopyObject(&s3.CopyObjectInput{
		Bucket:               aws.String(bucket),
		Key:                  aws.String(key),
		CopySource:           aws.String("/" + bucket + "/" + key_copy),
		MetadataDirective:    aws.String("REPLACE"),
		Metadata:             metadata,
		XAmzTagging:          aws.String(XAmzTagging),
		XAmzTaggingDirective: aws.String("REPLACE"),
	})
	fmt.Println("结果：\n", awsutil.StringValue(resp))
}

//抓取第三方URL上传到KS3
func (s *Ks3utilCommandSuite) TestFetchObj(c *C) {

	//指定目标Object对象标签，可同时设置多个标签，如：TagA=A&TagB=B。
	//说明 Key和Value需要先进行URL编码，如果某项没有“=”，则看作Value为空字符串。详情请见对象标签（https://docs.ksyun.com/documents/39576）。
	v := url.Values{}
	v.Add("schoole", "bbvvvvvv")
	v.Add("class", "123123123123")
	XAmzTagging := v.Encode()

	//源站url
	sourceUrl := "https://img0.pconline.com.cn/pconline/1111/04/2483449_20061139501.jpg"
	input := s3.FetchObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String("dst/testa"),
		SourceUrl:   aws.String(sourceUrl),
		XAmzTagging: aws.String(XAmzTagging),   //对象tag
		ACL:         aws.String("public-read"), //对象acl
	}
	resp, _ := client.FetchObject(&input)
	fmt.Println("结果：\n", awsutil.StringValue(resp))
}

//修改元数据信息
func (s *Ks3utilCommandSuite) TestModifyObjectMeta(c *C) {
	key_modify_meta := string("yourkey")

	metadata := make(map[string]*string)
	metadata["yourmetakey1"] = aws.String("yourmetavalue1")
	metadata["yourmetakey2"] = aws.String("yourmetavalue2")

	resp, _ := client.CopyObject(&s3.CopyObjectInput{
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
	fmt.Println("结果：\n", awsutil.StringValue(resp))
}

//分块上传
//此操作将启动一个分块上传任务并返回 upload ID。在一个确定的分块上传任务中，upload ID用于关联所有分块。连续分块上传请求中的 upload ID由用户指定。在Complete Multipart Upload 和 Abort Multipart Upload请求中同样包含 upload ID。
//关于请求签名的问题，分块上传为一系列的请求（初始化分块上传，上传块，完成分块上传，终止分块上传），用户启动任务，发送一个或多个分块，最终完成任务。用户需要对每一个请求单独签名。
//
//注意: 当你启动分块上传后，并开始上传分块，你必须完成或者放弃上传任务，才能终止因为存储造成的收费。
func (s *Ks3utilCommandSuite) TestMultipartUpload(c *C) {

	key = "jdexe"
	fileName := "d:/upload-test/jd.exe"
	initRet, _ := client.CreateMultipartUpload(&s3.CreateMultipartUploadInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		ACL:         aws.String("public-read"),
		ContentType: aws.String("application/octet-stream"),
	})
	//获取分块Id
	uploadId := *initRet.UploadID
	fmt.Printf("%s %s", "uploadId=", uploadId)

	f, err := os.Open(fileName)
	if err != nil {
		fmt.Println("can't opened this file")
		return
	}

	defer f.Close()
	var i int64 = 1
	//组装分块参数
	compParts := []*s3.CompletedPart{}
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
			resp, _ := client.UploadPart(&s3.UploadPartInput{
				Bucket:        aws.String(bucket),
				Key:           aws.String(key),
				PartNumber:    aws.Long(i),
				UploadID:      aws.String(uploadId),
				Body:          bytes.NewReader(sc[0:nr]),
				ContentLength: aws.Long(int64(len(sc[0:nr]))),
			})
			partsNum = append(partsNum, i)
			compParts = append(compParts, &s3.CompletedPart{PartNumber: &partsNum[i], ETag: resp.ETag})
			i++
			fmt.Println("结果：\n", awsutil.StringValue(resp))
		}
	}

	//此操作将完成对象装配之前的块上传任务 。
	//
	//用户启动一个分块上传任务后，会使用 Upload Parts 接口上传所有的块。成功上传所有相关块之后，用户需要调用此接口来完成分块上传。收到完成请求后，KS3将会根据块序号将所有的块组装起来创建一个新的对象。在用户的完成任务请求中需要用户提供分块列表，由于KS3将会按照列表将所有块连接起来，所以要求用户保证所有的块已经完成上传。对于分块列表中的每一个块，用户需要在上传块时添加块序号以及对象的 ETag 头部，KS3则会在块完成上传后回复完成响应。
	//
	//请注意，如果 Complete Multipart Upload 请求失败了，用户应用应当能够进行重试操作。
	compRet, _ := client.CompleteMultipartUpload(&s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(bucket),
		Key:      aws.String(key),
		UploadID: aws.String(uploadId),
		MultipartUpload: &s3.CompletedMultipartUpload{
			Parts: compParts,
		},
	})
	fmt.Println("结果：\n", awsutil.StringValue(compRet))

}

//上传加密
//服务器端加密关乎静态数据加密，即 KS3 在将您的数据写入数据中心内的磁盘时会在对象级别上加密这些数据，并在您访问这些数据时为您解密这些数据。
//只要您验证了您的请求并且拥有访问权限，您访问加密和未加密数据元的方式就没有区别。
//例如，如果您使用预签名的 URL 来共享您的对象，那么对于加密和解密对象，该 URL 的工作方式是相同的。
func (s *Ks3utilCommandSuite) TestPutObjectWithSSEC(c *C) {
	resp, _ := client.PutObject(&s3.PutObjectInput{
		Bucket:               aws.String(bucket),
		Key:                  aws.String(key),
		SSECustomerAlgorithm: aws.String("AES256"), //加密类型
		SSECustomerKey:       aws.String("12345678901234567890123456789012"),
	})
	fmt.Println("结果：\n", awsutil.StringValue(resp))
}

//生成上传外链
func (s *Ks3utilCommandSuite) TestPutObjectPresignedUrl(c *C) {
	params := &s3.PutObjectInput{
		Bucket:           aws.String(bucket),                    // bucket名称
		Key:              aws.String(key),                       // object key
		ACL:              aws.String("public-read"),             //设置ACL
		ContentType:      aws.String("application/ocet-stream"), //设置文件的content-type
		ContentMaxLength: aws.Long(20),                          //设置允许的最大长度，对应的header：x-amz-content-maxlength
	}
	resp, _ := client.PutObjectPresignedUrl(params, 1444370289000000000) //第二个参数为外链过期时间，第二个参数为time.Duration类型
	fmt.Println("结果：\n", awsutil.StringValue(resp))

	//简单上传示例
	date := time.Now().UTC().Format(http.TimeFormat)
	httpReq, _ := http.NewRequest("PUT", "", strings.NewReader("123"))
	httpReq.URL = resp
	httpReq.Header["x-amz-acl"] = []string{"public-read"}
	httpReq.Header["x-amz-content-maxlength"] = []string{"20"}
	httpReq.Header.Add("Content-Type", "application/ocet-stream")
	httpReq.Header["Date"] = []string{date}
	upLoadResp, _ := http.DefaultClient.Do(httpReq)
	fmt.Println("结果：\n", awsutil.StringValue(upLoadResp))
}

//判断文件是否存在
func objectExists(bucket, key string) bool {
	_, err := client.HeadObject(
		&s3.HeadObjectInput{
			Bucket: &bucket,
			Key:    &key,
		},
	)
	if err != nil {
		aerr, ok := err.(awserr.Error)
		if ok && (aerr.Code() == strconv.Itoa(403) || aerr.Code() == strconv.Itoa(404)) {
			// Specific error code handling
			return false
		}
	}
	return true
}

/**
批量删除对象
*/
func (s *Ks3utilCommandSuite) DeleteObjects() {

	params := &s3.DeleteObjectsInput{
		Bucket: aws.String(""), // Required
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
	resp := client.DeleteObjects(params)
	fmt.Println("error keys:", resp.Errors)
	fmt.Println("deleted keys:", resp.Deleted)
}

/**
删除前缀
*/
func (s *Ks3utilCommandSuite) DeleteBucketPrefix(prefix string) {

	params := &s3.DeleteBucketPrefixInput{
		Bucket: aws.String(""), // Required
		Prefix: aws.String(prefix),
	}
	resp, _ := client.DeleteBucketPrefix(params)
	fmt.Println("error keys:", resp.Errors)
	fmt.Println("deleted keys:", resp.Deleted)
}

/**
删除前缀(包含三次重试)
*/
func (s *Ks3utilCommandSuite) TryDeleteBucketPrefix(prefix string) {

	params := &s3.DeleteBucketPrefixInput{
		Bucket: aws.String(""),
		Prefix: aws.String(prefix),
	}
	resp, _ := client.TryDeleteBucketPrefix(params)
	fmt.Println("error keys:", resp.Errors)
	fmt.Println("deleted keys:", resp.Deleted)
}

//文件解冻
func RestoreObject() {

	params := &s3.RestoreObjectInput{
		Bucket: aws.String("ks3tools-test"),    // bucket名称
		Key:    aws.String("/restore/big.txt"), // object key
	}
	resp, err := client.RestoreObject(params)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(resp.HttpCode)
	fmt.Println(resp.Message)
}

//delObjectTagging
func (s *Ks3utilCommandSuite) DelTag(c *C) {

	params := &s3.DeleteObjectTaggingInput{
		Bucket: aws.String(bucket), // Required
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

//getObjectTagging
func (s *Ks3utilCommandSuite) GetTag(c *C) {

	params := &s3.GetObjectTaggingInput{
		Bucket: aws.String(bucket), // Required
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

//设置对象Tag
func (s *Ks3utilCommandSuite) PutTag(c *C) {

	//指定目标Object对象标签
	objTagging := s3.Tagging{
		TagSet: []*s3.Tag{&s3.Tag{
			Key:   aws.String("name"),
			Value: aws.String("yz"),
		}, &s3.Tag{
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
	resp, _ := client.PutObjectTagging(params)
	fmt.Println("结果：\n", awsutil.StringValue(resp))
}
