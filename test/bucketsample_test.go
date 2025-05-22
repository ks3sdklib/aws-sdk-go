package lib

import (
	"github.com/ks3sdklib/aws-sdk-go/aws"
	"github.com/ks3sdklib/aws-sdk-go/service/s3"
	. "gopkg.in/check.v1"
	"time"
)

// TestBucket 创建bucket
func (s *Ks3utilCommandSuite) TestBucket(c *C) {
	// 创建bucket
	bucketName := commonNamePrefix + randLowStr(10)
	_, err := client.CreateBucket(&s3.CreateBucketInput{
		ACL:    aws.String("public-read"),
		Bucket: aws.String(bucketName),
		//ProjectId:  aws.String("1232"), //项目ID
		BucketType: aws.String(s3.BucketTypeNormal),
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

// TestListBuckets 获取bucket列表
func (s *Ks3utilCommandSuite) TestListBuckets(c *C) {
	_, err := client.ListBuckets(&s3.ListBucketsInput{})
	c.Assert(err, IsNil)
}

// TestBucketAcl bucket acl
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
	c.Assert(s3.GetCannedACL(resp.Grants), Equals, s3.ACLPublicRead)
}

// TestBucketLifecycle bucket lifecycle
func (s *Ks3utilCommandSuite) TestBucketLifecycle(c *C) {
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
	_, err := client.PutBucketLifecycle(&s3.PutBucketLifecycleInput{
		Bucket:                 aws.String(bucket),
		LifecycleConfiguration: lifecycleConfiguration,
	})
	c.Assert(err, IsNil)

	// 获取生命周期规则
	resp, err := client.GetBucketLifecycle(&s3.GetBucketLifecycleInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, IsNil)
	c.Assert(len(resp.Rules), Equals, 1)
	c.Assert(*resp.Rules[0].ID, Equals, "rule1")
	c.Assert(*resp.Rules[0].Filter.Prefix, Equals, "prefix1")
	c.Assert(*resp.Rules[0].Status, Equals, "Enabled")
	c.Assert(*resp.Rules[0].Expiration.Days, Equals, int64(90))
	c.Assert(*resp.Rules[0].Transitions[0].StorageClass, Equals, s3.StorageClassIA)
	c.Assert(*resp.Rules[0].Transitions[0].Days, Equals, int64(30))
	c.Assert(*resp.Rules[0].AbortIncompleteMultipartUpload.DaysAfterInitiation, Equals, int64(60))

	// 删除生命周期规则
	_, err = client.DeleteBucketLifecycle(&s3.DeleteBucketLifecycleInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, IsNil)
}

// TestBucketCors bucket cors
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

// TestSetBucketLog bucket log
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

// TestBucketMirror bucket mirror
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
	resp, err := client.GetBucketMirror(&s3.GetBucketMirrorInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, IsNil)
	c.Assert(*resp.BucketMirror.Version, Equals, "V3")

	// 删除桶的镜像回源规则
	_, err = client.DeleteBucketMirror(&s3.DeleteBucketMirrorInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, IsNil)
}

// TestBucketPolicy bucket policy
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

// TestBucketDecompressPolicy bucket decompress policy
func (s *Ks3utilCommandSuite) TestBucketDecompressPolicy(c *C) {
	_, err := client.PutBucketDecompressPolicy(&s3.PutBucketDecompressPolicyInput{
		Bucket: aws.String(bucket),
		BucketDecompressPolicy: &s3.BucketDecompressPolicy{
			Rules: []*s3.DecompressPolicyRule{
				{
					Id:                 aws.String("test"),
					Events:             aws.String("ObjectCreated:*"),
					Prefix:             aws.String("prefix"),
					Suffix:             []*string{aws.String(".zip")},
					Overwrite:          aws.Long(0),
					Callback:           aws.String("http://callback.demo.com"),
					CallbackFormat:     aws.String("JSON"),
					PathPrefix:         aws.String("test/"),
					PathPrefixReplaced: aws.Long(0),
					PolicyType:         aws.String("decompress"),
				},
			},
		},
	})
	c.Assert(err, IsNil)

	resp, err := client.GetBucketDecompressPolicy(&s3.GetBucketDecompressPolicyInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, IsNil)
	c.Assert(len(resp.BucketDecompressPolicy.Rules), Equals, 1)
	c.Assert(*resp.BucketDecompressPolicy.Rules[0].Id, Equals, "test")
	c.Assert(*resp.BucketDecompressPolicy.Rules[0].Events, Equals, "ObjectCreated:*")
	c.Assert(*resp.BucketDecompressPolicy.Rules[0].Prefix, Equals, "prefix")
	c.Assert(*resp.BucketDecompressPolicy.Rules[0].Suffix[0], Equals, ".zip")
	c.Assert(*resp.BucketDecompressPolicy.Rules[0].Overwrite, Equals, int64(0))
	c.Assert(*resp.BucketDecompressPolicy.Rules[0].Callback, Equals, "http://callback.demo.com")
	c.Assert(*resp.BucketDecompressPolicy.Rules[0].CallbackFormat, Equals, "JSON")
	c.Assert(*resp.BucketDecompressPolicy.Rules[0].PathPrefix, Equals, "test/")
	c.Assert(*resp.BucketDecompressPolicy.Rules[0].PathPrefixReplaced, Equals, int64(0))
	c.Assert(*resp.BucketDecompressPolicy.Rules[0].PolicyType, Equals, "decompress")

	_, err = client.DeleteBucketDecompressPolicy(&s3.DeleteBucketDecompressPolicyInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, IsNil)
}

// TestBucketRetention bucket retention
func (s *Ks3utilCommandSuite) TestBucketRetention(c *C) {
	retentionBucket := commonNamePrefix + randLowStr(10)
	s.CreateBucket(retentionBucket, c)
	_, err := client.PutBucketRetention(&s3.PutBucketRetentionInput{
		Bucket: aws.String(retentionBucket),
		RetentionConfiguration: &s3.BucketRetentionConfiguration{
			Rule: &s3.RetentionRule{
				Status: aws.String("Enabled"),
				Days:   aws.Long(30),
			},
		},
	})
	c.Assert(err, IsNil)

	resp, err := client.GetBucketRetention(&s3.GetBucketRetentionInput{
		Bucket: aws.String(retentionBucket),
	})
	c.Assert(err, IsNil)
	c.Assert(*resp.RetentionConfiguration.Rule.Status, Equals, "Enabled")
	c.Assert(*resp.RetentionConfiguration.Rule.Days, Equals, int64(30))

	_, err = client.ListRetention(&s3.ListRetentionInput{
		Bucket: aws.String(retentionBucket),
	})
	c.Assert(err, IsNil)
	s.DeleteBucket(retentionBucket, c)
}

// TestBucketReplication bucket replication
func (s *Ks3utilCommandSuite) TestBucketReplication(c *C) {
	_, err := client.PutBucketReplication(&s3.PutBucketReplicationInput{
		Bucket: aws.String(bucket),
		ReplicationConfiguration: &s3.ReplicationConfiguration{
			Prefix:                      []*string{aws.String("test/")},
			DeleteMarkerStatus:          aws.String("Disabled"),
			TargetBucket:                aws.String(bucket),
			HistoricalObjectReplication: aws.String("Enabled"),
		},
	})
	c.Assert(err, IsNil)

	resp, err := client.GetBucketReplication(&s3.GetBucketReplicationInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, IsNil)
	c.Assert(len(resp.ReplicationConfiguration.Prefix), Equals, 1)
	c.Assert(*resp.ReplicationConfiguration.Prefix[0], Equals, "test/")
	c.Assert(*resp.ReplicationConfiguration.DeleteMarkerStatus, Equals, "Disabled")
	c.Assert(*resp.ReplicationConfiguration.TargetBucket, Equals, bucket)
	c.Assert(*resp.ReplicationConfiguration.HistoricalObjectReplication, Equals, "Enabled")

	_, err = client.DeleteBucketReplication(&s3.DeleteBucketReplicationInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, IsNil)
}

func (s *Ks3utilCommandSuite) TestBucketInventory(c *C) {
	id := randLowStr(8)
	_, err := client.PutBucketInventory(&s3.PutBucketInventoryInput{
		Bucket: aws.String(bucket),
		Id:     aws.String(id),
		InventoryConfiguration: &s3.InventoryConfiguration{
			Id:        aws.String(id),
			IsEnabled: aws.Boolean(true),
			Filter: &s3.InventoryFilter{
				Prefix: aws.String("abc/"),
			},
			Destination: &s3.Destination{
				KS3BucketDestination: &s3.KS3BucketDestination{
					Format: aws.String("CSV"),
					Bucket: aws.String(bucket),
					Prefix: aws.String("prefix/"),
				},
			},
			Schedule: &s3.Schedule{
				Frequency: aws.String("Once"),
			},
			OptionalFields: &s3.OptionalFields{
				Field: []*string{
					aws.String("Size"),
					aws.String("LastModifiedDate"),
					aws.String("ETag"),
					aws.String("StorageClass"),
					aws.String("IsMultipartUploaded"),
					aws.String("EncryptionStatus"),
				},
			},
		},
	})
	c.Assert(err, IsNil)

	resp, err := client.GetBucketInventory(&s3.GetBucketInventoryInput{
		Bucket: aws.String(bucket),
		Id:     aws.String(id),
	})
	c.Assert(err, IsNil)
	c.Assert(*resp.InventoryConfiguration.Id, Equals, id)
	c.Assert(*resp.InventoryConfiguration.IsEnabled, Equals, true)
	c.Assert(*resp.InventoryConfiguration.Filter.Prefix, Equals, "abc/")
	c.Assert(*resp.InventoryConfiguration.Destination.KS3BucketDestination.Format, Equals, "CSV")
	c.Assert(*resp.InventoryConfiguration.Destination.KS3BucketDestination.Bucket, Equals, bucket)
	c.Assert(*resp.InventoryConfiguration.Destination.KS3BucketDestination.Prefix, Equals, "prefix/")
	c.Assert(*resp.InventoryConfiguration.Schedule.Frequency, Equals, "Once")
	c.Assert(len(resp.InventoryConfiguration.OptionalFields.Field), Equals, 6)

	listResp, err := client.ListBucketInventory(&s3.ListBucketInventoryInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, IsNil)
	c.Assert(len(listResp.InventoryConfigurationsResult.InventoryConfigurations), Equals, 1)

	_, err = client.DeleteBucketInventory(&s3.DeleteBucketInventoryInput{
		Bucket: aws.String(bucket),
		Id:     aws.String(id),
	})
	c.Assert(err, IsNil)
}

func (s *Ks3utilCommandSuite) TestCreateBucketDataRedundancy(c *C) {
	// 创建bucket，使用数据冗余类型LRS
	bucketName1 := commonNamePrefix + randLowStr(10)
	_, err := client.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(bucketName1),
		CreateBucketConfiguration: &s3.CreateBucketConfiguration{
			DataRedundancyType: aws.String(s3.DataRedundancyTypeLRS),
		},
	})
	c.Assert(err, IsNil)

	// 获取bucket信息
	resp, err := client.HeadBucket(&s3.HeadBucketInput{
		Bucket: aws.String(bucketName1),
	})
	c.Assert(err, IsNil)
	c.Assert(*resp.Metadata[s3.HTTPHeaderAmzDataRedundancyType], Equals, s3.DataRedundancyTypeLRS)
	c.Assert(*resp.Metadata[s3.HTTPHeaderAmzZRSSwitchEnable], Equals, "none")

	// 创建bucket，使用数据冗余类型ZRS
	bucketName2 := commonNamePrefix + randLowStr(10)
	_, err = client.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(bucketName2),
		CreateBucketConfiguration: &s3.CreateBucketConfiguration{
			DataRedundancyType: aws.String(s3.DataRedundancyTypeZRS),
		},
	})
	c.Assert(err, IsNil)

	// 获取bucket信息
	resp, err = client.HeadBucket(&s3.HeadBucketInput{
		Bucket: aws.String(bucketName2),
	})
	c.Assert(err, IsNil)
	c.Assert(*resp.Metadata[s3.HTTPHeaderAmzDataRedundancyType], Equals, s3.DataRedundancyTypeZRS)
	c.Assert(*resp.Metadata[s3.HTTPHeaderAmzZRSSwitchEnable], Equals, "none")

	// 获取bucket列表
	listResp, err := client.ListBuckets(&s3.ListBucketsInput{})
	c.Assert(err, IsNil)
	for _, bucketInfo := range listResp.Buckets {
		if *bucketInfo.Name == bucketName1 {
			c.Assert(*bucketInfo.DataRedundancyType, Equals, s3.DataRedundancyTypeLRS)
		}

		if *bucketInfo.Name == bucketName2 {
			c.Assert(*bucketInfo.DataRedundancyType, Equals, s3.DataRedundancyTypeZRS)
		}
	}

	// 删除bucket
	s.DeleteBucket(bucketName1, c)
	s.DeleteBucket(bucketName2, c)
}

