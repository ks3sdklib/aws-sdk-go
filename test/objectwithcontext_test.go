package lib

import (
	"bytes"
	"context"
	"fmt"
	"github.com/ks3sdklib/aws-sdk-go/aws"
	"github.com/ks3sdklib/aws-sdk-go/aws/awsutil"
	"github.com/ks3sdklib/aws-sdk-go/service/s3"
	"github.com/ks3sdklib/aws-sdk-go/service/s3/s3manager"
	. "gopkg.in/check.v1"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

// PUT Object
func (s *Ks3utilCommandSuite) TestPutObjectWithContext(c *C) {
	object := randLowStr(10)
	createFile(object, 1024*1024*10)
	fd, _ := os.Open(object)
	// 上传对象，不通过context取消
	_, err := client.PutObjectWithContext(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
		Body:   fd,
	})
	c.Assert(err, IsNil)
	// head
	resp, err := client.HeadObjectWithContext(context.Background(), &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(*resp.StatusCode, Equals, int64(200))
	// delete
	_, err = client.DeleteObjectWithContext(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(err, IsNil)
	os.Remove(object)
	object = randLowStr(10)
	createFile(object, 1024*1024*10)
	fd, _ = os.Open(object)
	// 上传对象，通过context取消
	ctx, cancelFunc := context.WithTimeout(context.Background(), timeout)
	defer cancelFunc()
	_, err = client.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
		Body:   fd,
	})
	c.Assert(err, NotNil)
	// head
	resp, err = client.HeadObjectWithContext(context.Background(), &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(*resp.StatusCode, Equals, int64(404))
	os.Remove(object)
}

// GET Object
func (s *Ks3utilCommandSuite) TestGetObjectWithContext(c *C) {
	object := randLowStr(10)
	createFile(object, 1024*1024*10)
	fd, _ := os.Open(object)
	// put
	_, err := client.PutObjectWithContext(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
		Body:   fd,
	})
	c.Assert(err, IsNil)
	// head
	_, err = client.GetObjectWithContext(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(err, IsNil)
	os.Remove(object)
	tempFilePath := object + ".temp"
	// 下载文件，不通过context取消
	resp, err := client.GetObjectWithContext(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(err, IsNil)
	fd, err = os.OpenFile(tempFilePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.FileMode(0664))
	c.Assert(err, IsNil)
	_, err = io.Copy(fd, resp.Body)
	fd.Close()
	c.Assert(err, IsNil)
	os.Rename(tempFilePath, object)
	os.Remove(object)
	// 下载文件，通过context取消
	ctx, cancelFunc := context.WithTimeout(context.Background(), timeout)
	defer cancelFunc()
	resp, err = client.GetObjectWithContext(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(err, NotNil)
	os.Rename(tempFilePath, object)
	os.Remove(object)
	// delete
	_, err = client.DeleteObjectWithContext(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(err, IsNil)
}

// HEAD Object
func (s *Ks3utilCommandSuite) TestHeadObjectWithContext(c *C) {
	object := randLowStr(10)
	createFile(object, 1024*1024*1)
	fd, _ := os.Open(object)
	// put
	_, err := client.PutObjectWithContext(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
		Body:   fd,
	})
	c.Assert(err, IsNil)
	// head，不通过context取消
	resp, err := client.HeadObjectWithContext(context.Background(), &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(*resp.StatusCode, Equals, int64(200))
	// head，通过context取消
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Microsecond*10)
	defer cancelFunc()
	_, err = client.HeadObjectWithContext(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(err, NotNil)
	// delete
	_, err = client.DeleteObjectWithContext(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(err, IsNil)
	os.Remove(object)
}

// DELETE Object
func (s *Ks3utilCommandSuite) TestDeleteObjectWithContext(c *C) {
	object := randLowStr(10)
	createFile(object, 1024*1024*1)
	fd, _ := os.Open(object)
	// put
	_, err := client.PutObjectWithContext(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
		Body:   fd,
	})
	c.Assert(err, IsNil)
	// delete，通过context取消
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Microsecond*10)
	defer cancelFunc()
	_, err = client.HeadObjectWithContext(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(err, NotNil)
	// head
	resp, err := client.HeadObjectWithContext(context.Background(), &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(*resp.StatusCode, Equals, int64(200))
	// delete，不通过context取消
	_, err = client.DeleteObjectWithContext(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(err, IsNil)
	os.Remove(object)
}

// PUT Fetch Object
func (s *Ks3utilCommandSuite) TestFetchObjectWithContext(c *C) {
	sourceObject := randLowStr(10)
	createFile(sourceObject, 1024*1024*1)
	fd, _ := os.Open(sourceObject)
	// put
	_, err := client.PutObjectWithContext(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(sourceObject),
		ACL:    aws.String(s3.ACLPublicRead),
		Body:   fd,
	})
	c.Assert(err, IsNil)
	object := randLowStr(10)
	sourceUrl := fmt.Sprintf("https://%s.%s/%s", bucket, endpoint, sourceObject)
	// put fetch，不通过context取消
	_, err = client.FetchObjectWithContext(context.Background(), &s3.FetchObjectInput{
		Bucket:    aws.String(bucket),
		Key:       aws.String(object),
		SourceUrl: aws.String(sourceUrl),
	})
	c.Assert(err, IsNil)
	// put fetch 异步操作，等待10秒 head
	time.Sleep(time.Second * 10)
	// head
	resp, err := client.HeadObjectWithContext(context.Background(), &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(*resp.StatusCode, Equals, int64(200))
	// delete
	_, err = client.DeleteObjectWithContext(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(err, IsNil)
	object = randLowStr(10)
	// put fetch，通过context取消
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Microsecond*10)
	defer cancelFunc()
	_, err = client.FetchObjectWithContext(ctx, &s3.FetchObjectInput{
		Bucket:    aws.String(bucket),
		Key:       aws.String(object),
		SourceUrl: aws.String(sourceUrl),
	})
	c.Assert(err, NotNil)
	// head
	resp, err = client.HeadObjectWithContext(context.Background(), &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(*resp.StatusCode, Equals, int64(404))
	// delete
	_, err = client.DeleteObjectWithContext(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(sourceObject),
	})
	c.Assert(err, IsNil)
	os.Remove(sourceObject)
}

// PUT Object Copy
func (s *Ks3utilCommandSuite) TestCopyObjectWithContext(c *C) {
	srcObject := randLowStr(10)
	createFile(srcObject, 1024*1024*1)
	fd, _ := os.Open(srcObject)
	// put
	_, err := client.PutObjectWithContext(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(srcObject),
		Body:   fd,
	})
	c.Assert(err, IsNil)
	// head
	resp, err := client.HeadObjectWithContext(context.Background(), &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(srcObject),
	})
	c.Assert(*resp.StatusCode, Equals, int64(200))
	dstObject := randLowStr(10)
	// put copy，通过context取消
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Microsecond*10)
	defer cancelFunc()
	_, err = client.CopyObjectWithContext(ctx, &s3.CopyObjectInput{
		Bucket:     aws.String(bucket),
		Key:        aws.String(dstObject),
		CopySource: aws.String("/" + bucket + "/" + srcObject),
	})
	c.Assert(err, NotNil)
	// head
	resp, err = client.HeadObjectWithContext(context.Background(), &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(dstObject),
	})
	c.Assert(*resp.StatusCode, Equals, int64(404))
	// put copy，不通过context取消
	_, err = client.CopyObjectWithContext(context.Background(), &s3.CopyObjectInput{
		Bucket:     aws.String(bucket),
		Key:        aws.String(dstObject),
		CopySource: aws.String("/" + bucket + "/" + srcObject),
	})
	c.Assert(err, IsNil)
	// head
	resp, err = client.HeadObjectWithContext(context.Background(), &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(dstObject),
	})
	c.Assert(*resp.StatusCode, Equals, int64(200))
	// delete
	_, err = client.DeleteObjectWithContext(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(srcObject),
	})
	c.Assert(err, IsNil)
	// delete
	_, err = client.DeleteObjectWithContext(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(dstObject),
	})
	c.Assert(err, IsNil)
	os.Remove(srcObject)
}

