package ks3test

import (
	"bufio"
	"bytes"
	"github.com/ks3sdklib/aws-sdk-go/aws/awserr"
	"ks3sdklib/aws-sdk-go/internal/apierr"
	"log"
	"strconv"
	"strings"
	"testing"
	"time"
	//	"io"
	"fmt"
	"github.com/ks3sdklib/aws-sdk-go/aws"
	"github.com/ks3sdklib/aws-sdk-go/aws/credentials"
	"github.com/ks3sdklib/aws-sdk-go/service/s3"
	"github.com/stretchr/testify/assert"
	"net/http"
	"os"
)

var bucket = string("yourbucket")
var key = string("yourkey")
var key_encode = string("yourkey")
var key_copy = string("yourkey")
var content = string("content")
var cre = credentials.NewStaticCredentials("ak", "sk", "") //online
var svc = s3.New(&aws.Config{
	Region:      "BEIJING",
	Credentials: cre,
	//Endpoint:"ks3-sgp.ksyun.com",
	Endpoint:         "ks3-cn-beijing.ksyun.com",
	DisableSSL:       true,
	LogLevel:         1,
	S3ForcePathStyle: true,
	LogHTTPBody:      true,
})

func TestCreateBucket(t *testing.T) {
	_, err := svc.CreateBucket(&s3.CreateBucketInput{
		ACL:    aws.String("public-read"),
		Bucket: aws.String(bucket),
	})
	assert.Error(t, err)
	assert.Equal(t, "BucketAlreadyExists", err.(*apierr.RequestError).Code())
}
func TestBucketAcl(t *testing.T) {
	_, err := svc.PutBucketACL(&s3.PutBucketACLInput{
		Bucket: aws.String(bucket),
		ACL:    aws.String("public-read"),
	})
	assert.NoError(t, err)

	acp, err := svc.GetBucketACL(&s3.GetBucketACLInput{
		Bucket: aws.String(bucket),
	})
	assert.NoError(t, err)
	grants := acp.Grants
	assert.Equal(t, 2, len(grants), "size of grants")

	foundFull := false
	foundRead := false
	for i := 0; i < len(grants); i++ {
		grant := grants[i]
		if *grant.Permission == "FULL_CONTROL" {
			foundFull = true
			assert.NotNil(t, *grant.Grantee.ID, "grantee userid should not null")
			assert.NotNil(t, *grant.Grantee.DisplayName, "grantee displayname should not null")
		} else if *grant.Permission == "READ" {
			foundRead = true
			assert.NotNil(t, *grant.Grantee.URI, "grantee uri should not null")
		}
	}
	assert.True(t, foundRead, "acp should contains READ")
	assert.True(t, foundFull, "acp should contains FULL_CONTROL")

	_, putaclErr := svc.PutBucketACL(&s3.PutBucketACLInput{
		Bucket: aws.String(bucket),
		ACL:    aws.String("private"),
	})
	assert.NoError(t, putaclErr)

	acp, getaclErr := svc.GetBucketACL(&s3.GetBucketACLInput{
		Bucket: aws.String(bucket),
	})
	assert.NoError(t, getaclErr)
	privategrants := acp.Grants
	assert.Equal(t, 1, len(privategrants), "size of grants")
}
func TestListBuckets(t *testing.T) {
	out, err := svc.ListBuckets(nil)
	assert.NoError(t, err)
	buckets := out.Buckets
	found := false
	for i := 0; i < len(buckets); i++ {
		fmt.Println(*buckets[i].Name)
		fmt.Println(*buckets[i].Region)
	}
	assert.True(t, found, "list buckets expected contains "+bucket)
}
func TestHeadBucket(t *testing.T) {
	_, err := svc.HeadBucket(&s3.HeadBucketInput{
		Bucket: aws.String(bucket),
	})
	assert.NoError(t, err)
}
func TestDeleteBucket(t *testing.T) {
	putObjectSimple()
	_, err := svc.DeleteBucket(&s3.DeleteBucketInput{
		Bucket: aws.String(bucket),
	})
	assert.Error(t, err)
}
func TestListObjects(t *testing.T) {
	//putObjectSimple()
	objects1, err := svc.ListObjects(&s3.ListObjectsInput{
		Bucket: aws.String(bucket),
	})
	assert.NoError(t, err)
	objectList := objects1.Contents
	for i := 0; i < len(objectList); i++ {
		object := objectList[i]
		println(*object.Key)
	}
}

