package lib

import (
	"github.com/ks3sdklib/aws-sdk-go/aws"
	"github.com/ks3sdklib/aws-sdk-go/service/s3"
	. "gopkg.in/check.v1"
)

// 创建bucket并关联项目
func (s *Ks3utilCommandSuite) TestBucket(c *C) {
	// 创建bucket
	bucketName := commonNamePrefix + randLowStr(10)
	_, err := client.CreateBucket(&s3.CreateBucketInput{
		ACL:    aws.String("public-read"),
		Bucket: aws.String(bucketName),
		//ProjectId:  aws.String("1232"), //项目ID
		BucketType: aws.String(s3.BucketTypeIA),
	})
	c.Assert(err, IsNil)

	// 判断bucket桶是否存在
	exist, err := client.HeadBucketExist(bucketName)
	c.Assert(err, IsNil)
	c.Assert(exist, Equals, true)

	// 获取bucket信息
	_, err = client.HeadBucket(&s3.HeadBucketInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, IsNil)

	// 删除bucket
	_, err = client.DeleteBucket(&s3.DeleteBucketInput{
		Bucket: aws.String(bucketName),
	})
	c.Assert(err, IsNil)
}

// 获取bucket列表
func (s *Ks3utilCommandSuite) TestListBuckets(c *C) {
	_, err := client.ListBuckets(&s3.ListBucketsInput{})
	c.Assert(err, IsNil)
}

// bucket acl
func (s *Ks3utilCommandSuite) TestBucketAcl(c *C) {
	_, err := client.PutBucketACL(&s3.PutBucketACLInput{
		Bucket: aws.String(bucket),
		ACL:    aws.String(s3.ACLPublicRead),
	})
	c.Assert(err, IsNil)

	resp, err := client.GetBucketACL(&s3.GetBucketACLInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, IsNil)
	c.Assert(s3.GetBucketAcl(*resp), Equals, s3.PublicRead)
}

// bucket lifecycle
func (s *Ks3utilCommandSuite) TestBucketLifecycle(c *C) {
	// 配置生命周期规则
	lifecycleConfiguration := &s3.LifecycleConfiguration{
		Rules: []*s3.LifecycleRule{
			{
				ID:     aws.String("rule1"),
				Status: aws.String("Enabled"),
				Expiration: &s3.LifecycleExpiration{
					Days: aws.Long(30),
				},
			},
		},
	}
	_, err := client.PutBucketLifecycle(&s3.PutBucketLifecycleInput{
		Bucket:                 aws.String(bucket),
		LifecycleConfiguration: lifecycleConfiguration,
	})
	c.Assert(err, IsNil)

	// 获取生命周期规则
	_, err = client.GetBucketLifecycle(&s3.GetBucketLifecycleInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, IsNil)

	// 删除生命周期规则
	_, err = client.DeleteBucketLifecycle(&s3.DeleteBucketLifecycleInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, IsNil)
}

// bucket cors
func (s *Ks3utilCommandSuite) TestBucketCors(c *C) {
	// 配置CORS规则
	corsConfiguration := &s3.CORSConfiguration{
		Rules: []*s3.CORSRule{
			{
				AllowedHeader: []string{
					"*",
				},
				AllowedMethod: []string{
					"GET",
				},
				AllowedOrigin: []string{
					"*",
				},
				MaxAgeSeconds: 100,
			},
		},
	}
	// 设置桶的CORS配置
	_, err := client.PutBucketCORS(&s3.PutBucketCORSInput{
		Bucket:            aws.String(bucket),
		CORSConfiguration: corsConfiguration,
	})
	c.Assert(err, IsNil)

	// 获取桶的CORS配置
	_, err = client.GetBucketCORS(&s3.GetBucketCORSInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, IsNil)

	// 删除桶的CORS配置
	_, err = client.DeleteBucketCORS(&s3.DeleteBucketCORSInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, IsNil)
}

// bucket log
func (s *Ks3utilCommandSuite) TestSetBucketLog(c *C) {
	logStatus := &s3.BucketLoggingStatus{
		LoggingEnabled: &s3.LoggingEnabled{
			TargetBucket: aws.String(bucket),
			TargetPrefix: aws.String(bucket),
		},
	}
	_, err := client.PutBucketLogging(&s3.PutBucketLoggingInput{
		Bucket:              aws.String(bucket),
		BucketLoggingStatus: logStatus,
	})
	c.Assert(err, IsNil)

	_, err = client.GetBucketLogging(&s3.GetBucketLoggingInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, IsNil)
}

