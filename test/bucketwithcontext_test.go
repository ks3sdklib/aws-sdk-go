package lib

import (
	"context"
	"github.com/ks3sdklib/aws-sdk-go/aws"
	"github.com/ks3sdklib/aws-sdk-go/aws/awsutil"
	"github.com/ks3sdklib/aws-sdk-go/aws/credentials"
	"github.com/ks3sdklib/aws-sdk-go/service/s3"
	. "gopkg.in/check.v1"
	"time"
)

var bucketTimeout = time.Microsecond * 10

// PUT Bucket CORS
func (s *Ks3utilCommandSuite) TestPutBucketCORSWithContext(c *C) {
	// 配置CORS规则
	corsConfiguration := &s3.CORSConfiguration{
		Rules: []*s3.CORSRule{{
			AllowedHeader: []string{"*"},
			AllowedMethod: []string{"GET"},
			AllowedOrigin: []string{"*"},
			MaxAgeSeconds: 100},
		},
	}
	// put,不通过context取消
	_, err := client.PutBucketCORSWithContext(context.Background(), &s3.PutBucketCORSInput{
		Bucket:            aws.String(bucket),
		CORSConfiguration: corsConfiguration,
	})
	c.Assert(err, IsNil)
	// get
	resp, err := client.GetBucketCORSWithContext(context.Background(), &s3.GetBucketCORSInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, IsNil)
	c.Assert(awsutil.StringValue(resp.Rules), Equals, awsutil.StringValue(corsConfiguration.Rules))
	// delete
	_, err = client.DeleteBucketCORSWithContext(context.Background(), &s3.DeleteBucketCORSInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, IsNil)
	// put,通过context取消
	ctx, cancelFunc := context.WithTimeout(context.Background(), bucketTimeout)
	defer cancelFunc()
	_, err = client.PutBucketCORSWithContext(ctx, &s3.PutBucketCORSInput{
		Bucket:            aws.String(bucket),
		CORSConfiguration: corsConfiguration,
	})
	c.Assert(err, NotNil)
	// get
	resp, err = client.GetBucketCORSWithContext(context.Background(), &s3.GetBucketCORSInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, IsNil)
	c.Assert(awsutil.StringValue(resp.Rules), Equals, awsutil.StringValue([]string{}))
}

// GET Bucket CORS
func (s *Ks3utilCommandSuite) TestGetBucketCORSWithContext(c *C) {
	// 配置CORS规则
	corsConfiguration := &s3.CORSConfiguration{
		Rules: []*s3.CORSRule{{
			AllowedHeader: []string{"*"},
			AllowedMethod: []string{"GET"},
			AllowedOrigin: []string{"*"},
			MaxAgeSeconds: 100},
		},
	}
	// put
	_, err := client.PutBucketCORSWithContext(context.Background(), &s3.PutBucketCORSInput{
		Bucket:            aws.String(bucket),
		CORSConfiguration: corsConfiguration,
	})
	c.Assert(err, IsNil)
	// get，不通过context取消
	resp, err := client.GetBucketCORSWithContext(context.Background(), &s3.GetBucketCORSInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, IsNil)
	c.Assert(awsutil.StringValue(resp.Rules), Equals, awsutil.StringValue(corsConfiguration.Rules))
	// get，通过context取消
	ctx, cancelFunc := context.WithTimeout(context.Background(), bucketTimeout)
	defer cancelFunc()
	_, err = client.GetBucketCORSWithContext(ctx, &s3.GetBucketCORSInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, NotNil)
	// delete
	_, err = client.DeleteBucketCORSWithContext(context.Background(), &s3.DeleteBucketCORSInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, IsNil)
}

// DELETE Bucket CORS
func (s *Ks3utilCommandSuite) TestDeleteBucketCORSWithContext(c *C) {
	// 配置CORS规则
	corsConfiguration := &s3.CORSConfiguration{
		Rules: []*s3.CORSRule{{
			AllowedHeader: []string{"*"},
			AllowedMethod: []string{"GET"},
			AllowedOrigin: []string{"*"},
			MaxAgeSeconds: 100},
		},
	}
	// put
	_, err := client.PutBucketCORSWithContext(context.Background(), &s3.PutBucketCORSInput{
		Bucket:            aws.String(bucket),
		CORSConfiguration: corsConfiguration,
	})
	c.Assert(err, IsNil)
	// get
	resp, err := client.GetBucketCORSWithContext(context.Background(), &s3.GetBucketCORSInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, IsNil)
	c.Assert(awsutil.StringValue(resp.Rules), Equals, awsutil.StringValue(corsConfiguration.Rules))
	// delete，通过context取消
	ctx, cancelFunc := context.WithTimeout(context.Background(), bucketTimeout)
	defer cancelFunc()
	_, err = client.DeleteBucketCORSWithContext(ctx, &s3.DeleteBucketCORSInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, NotNil)
	// delete，不通过context取消
	_, err = client.DeleteBucketCORSWithContext(context.Background(), &s3.DeleteBucketCORSInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, IsNil)
}