func TestListObjectPages(t *testing.T) {
	totalNum := 0
	err := svc.ListObjectsPages(&s3.ListObjectsInput{
		Bucket:    aws.String(bucket),
		Delimiter: aws.String("/"),
		MaxKeys:   aws.Long(10),
		Prefix:    aws.String(""),
	}, func(p *s3.ListObjectsOutput, lastPage bool) (shouldContinue bool) {
		for _, obj := range p.Contents {
			fmt.Println("Object:", *obj.Key)
			totalNum++
		}
		if lastPage {
			return false
		} else {
			return true
		}
	})
	assert.NoError(t, err)
	println(totalNum)
}

func TestDelObject(t *testing.T) {
	putObjectSimple()
	assert.True(t, objectExists(bucket, key))
	_, err := svc.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	assert.NoError(t, err)
	assert.False(t, objectExists(bucket, key))
}

func TestGetObject(t *testing.T) {
	putObjectSimple()
	out, err := svc.GetObject(&s3.GetObjectInput{
		Bucket:              aws.String(bucket),
		Key:                 aws.String(key),
		ResponseContentType: aws.String("application/pdf"),
		Range:               aws.String("bytes=0-1"),
	})
	assert.NoError(t, err)
	assert.True(t, strings.HasPrefix(*out.ContentRange, "bytes 0-1/"))
	assert.Equal(t, *aws.Long(2), *out.ContentLength)
	assert.Equal(t, "application/pdf", *out.ContentType)
	br := bufio.NewReader(out.Body)
	w, _ := br.ReadString('\n')
	assert.Equal(t, content[:2], w)
}

func TestGetObjectPresignedUrl(t *testing.T) {
	//putObjectSimple();
	rl, err := svc.GetObjectPresignedUrl(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}, time.Second*time.Duration(time.Now().Add(time.Second*600).Unix()))
	assert.NoError(t, err)
	println(rl)
}

func TestObjectAcl(t *testing.T) {
	putObjectSimple()
	_, err := svc.PutObjectACL(&s3.PutObjectACLInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		ACL:    aws.String("public-read"),
	})
	assert.NoError(t, err)

	acp, err := svc.GetObjectACL(&s3.GetObjectACLInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	assert.NoError(t, err)
	grants := acp.Grants
	assert.Equal(t, 2, len(grants), "size of grants")

	foundFull := false
	foundRead := false
	for i := 0; i < len(grants); i++ {
		grant := grants[i]
		if *grant.Permission == "FULL_CONTROL" {
			foundFull = true
			assert.NotNil(t, *grant.Grantee.ID, "grantee userid should not null")
			assert.NotNil(t, *grant.Grantee.DisplayName, "grantee displayname should not null")
		} else if *grant.Permission == "READ" {
			foundRead = true
			assert.NotNil(t, *grant.Grantee.URI, "grantee uri should not null")
		}
	}
	assert.True(t, foundRead, "acp should contains READ")
	assert.True(t, foundFull, "acp should contains FULL_CONTROL")

	_, putaclErr := svc.PutObjectACL(&s3.PutObjectACLInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		ACL:    aws.String("private"),
	})
	assert.NoError(t, putaclErr)

	acp, getaclErr := svc.GetObjectACL(&s3.GetObjectACLInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	assert.NoError(t, getaclErr)
	privategrants := acp.Grants
	assert.Equal(t, 1, len(privategrants), "size of grants")
}

