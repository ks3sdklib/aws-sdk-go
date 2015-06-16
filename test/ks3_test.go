package ks3test

import (
	"testing"
	"strings"
	"github.com/ks3sdklib/aws-sdk-go/aws"
	"github.com/ks3sdklib/aws-sdk-go/internal/apierr"
	"github.com/ks3sdklib/aws-sdk-go/aws/credentials"
	"github.com/ks3sdklib/aws-sdk-go/service/s3"
	"github.com/stretchr/testify/assert"
)
var bucket =string("aa-go-sdk")
var key = string("中文/test.go")
var cre = credentials.NewStaticCredentials("lMQTr0hNlMpB0iOk/i+x","D4CsYLs75JcWEjbiI22zR3P7kJ/+5B1qdEje7A7I","")
var svc = s3.New(&aws.Config{
		Region: "HANGZHOU",
		Credentials: cre,
		Endpoint:"kss.ksyun.com",
		DisableSSL:true,
		LogLevel:1,
		S3ForcePathStyle:true,
		LogHTTPBody:true,
		})

func TestCreateBucket(t *testing.T){
	_,err := svc.CreateBucket(&s3.CreateBucketInput{
		ACL:aws.String("public-read"),
		Bucket:aws.String(bucket),
		})
	assert.Error(t,err)
	assert.Equal(t,"BucketAlreadyExists",err.(*apierr.RequestError).Code())	
}
func TestBucketAcl(t *testing.T){
	_,err := svc.PutBucketACL(&s3.PutBucketACLInput{
		Bucket:aws.String(bucket),
		ACL:aws.String("public-read"),
		})
	assert.NoError(t,err)

	acp,err := svc.GetBucketACL(&s3.GetBucketACLInput{
		Bucket:aws.String(bucket),
		})
	assert.NoError(t,err)
	grants := acp.Grants
	assert.Equal(t,2,len(grants),"size of grants")

	foundFull := false
	foundRead := false
	for i:=0;i <len(grants);i++{
		grant := grants[i]
		if *grant.Permission == "FULL_CONTROL"{
			foundFull = true
			assert.NotNil(t,*grant.Grantee.ID,"grantee userid should not null")
			assert.NotNil(t,*grant.Grantee.DisplayName,"grantee displayname should not null")
		}else if *grant.Permission == "READ"{
			foundRead = true
			assert.NotNil(t,*grant.Grantee.URI,"grantee uri should not null")
		}
	}
	assert.True(t,foundRead,"acp should contains READ")
	assert.True(t,foundFull,"acp should contains FULL_CONTROL")

	_,putaclErr := svc.PutBucketACL(&s3.PutBucketACLInput{
		Bucket:aws.String(bucket),
		ACL:aws.String("private"),
		})
	assert.NoError(t,putaclErr)

	acp,getaclErr := svc.GetBucketACL(&s3.GetBucketACLInput{
		Bucket:aws.String(bucket),
		})
	assert.NoError(t,getaclErr)
	privategrants := acp.Grants
	assert.Equal(t,1,len(privategrants),"size of grants")
}
func TestListBuckets(t *testing.T) {
	out,err := svc.ListBuckets(nil)
	assert.NoError(t,err)
	buckets := out.Buckets
	found := false
	for i:=0;i<len(buckets);i++{
		if *buckets[i].Name == bucket{
			found = true
		}
	}
	assert.True(t,found,"list buckets expected contains "+bucket)
}
func TestHeadBucket(t *testing.T) {
	_,err := svc.HeadBucket(&s3.HeadBucketInput{
		Bucket:aws.String(bucket),
	})
	assert.NoError(t,err)
}
func TestDeleteBucket(t *testing.T) {
	putObjectSimple()
	_,err := svc.DeleteBucket(&s3.DeleteBucketInput{
		Bucket:aws.String(bucket),
	})
	assert.Error(t,err)
	assert.Equal(t,"BucketNotEmpty",err.(*apierr.RequestError).Code())	
}
func TestListObjects(t *testing.T) {
	putObjectSimple()
	objects,err := svc.ListObjects(&s3.ListObjectsInput{
		Bucket:aws.String(bucket),
		Delimiter:aws.String("/"),
		MaxKeys:aws.Long(10),
		Prefix:aws.String(""),
	})
	assert.NoError(t,err)
	assert.Equal(t,"/",*objects.Delimiter)
	assert.Equal(t,*aws.Long(10),*objects.MaxKeys)
	assert.Equal(t,"",*objects.Prefix)
	assert.False(t,*objects.IsTruncated)

	objects1,err := svc.ListObjects(&s3.ListObjectsInput{
		Bucket:aws.String(bucket),
		})
	assert.NoError(t,err)
	objectList := objects1.Contents
	found := false
	for i:=0;i <len(objectList);i++{
		object := objectList[i]
		assert.Equal(t,"STANDARD",*object.StorageClass)
		if *object.Key == key{
			found = true
		}
	}
	assert.True(t,found,"expected found "+key+"in object listing")
}
func putObjectSimple() {
	svc.PutObject(
		&s3.PutObjectInput{
			Bucket:aws.String(bucket),
			Key:aws.String(key),
			Body:strings.NewReader("content"),
		},
	)
}