// Restore Object
func (s *Ks3utilCommandSuite) TestRestoreObjectWithContext(c *C) {
	object := randLowStr(10)
	createFile(object, 1024*1024*1)
	fd, _ := os.Open(object)
	// put
	_, err := client.PutObjectWithContext(context.Background(), &s3.PutObjectInput{
		Bucket:       aws.String(bucket),
		Key:          aws.String(object),
		Body:         fd,
		StorageClass: aws.String("ARCHIVE"),
	})
	c.Assert(err, IsNil)
	// head
	resp, err := client.HeadObjectWithContext(context.Background(), &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(*resp.Metadata["X-Amz-Storage-Class"], Equals, "ARCHIVE")
	c.Assert(*resp.StatusCode, Equals, int64(200))
	// get
	_, err = client.GetObjectWithContext(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(err, NotNil)
	// restore，通过context取消
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Microsecond*1)
	defer cancelFunc()
	_, err = client.RestoreObjectWithContext(ctx, &s3.RestoreObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(err, NotNil)
	// restore，不通过context取消
	_, err = client.RestoreObjectWithContext(context.Background(), &s3.RestoreObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(err, IsNil)
	// delete
	_, err = client.DeleteObjectWithContext(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(err, IsNil)
	os.Remove(object)
}

// GET Object To File
func (s *Ks3utilCommandSuite) TestGetObjectToFileWithContext(c *C) {
	object := randLowStr(10)
	createFile(object, 1024*1024*10)
	fd, _ := os.Open(object)
	// put
	_, err := client.PutObjectWithContext(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
		Body:   fd,
	})
	c.Assert(err, IsNil)
	// head
	resp, err := client.HeadObjectWithContext(context.Background(), &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(*resp.StatusCode, Equals, int64(200))
	os.Remove(object)
	filePath := object
	// 下载文件，不通过context取消
	err = client.GetObjectToFileWithContext(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	}, filePath)
	c.Assert(err, IsNil)
	os.Remove(filePath)
	// Range下载
	err = client.GetObjectToFileWithContext(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
		Range:  aws.String("bytes=0-100000"),
	}, filePath)
	c.Assert(err, IsNil)
	os.Remove(filePath)
	// 下载文件，通过context取消
	ctx, cancelFunc := context.WithTimeout(context.Background(), timeout)
	defer cancelFunc()
	err = client.GetObjectToFileWithContext(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	}, filePath)
	c.Assert(err, NotNil)
	os.Remove(filePath + ".temp")
	// delete
	_, err = client.DeleteObjectWithContext(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(err, IsNil)
}

// s3manager Upload
func (s *Ks3utilCommandSuite) TestUploadWithContext(c *C) {
	sourceObject := randLowStr(10)
	createFile(sourceObject, 1024*1024*10)
	fd, _ := os.Open(sourceObject)
	// 初始化配置
	uploader := s3manager.NewUploader(&s3manager.UploadOptions{
		S3: client, // S3Client实例，必填
	})
	// s3manager Upload 上传
	_, err := uploader.UploadWithContext(context.Background(), &s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(sourceObject),
		ACL:    aws.String(s3.ACLPublicRead),
		Body:   fd,
	})
	c.Assert(err, IsNil)
	object := randLowStr(10)
	sourceUrl := fmt.Sprintf("https://%s.%s/%s", bucket, endpoint, sourceObject)
	result, err := http.Get(sourceUrl)
	c.Assert(err, IsNil)
	// 上传网络流，不通过context取消
	_, err = uploader.UploadWithContext(context.Background(), &s3manager.UploadInput{
		Bucket: aws.String(bucket), // 存储空间名称，必填
		Key:    aws.String(object), // 对象的key，必填、
		Body:   result.Body,        // 要上传的文件，必填
	})
	c.Assert(err, IsNil)
	// head
	resp, err := client.GetObjectWithContext(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(*resp.StatusCode, Equals, int64(200))
	// delete
	_, err = client.DeleteObjectWithContext(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(err, IsNil)
	// 上传网络流，通过context取消
	result, err = http.Get(sourceUrl)
	c.Assert(err, IsNil)
	ctx, cancelFunc := context.WithTimeout(context.Background(), timeout)
	defer cancelFunc()
	_, err = uploader.UploadWithContext(ctx, &s3manager.UploadInput{
		Bucket: aws.String(bucket), // 存储空间名称，必填
		Key:    aws.String(object), // 对象的key，必填、
		Body:   result.Body,        // 要上传的文件，必填
	})
	c.Assert(err, NotNil)
	// head
	resp, err = client.GetObjectWithContext(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(*resp.StatusCode, Equals, int64(404))
	// delete
	_, err = client.DeleteObjectWithContext(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(sourceObject),
	})
	c.Assert(err, IsNil)
	os.Remove(sourceObject)
}

// s3manager Upload Dir
func (s *Ks3utilCommandSuite) TestUploadDirWithContext(c *C) {
	dir := randLowStr(8) + "/"
	os.Mkdir(dir, 0777)
	prefix := randLowStr(6) + "/"
	object1 := randLowStr(10)
	createFile(dir+object1, 1024*1024*1)
	object2 := randLowStr(10)
	createFile(dir+object2, 1024*1024*12)
	object3 := randLowStr(10)
	createFile(dir+object3, 1024*1024*20)
	// 初始化配置
	uploader := s3manager.NewUploader(&s3manager.UploadOptions{
		S3: client, // S3Client实例，必填
	})
	// upload dir，通过context取消
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Millisecond*1)
	defer cancelFunc()
	err := uploader.UploadDirWithContext(ctx, &s3manager.UploadDirInput{
		RootDir:      dir,
		Bucket:       bucket,
		Prefix:       prefix,
		ACL:          s3.ACLPublicRead,
		StorageClass: s3.StorageClassIA,
	})
	c.Assert(err, IsNil)
	// list
	resp, err := client.ListObjectsWithContext(context.Background(), &s3.ListObjectsInput{
		Bucket: aws.String(bucket),
		Prefix: aws.String(prefix),
	})
	c.Assert(err, IsNil)
	c.Assert(len(resp.Contents), Equals, 0)
	// upload dir，不通过context取消
	err = uploader.UploadDirWithContext(context.Background(), &s3manager.UploadDirInput{
		RootDir:      dir,
		Bucket:       bucket,
		Prefix:       prefix,
		ACL:          s3.ACLPublicRead,
		StorageClass: s3.StorageClassIA,
	})
	c.Assert(err, IsNil)
	// list
	resp, err = client.ListObjectsWithContext(context.Background(), &s3.ListObjectsInput{
		Bucket: aws.String(bucket),
		Prefix: aws.String(prefix),
	})
	c.Assert(err, IsNil)
	c.Assert(len(resp.Contents), Equals, 3)
	// delete objects
	_, err = client.DeleteBucketPrefixWithContext(context.Background(), &s3.DeleteBucketPrefixInput{
		Bucket: aws.String(bucket),
		Prefix: aws.String(prefix),
	})
	c.Assert(err, IsNil)
	// list
	resp, err = client.ListObjectsWithContext(context.Background(), &s3.ListObjectsInput{
		Bucket: aws.String(bucket),
		Prefix: aws.String(prefix),
	})
	c.Assert(err, IsNil)
	c.Assert(len(resp.Contents), Equals, 0)
	os.RemoveAll(dir)
}

// DELETE Objects
func (s *Ks3utilCommandSuite) TestDeleteObjectsWithContext(c *C) {
	object := randLowStr(10)
	createFile(object, 1024*1024*1)
	fd, _ := os.Open(object)
	// put
	_, err := client.PutObjectWithContext(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
		Body:   fd,
	})
	c.Assert(err, IsNil)
	// delete，通过context取消
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Microsecond*10)
	defer cancelFunc()
	resp, err := client.DeleteObjectsWithContext(ctx, &s3.DeleteObjectsInput{
		Bucket:          aws.String(bucket),
		IsReTurnResults: aws.Boolean(true),
		Delete: &s3.Delete{
			Objects: []*s3.ObjectIdentifier{
				{Key: aws.String(object)},
			},
		},
	})
	c.Assert(len(resp.Errors), Equals, 1)
	// delete，不通过context取消
	resp, err = client.DeleteObjectsWithContext(context.Background(), &s3.DeleteObjectsInput{
		Bucket:          aws.String(bucket),
		IsReTurnResults: aws.Boolean(true),
		Delete: &s3.Delete{
			Objects: []*s3.ObjectIdentifier{
				{Key: aws.String(object)},
			},
		},
	})
	c.Assert(len(resp.Deleted), Equals, 1)
	os.Remove(object)
}