func TestCopyObject(t *testing.T) {
	//key_modify_meta := string("yourkey")

	metadata := make(map[string]*string)
	metadata["yourmetakey1"] = aws.String("yourmetavalue1")
	metadata["yourmetakey2"] = aws.String("yourmetavalue2")

	_, err := svc.CopyObject(&s3.CopyObjectInput{
		Bucket:            aws.String(bucket),
		Key:               aws.String(key_copy),
		CopySource:        aws.String("/" + bucket + "/" + key_copy),
		MetadataDirective: aws.String("REPLACE"),
		Metadata:          metadata,
	})
	assert.NoError(t, err)
	assert.True(t, objectExists(bucket, key))
}

func TestModifyObjectMeta(t *testing.T) {
	key_modify_meta := string("yourkey")

	metadata := make(map[string]*string)
	metadata["yourmetakey1"] = aws.String("yourmetavalue1")
	metadata["yourmetakey2"] = aws.String("yourmetavalue2")

	_, err := svc.CopyObject(&s3.CopyObjectInput{
		Bucket:            aws.String(bucket),
		Key:               aws.String(key_modify_meta),
		CopySource:        aws.String("/" + bucket + "/" + key_modify_meta),
		MetadataDirective: aws.String("REPLACE"),
		Metadata:          metadata,
	})
	assert.NoError(t, err)
	assert.True(t, objectExists(bucket, key))
}

func TestMultipartUpload(t *testing.T) {
	key = "jdexe"
	fileName := "d:/upload-test/jd.exe"
	initRet, initErr := svc.CreateMultipartUpload(&s3.CreateMultipartUploadInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		ACL:         aws.String("public-read"),
		ContentType: aws.String("application/octet-stream"),
	})
	assert.NoError(t, initErr)
	assert.Equal(t, bucket, *initRet.Bucket)
	assert.Equal(t, key, *initRet.Key)
	assert.NotNil(t, *initRet.UploadID)

	uploadId := *initRet.UploadID
	fmt.Printf("%s %s", "uploadId=", uploadId)

	f, err := os.Open(fileName)
	if err != nil {
		fmt.Println("can't opened this file")
		return
	}

	defer f.Close()
	var i int64 = 1
	compParts := []*s3.CompletedPart{}
	partsNum := []int64{0}
	s := make([]byte, 52428800)

	for {
		nr, err := f.Read(s[:])
		if nr < 0 {
			fmt.Fprintf(os.Stderr, "cat: error reading: %s\n", err.Error())
			os.Exit(1)
		} else if nr == 0 {
			break
		} else {
			upRet, upErr := svc.UploadPart(&s3.UploadPartInput{
				Bucket:        aws.String(bucket),
				Key:           aws.String(key),
				PartNumber:    aws.Long(i),
				UploadID:      aws.String(uploadId),
				Body:          bytes.NewReader(s[0:nr]),
				ContentLength: aws.Long(int64(len(s[0:nr]))),
			})
			partsNum = append(partsNum, i)
			compParts = append(compParts, &s3.CompletedPart{PartNumber: &partsNum[i], ETag: upRet.ETag})
			i++
			assert.NoError(t, upErr)
			assert.NotNil(t, upRet.ETag)
		}
	}

	compRet, compErr := svc.CompleteMultipartUpload(&s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(bucket),
		Key:      aws.String(key),
		UploadID: aws.String(uploadId),
		MultipartUpload: &s3.CompletedMultipartUpload{
			Parts: compParts,
		},
	})
	assert.NoError(t, compErr)
	assert.Equal(t, bucket, *compRet.Bucket)
	assert.Equal(t, key, *compRet.Key)
}

