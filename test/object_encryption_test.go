package lib

import (
	"bytes"
	"context"
	"github.com/ks3sdklib/aws-sdk-go/aws"
	"github.com/ks3sdklib/aws-sdk-go/service/s3"
	"github.com/ks3sdklib/aws-sdk-go/service/s3/s3util"
	. "gopkg.in/check.v1"
	"io"
	"os"
)

var customerKey = "<encryption_key>"

// PUT Object with SSE-S3
func (s *Ks3utilCommandSuite) TestPutObjectWithSSE_S3(c *C) {
	// 上传加密对象，SSE-S3
	object := randLowStr(10)
	createFile(object, 1024*1024*1)
	fd, _ := os.Open(object)
	defer fd.Close()
	sseResp, err := client.PutObjectWithContext(context.Background(), &s3.PutObjectInput{
		Bucket:               aws.String(bucket),
		Key:                  aws.String(object),
		Body:                 fd,
		ServerSideEncryption: aws.String("AES256"),
	})
	c.Assert(err, IsNil)
	c.Assert(*sseResp.ServerSideEncryption, Equals, "AES256")
	// head
	headResp, err := client.HeadObjectWithContext(context.Background(), &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(err, IsNil)
	c.Assert(*headResp.ServerSideEncryption, Equals, "AES256")
	// get
	getResp, err := client.GetObjectWithContext(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(err, IsNil)
	c.Assert(*getResp.ServerSideEncryption, Equals, "AES256")
	// delete
	_, err = client.DeleteObjectWithContext(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(err, IsNil)
	os.Remove(object)
}

// PUT Object with SSE-C
func (s *Ks3utilCommandSuite) TestPutObjectWithSSE_C(c *C) {
	// 上传加密对象，SSE-C
	object := randLowStr(10)
	createFile(object, 1024*1024*1)
	fd, _ := os.Open(object)
	defer fd.Close()
	sseResp, err := client.PutObjectWithContext(context.Background(), &s3.PutObjectInput{
		Bucket:               aws.String(bucket),
		Key:                  aws.String(object),
		Body:                 fd,
		SSECustomerAlgorithm: aws.String("AES256"),
		SSECustomerKey:       aws.String(s3util.GetBase64Str(customerKey)),
		SSECustomerKeyMD5:    aws.String(s3util.GetBase64MD5Str(customerKey)),
	})
	c.Assert(err, IsNil)
	c.Assert(*sseResp.SSECustomerAlgorithm, Equals, "AES256")
	c.Assert(*sseResp.SSECustomerKeyMD5, Equals, s3util.GetBase64MD5Str(customerKey))
	// head
	headResp, err := client.HeadObjectWithContext(context.Background(), &s3.HeadObjectInput{
		Bucket:               aws.String(bucket),
		Key:                  aws.String(object),
		SSECustomerAlgorithm: aws.String("AES256"),
		SSECustomerKey:       aws.String(s3util.GetBase64Str(customerKey)),
		SSECustomerKeyMD5:    aws.String(s3util.GetBase64MD5Str(customerKey)),
	})
	c.Assert(err, IsNil)
	c.Assert(*headResp.SSECustomerAlgorithm, Equals, "AES256")
	c.Assert(*sseResp.SSECustomerKeyMD5, Equals, s3util.GetBase64MD5Str(customerKey))
	// get
	getResp, err := client.GetObjectWithContext(context.Background(), &s3.GetObjectInput{
		Bucket:               aws.String(bucket),
		Key:                  aws.String(object),
		SSECustomerAlgorithm: aws.String("AES256"),
		SSECustomerKey:       aws.String(s3util.GetBase64Str(customerKey)),
		SSECustomerKeyMD5:    aws.String(s3util.GetBase64MD5Str(customerKey)),
	})
	c.Assert(err, IsNil)
	c.Assert(*getResp.SSECustomerAlgorithm, Equals, "AES256")
	c.Assert(*sseResp.SSECustomerKeyMD5, Equals, s3util.GetBase64MD5Str(customerKey))
	// delete
	_, err = client.DeleteObjectWithContext(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(err, IsNil)
	os.Remove(object)
}

// Copy Object with SSE-S3
func (s *Ks3utilCommandSuite) TestCopyObjectWithSSE_S3(c *C) {
	// 上传加密对象，SSE-S3
	object := randLowStr(10)
	createFile(object, 1024*1024*1)
	fd, _ := os.Open(object)
	defer fd.Close()
	// put
	_, err := client.PutObjectWithContext(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
		Body:   fd,
	})
	c.Assert(err, IsNil)
	dstObject := randLowStr(10)
	// put copy
	copyResp, err := client.CopyObjectWithContext(context.Background(), &s3.CopyObjectInput{
		Bucket:               aws.String(bucket),
		Key:                  aws.String(dstObject),
		CopySource:           aws.String("/" + bucket + "/" + object),
		ServerSideEncryption: aws.String("AES256"),
	})
	c.Assert(err, IsNil)
	c.Assert(*copyResp.ServerSideEncryption, Equals, "AES256")
	// delete
	_, err = client.DeleteObjectWithContext(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(err, IsNil)
	_, err = client.DeleteObjectWithContext(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(dstObject),
	})
	c.Assert(err, IsNil)
	os.Remove(object)
}

// Copy Object with SSE-C
func (s *Ks3utilCommandSuite) TestCopyObjectWithSSE_C(c *C) {
	// 上传加密对象，SSE-C
	object := randLowStr(10)
	createFile(object, 1024*1024*1)
	fd, _ := os.Open(object)
	defer fd.Close()
	// put
	_, err := client.PutObjectWithContext(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
		Body:   fd,
	})
	c.Assert(err, IsNil)
	dstObject := randLowStr(10)
	// put copy
	copyResp, err := client.CopyObjectWithContext(context.Background(), &s3.CopyObjectInput{
		Bucket:               aws.String(bucket),
		Key:                  aws.String(dstObject),
		CopySource:           aws.String("/" + bucket + "/" + object),
		SSECustomerAlgorithm: aws.String("AES256"),
		SSECustomerKey:       aws.String(s3util.GetBase64Str(customerKey)),
		SSECustomerKeyMD5:    aws.String(s3util.GetBase64MD5Str(customerKey)),
	})
	c.Assert(err, IsNil)
	c.Assert(*copyResp.SSECustomerAlgorithm, Equals, "AES256")
	c.Assert(*copyResp.SSECustomerKeyMD5, Equals, s3util.GetBase64MD5Str(customerKey))
	// delete
	_, err = client.DeleteObjectWithContext(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(err, IsNil)
	_, err = client.DeleteObjectWithContext(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(dstObject),
	})
	c.Assert(err, IsNil)
	os.Remove(object)
}

// Multipart Upload with SSE-S3
func (s *Ks3utilCommandSuite) TestMultipartUploadWithSSE_S3(c *C) {
	// 上传加密对象，SSE-S3
	object := randLowStr(10)
	createFile(object, 1024*1024*12)
	fd, _ := os.Open(object)
	defer fd.Close()
	// init
	initRet, err := client.CreateMultipartUploadWithContext(context.Background(), &s3.CreateMultipartUploadInput{
		Bucket:               aws.String(bucket),
		Key:                  aws.String(object),
		ServerSideEncryption: aws.String("AES256"),
	})
	c.Assert(err, IsNil)
	c.Assert(*initRet.ServerSideEncryption, Equals, "AES256")
	// 获取分块上传Id
	uploadId := *initRet.UploadID
	var i int64 = 1
	// 待合并分块
	compParts := []*s3.CompletedPart{}
	partsNum := []int64{0}
	// 缓冲区，分块大小为5MB
	buffer := make([]byte, 5*1024*1024)
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
				PartNumber: aws.Long(i),
				UploadID:   aws.String(uploadId),
				Body:       bytes.NewReader(buffer[:n]),
			})
			c.Assert(err, IsNil)
			c.Assert(*resp.ServerSideEncryption, Equals, "AES256")
			partsNum = append(partsNum, i)
			compParts = append(compParts, &s3.CompletedPart{PartNumber: &partsNum[i], ETag: resp.ETag})
			i++
		}
	}
	// complete
	comResp, err := client.CompleteMultipartUploadWithContext(context.Background(), &s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(bucket),
		Key:      aws.String(object),
		UploadID: aws.String(uploadId),
		MultipartUpload: &s3.CompletedMultipartUpload{
			Parts: compParts,
		},
	})
	c.Assert(err, IsNil)
	c.Assert(*comResp.ServerSideEncryption, Equals, "AES256")
	// delete
	_, err = client.DeleteObjectWithContext(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(err, IsNil)
	os.Remove(object)
}