// DELETE Bucket Prefix
func (s *Ks3utilCommandSuite) TestDeleteBucketPrefixWithContext(c *C) {
	object := randLowStr(10)
	prefix := randLowStr(6) + "/"
	createFile(object, 1024*1024*1)
	fd, _ := os.Open(object)
	// put
	_, err := client.PutObjectWithContext(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(prefix + object),
		Body:   fd,
	})
	c.Assert(err, IsNil)
	// delete，通过context取消
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Microsecond*10)
	defer cancelFunc()
	_, err = client.DeleteBucketPrefixWithContext(ctx, &s3.DeleteBucketPrefixInput{
		Bucket:          aws.String(bucket),
		Prefix:          aws.String(prefix),
		MaxKeys:         aws.Long(20),
		IsReTurnResults: aws.Boolean(true),
	})
	c.Assert(err, NotNil)
	// delete，不通过context取消
	resp, err := client.DeleteBucketPrefixWithContext(context.Background(), &s3.DeleteBucketPrefixInput{
		Bucket:          aws.String(bucket),
		Prefix:          aws.String(prefix),
		MaxKeys:         aws.Long(20),
		IsReTurnResults: aws.Boolean(true),
	})
	c.Assert(err, IsNil)
	c.Assert(len(resp.Deleted), Equals, 1)
	os.Remove(object)
}