func TestPutObjectWithUserMeta(t *testing.T) {
	meta := make(map[string]*string)
	meta["user"] = aws.String("lijunwei")
	meta["edit-date"] = aws.String("20150623")
	_, putErr := svc.PutObject(&s3.PutObjectInput{
		Bucket:   aws.String(bucket),
		Key:      aws.String(key),
		Metadata: meta,
	})
	assert.NoError(t, putErr)

	headRet, headErr := svc.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	assert.NoError(t, headErr)

	outMeta := headRet.Metadata
	user := outMeta["User"]
	date := outMeta["Edit-Date"]
	assert.NotNil(t, user)
	assert.NotNil(t, date)
	if user != nil {
		assert.Equal(t, "lijunwei", *user)
	}
	if date != nil {
		assert.Equal(t, "20150623", *date)
	}
}
func TestPutObjectWithSSEC(t *testing.T) {
	putRet, putErr := svc.PutObject(&s3.PutObjectInput{
		Bucket:               aws.String(bucket),
		Key:                  aws.String(key),
		SSECustomerAlgorithm: aws.String("AES256"),
		SSECustomerKey:       aws.String("12345678901234567890123456789012"),
	})
	assert.NoError(t, putErr)
	assert.NotNil(t, putRet.SSECustomerAlgorithm)
	if putRet.SSECustomerAlgorithm != nil {
		assert.Equal(t, "AES256", *putRet.SSECustomerAlgorithm)
	}
	assert.NotNil(t, putRet.SSECustomerKeyMD5)
	if putRet.SSECustomerKeyMD5 != nil {
		assert.NotNil(t, *putRet.SSECustomerKeyMD5)
	}

	headRet, headErr := svc.HeadObject(&s3.HeadObjectInput{
		Bucket:               aws.String(bucket),
		Key:                  aws.String(key),
		SSECustomerAlgorithm: aws.String("AES256"),
		SSECustomerKey:       aws.String("12345678901234567890123456789012"),
	})
	assert.NoError(t, headErr)
	assert.NotNil(t, headRet.SSECustomerAlgorithm)
	if headRet.SSECustomerAlgorithm != nil {
		assert.Equal(t, "AES256", *headRet.SSECustomerAlgorithm)
	}
	assert.NotNil(t, headRet.SSECustomerKeyMD5)
	if headRet.SSECustomerKeyMD5 != nil {
		assert.NotNil(t, *headRet.SSECustomerKeyMD5)
	}
}
func TestPutObjectAclPresignedUrl(t *testing.T) {
	params := &s3.PutObjectACLInput{
		Bucket:      aws.String(bucket),    // bucket名称
		Key:         aws.String(key),       // object key
		ACL:         aws.String("private"), //设置ACL
		ContentType: aws.String("text/plain"),
	}
	resp, err := svc.PutObjectACLPresignedUrl(params, 1444370289000000000) //第二个参数为外链过期时间，第二个参数为time.Duration类型
	if err != nil {
		panic(err)
	}

	httpReq, _ := http.NewRequest("PUT", "", nil)
	httpReq.URL = resp
	httpReq.Header["x-amz-acl"] = []string{"private"}
	httpReq.Header.Add("Content-Type", "text/plain")
	httpRep, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		panic(err)
	}
	assert.Equal(t, "200 OK", httpRep.Status)
}
func TestPutObjectPresignedUrl(t *testing.T) {
	params := &s3.PutObjectInput{
		Bucket:           aws.String(bucket),                    // bucket名称
		Key:              aws.String(key),                       // object key
		ACL:              aws.String("public-read"),             //设置ACL
		ContentType:      aws.String("application/ocet-stream"), //设置文件的content-type
		ContentMaxLength: aws.Long(20),                          //设置允许的最大长度，对应的header：x-amz-content-maxlength
	}
	resp, err := svc.PutObjectPresignedUrl(params, 1444370289000000000) //第二个参数为外链过期时间，第二个参数为time.Duration类型
	if err != nil {
		panic(err)
	}

	httpReq, _ := http.NewRequest("PUT", "", strings.NewReader("123"))
	httpReq.URL = resp
	httpReq.Header["x-amz-acl"] = []string{"public-read"}
	httpReq.Header["x-amz-content-maxlength"] = []string{"20"}
	httpReq.Header.Add("Content-Type", "application/ocet-stream")
	fmt.Println(httpReq)
	httpRep, err := http.DefaultClient.Do(httpReq)
	fmt.Println(httpRep)
	if err != nil {
		panic(err)
	}
	assert.Equal(t, "200 OK", httpRep.Status)
}
func putObjectSimple() {
	svc.PutObject(
		&s3.PutObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
			Body:   strings.NewReader(content),
		},
	)
}