// Multipart Upload with SSE-C
func (s *Ks3utilCommandSuite) TestMultipartUploadWithSSE_C(c *C) {
	// 上传加密对象，SSE-C
	object := randLowStr(10)
	createFile(object, 1024*1024*12)
	fd, _ := os.Open(object)
	defer fd.Close()
	// init
	initRet, err := client.CreateMultipartUploadWithContext(context.Background(), &s3.CreateMultipartUploadInput{
		Bucket:               aws.String(bucket),
		Key:                  aws.String(object),
		SSECustomerAlgorithm: aws.String("AES256"),
		SSECustomerKey:       aws.String(s3util.GetBase64Str(customerKey)),
		SSECustomerKeyMD5:    aws.String(s3util.GetBase64MD5Str(customerKey)),
	})
	c.Assert(err, IsNil)
	c.Assert(*initRet.SSECustomerAlgorithm, Equals, "AES256")
	c.Assert(*initRet.SSECustomerKeyMD5, Equals, s3util.GetBase64MD5Str(customerKey))
	// 获取分块上传Id
	uploadId := *initRet.UploadID
	var i int64 = 1
	// 待合并分块
	compParts := []*s3.CompletedPart{}
	partsNum := []int64{0}
	// 缓冲区，分块大小为5MB
	buffer := make([]byte, 5*1024*1024)
	for {
		n, err := fd.Read(buffer)
		if err != nil && err != io.EOF {
			panic(err)
		} else if n == 0 {
			break
		} else {
			// upload part，不通过context取消
			resp, err := client.UploadPartWithContext(context.Background(), &s3.UploadPartInput{
				Bucket:               aws.String(bucket),
				Key:                  aws.String(object),
				PartNumber:           aws.Long(i),
				UploadID:             aws.String(uploadId),
				Body:                 bytes.NewReader(buffer[:n]),
				SSECustomerAlgorithm: aws.String("AES256"),
				SSECustomerKey:       aws.String(s3util.GetBase64Str(customerKey)),
				SSECustomerKeyMD5:    aws.String(s3util.GetBase64MD5Str(customerKey)),
			})
			c.Assert(err, IsNil)
			c.Assert(*resp.SSECustomerAlgorithm, Equals, "AES256")
			c.Assert(*resp.SSECustomerKeyMD5, Equals, s3util.GetBase64MD5Str(customerKey))
			partsNum = append(partsNum, i)
			compParts = append(compParts, &s3.CompletedPart{PartNumber: &partsNum[i], ETag: resp.ETag})
			i++
		}
	}
	// complete
	comResp, err := client.CompleteMultipartUploadWithContext(context.Background(), &s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(bucket),
		Key:      aws.String(object),
		UploadID: aws.String(uploadId),
		MultipartUpload: &s3.CompletedMultipartUpload{
			Parts: compParts,
		},
	})
	c.Assert(err, IsNil)
	c.Assert(*comResp.SSECustomerAlgorithm, Equals, "AES256")
	c.Assert(*comResp.SSECustomerKeyMD5, Equals, s3util.GetBase64MD5Str(customerKey))
	// delete
	_, err = client.DeleteObjectWithContext(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(err, IsNil)
	os.Remove(object)
}