// DELETE Bucket Prefix Try 3
func (s *Ks3utilCommandSuite) TestTryDeleteBucketPrefixWithContext(c *C) {
	object := randLowStr(10)
	prefix := randLowStr(6) + "/"
	createFile(object, 1024*1024*1)
	fd, _ := os.Open(object)
	// put
	_, err := client.PutObjectWithContext(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(prefix + object),
		Body:   fd,
	})
	c.Assert(err, IsNil)
	// delete，通过context取消
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Microsecond*10)
	defer cancelFunc()
	_, err = client.TryDeleteBucketPrefixWithContext(ctx, &s3.DeleteBucketPrefixInput{
		Bucket:          aws.String(bucket),
		Prefix:          aws.String(prefix),
		MaxKeys:         aws.Long(20),
		IsReTurnResults: aws.Boolean(true),
	})
	c.Assert(err, NotNil)
	// delete，不通过context取消
	resp, err := client.TryDeleteBucketPrefixWithContext(context.Background(), &s3.DeleteBucketPrefixInput{
		Bucket:          aws.String(bucket),
		Prefix:          aws.String(prefix),
		MaxKeys:         aws.Long(20),
		IsReTurnResults: aws.Boolean(true),
	})
	c.Assert(len(resp.Deleted), Equals, 1)
	os.Remove(object)
}

// Initiate Multipart Upload
func (s *Ks3utilCommandSuite) TestCreateMultipartUploadWithContext(c *C) {
	object := randLowStr(10)
	createFile(object, 1024*1024*10)
	fd, _ := os.Open(object)
	// init，通过context取消
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Microsecond*10)
	defer cancelFunc()
	_, err := client.CreateMultipartUploadWithContext(ctx, &s3.CreateMultipartUploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(err, NotNil)
	// init，不通过context取消
	initRet, err := client.CreateMultipartUploadWithContext(context.Background(), &s3.CreateMultipartUploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(err, IsNil)
	// 获取分块上传Id
	uploadId := *initRet.UploadID
	var partNum int64 = 1
	// 待合并分块
	var compParts []*s3.CompletedPart
	// 缓冲区，分块大小为5MB
	buffer := make([]byte, 5*1024*1024)
	for {
		n, err := fd.Read(buffer)
		if err != nil && err != io.EOF {
			panic(err)
		} else if n == 0 {
			break
		} else {
			// part
			resp, err := client.UploadPartWithContext(context.Background(), &s3.UploadPartInput{
				Bucket:     aws.String(bucket),
				Key:        aws.String(object),
				PartNumber: aws.Long(partNum),
				UploadID:   aws.String(uploadId),
				Body:       bytes.NewReader(buffer[:n]),
			})
			c.Assert(err, IsNil)
			compParts = append(compParts, &s3.CompletedPart{PartNumber: aws.Long(partNum), ETag: resp.ETag})
			partNum++
		}
	}
	// complete
	_, err = client.CompleteMultipartUploadWithContext(context.Background(), &s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(bucket),
		Key:      aws.String(object),
		UploadID: aws.String(uploadId),
		MultipartUpload: &s3.CompletedMultipartUpload{
			Parts: compParts,
		},
	})
	c.Assert(err, IsNil)
	// delete
	_, err = client.DeleteObjectWithContext(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(err, IsNil)
	os.Remove(object)
}

// Upload Part
func (s *Ks3utilCommandSuite) TestUploadPartWithContext(c *C) {
	object := randLowStr(10)
	createFile(object, 1024*1024*10)
	fd, _ := os.Open(object)
	// init
	initRet, err := client.CreateMultipartUploadWithContext(context.Background(), &s3.CreateMultipartUploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(err, IsNil)
	// 获取分块上传Id
	uploadId := *initRet.UploadID
	var partNum int64 = 1
	// 待合并分块
	var compParts []*s3.CompletedPart
	// 缓冲区，分块大小为5MB
	buffer := make([]byte, 5*1024*1024)
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Millisecond*500)
	defer cancelFunc()
	for {
		n, err := fd.Read(buffer)
		if err != nil && err != io.EOF {
			panic(err)
		} else if n == 0 {
			break
		} else {
			// upload part，通过context取消
			_, err := client.UploadPartWithContext(ctx, &s3.UploadPartInput{
				Bucket:     aws.String(bucket),
				Key:        aws.String(object),
				PartNumber: aws.Long(partNum),
				UploadID:   aws.String(uploadId),
				Body:       bytes.NewReader(buffer[:n]),
			})
			c.Assert(err, NotNil)
		}
	}
	_, err = fd.Seek(0, 0)
	c.Assert(err, IsNil)
	for {
		n, err := fd.Read(buffer)
		if err != nil && err != io.EOF {
			panic(err)
		} else if n == 0 {
			break
		} else {
			// upload part，不通过context取消
			resp, err := client.UploadPartWithContext(context.Background(), &s3.UploadPartInput{
				Bucket:     aws.String(bucket),
				Key:        aws.String(object),
				PartNumber: aws.Long(partNum),
				UploadID:   aws.String(uploadId),
				Body:       bytes.NewReader(buffer[:n]),
			})
			c.Assert(err, IsNil)
			compParts = append(compParts, &s3.CompletedPart{PartNumber: aws.Long(partNum), ETag: resp.ETag})
			partNum++
		}
	}
	// complete
	_, err = client.CompleteMultipartUploadWithContext(context.Background(), &s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(bucket),
		Key:      aws.String(object),
		UploadID: aws.String(uploadId),
		MultipartUpload: &s3.CompletedMultipartUpload{
			Parts: compParts,
		},
	})
	c.Assert(err, IsNil)
	// delete
	_, err = client.DeleteObjectWithContext(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(err, IsNil)
	os.Remove(object)
}

// Complete Multipart Upload
func (s *Ks3utilCommandSuite) TestCompleteMultipartUploadWithContext(c *C) {
	object := randLowStr(10)
	createFile(object, 1024*1024*10)
	fd, _ := os.Open(object)
	// init
	initRet, err := client.CreateMultipartUploadWithContext(context.Background(), &s3.CreateMultipartUploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(err, IsNil)
	// 获取分块上传Id
	uploadId := *initRet.UploadID
	var partNum int64 = 1
	// 待合并分块
	var compParts []*s3.CompletedPart
	// 缓冲区，分块大小为5MB
	buffer := make([]byte, 5*1024*1024)
	for {
		n, err := fd.Read(buffer)
		if err != nil && err != io.EOF {
			panic(err)
		} else if n == 0 {
			break
		} else {
			// part
			resp, err := client.UploadPartWithContext(context.Background(), &s3.UploadPartInput{
				Bucket:     aws.String(bucket),
				Key:        aws.String(object),
				PartNumber: aws.Long(partNum),
				UploadID:   aws.String(uploadId),
				Body:       bytes.NewReader(buffer[:n]),
			})
			c.Assert(err, IsNil)
			compParts = append(compParts, &s3.CompletedPart{PartNumber: aws.Long(partNum), ETag: resp.ETag})
			partNum++
		}
	}
	// complete，通过context取消
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Microsecond*10)
	defer cancelFunc()
	_, err = client.CompleteMultipartUploadWithContext(ctx, &s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(bucket),
		Key:      aws.String(object),
		UploadID: aws.String(uploadId),
		MultipartUpload: &s3.CompletedMultipartUpload{
			Parts: compParts,
		},
	})
	c.Assert(err, NotNil)
	// complete，不通过context取消
	_, err = client.CompleteMultipartUploadWithContext(context.Background(), &s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(bucket),
		Key:      aws.String(object),
		UploadID: aws.String(uploadId),
		MultipartUpload: &s3.CompletedMultipartUpload{
			Parts: compParts,
		},
	})
	c.Assert(err, IsNil)
	listMulRes, err := client.ListMultipartUploadsWithContext(context.Background(), &s3.ListMultipartUploadsInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, IsNil)
	uploadIdExits := false
	for _, upload := range listMulRes.Uploads {
		if *upload.UploadID == uploadId {
			uploadIdExits = true
		}
	}
	c.Assert(uploadIdExits, Equals, false)
	// delete
	_, err = client.DeleteObjectWithContext(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(err, IsNil)
	os.Remove(object)
}