func (s *Ks3utilCommandSuite) TestBucketDataRedundancySwitch(c *C) {
	// 创建bucket，使用数据冗余类型ZRS
	bucketName := commonNamePrefix + randLowStr(10)
	_, err := client.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
		CreateBucketConfiguration: &s3.CreateBucketConfiguration{
			DataRedundancyType: aws.String(s3.DataRedundancyTypeZRS),
		},
	})
	c.Assert(err, IsNil)

	// 获取bucket信息
	resp, err := client.HeadBucket(&s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	})
	c.Assert(err, IsNil)
	c.Assert(*resp.Metadata[s3.HTTPHeaderAmzDataRedundancyType], Equals, s3.DataRedundancyTypeZRS)
	c.Assert(*resp.Metadata[s3.HTTPHeaderAmzZRSSwitchEnable], Equals, "none")

	// 修改数据冗余类型为LRS
	_, err = client.PutBucketDataRedundancySwitch(&s3.PutBucketDataRedundancySwitchInput{
		Bucket:             aws.String(bucketName),
		DataRedundancyType: aws.String(s3.DataRedundancyTypeLRS),
	})
	c.Assert(err, IsNil)

	// 等待数据冗余类型切换完成
	time.Sleep(time.Second * 120)

	// 获取数据冗余类型
	switchResp, err := client.GetBucketDataRedundancySwitch(&s3.GetBucketDataRedundancySwitchInput{
		Bucket: aws.String(bucketName),
	})
	c.Assert(err, IsNil)
	c.Assert(*switchResp.DataRedundancySwitch.DataRedundancyType, Equals, s3.DataRedundancyTypeLRS)

	// 获取bucket信息
	resp, err = client.HeadBucket(&s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	})
	c.Assert(err, IsNil)
	c.Assert(*resp.Metadata[s3.HTTPHeaderAmzDataRedundancyType], Equals, s3.DataRedundancyTypeZRS)
	c.Assert(*resp.Metadata[s3.HTTPHeaderAmzZRSSwitchEnable], Equals, "false")

	// 修改数据冗余类型为ZRS
	_, err = client.PutBucketDataRedundancySwitch(&s3.PutBucketDataRedundancySwitchInput{
		Bucket:             aws.String(bucketName),
		DataRedundancyType: aws.String(s3.DataRedundancyTypeZRS),
	})
	c.Assert(err, IsNil)

	// 等待数据冗余类型切换完成
	time.Sleep(time.Second * 120)

	// 获取数据冗余类型
	switchResp, err = client.GetBucketDataRedundancySwitch(&s3.GetBucketDataRedundancySwitchInput{
		Bucket: aws.String(bucketName),
	})
	c.Assert(err, IsNil)
	c.Assert(*switchResp.DataRedundancySwitch.DataRedundancyType, Equals, s3.DataRedundancyTypeZRS)

	// 获取bucket信息
	resp, err = client.HeadBucket(&s3.HeadBucketInput{
		Bucket: aws.String(bucketName),
	})
	c.Assert(err, IsNil)
	c.Assert(*resp.Metadata[s3.HTTPHeaderAmzDataRedundancyType], Equals, s3.DataRedundancyTypeZRS)
	c.Assert(*resp.Metadata[s3.HTTPHeaderAmzZRSSwitchEnable], Equals, "true")

	// 删除bucket
	s.DeleteBucket(bucketName, c)
}