// bucket mirror
// 详情见API(https://docs.ksyun.com/documents/39134)
func (s *Ks3utilCommandSuite) TestBucketMirror(c *C) {
	bucketMirror := &s3.BucketMirror{
		Version:          aws.String("V3"),   //回源类型
		UseDefaultRobots: aws.Boolean(false), //是否使用默认的robots.txt，如果为true则会在bucket下生成一个robots.txt。
		//异步回源规则，该字段与sync_mirror_rules必须至少有一个，可同时存在。
		AsyncMirrorRule: &s3.AsyncMirrorRule{
			//一组源站url，数量不超过10个，url必须以http或者https开头，域名部分最多不超过256个字符，path部分最多不超过1024个字符。
			MirrorUrls: []*string{
				aws.String("http://abc.om"),
				aws.String("http://abc.om"),
			},
			SavingSetting: &s3.SavingSetting{
				ACL: aws.String("private"),
			},
		},
		//一组同步回源规则，最多可配置20个。该字段与async_mirror_rule必须至少有一个，可同时存在。
		SyncMirrorRules: []*s3.SyncMirrorRules{
			{
				//回源触发条件，可不填，不填表示对该bucket中不存在的object发送get请求时，将会触发回源。
				MatchCondition: s3.MatchCondition{
					//触发回源的http状态码，目前仅支持404一种。
					HTTPCodes: []*string{
						aws.String("404"),
					},
					//当请求的object key的前缀与任意一个key_prefix匹配时触发回源，仅支持一个前缀
					KeyPrefixes: []*string{
						aws.String("abc"),
					},
				},
				//源站url,必须以http或者https开头，域名部分最多不超过256个字符，path部分最多不超过1024个字符。
				MirrorURL: aws.String("http://v-ks-a-i.originalvod.com"),
				//ks3请求源站时的配置，可不填。
				MirrorRequestSetting: &s3.MirrorRequestSetting{
					//ks3请求源站时是否将客户端请求ks3时的query string透传给源站。
					PassQueryString: aws.Boolean(false),
					//设置访问源站时，是否follow 302/301。
					Follow3Xx: aws.Boolean(false),
					//ks3请求源站时的header配置，注意以下的属性有优先级:set_headers > remove_headers > pass_all > pass_headers。
					HeaderSetting: &s3.HeaderSetting{
						//自定义header，这些header的key和value均是固定的，ks3请求源站时会带上这些header。
						SetHeaders: []*s3.SetHeaders{
							{
								Key:   aws.String("a"),
								Value: aws.String("b"),
							},
						},
						//从客户端发给ks3的header中移除以下指定的header，通常与pass_all或者pass_headers配合使用，只能指定header中的key，不能指定value
						RemoveHeaders: []*s3.RemoveHeaders{
							{
								Key: aws.String("c"),
							},
							{
								Key: aws.String("d"),
							},
						},
						//将客户端发给ks3的header全部透传给源站，该字段与pass_headers互斥。
						PassAll: aws.Boolean(false),
						//将客户端发给ks3的header中指定的几个透传给源站，只能指定header中的key，不能指定value。
						PassHeaders: []*s3.PassHeaders{
							{
								Key: aws.String("key"),
							},
						},
					},
				},
				//
				SavingSetting: &s3.SavingSetting{
					ACL: aws.String("private"),
				},
			},
		},
	}

	// 设置桶的镜像回源规则
	_, err := client.PutBucketMirror(&s3.PutBucketMirrorInput{
		Bucket:       aws.String(bucket), // Required
		BucketMirror: bucketMirror,
	})
	c.Assert(err, IsNil)

	// 获取桶的镜像回源规则
	_, err = client.GetBucketMirror(&s3.GetBucketMirrorInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, IsNil)

	// 删除桶的镜像回源规则
	_, err = client.DeleteBucketMirror(&s3.DeleteBucketMirrorInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, IsNil)
}

// bucket policy
func (s *Ks3utilCommandSuite) TestBucketPolicy(c *C) {
	_, err := client.PutBucketPolicy(&s3.PutBucketPolicyInput{
		Bucket: aws.String(bucket),
		Policy: aws.String("{\"Statement\":[{\"Resource\":[\"krn:ksc:ks3:::" + bucket + "/中文22prefix\"],\"Action\":[\"ks3:AbortMultipartUpload\",\"ks3:DeleteObject\"],\"Principal\":{\"KSC\":[\"*\"]},\"Effect\":\"Allow\"}]}"), //项目ID
	})
	c.Assert(err, IsNil)

	_, err = client.GetBucketPolicy(&s3.GetBucketPolicyInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, IsNil)

	_, err = client.DeleteBucketPolicy(&s3.DeleteBucketPolicyInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, IsNil)
}