// Abort Multipart Upload
func (s *Ks3utilCommandSuite) TestAbortMultipartUploadWithContext(c *C) {
	object := randLowStr(10)
	createFile(object, 1024*1024*10)
	fd, _ := os.Open(object)
	// init
	initRet, err := client.CreateMultipartUploadWithContext(context.Background(), &s3.CreateMultipartUploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(err, IsNil)
	// 获取分块上传Id
	uploadId := *initRet.UploadID
	var partNum int64 = 1
	// 待合并分块
	var compParts []*s3.CompletedPart
	// 缓冲区，分块大小为5MB
	buffer := make([]byte, 5*1024*1024)
	for {
		n, err := fd.Read(buffer)
		if err != nil && err != io.EOF {
			panic(err)
		} else if n == 0 {
			break
		} else {
			// part
			resp, err := client.UploadPartWithContext(context.Background(), &s3.UploadPartInput{
				Bucket:     aws.String(bucket),
				Key:        aws.String(object),
				PartNumber: aws.Long(partNum),
				UploadID:   aws.String(uploadId),
				Body:       bytes.NewReader(buffer[:n]),
			})
			c.Assert(err, IsNil)
			compParts = append(compParts, &s3.CompletedPart{PartNumber: aws.Long(partNum), ETag: resp.ETag})
			partNum++
		}
	}
	// abort，通过context取消
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Microsecond*10)
	defer cancelFunc()
	_, err = client.AbortMultipartUploadWithContext(ctx, &s3.AbortMultipartUploadInput{
		Bucket:   aws.String(bucket),
		Key:      aws.String(object),
		UploadID: aws.String(uploadId),
	})
	c.Assert(err, NotNil)
	// abort，不通过context取消
	_, err = client.AbortMultipartUploadWithContext(context.Background(), &s3.AbortMultipartUploadInput{
		Bucket:   aws.String(bucket),
		Key:      aws.String(object),
		UploadID: aws.String(uploadId),
	})
	c.Assert(err, IsNil)
	listMulRes, err := client.ListMultipartUploadsWithContext(context.Background(), &s3.ListMultipartUploadsInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, IsNil)
	uploadIdExits := false
	for _, upload := range listMulRes.Uploads {
		if *upload.UploadID == uploadId {
			uploadIdExits = true
		}
	}
	c.Assert(uploadIdExits, Equals, false)
	os.Remove(object)
}