func objectExists(bucket, key string) bool {
	_, err := svc.HeadObject(
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
func TestBug(t *testing.T) {

	input := &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String("testkey&122"),
		//Body: bytes.NewReader([]byte("PAYLOAD")),
		ACL:          aws.String("Private"),
		CallbackUrl:  aws.String("http://ip:port"),
		CallbackBody: aws.String("objectKey=${etag}${encodedKey}&etag=${etag}&objsize=${objectSize}"),
	}

	now := time.Now()
	now = now.Add(100 * time.Second)
	url, err := svc.PutObjectPresignedUrl(input, time.Duration(now.UnixNano()))
	if err != nil {
		panic(err)
	}

	log.Println(url)

	httpReq, _ := http.NewRequest("PUT", "ks3-sgp.ksyun.com/testkey&122", strings.NewReader("123123413412341241241241241241243124123412412341241343242342134"))
	httpReq.URL = url
	httpReq.Header["x-amz-acl"] = []string{"Private"}
	httpReq.Header["x-kss-callbackurl"] = []string{"http://ip:port"}
	httpReq.Header["x-kss-callbackbody"] = []string{"objectKey=${etag}${encodedKey}&etag=${etag}&objsize=${objectSize}"}
	httpResp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		panic(err)
	}
	log.Println(httpResp)
}

/**
批量删除对象
*/
func DeleteObjects() {

	params := &s3.DeleteObjectsInput{
		Bucket: aws.String(""), // Required
		Delete: &s3.Delete{ // Required
			Objects: []*s3.ObjectIdentifier{
				{
					Key:       aws.String("1"), // Required
				},
				{
					Key:       aws.String("2"), // Required
				},
				// More values...
			},
		},
	}
	resp := svc.DeleteObjects(params)
	fmt.Println("error keys:",resp.Errors)
	fmt.Println("deleted keys:",resp.Deleted)
}

/**
删除前缀
*/
func DeleteBucketPrefix(prefix string) {

	params := &s3.DeleteBucketPrefixInput{
		Bucket: aws.String(""), // Required
		Prefix: aws.String(prefix),
	}
	resp, _ := svc.DeleteBucketPrefix(params)
	fmt.Println("error keys:",resp.Errors)
	fmt.Println("deleted keys:",resp.Deleted)
}

/**
 删除前缀(包含三次重试)
*/
func TryDeleteBucketPrefix(prefix string) {

	params := &s3.DeleteBucketPrefixInput{
		Bucket: aws.String(""),
		Prefix: aws.String(prefix),
	}
	resp, _ := svc.TryDeleteBucketPrefix(params)
	fmt.Println("error keys:",resp.Errors)
	fmt.Println("deleted keys:",resp.Deleted)
}

func PutFile() {

	filename := "/Users/cqc/Desktop/zs.java"
	// 读取本地文件。
	fd, err := os.Open(filename)
	params := &s3.PutReaderRequest{
		Bucket:      aws.String("ks3tools-test"),                // bucket名称
		Key:         aws.String("go-demo/test"),            // object key
		ACL:         aws.String("private"),             //权限，支持private(私有)，public-read(公开读)
		Body:        fd,                 //要上传的内容
		ContentType: aws.String("application/ocet-stream"), //设置content-type
		Metadata: map[string]*string{
			//"Key": aws.String("MetadataValue"), // 设置用户元数据
			// More values...
		},
	}
	resp, err := client.PutReader(params)
	if err != nil {
		panic(err)
	}
	fmt.Println(resp)

}