// Append Object with SSE-S3
func (s *Ks3utilCommandSuite) TestAppendObjectWithSSE_S3(c *C) {
	// 上传加密对象，SSE-S3
	object := randLowStr(10)
	createFile(object, 1024*1024*1)
	fd, _ := os.Open(object)
	defer fd.Close()
	sseResp, err := client.AppendObjectWithContext(context.Background(), &s3.AppendObjectInput{
		Bucket:               aws.String(bucket),
		Key:                  aws.String(object),
		Position:             aws.Long(0),
		Body:                 fd,
		ServerSideEncryption: aws.String("AES256"),
	})
	c.Assert(err, IsNil)
	c.Assert(*sseResp.ServerSideEncryption, Equals, "AES256")
	// head
	headResp, err := client.HeadObjectWithContext(context.Background(), &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(err, IsNil)
	c.Assert(*headResp.ServerSideEncryption, Equals, "AES256")
	// get
	getResp, err := client.GetObjectWithContext(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(err, IsNil)
	c.Assert(*getResp.ServerSideEncryption, Equals, "AES256")
	// delete
	_, err = client.DeleteObjectWithContext(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(err, IsNil)
	os.Remove(object)
}

// Append Object with SSE-C
func (s *Ks3utilCommandSuite) TestAppendObjectWithSSE_C(c *C) {
	// 上传加密对象，SSE-C
	object := randLowStr(10)
	createFile(object, 1024*1024*1)
	fd, _ := os.Open(object)
	defer fd.Close()
	sseResp, err := client.AppendObjectWithContext(context.Background(), &s3.AppendObjectInput{
		Bucket:               aws.String(bucket),
		Key:                  aws.String(object),
		Position:             aws.Long(0),
		Body:                 fd,
		SSECustomerAlgorithm: aws.String("AES256"),
		SSECustomerKey:       aws.String(s3util.GetBase64Str(customerKey)),
		SSECustomerKeyMD5:    aws.String(s3util.GetBase64MD5Str(customerKey)),
	})
	c.Assert(err, IsNil)
	c.Assert(*sseResp.SSECustomerAlgorithm, Equals, "AES256")
	c.Assert(*sseResp.SSECustomerKeyMD5, Equals, s3util.GetBase64MD5Str(customerKey))
	// head
	headResp, err := client.HeadObjectWithContext(context.Background(), &s3.HeadObjectInput{
		Bucket:               aws.String(bucket),
		Key:                  aws.String(object),
		SSECustomerAlgorithm: aws.String("AES256"),
		SSECustomerKey:       aws.String(s3util.GetBase64Str(customerKey)),
		SSECustomerKeyMD5:    aws.String(s3util.GetBase64MD5Str(customerKey)),
	})
	c.Assert(err, IsNil)
	c.Assert(*headResp.SSECustomerAlgorithm, Equals, "AES256")
	c.Assert(*sseResp.SSECustomerKeyMD5, Equals, s3util.GetBase64MD5Str(customerKey))
	// get
	getResp, err := client.GetObjectWithContext(context.Background(), &s3.GetObjectInput{
		Bucket:               aws.String(bucket),
		Key:                  aws.String(object),
		SSECustomerAlgorithm: aws.String("AES256"),
		SSECustomerKey:       aws.String(s3util.GetBase64Str(customerKey)),
		SSECustomerKeyMD5:    aws.String(s3util.GetBase64MD5Str(customerKey)),
	})
	c.Assert(err, IsNil)
	c.Assert(*getResp.SSECustomerAlgorithm, Equals, "AES256")
	c.Assert(*sseResp.SSECustomerKeyMD5, Equals, s3util.GetBase64MD5Str(customerKey))
	// delete
	_, err = client.DeleteObjectWithContext(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(object),
	})
	c.Assert(err, IsNil)
	os.Remove(object)
}