// List Parts
func (s *Ks3utilCommandSuite) TestListPartsWithContext(c *C) {
	object := randLowStr(10)
	createFile(object, 1024*1024*10)
	fd, _ := os.Open(object)
	// init
	initRet, err := client.CreateMultipartUploadWithContext(context.Background(), &s3.CreateMultipartUploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(err, IsNil)
	// 获取分块上传Id
	uploadId := *initRet.UploadID
	var partNum int64 = 1
	// 待合并分块
	var compParts []*s3.CompletedPart
	// 缓冲区，分块大小为5MB
	buffer := make([]byte, 5*1024*1024)
	for {
		n, err := fd.Read(buffer)
		if err != nil && err != io.EOF {
			panic(err)
		} else if n == 0 {
			break
		} else {
			// part
			resp, err := client.UploadPartWithContext(context.Background(), &s3.UploadPartInput{
				Bucket:     aws.String(bucket),
				Key:        aws.String(object),
				PartNumber: aws.Long(partNum),
				UploadID:   aws.String(uploadId),
				Body:       bytes.NewReader(buffer[:n]),
			})
			c.Assert(err, IsNil)
			compParts = append(compParts, &s3.CompletedPart{PartNumber: aws.Long(partNum), ETag: resp.ETag})
			partNum++
		}
	}
	// list，通过context取消
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Microsecond*10)
	defer cancelFunc()
	_, err = client.ListPartsWithContext(ctx, &s3.ListPartsInput{
		Bucket:   aws.String(bucket),
		Key:      aws.String(object),
		UploadID: aws.String(uploadId),
	})
	c.Assert(err, NotNil)
	// list，不通过context取消
	listPartRes, err := client.ListPartsWithContext(context.Background(), &s3.ListPartsInput{
		Bucket:   aws.String(bucket),
		Key:      aws.String(object),
		UploadID: aws.String(uploadId),
	})
	c.Assert(err, IsNil)
	c.Assert(len(listPartRes.Parts), Equals, 2)
	// abort
	_, err = client.AbortMultipartUploadWithContext(context.Background(), &s3.AbortMultipartUploadInput{
		Bucket:   aws.String(bucket),
		Key:      aws.String(object),
		UploadID: aws.String(uploadId),
	})
	c.Assert(err, IsNil)
	os.Remove(object)
}

// List Multipart Uploads
func (s *Ks3utilCommandSuite) TestListMultipartUploadsWithContext(c *C) {
	object := randLowStr(10)
	createFile(object, 1024*1024*10)
	fd, _ := os.Open(object)
	// init
	initRet, err := client.CreateMultipartUploadWithContext(context.Background(), &s3.CreateMultipartUploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(err, IsNil)
	// 获取分块上传Id
	uploadId := *initRet.UploadID
	var partNum int64 = 1
	// 待合并分块
	var compParts []*s3.CompletedPart
	// 缓冲区，分块大小为5MB
	buffer := make([]byte, 5*1024*1024)
	for {
		n, err := fd.Read(buffer)
		if err != nil && err != io.EOF {
			panic(err)
		} else if n == 0 {
			break
		} else {
			// part
			resp, err := client.UploadPartWithContext(context.Background(), &s3.UploadPartInput{
				Bucket:     aws.String(bucket),
				Key:        aws.String(object),
				PartNumber: aws.Long(partNum),
				UploadID:   aws.String(uploadId),
				Body:       bytes.NewReader(buffer[:n]),
			})
			c.Assert(err, IsNil)
			compParts = append(compParts, &s3.CompletedPart{PartNumber: aws.Long(partNum), ETag: resp.ETag})
			partNum++
		}
	}
	// list mul，通过context取消
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Microsecond*10)
	defer cancelFunc()
	_, err = client.ListMultipartUploadsWithContext(ctx, &s3.ListMultipartUploadsInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, NotNil)
	// list mul，不通过context取消
	listMulRes, err := client.ListMultipartUploadsWithContext(context.Background(), &s3.ListMultipartUploadsInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, IsNil)
	uploadIdExits := false
	for _, upload := range listMulRes.Uploads {
		if *upload.UploadID == uploadId {
			uploadIdExits = true
		}
	}
	c.Assert(uploadIdExits, Equals, true)
	// abort
	_, err = client.AbortMultipartUploadWithContext(context.Background(), &s3.AbortMultipartUploadInput{
		Bucket:   aws.String(bucket),
		Key:      aws.String(object),
		UploadID: aws.String(uploadId),
	})
	c.Assert(err, IsNil)
	os.Remove(object)
}

// Upload Part Copy
func (s *Ks3utilCommandSuite) TestPartWithContext(c *C) {
	srcObject := randLowStr(10)
	createFile(srcObject, 1024*1024*10)
	fd, _ := os.Open(srcObject)
	// put
	_, err := client.PutObjectWithContext(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(srcObject),
		Body:   fd,
	})
	c.Assert(err, IsNil)
	headObjectResp, err := client.HeadObjectWithContext(context.Background(), &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(srcObject),
	})
	c.Assert(err, IsNil)
	c.Assert(*headObjectResp.StatusCode, Equals, int64(200))
	contentLength := *headObjectResp.ContentLength
	var partSize int64 = 5 * 1024 * 1024
	var partNum int64 = 1
	// 待合并分块
	var compParts []*s3.CompletedPart
	var start int64 = 0
	var end int64 = 0
	dstObject := randLowStr(10)
	// init
	initRet, err := client.CreateMultipartUploadWithContext(context.Background(), &s3.CreateMultipartUploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(dstObject),
	})
	c.Assert(err, IsNil)
	// 获取分块上传Id
	uploadId := *initRet.UploadID
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Microsecond*10)
	defer cancelFunc()
	for {
		if end >= contentLength {
			break
		}
		if start+partSize >= contentLength {
			end = contentLength
		} else {
			end = start + partSize
		}
		// upload part copy，通过context取消
		_, err := client.UploadPartCopyWithContext(ctx, &s3.UploadPartCopyInput{
			Bucket:          aws.String(bucket),
			Key:             aws.String(dstObject),
			CopySource:      aws.String("/" + bucket + "/" + srcObject),
			UploadID:        aws.String(uploadId),
			PartNumber:      aws.Long(partNum),
			CopySourceRange: aws.String("bytes=" + strconv.FormatInt(start, 10) + "-" + strconv.FormatInt(end-1, 10)),
		})
		c.Assert(err, NotNil)
		partNum++
		start = end
	}
	partNum = 1
	start = 0
	end = 0
	for {
		if end >= contentLength {
			break
		}
		if start+partSize >= contentLength {
			end = contentLength
		} else {
			end = start + partSize
		}
		// upload part copy，不通过context取消
		resp, err := client.UploadPartCopyWithContext(context.Background(), &s3.UploadPartCopyInput{
			Bucket:          aws.String(bucket),
			Key:             aws.String(dstObject),
			CopySource:      aws.String("/" + bucket + "/" + srcObject),
			UploadID:        aws.String(uploadId),
			PartNumber:      aws.Long(partNum),
			CopySourceRange: aws.String("bytes=" + strconv.FormatInt(start, 10) + "-" + strconv.FormatInt(end-1, 10)),
		})
		c.Assert(err, IsNil)
		compParts = append(compParts, &s3.CompletedPart{PartNumber: aws.Long(partNum), ETag: resp.CopyPartResult.ETag})
		partNum++
		start = end
	}
	// complete
	_, err = client.CompleteMultipartUploadWithContext(context.Background(), &s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(bucket),
		Key:      aws.String(dstObject),
		UploadID: aws.String(uploadId),
		MultipartUpload: &s3.CompletedMultipartUpload{
			Parts: compParts,
		},
	})
	c.Assert(err, IsNil)
	// delete
	_, err = client.DeleteObjectWithContext(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(srcObject),
	})
	c.Assert(err, IsNil)
	// delete
	_, err = client.DeleteObjectWithContext(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(dstObject),
	})
	c.Assert(err, IsNil)
	os.Remove(srcObject)
}

