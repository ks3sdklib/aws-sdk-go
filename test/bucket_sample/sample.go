package bucket_sample

import (
	"aws-sdk-go/aws/awserr"
	"fmt"
	"github.com/ks3sdklib/aws-sdk-go/aws"
	"github.com/ks3sdklib/aws-sdk-go/aws/credentials"
	"github.com/ks3sdklib/aws-sdk-go/service/s3"
	"github.com/stretchr/testify/assert"
	"testing"
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
	Endpoint:         "ks3-cn-beijing.ksyuncs.com",
	DisableSSL:       true,
	LogLevel:         1,
	S3ForcePathStyle: true,
	LogHTTPBody:      true,
})

func TestCreateBucket(t *testing.T) {
	_, err := svc.CreateBucket(&s3.CreateBucketInput{
		ACL:    aws.String("public-read"),
		Bucket: aws.String(bucket),
		ProjectId: aws.String("1232"),
	})
	assert.Error(t, err)
	assert.Equal(t, "BucketAlreadyExists", err)
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
	_, err := svc.DeleteBucket(&s3.DeleteBucketInput{
		Bucket: aws.String(bucket),
	})
	assert.Error(t, err)
}


//设置镜像回源规则
func PutBucketMirrorRules(client *s3.S3) {

	params := &s3.PutBucketMirrorInput{
		Bucket: aws.String(bucket), // Required
		BucketMirror: &s3.BucketMirror{
			Version:          aws.String("V3"),
			UseDefaultRobots: aws.Boolean(false),
			AsyncMirrorRule: &s3.AsyncMirrorRule{
				MirrorUrls: []*string{
					aws.String("http://abc.om"),
					aws.String("http://abc.om"),
				},
				SavingSetting: &s3.SavingSetting{
					ACL: aws.String("private"),
				},
			},
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
						PassQueryString: aws.Boolean(false),
						Follow3Xx:       aws.Boolean(false),
						HeaderSetting: &s3.HeaderSetting{
							SetHeaders: []*s3.SetHeaders{
								{
									Key:   aws.String("a"),
									Value: aws.String("b"),
								},
							},
							RemoveHeaders: []*s3.RemoveHeaders{
								{
									Key: aws.String("c"),
								},
								{
									Key: aws.String("d"),
								},
							},
							PassAll: aws.Boolean(false),
							PassHeaders: []*s3.PassHeaders{
								{
									Key: aws.String("key"),
								},
							},
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

//获取镜像回源规则
func GetBucketMirrorRules(client *s3.S3) {

	params := &s3.GetBucketMirrorInput{
		Bucket: aws.String(bucket),
	}
	resp, _ := client.GetBucketMirror(params)
	fmt.Println("resp.code is:", resp.HttpCode)
	fmt.Println("resp.Header is:", resp.Header)
	// Pretty-print the response data.
	var bodyStr = string(resp.Body[:])
	fmt.Println("resp.Body is:", bodyStr)

}

//删除镜像回源规则
func DeleteBucketMirrorRules(client *s3.S3) {

	params := &s3.DeleteBucketMirrorInput{
		Bucket: aws.String(bucket),
	}
	resp, _ := client.DeleteBucketMirror(params)
	fmt.Println("resp.code is:", resp.HttpCode)
	fmt.Println("resp.Header is:", resp.Header)
	// Pretty-print the response data.
	var bodyStr = string(resp.Body[:])
	fmt.Println("resp.Body is:", bodyStr)

}