// PUT Bucket Mirror
func (s *Ks3utilCommandSuite) TestPutBucketMirrorWithContext(c *C) {
	// 配置Mirror规则
	BucketMirror := &s3.BucketMirror{
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
	// put,不通过context取消
	_, err := client.PutBucketMirrorWithContext(context.Background(), &s3.PutBucketMirrorInput{
		Bucket:       aws.String(bucket),
		BucketMirror: BucketMirror,
	})
	c.Assert(err, IsNil)
	// get
	resp, err := client.GetBucketMirrorWithContext(context.Background(), &s3.GetBucketMirrorInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, IsNil)
	c.Assert(awsutil.StringValue(resp.BucketMirror), Equals, awsutil.StringValue(BucketMirror))
	// delete
	_, err = client.DeleteBucketMirrorWithContext(context.Background(), &s3.DeleteBucketMirrorInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, IsNil)
	// put,通过context取消
	ctx, cancelFunc := context.WithTimeout(context.Background(), bucketTimeout)
	defer cancelFunc()
	_, err = client.PutBucketMirrorWithContext(ctx, &s3.PutBucketMirrorInput{
		Bucket:       aws.String(bucket),
		BucketMirror: BucketMirror,
	})
	c.Assert(err, NotNil)
	// get
	resp, err = client.GetBucketMirrorWithContext(context.Background(), &s3.GetBucketMirrorInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, NotNil)
	c.Assert(awsutil.StringValue(resp.BucketMirror), Not(Equals), awsutil.StringValue(BucketMirror))
}

// GET Bucket Mirror
func (s *Ks3utilCommandSuite) TestGetBucketMirrorWithContext(c *C) {
	// 配置Mirror规则
	BucketMirror := &s3.BucketMirror{
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
	// put
	_, err := client.PutBucketMirrorWithContext(context.Background(), &s3.PutBucketMirrorInput{
		Bucket:       aws.String(bucket),
		BucketMirror: BucketMirror,
	})
	c.Assert(err, IsNil)
	// get，不通过context取消
	resp, err := client.GetBucketMirrorWithContext(context.Background(), &s3.GetBucketMirrorInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, IsNil)
	c.Assert(awsutil.StringValue(resp.BucketMirror), Equals, awsutil.StringValue(BucketMirror))
	// get，通过context取消
	ctx, cancelFunc := context.WithTimeout(context.Background(), bucketTimeout)
	defer cancelFunc()
	_, err = client.GetBucketMirrorWithContext(ctx, &s3.GetBucketMirrorInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, NotNil)
	// delete
	_, err = client.DeleteBucketMirrorWithContext(context.Background(), &s3.DeleteBucketMirrorInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, IsNil)
}

// Delete Bucket Mirror
func (s *Ks3utilCommandSuite) TestDeleteBucketMirrorWithContext(c *C) {
	// 配置Mirror规则
	BucketMirror := &s3.BucketMirror{
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
	// put
	_, err := client.PutBucketMirrorWithContext(context.Background(), &s3.PutBucketMirrorInput{
		Bucket:       aws.String(bucket),
		BucketMirror: BucketMirror,
	})
	c.Assert(err, IsNil)
	// get
	resp, err := client.GetBucketMirrorWithContext(context.Background(), &s3.GetBucketMirrorInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, IsNil)
	c.Assert(awsutil.StringValue(resp.BucketMirror), Equals, awsutil.StringValue(BucketMirror))
	// delete，通过context取消
	ctx, cancelFunc := context.WithTimeout(context.Background(), bucketTimeout)
	defer cancelFunc()
	_, err = client.DeleteBucketMirrorWithContext(ctx, &s3.DeleteBucketMirrorInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, NotNil)
	// delete，不通过context取消
	_, err = client.DeleteBucketMirrorWithContext(context.Background(), &s3.DeleteBucketMirrorInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, IsNil)
}

// PUT Bucket Lifecycle
func (s *Ks3utilCommandSuite) TestPutBucketLifecycleWithContext(c *C) {
	// 配置生命周期规则
	lifecycleConfiguration := &s3.LifecycleConfiguration{
		Rules: []*s3.LifecycleRule{
			{
				ID: aws.String("rule1"),
				Filter: &s3.LifecycleFilter{
					Prefix: aws.String("prefix1"),
				},
				Status: aws.String("Enabled"),
				Expiration: &s3.LifecycleExpiration{
					Days: aws.Long(90),
				},
				Transitions: []*s3.Transition{
					{
						StorageClass: aws.String(s3.StorageClassIA),
						Days:         aws.Long(30),
					},
				},
				AbortIncompleteMultipartUpload: &s3.AbortIncompleteMultipartUpload{
					DaysAfterInitiation: aws.Long(60),
				},
			},
		},
	}
	// put,不通过context取消
	_, err := client.PutBucketLifecycleWithContext(context.Background(), &s3.PutBucketLifecycleInput{
		Bucket:                 aws.String(bucket),
		LifecycleConfiguration: lifecycleConfiguration,
	})
	c.Assert(err, IsNil)
	// get
	resp, err := client.GetBucketLifecycleWithContext(context.Background(), &s3.GetBucketLifecycleInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, IsNil)
	c.Assert(awsutil.StringValue(resp.Rules), Equals, awsutil.StringValue(lifecycleConfiguration.Rules))
	// delete
	_, err = client.DeleteBucketLifecycleWithContext(context.Background(), &s3.DeleteBucketLifecycleInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, IsNil)
	// put,通过context取消
	ctx, cancelFunc := context.WithTimeout(context.Background(), bucketTimeout)
	defer cancelFunc()
	_, err = client.PutBucketLifecycleWithContext(ctx, &s3.PutBucketLifecycleInput{
		Bucket:                 aws.String(bucket),
		LifecycleConfiguration: lifecycleConfiguration,
	})
	c.Assert(err, NotNil)
	// get
	resp, err = client.GetBucketLifecycleWithContext(context.Background(), &s3.GetBucketLifecycleInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, NotNil)
	c.Assert(awsutil.StringValue(resp.Rules), Not(Equals), awsutil.StringValue(lifecycleConfiguration.Rules))
}

// GET Bucket Lifecycle
func (s *Ks3utilCommandSuite) TestGetBucketLifecycleWithContext(c *C) {
	// 配置生命周期规则
	lifecycleConfiguration := &s3.LifecycleConfiguration{
		Rules: []*s3.LifecycleRule{
			{
				ID: aws.String("rule1"),
				Filter: &s3.LifecycleFilter{
					Prefix: aws.String("prefix1"),
				},
				Status: aws.String("Enabled"),
				Expiration: &s3.LifecycleExpiration{
					Days: aws.Long(90),
				},
				Transitions: []*s3.Transition{
					{
						StorageClass: aws.String(s3.StorageClassIA),
						Days:         aws.Long(30),
					},
				},
				AbortIncompleteMultipartUpload: &s3.AbortIncompleteMultipartUpload{
					DaysAfterInitiation: aws.Long(60),
				},
			},
		},
	}
	// put
	_, err := client.PutBucketLifecycleWithContext(context.Background(), &s3.PutBucketLifecycleInput{
		Bucket:                 aws.String(bucket),
		LifecycleConfiguration: lifecycleConfiguration,
	})
	c.Assert(err, IsNil)
	// get，不通过context取消
	resp, err := client.GetBucketLifecycleWithContext(context.Background(), &s3.GetBucketLifecycleInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, IsNil)
	c.Assert(awsutil.StringValue(resp.Rules), Equals, awsutil.StringValue(lifecycleConfiguration.Rules))
	// get，通过context取消
	ctx, cancelFunc := context.WithTimeout(context.Background(), bucketTimeout)
	defer cancelFunc()
	_, err = client.GetBucketLifecycleWithContext(ctx, &s3.GetBucketLifecycleInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, NotNil)
	// delete
	_, err = client.DeleteBucketLifecycleWithContext(context.Background(), &s3.DeleteBucketLifecycleInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, IsNil)
}

// Delete Bucket Lifecycle
func (s *Ks3utilCommandSuite) TestDeleteBucketLifecycleWithContext(c *C) {
	// 配置生命周期规则
	lifecycleConfiguration := &s3.LifecycleConfiguration{
		Rules: []*s3.LifecycleRule{
			{
				ID: aws.String("rule1"),
				Filter: &s3.LifecycleFilter{
					Prefix: aws.String("prefix1"),
				},
				Status: aws.String("Enabled"),
				Expiration: &s3.LifecycleExpiration{
					Days: aws.Long(90),
				},
				Transitions: []*s3.Transition{
					{
						StorageClass: aws.String(s3.StorageClassIA),
						Days:         aws.Long(30),
					},
				},
				AbortIncompleteMultipartUpload: &s3.AbortIncompleteMultipartUpload{
					DaysAfterInitiation: aws.Long(60),
				},
			},
		},
	}
	// put
	_, err := client.PutBucketLifecycleWithContext(context.Background(), &s3.PutBucketLifecycleInput{
		Bucket:                 aws.String(bucket),
		LifecycleConfiguration: lifecycleConfiguration,
	})
	c.Assert(err, IsNil)
	// get
	resp, err := client.GetBucketLifecycleWithContext(context.Background(), &s3.GetBucketLifecycleInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, IsNil)
	c.Assert(awsutil.StringValue(resp.Rules), Equals, awsutil.StringValue(lifecycleConfiguration.Rules))
	// delete，通过context取消
	ctx, cancelFunc := context.WithTimeout(context.Background(), bucketTimeout)
	defer cancelFunc()
	_, err = client.DeleteBucketLifecycleWithContext(ctx, &s3.DeleteBucketLifecycleInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, NotNil)
	// delete，不通过context取消
	_, err = client.DeleteBucketLifecycleWithContext(context.Background(), &s3.DeleteBucketLifecycleInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, IsNil)
}

// PUT Bucket Policy
func (s *Ks3utilCommandSuite) TestPutBucketPolicyWithContext(c *C) {
	// 配置Policy规则
	policy := "{\"Statement\":[{\"Resource\":[\"krn:ksc:ks3:::" + bucket + "/中文22prefix\"],\"Action\":[\"ks3:AbortMultipartUpload\",\"ks3:DeleteObject\"],\"Principal\":{\"KSC\":[\"*\"]},\"Effect\":\"Allow\"}]}"
	// put,不通过context取消
	_, err := client.PutBucketPolicyWithContext(context.Background(), &s3.PutBucketPolicyInput{
		Bucket: aws.String(bucket),
		Policy: aws.String(policy),
	})
	c.Assert(err, IsNil)
	// get
	resp, err := client.GetBucketPolicyWithContext(context.Background(), &s3.GetBucketPolicyInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, IsNil)
	c.Assert(*resp.Policy, Equals, policy)
	// delete
	_, err = client.DeleteBucketPolicyWithContext(context.Background(), &s3.DeleteBucketPolicyInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, IsNil)
	// put,通过context取消
	ctx, cancelFunc := context.WithTimeout(context.Background(), bucketTimeout)
	defer cancelFunc()
	_, err = client.PutBucketPolicyWithContext(ctx, &s3.PutBucketPolicyInput{
		Bucket: aws.String(bucket),
		Policy: aws.String(policy),
	})
	c.Assert(err, NotNil)
	// get
	resp, err = client.GetBucketPolicyWithContext(context.Background(), &s3.GetBucketPolicyInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, NotNil)
}

// GET Bucket Policy
func (s *Ks3utilCommandSuite) TestGetBucketPolicyWithContext(c *C) {
	// 配置Policy规则
	policy := "{\"Statement\":[{\"Resource\":[\"krn:ksc:ks3:::" + bucket + "/中文22prefix\"],\"Action\":[\"ks3:AbortMultipartUpload\",\"ks3:DeleteObject\"],\"Principal\":{\"KSC\":[\"*\"]},\"Effect\":\"Allow\"}]}"
	// put
	_, err := client.PutBucketPolicyWithContext(context.Background(), &s3.PutBucketPolicyInput{
		Bucket: aws.String(bucket),
		Policy: aws.String(policy),
	})
	c.Assert(err, IsNil)
	// get，不通过context取消
	resp, err := client.GetBucketPolicyWithContext(context.Background(), &s3.GetBucketPolicyInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, IsNil)
	c.Assert(*resp.Policy, Equals, policy)
	// get，通过context取消
	ctx, cancelFunc := context.WithTimeout(context.Background(), bucketTimeout)
	defer cancelFunc()
	_, err = client.GetBucketPolicyWithContext(ctx, &s3.GetBucketPolicyInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, NotNil)
	// delete
	_, err = client.DeleteBucketPolicyWithContext(context.Background(), &s3.DeleteBucketPolicyInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, IsNil)
}

// Delete Bucket Policy
func (s *Ks3utilCommandSuite) TestDeleteBucketPolicyWithContext(c *C) {
	// 配置Policy规则
	policy := "{\"Statement\":[{\"Resource\":[\"krn:ksc:ks3:::" + bucket + "/中文22prefix\"],\"Action\":[\"ks3:AbortMultipartUpload\",\"ks3:DeleteObject\"],\"Principal\":{\"KSC\":[\"*\"]},\"Effect\":\"Allow\"}]}"
	// put
	_, err := client.PutBucketPolicyWithContext(context.Background(), &s3.PutBucketPolicyInput{
		Bucket: aws.String(bucket),
		Policy: aws.String(policy),
	})
	c.Assert(err, IsNil)
	// get
	resp, err := client.GetBucketPolicyWithContext(context.Background(), &s3.GetBucketPolicyInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, IsNil)
	c.Assert(*resp.Policy, Equals, policy)
	// delete，通过context取消
	ctx, cancelFunc := context.WithTimeout(context.Background(), bucketTimeout)
	defer cancelFunc()
	_, err = client.DeleteBucketPolicyWithContext(ctx, &s3.DeleteBucketPolicyInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, NotNil)
	// delete，不通过context取消
	_, err = client.DeleteBucketPolicyWithContext(context.Background(), &s3.DeleteBucketPolicyInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, IsNil)
}

// PUT Bucket Logging
func (s *Ks3utilCommandSuite) TestPutBucketLoggingWithContext(c *C) {
	logStatus := &s3.BucketLoggingStatus{
		LoggingEnabled: &s3.LoggingEnabled{
			TargetBucket: aws.String(bucket),
			TargetPrefix: aws.String(bucket),
		},
	}
	// put,不通过context取消
	_, err := client.PutBucketLoggingWithContext(context.Background(), &s3.PutBucketLoggingInput{
		Bucket:              aws.String(bucket),
		BucketLoggingStatus: logStatus,
	})
	c.Assert(err, IsNil)
	// get
	resp, err := client.GetBucketLoggingWithContext(context.Background(), &s3.GetBucketLoggingInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, IsNil)
	c.Assert(awsutil.StringValue(resp.LoggingEnabled), Equals, awsutil.StringValue(logStatus.LoggingEnabled))
	// put,通过context取消
	ctx, cancelFunc := context.WithTimeout(context.Background(), bucketTimeout)
	defer cancelFunc()
	_, err = client.PutBucketLoggingWithContext(ctx, &s3.PutBucketLoggingInput{
		Bucket:              aws.String(bucket),
		BucketLoggingStatus: logStatus,
	})
	c.Assert(err, NotNil)
	// get
	resp, err = client.GetBucketLoggingWithContext(context.Background(), &s3.GetBucketLoggingInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, IsNil)
	c.Assert(awsutil.StringValue(resp.LoggingEnabled), Equals, awsutil.StringValue(logStatus.LoggingEnabled))
}

// GET Bucket Logging
func (s *Ks3utilCommandSuite) TestGetBucketLoggingWithContext(c *C) {
	logStatus := &s3.BucketLoggingStatus{
		LoggingEnabled: &s3.LoggingEnabled{
			TargetBucket: aws.String(bucket),
			TargetPrefix: aws.String(bucket),
		},
	}
	// put
	_, err := client.PutBucketLoggingWithContext(context.Background(), &s3.PutBucketLoggingInput{
		Bucket:              aws.String(bucket),
		BucketLoggingStatus: logStatus,
	})
	c.Assert(err, IsNil)
	// get,不通过context取消
	resp, err := client.GetBucketLoggingWithContext(context.Background(), &s3.GetBucketLoggingInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, IsNil)
	c.Assert(awsutil.StringValue(resp.LoggingEnabled), Equals, awsutil.StringValue(logStatus.LoggingEnabled))
	// get,通过context取消
	ctx, cancelFunc := context.WithTimeout(context.Background(), bucketTimeout)
	defer cancelFunc()
	resp, err = client.GetBucketLoggingWithContext(ctx, &s3.GetBucketLoggingInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, NotNil)
}

// PUT Bucket Decompress Policy
func (s *Ks3utilCommandSuite) TestPutBucketDecompressPolicyWithContext(c *C) {
	// put,不通过context取消
	_, err := client.PutBucketDecompressPolicyWithContext(context.Background(), &s3.PutBucketDecompressPolicyInput{
		Bucket: aws.String(bucket),
		BucketDecompressPolicy: &s3.BucketDecompressPolicy{
			Rules: []*s3.DecompressPolicyRule{
				{
					Id:         aws.String("test"),
					Events:     aws.String("ObjectCreated:*"),
					Prefix:     aws.String("prefix"),
					Suffix:     []*string{aws.String(".zip")},
					Overwrite:  aws.Long(0),
					PolicyType: aws.String("decompress"),
				},
			},
		},
	})
	c.Assert(err, IsNil)
	// get
	resp, err := client.GetBucketDecompressPolicyWithContext(context.Background(), &s3.GetBucketDecompressPolicyInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, IsNil)
	c.Assert(len(resp.BucketDecompressPolicy.Rules), Equals, 1)
	c.Assert(*resp.BucketDecompressPolicy.Rules[0].Id, Equals, "test")
	c.Assert(*resp.BucketDecompressPolicy.Rules[0].Events, Equals, "ObjectCreated:*")
	c.Assert(*resp.BucketDecompressPolicy.Rules[0].Prefix, Equals, "prefix")
	c.Assert(*resp.BucketDecompressPolicy.Rules[0].Suffix[0], Equals, ".zip")
	c.Assert(*resp.BucketDecompressPolicy.Rules[0].Overwrite, Equals, int64(0))
	c.Assert(*resp.BucketDecompressPolicy.Rules[0].PolicyType, Equals, "decompress")
	// put,通过context取消
	ctx, cancelFunc := context.WithTimeout(context.Background(), bucketTimeout)
	defer cancelFunc()
	_, err = client.PutBucketDecompressPolicyWithContext(ctx, &s3.PutBucketDecompressPolicyInput{
		Bucket: aws.String(bucket),
		BucketDecompressPolicy: &s3.BucketDecompressPolicy{
			Rules: []*s3.DecompressPolicyRule{
				{
					Id:         aws.String("test"),
					Events:     aws.String("ObjectCreated:*"),
					Prefix:     aws.String("prefix"),
					Suffix:     []*string{aws.String(".zip")},
					Overwrite:  aws.Long(0),
					PolicyType: aws.String("decompress"),
				},
			},
		},
	})
	c.Assert(err, NotNil)
	// get
	resp, err = client.GetBucketDecompressPolicyWithContext(context.Background(), &s3.GetBucketDecompressPolicyInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, IsNil)
	c.Assert(len(resp.BucketDecompressPolicy.Rules), Equals, 1)
	c.Assert(*resp.BucketDecompressPolicy.Rules[0].Id, Equals, "test")
	c.Assert(*resp.BucketDecompressPolicy.Rules[0].Events, Equals, "ObjectCreated:*")
	c.Assert(*resp.BucketDecompressPolicy.Rules[0].Prefix, Equals, "prefix")
	c.Assert(*resp.BucketDecompressPolicy.Rules[0].Suffix[0], Equals, ".zip")
	c.Assert(*resp.BucketDecompressPolicy.Rules[0].Overwrite, Equals, int64(0))
	c.Assert(*resp.BucketDecompressPolicy.Rules[0].PolicyType, Equals, "decompress")
}

// GET Bucket Decompress Policy
func (s *Ks3utilCommandSuite) TestGetBucketDecompressPolicyWithContext(c *C) {
	// put
	_, err := client.PutBucketDecompressPolicyWithContext(context.Background(), &s3.PutBucketDecompressPolicyInput{
		Bucket: aws.String(bucket),
		BucketDecompressPolicy: &s3.BucketDecompressPolicy{
			Rules: []*s3.DecompressPolicyRule{
				{
					Id:         aws.String("test"),
					Events:     aws.String("ObjectCreated:*"),
					Prefix:     aws.String("prefix"),
					Suffix:     []*string{aws.String(".zip")},
					Overwrite:  aws.Long(0),
					PolicyType: aws.String("decompress"),
				},
			},
		},
	})
	c.Assert(err, IsNil)
	// get,不通过context取消
	resp, err := client.GetBucketDecompressPolicyWithContext(context.Background(), &s3.GetBucketDecompressPolicyInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, IsNil)
	c.Assert(len(resp.BucketDecompressPolicy.Rules), Equals, 1)
	c.Assert(*resp.BucketDecompressPolicy.Rules[0].Id, Equals, "test")
	c.Assert(*resp.BucketDecompressPolicy.Rules[0].Events, Equals, "ObjectCreated:*")
	c.Assert(*resp.BucketDecompressPolicy.Rules[0].Prefix, Equals, "prefix")
	c.Assert(*resp.BucketDecompressPolicy.Rules[0].Suffix[0], Equals, ".zip")
	c.Assert(*resp.BucketDecompressPolicy.Rules[0].Overwrite, Equals, int64(0))
	c.Assert(*resp.BucketDecompressPolicy.Rules[0].PolicyType, Equals, "decompress")
	// get,通过context取消
	ctx, cancelFunc := context.WithTimeout(context.Background(), bucketTimeout)
	defer cancelFunc()
	resp, err = client.GetBucketDecompressPolicyWithContext(ctx, &s3.GetBucketDecompressPolicyInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, NotNil)
}

// DELETE Bucket Decompress Policy
func (s *Ks3utilCommandSuite) TestDeleteBucketDecompressPolicyWithContext(c *C) {
	// put
	_, err := client.PutBucketDecompressPolicyWithContext(context.Background(), &s3.PutBucketDecompressPolicyInput{
		Bucket: aws.String(bucket),
		BucketDecompressPolicy: &s3.BucketDecompressPolicy{
			Rules: []*s3.DecompressPolicyRule{
				{
					Id:         aws.String("test"),
					Events:     aws.String("ObjectCreated:*"),
					Prefix:     aws.String("prefix"),
					Suffix:     []*string{aws.String(".zip")},
					Overwrite:  aws.Long(0),
					PolicyType: aws.String("decompress"),
				},
			},
		},
	})
	c.Assert(err, IsNil)
	// get
	resp, err := client.GetBucketDecompressPolicyWithContext(context.Background(), &s3.GetBucketDecompressPolicyInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, IsNil)
	c.Assert(len(resp.BucketDecompressPolicy.Rules), Equals, 1)
	c.Assert(*resp.BucketDecompressPolicy.Rules[0].Id, Equals, "test")
	c.Assert(*resp.BucketDecompressPolicy.Rules[0].Events, Equals, "ObjectCreated:*")
	c.Assert(*resp.BucketDecompressPolicy.Rules[0].Prefix, Equals, "prefix")
	c.Assert(*resp.BucketDecompressPolicy.Rules[0].Suffix[0], Equals, ".zip")
	c.Assert(*resp.BucketDecompressPolicy.Rules[0].Overwrite, Equals, int64(0))
	c.Assert(*resp.BucketDecompressPolicy.Rules[0].PolicyType, Equals, "decompress")
	// delete，通过context取消
	ctx, cancelFunc := context.WithTimeout(context.Background(), bucketTimeout)
	defer cancelFunc()
	_, err = client.DeleteBucketDecompressPolicyWithContext(ctx, &s3.DeleteBucketDecompressPolicyInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, NotNil)
	// delete，不通过context取消
	_, err = client.DeleteBucketDecompressPolicyWithContext(context.Background(), &s3.DeleteBucketDecompressPolicyInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, IsNil)
}

// PUT Bucket Replication
func (s *Ks3utilCommandSuite) TestPutBucketReplicationWithContext(c *C) {
	// put,不通过context取消
	_, err := client.PutBucketReplicationWithContext(context.Background(), &s3.PutBucketReplicationInput{
		Bucket: aws.String(bucket),
		ReplicationConfiguration: &s3.ReplicationConfiguration{
			Prefix:                      []*string{aws.String("test/")},
			DeleteMarkerStatus:          aws.String("Disabled"),
			TargetBucket:                aws.String(bucket),
			HistoricalObjectReplication: aws.String("Enabled"),
		},
	})
	c.Assert(err, IsNil)
	// get
	resp, err := client.GetBucketReplicationWithContext(context.Background(), &s3.GetBucketReplicationInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, IsNil)
	c.Assert(len(resp.ReplicationConfiguration.Prefix), Equals, 1)
	c.Assert(*resp.ReplicationConfiguration.Prefix[0], Equals, "test/")
	c.Assert(*resp.ReplicationConfiguration.DeleteMarkerStatus, Equals, "Disabled")
	c.Assert(*resp.ReplicationConfiguration.TargetBucket, Equals, bucket)
	c.Assert(*resp.ReplicationConfiguration.HistoricalObjectReplication, Equals, "Enabled")
	// put,通过context取消
	ctx, cancelFunc := context.WithTimeout(context.Background(), bucketTimeout)
	defer cancelFunc()
	_, err = client.PutBucketReplicationWithContext(ctx, &s3.PutBucketReplicationInput{
		Bucket: aws.String(bucket),
		ReplicationConfiguration: &s3.ReplicationConfiguration{
			Prefix:                      []*string{aws.String("test2/")},
			DeleteMarkerStatus:          aws.String("Enabled"),
			TargetBucket:                aws.String(bucket),
			HistoricalObjectReplication: aws.String("Disabled"),
		},
	})
	c.Assert(err, NotNil)
	// get
	resp, err = client.GetBucketReplicationWithContext(context.Background(), &s3.GetBucketReplicationInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, IsNil)
	c.Assert(len(resp.ReplicationConfiguration.Prefix), Equals, 1)
	c.Assert(*resp.ReplicationConfiguration.Prefix[0], Equals, "test/")
	c.Assert(*resp.ReplicationConfiguration.DeleteMarkerStatus, Equals, "Disabled")
	c.Assert(*resp.ReplicationConfiguration.TargetBucket, Equals, bucket)
	c.Assert(*resp.ReplicationConfiguration.HistoricalObjectReplication, Equals, "Enabled")
	// delete
	_, err = client.DeleteBucketReplicationWithContext(context.Background(), &s3.DeleteBucketReplicationInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, IsNil)
}

// GET Bucket Replication
func (s *Ks3utilCommandSuite) TestGetBucketReplicationWithContext(c *C) {
	// put
	_, err := client.PutBucketReplicationWithContext(context.Background(), &s3.PutBucketReplicationInput{
		Bucket: aws.String(bucket),
		ReplicationConfiguration: &s3.ReplicationConfiguration{
			Prefix:                      []*string{aws.String("test/")},
			DeleteMarkerStatus:          aws.String("Disabled"),
			TargetBucket:                aws.String(bucket),
			HistoricalObjectReplication: aws.String("Enabled"),
		},
	})
	c.Assert(err, IsNil)
	// get,不通过context取消
	resp, err := client.GetBucketReplicationWithContext(context.Background(), &s3.GetBucketReplicationInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, IsNil)
	c.Assert(len(resp.ReplicationConfiguration.Prefix), Equals, 1)
	c.Assert(*resp.ReplicationConfiguration.Prefix[0], Equals, "test/")
	c.Assert(*resp.ReplicationConfiguration.DeleteMarkerStatus, Equals, "Disabled")
	c.Assert(*resp.ReplicationConfiguration.TargetBucket, Equals, bucket)
	c.Assert(*resp.ReplicationConfiguration.HistoricalObjectReplication, Equals, "Enabled")
	// get,通过context取消
	ctx, cancelFunc := context.WithTimeout(context.Background(), bucketTimeout)
	defer cancelFunc()
	resp, err = client.GetBucketReplicationWithContext(ctx, &s3.GetBucketReplicationInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, NotNil)
	// delete
	_, err = client.DeleteBucketReplicationWithContext(context.Background(), &s3.DeleteBucketReplicationInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, IsNil)
}

// DELETE Bucket Replication
func (s *Ks3utilCommandSuite) TestDeleteBucketReplicationWithContext(c *C) {
	// put
	_, err := client.PutBucketReplicationWithContext(context.Background(), &s3.PutBucketReplicationInput{
		Bucket: aws.String(bucket),
		ReplicationConfiguration: &s3.ReplicationConfiguration{
			Prefix:                      []*string{aws.String("test/")},
			DeleteMarkerStatus:          aws.String("Disabled"),
			TargetBucket:                aws.String(bucket),
			HistoricalObjectReplication: aws.String("Enabled"),
		},
	})
	c.Assert(err, IsNil)
	// get
	resp, err := client.GetBucketReplicationWithContext(context.Background(), &s3.GetBucketReplicationInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, IsNil)
	c.Assert(len(resp.ReplicationConfiguration.Prefix), Equals, 1)
	c.Assert(*resp.ReplicationConfiguration.Prefix[0], Equals, "test/")
	c.Assert(*resp.ReplicationConfiguration.DeleteMarkerStatus, Equals, "Disabled")
	c.Assert(*resp.ReplicationConfiguration.TargetBucket, Equals, bucket)
	c.Assert(*resp.ReplicationConfiguration.HistoricalObjectReplication, Equals, "Enabled")
	// delete，通过context取消
	ctx, cancelFunc := context.WithTimeout(context.Background(), bucketTimeout)
	defer cancelFunc()
	_, err = client.DeleteBucketReplicationWithContext(ctx, &s3.DeleteBucketReplicationInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, NotNil)
	// delete，不通过context取消
	_, err = client.DeleteBucketReplicationWithContext(context.Background(), &s3.DeleteBucketReplicationInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, IsNil)
}

// PUT Bucket ACL
func (s *Ks3utilCommandSuite) TestPutBucketACLWithContext(c *C) {
	// put,不通过context取消
	_, err := client.PutBucketACLWithContext(context.Background(), &s3.PutBucketACLInput{
		Bucket: aws.String(bucket),
		ACL:    aws.String("public-read"),
	})
	c.Assert(err, IsNil)
	// get
	resp, err := client.GetBucketACLWithContext(context.Background(), &s3.GetBucketACLInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, IsNil)
	c.Assert(s3.GetBucketAcl(*resp), Equals, s3.PublicRead)
	// put,通过context取消
	ctx, cancelFunc := context.WithTimeout(context.Background(), bucketTimeout)
	defer cancelFunc()
	_, err = client.PutBucketACLWithContext(ctx, &s3.PutBucketACLInput{
		Bucket: aws.String(bucket),
		ACL:    aws.String("public-read-write"),
	})
	// get
	resp, err = client.GetBucketACLWithContext(context.Background(), &s3.GetBucketACLInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, IsNil)
	c.Assert(s3.GetBucketAcl(*resp), Equals, s3.PublicRead)
}

// GET Bucket ACL
func (s *Ks3utilCommandSuite) TestGetBucketACLWithContext(c *C) {
	// put
	_, err := client.PutBucketACLWithContext(context.Background(), &s3.PutBucketACLInput{
		Bucket: aws.String(bucket),
		ACL:    aws.String("public-read"),
	})
	c.Assert(err, IsNil)
	// get,不通过context取消
	resp, err := client.GetBucketACLWithContext(context.Background(), &s3.GetBucketACLInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, IsNil)
	c.Assert(s3.GetBucketAcl(*resp), Equals, s3.PublicRead)
	// get,通过context取消
	ctx, cancelFunc := context.WithTimeout(context.Background(), bucketTimeout)
	defer cancelFunc()
	resp, err = client.GetBucketACLWithContext(ctx, &s3.GetBucketACLInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, NotNil)
}

// PUT Bucket
func (s *Ks3utilCommandSuite) TestCreateBucketWithContext(c *C) {
	tempBucket := commonNamePrefix + randLowStr(10)
	// put,通过context取消
	ctx, cancelFunc := context.WithTimeout(context.Background(), bucketTimeout)
	defer cancelFunc()
	_, err := client.CreateBucketWithContext(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(tempBucket),
	})
	c.Assert(err, NotNil)
	// head
	resp, err := client.HeadBucketWithContext(context.Background(), &s3.HeadBucketInput{
		Bucket: aws.String(tempBucket),
	})
	c.Assert(*resp.StatusCode, Equals, int64(404))
	// put,不通过context取消
	_, err = client.CreateBucketWithContext(context.Background(), &s3.CreateBucketInput{
		Bucket: aws.String(tempBucket),
	})
	c.Assert(err, IsNil)
	// head
	resp, err = client.HeadBucketWithContext(context.Background(), &s3.HeadBucketInput{
		Bucket: aws.String(tempBucket),
	})
	c.Assert(*resp.StatusCode, Equals, int64(200))
	// delete
	_, err = client.DeleteBucketWithContext(context.Background(), &s3.DeleteBucketInput{
		Bucket: aws.String(tempBucket),
	})
	c.Assert(err, IsNil)
}

// GET Bucket
func (s *Ks3utilCommandSuite) TestGetBucketWithContext(c *C) {
	tempBucket := bucket + "test-2"
	// put
	_, err := client.CreateBucketWithContext(context.Background(), &s3.CreateBucketInput{
		Bucket: aws.String(tempBucket),
	})
	c.Assert(err, IsNil)
	// head
	resp, err := client.HeadBucketWithContext(context.Background(), &s3.HeadBucketInput{
		Bucket: aws.String(tempBucket),
	})
	c.Assert(*resp.StatusCode, Equals, int64(200))
	// get,不通过context取消
	_, err = client.ListObjectsWithContext(context.Background(), &s3.ListObjectsInput{
		Bucket: aws.String(tempBucket),
	})
	c.Assert(err, IsNil)
	// get,通过context取消
	ctx, cancelFunc := context.WithTimeout(context.Background(), bucketTimeout)
	defer cancelFunc()
	_, err = client.ListObjectsWithContext(ctx, &s3.ListObjectsInput{
		Bucket: aws.String(tempBucket),
	})
	c.Assert(err, NotNil)
	// delete
	_, err = client.DeleteBucketWithContext(context.Background(), &s3.DeleteBucketInput{
		Bucket: aws.String(tempBucket),
	})
	c.Assert(err, IsNil)
}

// DELETE Bucket
func (s *Ks3utilCommandSuite) TestDeleteBucketWithContext(c *C) {
	tempBucket := bucket + "test-3"
	// put
	_, err := client.CreateBucketWithContext(context.Background(), &s3.CreateBucketInput{
		Bucket: aws.String(tempBucket),
	})
	c.Assert(err, IsNil)
	// head
	resp, err := client.HeadBucketWithContext(context.Background(), &s3.HeadBucketInput{
		Bucket: aws.String(tempBucket),
	})
	c.Assert(*resp.StatusCode, Equals, int64(200))
	// delete,通过context取消
	ctx, cancelFunc := context.WithTimeout(context.Background(), bucketTimeout)
	defer cancelFunc()
	_, err = client.DeleteBucketWithContext(ctx, &s3.DeleteBucketInput{
		Bucket: aws.String(tempBucket),
	})
	c.Assert(err, NotNil)
	// head
	resp, err = client.HeadBucketWithContext(context.Background(), &s3.HeadBucketInput{
		Bucket: aws.String(tempBucket),
	})
	c.Assert(*resp.StatusCode, Equals, int64(200))
	// delete,不通过context取消
	_, err = client.DeleteBucketWithContext(context.Background(), &s3.DeleteBucketInput{
		Bucket: aws.String(tempBucket),
	})
	c.Assert(err, IsNil)
}

// List Buckets
func (s *Ks3utilCommandSuite) TestListBucketsWithContext(c *C) {
	// list,不通过context取消
	_, err := client.ListBucketsWithContext(context.Background(), &s3.ListBucketsInput{})
	c.Assert(err, IsNil)
	// list,通过context取消
	ctx, cancelFunc := context.WithTimeout(context.Background(), bucketTimeout)
	defer cancelFunc()
	_, err = client.ListBucketsWithContext(ctx, &s3.ListBucketsInput{})
	c.Assert(err, NotNil)
}

// HEAD Bucket
func (s *Ks3utilCommandSuite) TestHeadBucketWithContext(c *C) {
	tempBucket := bucket + "test-4"
	// put
	_, err := client.CreateBucketWithContext(context.Background(), &s3.CreateBucketInput{
		Bucket: aws.String(tempBucket),
	})
	c.Assert(err, IsNil)
	// head,通过context取消
	ctx, cancelFunc := context.WithTimeout(context.Background(), bucketTimeout)
	defer cancelFunc()
	resp, err := client.HeadBucketWithContext(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(tempBucket),
	})
	c.Assert(err, NotNil)
	// head,不通过context取消
	resp, err = client.HeadBucketWithContext(context.Background(), &s3.HeadBucketInput{
		Bucket: aws.String(tempBucket),
	})
	c.Assert(*resp.StatusCode, Equals, int64(200))
	// delete
	_, err = client.DeleteBucketWithContext(context.Background(), &s3.DeleteBucketInput{
		Bucket: aws.String(tempBucket),
	})
	c.Assert(err, IsNil)
}

// HEAD Bucket Exist
func (s *Ks3utilCommandSuite) TestHeadBucketExistWithContext(c *C) {
	tempBucket := bucket + "test-5"
	// put
	_, err := client.CreateBucketWithContext(context.Background(), &s3.CreateBucketInput{
		Bucket: aws.String(tempBucket),
	})
	c.Assert(err, IsNil)
	// head,通过context取消
	ctx, cancelFunc := context.WithTimeout(context.Background(), bucketTimeout)
	defer cancelFunc()
	_, err = client.HeadBucketExistWithContext(ctx, tempBucket)
	c.Assert(err, NotNil)
	// head,不通过context取消
	exist, err := client.HeadBucketExistWithContext(context.Background(), tempBucket)
	c.Assert(exist, Equals, true)
	// delete
	_, err = client.DeleteBucketWithContext(context.Background(), &s3.DeleteBucketInput{
		Bucket: aws.String(tempBucket),
	})
	c.Assert(err, IsNil)
}

// Get Bucket Location
func (s *Ks3utilCommandSuite) TestGetBucketLocationWithContext(c *C) {
	tempBucket := bucket + "test-6"
	createBucketConfiguration := &s3.CreateBucketConfiguration{
		LocationConstraint: aws.String("BEIJING"),
	}
	// put
	_, err := client.CreateBucketWithContext(context.Background(), &s3.CreateBucketInput{
		Bucket:                    aws.String(tempBucket),
		CreateBucketConfiguration: createBucketConfiguration,
	})
	c.Assert(err, IsNil)
	// get，通过context取消
	ctx, cancelFunc := context.WithTimeout(context.Background(), bucketTimeout)
	defer cancelFunc()
	_, err = client.GetBucketLocationWithContext(ctx, &s3.GetBucketLocationInput{
		Bucket: aws.String(tempBucket),
	})
	c.Assert(err, NotNil)
	// get，不通过context取消
	resp, err := client.GetBucketLocationWithContext(context.Background(), &s3.GetBucketLocationInput{
		Bucket: aws.String(tempBucket),
	})
	c.Assert(err, IsNil)
	c.Assert(*resp.LocationConstraint, Equals, "BEIJING")
	// delete
	_, err = client.DeleteBucketWithContext(context.Background(), &s3.DeleteBucketInput{
		Bucket: aws.String(tempBucket),
	})
	c.Assert(err, IsNil)
}

// URL Redirect
func (s *Ks3utilCommandSuite) TestURLRedirect(c *C) {
	var cre = credentials.NewStaticCredentials(accessKeyID, accessKeySecret, "")
	client2 := s3.New(&aws.Config{
		Credentials: cre,
		Region:      "SHANGHAI",
		Endpoint:    "ks3-cn-shanghai.ksyuncs.com",
	})
	_, err := client2.GetBucketACLWithContext(context.Background(), &s3.GetBucketACLInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, IsNil)
}