// PUT Object ACL
func (s *Ks3utilCommandSuite) TestPutObjectACLWithContext(c *C) {
	object := randLowStr(10)
	createFile(object, 1024*1024*1)
	fd, _ := os.Open(object)
	// put
	_, err := client.PutObjectWithContext(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
		Body:   fd,
	})
	c.Assert(err, IsNil)
	// get acl
	resp, err := client.GetObjectACLWithContext(context.Background(), &s3.GetObjectACLInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(err, IsNil)
	c.Assert(s3.GetCannedACL(resp.Grants), Equals, s3.ACLPrivate)
	// put acl，通过context取消
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Microsecond*10)
	defer cancelFunc()
	_, err = client.PutObjectACLWithContext(ctx, &s3.PutObjectACLInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
		ACL:    aws.String("public-read"),
	})
	c.Assert(err, NotNil)
	// get acl
	resp, err = client.GetObjectACLWithContext(context.Background(), &s3.GetObjectACLInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(err, IsNil)
	c.Assert(s3.GetCannedACL(resp.Grants), Equals, s3.ACLPrivate)
	// put acl，不通过context取消
	_, err = client.PutObjectACLWithContext(context.Background(), &s3.PutObjectACLInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
		ACL:    aws.String("public-read"),
	})
	c.Assert(err, IsNil)
	// get acl
	resp, err = client.GetObjectACLWithContext(context.Background(), &s3.GetObjectACLInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(err, IsNil)
	c.Assert(s3.GetCannedACL(resp.Grants), Equals, s3.ACLPublicRead)
	// delete
	_, err = client.DeleteObjectWithContext(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(err, IsNil)
	os.Remove(object)
}

// GET Object ACL
func (s *Ks3utilCommandSuite) TestGetObjectACLWithContext(c *C) {
	object := randLowStr(10)
	createFile(object, 1024*1024*1)
	fd, _ := os.Open(object)
	// put
	_, err := client.PutObjectWithContext(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
		Body:   fd,
		ACL:    aws.String("public-read"),
	})
	c.Assert(err, IsNil)
	// get acl，通过context取消
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Microsecond*10)
	defer cancelFunc()
	_, err = client.GetObjectACLWithContext(ctx, &s3.GetObjectACLInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(err, NotNil)
	// get acl，不通过context取消
	resp, err := client.GetObjectACLWithContext(context.Background(), &s3.GetObjectACLInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(err, IsNil)
	c.Assert(s3.GetCannedACL(resp.Grants), Equals, s3.ACLPublicRead)
	// delete
	_, err = client.DeleteObjectWithContext(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(err, IsNil)
	os.Remove(object)
}

// PUT Object Tagging
func (s *Ks3utilCommandSuite) TestPutObjectTaggingWithContext(c *C) {
	object := randLowStr(10)
	createFile(object, 1024*1024*1)
	fd, _ := os.Open(object)
	// put
	_, err := client.PutObjectWithContext(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
		Body:   fd,
	})
	c.Assert(err, IsNil)
	// 构造对象标签
	objTagging := &s3.Tagging{
		TagSet: []*s3.Tag{
			{
				Key:   aws.String("tagKey"),
				Value: aws.String("tagValue"),
			},
		},
	}
	// put tagging，通过context取消
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Microsecond*10)
	defer cancelFunc()
	_, err = client.PutObjectTaggingWithContext(ctx, &s3.PutObjectTaggingInput{
		Bucket:  aws.String(bucket),
		Key:     aws.String(object),
		Tagging: objTagging,
	})
	c.Assert(err, NotNil)
	// put tagging，不通过context取消
	_, err = client.PutObjectTaggingWithContext(context.Background(), &s3.PutObjectTaggingInput{
		Bucket:  aws.String(bucket),
		Key:     aws.String(object),
		Tagging: objTagging,
	})
	c.Assert(err, IsNil)
	// get tagging
	resp, err := client.GetObjectTaggingWithContext(context.Background(), &s3.GetObjectTaggingInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(err, IsNil)
	c.Assert(awsutil.StringValue(objTagging), Equals, awsutil.StringValue(resp.Tagging))
	// delete
	_, err = client.DeleteObjectWithContext(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(err, IsNil)
	os.Remove(object)
}

// GET Object Tagging
func (s *Ks3utilCommandSuite) TestGetObjectTaggingWithContext(c *C) {
	object := randLowStr(10)
	createFile(object, 1024*1024*1)
	fd, _ := os.Open(object)
	// put
	_, err := client.PutObjectWithContext(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
		Body:   fd,
	})
	c.Assert(err, IsNil)
	// 构造对象标签
	objTagging := &s3.Tagging{
		TagSet: []*s3.Tag{
			{
				Key:   aws.String("tagKey"),
				Value: aws.String("tagValue"),
			},
		},
	}
	// put tagging
	_, err = client.PutObjectTaggingWithContext(context.Background(), &s3.PutObjectTaggingInput{
		Bucket:  aws.String(bucket),
		Key:     aws.String(object),
		Tagging: objTagging,
	})
	c.Assert(err, IsNil)
	// get tagging，通过context取消
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Microsecond*10)
	defer cancelFunc()
	_, err = client.GetObjectTaggingWithContext(ctx, &s3.GetObjectTaggingInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(err, NotNil)
	// get tagging，不通过context取消
	resp, err := client.GetObjectTaggingWithContext(context.Background(), &s3.GetObjectTaggingInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(err, IsNil)
	c.Assert(awsutil.StringValue(objTagging), Equals, awsutil.StringValue(resp.Tagging))
	// delete
	_, err = client.DeleteObjectWithContext(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(err, IsNil)
	os.Remove(object)
}

// DELETE Object Tagging
func (s *Ks3utilCommandSuite) TestDeleteObjectTaggingWithContext(c *C) {
	object := randLowStr(10)
	createFile(object, 1024*1024*1)
	fd, _ := os.Open(object)
	// put
	_, err := client.PutObjectWithContext(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
		Body:   fd,
	})
	c.Assert(err, IsNil)
	// 构造对象标签
	objTagging := &s3.Tagging{
		TagSet: []*s3.Tag{
			{
				Key:   aws.String("tagKey"),
				Value: aws.String("tagValue"),
			},
		},
	}
	// put tagging
	_, err = client.PutObjectTaggingWithContext(context.Background(), &s3.PutObjectTaggingInput{
		Bucket:  aws.String(bucket),
		Key:     aws.String(object),
		Tagging: objTagging,
	})
	c.Assert(err, IsNil)
	// get tagging
	resp, err := client.GetObjectTaggingWithContext(context.Background(), &s3.GetObjectTaggingInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(err, IsNil)
	c.Assert(awsutil.StringValue(objTagging), Equals, awsutil.StringValue(resp.Tagging))
	// delete tagging，通过context取消
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Microsecond*10)
	defer cancelFunc()
	_, err = client.DeleteObjectTaggingWithContext(ctx, &s3.DeleteObjectTaggingInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(err, NotNil)
	// get tagging
	resp, err = client.GetObjectTaggingWithContext(context.Background(), &s3.GetObjectTaggingInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(err, IsNil)
	c.Assert(awsutil.StringValue(objTagging), Equals, awsutil.StringValue(resp.Tagging))
	// delete tagging，不通过context取消
	_, err = client.DeleteObjectTaggingWithContext(context.Background(), &s3.DeleteObjectTaggingInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(err, IsNil)
	// get tagging
	resp, err = client.GetObjectTaggingWithContext(context.Background(), &s3.GetObjectTaggingInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(err, IsNil)
	c.Assert(awsutil.StringValue(objTagging), Not(Equals), awsutil.StringValue(resp.Tagging))
	// delete
	_, err = client.DeleteObjectWithContext(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(err, IsNil)
	os.Remove(object)
}

// Append Object
func (s *Ks3utilCommandSuite) TestAppendObjectWithContext(c *C) {
	object := randLowStr(10)
	createFile(object, 1024*1024*10)
	fd, _ := os.Open(object)
	// 追加上传对象，不通过context取消
	resp, err := client.AppendObjectWithContext(context.Background(), &s3.AppendObjectInput{
		Bucket:       aws.String(bucket),
		Key:          aws.String(object),
		Position:     aws.Long(0),
		Body:         fd,
		ACL:          aws.String(s3.ACLPublicRead),
		StorageClass: aws.String(s3.StorageClassDeepIA),
	})
	c.Assert(err, IsNil)
	position := *resp.NextAppendPosition
	object2 := randLowStr(10)
	createFile(object2, 1024*1024*10)
	fd2, _ := os.Open(object2)
	// 再次追加
	_, err = client.AppendObjectWithContext(context.Background(), &s3.AppendObjectInput{
		Bucket:   aws.String(bucket),
		Key:      aws.String(object),
		Position: aws.Long(position),
		Body:     fd2,
	})
	c.Assert(err, IsNil)
	// head
	headResp, err := client.HeadObjectWithContext(context.Background(), &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(*headResp.StatusCode, Equals, int64(200))
	contentLength, err := strconv.ParseInt(*headResp.Metadata["Content-Length"], 10, 64)
	c.Assert(err, IsNil)
	c.Assert(contentLength, Equals, int64(1024*1024*10*2))
	// delete
	_, err = client.DeleteObjectWithContext(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(err, IsNil)
	os.Remove(object)
	os.Remove(object2)
	object = randLowStr(10)
	createFile(object, 1024*1024*10)
	fd, _ = os.Open(object)
	ctx, cancelFunc := context.WithTimeout(context.Background(), timeout)
	defer cancelFunc()
	// 追加上传对象，通过context取消
	resp, err = client.AppendObjectWithContext(ctx, &s3.AppendObjectInput{
		Bucket:   aws.String(bucket),
		Key:      aws.String(object),
		Position: aws.Long(0),
		ACL:      aws.String(s3.ACLPublicRead),
		Body:     fd,
	})
	c.Assert(err, NotNil)
	// head
	headResp, err = client.HeadObjectWithContext(context.Background(), &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(*headResp.StatusCode, Equals, int64(404))
	os.Remove(object)
}
