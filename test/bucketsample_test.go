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
		Bucket: aws.String(bucketName),
		ACL:    aws.String(s3.ACLPublicRead),
		//ProjectId:  aws.String("1232"), //项目ID
		BucketType:      aws.String(s3.BucketTypeNormal),
		BucketVisitType: aws.String(s3.BucketVisitTypeNormal),
	})
	c.Assert(err, IsNil)

	// 判断bucket桶是否存在
	exist, err := client.HeadBucketExist(bucketName)
	c.Assert(err, IsNil)
	c.Assert(exist, Equals, true)

	// 获取bucket信息
	headResp, err := client.HeadBucket(&s3.HeadBucketInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, IsNil)
	c.Assert(*headResp.Metadata[s3.HTTPHeaderAmzBucketType], Equals, s3.BucketTypeNormal)
	c.Assert(*headResp.Metadata[s3.HTTPHeaderAmzBucketVisitType], Equals, s3.BucketVisitTypeNormal)

	// 获取bucket列表
	listResp, err := client.ListBuckets(&s3.ListBucketsInput{})
	c.Assert(err, IsNil)
	for _, bucketInfo := range listResp.Buckets {
		if *bucketInfo.Name == bucketName {
			c.Assert(*bucketInfo.Type, Equals, s3.BucketTypeNormal)
			c.Assert(*bucketInfo.VisitType, Equals, s3.BucketVisitTypeNormal)
			c.Assert(*bucketInfo.DataRedundancyType, Equals, s3.DataRedundancyTypeLRS)
		}
	}

	// 删除bucket
	_, err = client.DeleteBucket(&s3.DeleteBucketInput{
		Bucket: aws.String(bucketName),
	})
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
	// 设置桶访问追踪配置
	_, err := client.PutBucketAccessMonitor(&s3.PutBucketAccessMonitorInput{
		Bucket: aws.String(bucket),
		AccessMonitorConfiguration: &s3.AccessMonitorConfiguration{
			Status: aws.String(s3.StatusEnabled),
		},
	})
	c.Assert(err, IsNil)

	// 获取桶访问追踪配置
	accessMonitorResp, err := client.GetBucketAccessMonitor(&s3.GetBucketAccessMonitorInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, IsNil)
	c.Assert(*accessMonitorResp.AccessMonitorConfiguration.Status, Equals, s3.StatusEnabled)

	lifecycleConfiguration := &s3.LifecycleConfiguration{
		Rules: []*s3.LifecycleRule{
			{
				ID: aws.String("rule1"),
				Filter: &s3.LifecycleFilter{
					Prefix: aws.String("prefix1/"),
				},
				Status: aws.String(s3.StatusEnabled),
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
			{
				ID: aws.String("rule2"),
				Filter: &s3.LifecycleFilter{
					And: &s3.And{
						Prefix: aws.String("prefix2/"),
						Tag: []*s3.Tag{
							{
								Key:   aws.String("key1"),
								Value: aws.String("value1"),
							},
						},
					},
				},
				Status: aws.String(s3.StatusEnabled),
				Transitions: []*s3.Transition{
					{
						Days:                 aws.Long(30),
						StorageClass:         aws.String(s3.StorageClassIA),
						IsAccessTime:         aws.Boolean(true),
						ReturnToStdWhenVisit: aws.Boolean(true),
					},
				},
			},
		},
	}
	// 设置桶生命周期规则
	_, err = client.PutBucketLifecycle(&s3.PutBucketLifecycleInput{
		Bucket:                 aws.String(bucket),
		LifecycleConfiguration: lifecycleConfiguration,
		AllowSameActionOverlap: aws.Boolean(true),
	})
	c.Assert(err, IsNil)

	// 获取桶生命周期规则
	resp, err := client.GetBucketLifecycle(&s3.GetBucketLifecycleInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, IsNil)
	c.Assert(len(resp.Rules), Equals, 2)
	c.Assert(*resp.Rules[0].ID, Equals, "rule1")
	c.Assert(*resp.Rules[0].Filter.Prefix, Equals, "prefix1/")
	c.Assert(*resp.Rules[0].Status, Equals, s3.StatusEnabled)
	c.Assert(*resp.Rules[0].Expiration.Days, Equals, int64(90))
	c.Assert(*resp.Rules[0].Transitions[0].StorageClass, Equals, s3.StorageClassIA)
	c.Assert(*resp.Rules[0].Transitions[0].Days, Equals, int64(30))
	c.Assert(*resp.Rules[0].AbortIncompleteMultipartUpload.DaysAfterInitiation, Equals, int64(60))
	c.Assert(*resp.Rules[1].ID, Equals, "rule2")
	c.Assert(*resp.Rules[1].Filter.And.Prefix, Equals, "prefix2/")
	c.Assert(len(resp.Rules[1].Filter.And.Tag), Equals, 1)
	c.Assert(*resp.Rules[1].Filter.And.Tag[0].Key, Equals, "key1")
	c.Assert(*resp.Rules[1].Filter.And.Tag[0].Value, Equals, "value1")
	c.Assert(*resp.Rules[1].Status, Equals, s3.StatusEnabled)
	c.Assert(*resp.Rules[1].Transitions[0].StorageClass, Equals, s3.StorageClassIA)
	c.Assert(*resp.Rules[1].Transitions[0].Days, Equals, int64(30))
	c.Assert(*resp.Rules[1].Transitions[0].IsAccessTime, Equals, true)
	c.Assert(*resp.Rules[1].Transitions[0].ReturnToStdWhenVisit, Equals, true)
	c.Assert(*resp.Metadata[s3.HTTPHeaderAmzAllowSameActionOverlap], Equals, "true")

	// 删除桶生命周期规则
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
					"PUT", "GET", "HEAD", "DELETE",
				},
				AllowedOrigin: []string{
					"*",
				},
				ExposeHeader: []string{
					"ETag", "x-kss-meta-test",
				},
				MaxAgeSeconds: aws.Long(100),
			},
			{
				AllowedHeader: []string{
					"x-kss-meta-test1", "x-kss-meta-test2",
				},
				AllowedMethod: []string{
					"GET", "HEAD",
				},
				AllowedOrigin: []string{
					"https://example1.com", "https://example2.com",
				},
				ExposeHeader: []string{
					"ETag", "x-kss-acl",
				},
				MaxAgeSeconds: aws.Long(200),
			},
		},
		NonCrossOriginResponseVary: aws.Boolean(true),
	}
	// 设置桶的CORS配置
	_, err := client.PutBucketCORS(&s3.PutBucketCORSInput{
		Bucket:            aws.String(bucket),
		CORSConfiguration: corsConfiguration,
	})
	c.Assert(err, IsNil)

	// 获取桶的CORS配置
	resp, err := client.GetBucketCORS(&s3.GetBucketCORSInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, IsNil)
	c.Assert(len(resp.CORSConfiguration.Rules), Equals, 2)
	c.Assert(resp.CORSConfiguration.Rules[0].AllowedHeader, DeepEquals, []string{"*"})
	c.Assert(resp.CORSConfiguration.Rules[0].AllowedMethod, DeepEquals, []string{"PUT", "GET", "HEAD", "DELETE"})
	c.Assert(resp.CORSConfiguration.Rules[0].AllowedOrigin, DeepEquals, []string{"*"})
	c.Assert(resp.CORSConfiguration.Rules[0].ExposeHeader, DeepEquals, []string{"ETag", "x-kss-meta-test"})
	c.Assert(*resp.CORSConfiguration.Rules[0].MaxAgeSeconds, Equals, int64(100))
	c.Assert(resp.CORSConfiguration.Rules[1].AllowedHeader, DeepEquals, []string{"x-kss-meta-test1", "x-kss-meta-test2"})
	c.Assert(resp.CORSConfiguration.Rules[1].AllowedMethod, DeepEquals, []string{"GET", "HEAD"})
	c.Assert(resp.CORSConfiguration.Rules[1].AllowedOrigin, DeepEquals, []string{"https://example1.com", "https://example2.com"})
	c.Assert(resp.CORSConfiguration.Rules[1].ExposeHeader, DeepEquals, []string{"ETag", "x-kss-acl"})
	c.Assert(*resp.CORSConfiguration.Rules[1].MaxAgeSeconds, Equals, int64(200))
	c.Assert(*resp.CORSConfiguration.NonCrossOriginResponseVary, Equals, true)

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
			EnableMultipleVersion: aws.Boolean(true),
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
	c.Assert(*resp.RetentionConfiguration.EnableMultipleVersion, Equals, true)
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
	c.Skip("Skip TestBucketInventory")
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

func (s *Ks3utilCommandSuite) TestBucketQos(c *C) {
	c.Skip("Skip TestBucketQos")
	// 设置桶流控配置
	_, err := client.PutBucketQos(&s3.PutBucketQosInput{
		Bucket: aws.String(bucket),
		BucketQosConfiguration: &s3.BucketQosConfiguration{
			Quotas: []*s3.BucketQosQuota{
				{
					StorageMedium:             aws.String(s3.StorageMediumNormal),
					ExtranetUploadBandwidth:   aws.Long(10),
					IntranetUploadBandwidth:   aws.Long(10),
					ExtranetDownloadBandwidth: aws.Long(10),
					IntranetDownloadBandwidth: aws.Long(10),
				},
				{
					StorageMedium:             aws.String(s3.StorageMediumExtreme),
					ExtranetUploadBandwidth:   aws.Long(10),
					IntranetUploadBandwidth:   aws.Long(10),
					ExtranetDownloadBandwidth: aws.Long(10),
					IntranetDownloadBandwidth: aws.Long(10),
				},
			},
		},
	})
	c.Assert(err, IsNil)

	// 获取桶流控配置
	resp, err := client.GetBucketQos(&s3.GetBucketQosInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, IsNil)
	c.Assert(len(resp.BucketQosConfiguration.Quotas), Equals, 2)
	c.Assert(*resp.BucketQosConfiguration.Quotas[0].StorageMedium, Equals, s3.StorageMediumNormal)
	c.Assert(*resp.BucketQosConfiguration.Quotas[0].ExtranetUploadBandwidth, Equals, int64(10))
	c.Assert(*resp.BucketQosConfiguration.Quotas[0].IntranetUploadBandwidth, Equals, int64(10))
	c.Assert(*resp.BucketQosConfiguration.Quotas[0].ExtranetDownloadBandwidth, Equals, int64(10))
	c.Assert(*resp.BucketQosConfiguration.Quotas[0].IntranetDownloadBandwidth, Equals, int64(10))
	c.Assert(*resp.BucketQosConfiguration.Quotas[1].StorageMedium, Equals, s3.StorageMediumExtreme)
	c.Assert(*resp.BucketQosConfiguration.Quotas[1].ExtranetUploadBandwidth, Equals, int64(10))
	c.Assert(*resp.BucketQosConfiguration.Quotas[1].IntranetUploadBandwidth, Equals, int64(10))
	c.Assert(*resp.BucketQosConfiguration.Quotas[1].ExtranetDownloadBandwidth, Equals, int64(10))
	c.Assert(*resp.BucketQosConfiguration.Quotas[1].IntranetDownloadBandwidth, Equals, int64(10))

	// 删除桶流控配置
	_, err = client.DeleteBucketQos(&s3.DeleteBucketQosInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, IsNil)
}

func (s *Ks3utilCommandSuite) TestRequesterQos(c *C) {
	c.Skip("Skip TestRequesterQos")
	// 设置请求者流控配置
	_, err := client.PutRequesterQos(&s3.PutRequesterQosInput{
		Bucket: aws.String(bucket),
		RequesterQosConfiguration: &s3.RequesterQosConfiguration{
			Rules: []*s3.RequesterQosRule{
				{
					UserType: aws.String("User"),
					Krn:      aws.String("12345678/user1"),
					Quotas: []*s3.BucketQosQuota{
						{
							StorageMedium:             aws.String(s3.StorageMediumNormal),
							ExtranetUploadBandwidth:   aws.Long(1),
							IntranetUploadBandwidth:   aws.Long(1),
							ExtranetDownloadBandwidth: aws.Long(1),
							IntranetDownloadBandwidth: aws.Long(1),
						},
					},
				},
				{
					UserType: aws.String("Role"),
					Krn:      aws.String("12345678/role1"),
					Quotas: []*s3.BucketQosQuota{
						{
							StorageMedium:             aws.String(s3.StorageMediumExtreme),
							ExtranetUploadBandwidth:   aws.Long(1),
							IntranetUploadBandwidth:   aws.Long(1),
							ExtranetDownloadBandwidth: aws.Long(1),
							IntranetDownloadBandwidth: aws.Long(1),
						},
					},
				},
			},
		},
	})
	c.Assert(err, IsNil)

	// 获取请求者流控配置
	resp, err := client.GetRequesterQos(&s3.GetRequesterQosInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, IsNil)
	c.Assert(len(resp.RequesterQosConfiguration.Rules), Equals, 2)
	c.Assert(*resp.RequesterQosConfiguration.Rules[0].UserType, Equals, "User")
	c.Assert(*resp.RequesterQosConfiguration.Rules[0].Krn, Equals, "12345678/user1")
	c.Assert(len(resp.RequesterQosConfiguration.Rules[0].Quotas), Equals, 1)
	c.Assert(*resp.RequesterQosConfiguration.Rules[0].Quotas[0].StorageMedium, Equals, s3.StorageMediumNormal)
	c.Assert(*resp.RequesterQosConfiguration.Rules[0].Quotas[0].ExtranetUploadBandwidth, Equals, int64(1))
	c.Assert(*resp.RequesterQosConfiguration.Rules[0].Quotas[0].IntranetUploadBandwidth, Equals, int64(1))
	c.Assert(*resp.RequesterQosConfiguration.Rules[0].Quotas[0].ExtranetDownloadBandwidth, Equals, int64(1))
	c.Assert(*resp.RequesterQosConfiguration.Rules[0].Quotas[0].IntranetDownloadBandwidth, Equals, int64(1))
	c.Assert(*resp.RequesterQosConfiguration.Rules[1].UserType, Equals, "Role")
	c.Assert(*resp.RequesterQosConfiguration.Rules[1].Krn, Equals, "12345678/role1")
	c.Assert(len(resp.RequesterQosConfiguration.Rules[1].Quotas), Equals, 1)
	c.Assert(*resp.RequesterQosConfiguration.Rules[1].Quotas[0].StorageMedium, Equals, s3.StorageMediumExtreme)
	c.Assert(*resp.RequesterQosConfiguration.Rules[1].Quotas[0].ExtranetUploadBandwidth, Equals, int64(1))
	c.Assert(*resp.RequesterQosConfiguration.Rules[1].Quotas[0].IntranetUploadBandwidth, Equals, int64(1))
	c.Assert(*resp.RequesterQosConfiguration.Rules[1].Quotas[0].ExtranetDownloadBandwidth, Equals, int64(1))
	c.Assert(*resp.RequesterQosConfiguration.Rules[1].Quotas[0].IntranetDownloadBandwidth, Equals, int64(1))

	// 删除请求者流控配置
	_, err = client.DeleteRequesterQos(&s3.DeleteRequesterQosInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, IsNil)
}

func (s *Ks3utilCommandSuite) TestBucketTagging(c *C) {
	// 设置桶标签
	_, err := client.PutBucketTagging(&s3.PutBucketTaggingInput{
		Bucket: aws.String(bucket),
		Tagging: &s3.Tagging{
			TagSet: []*s3.Tag{
				{
					Key:   aws.String("key1"),
					Value: aws.String("value1"),
				},
				{
					Key:   aws.String("key2"),
					Value: aws.String("value2"),
				},
			},
		},
	})
	c.Assert(err, IsNil)

	// 获取桶标签
	resp, err := client.GetBucketTagging(&s3.GetBucketTaggingInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, IsNil)
	c.Assert(len(resp.Tagging.TagSet), Equals, 2)
	c.Assert(*resp.Tagging.TagSet[0].Key, Equals, "key1")
	c.Assert(*resp.Tagging.TagSet[0].Value, Equals, "value1")
	c.Assert(*resp.Tagging.TagSet[1].Key, Equals, "key2")
	c.Assert(*resp.Tagging.TagSet[1].Value, Equals, "value2")

	// 删除桶标签
	_, err = client.DeleteBucketTagging(&s3.DeleteBucketTaggingInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, IsNil)
}

func (s *Ks3utilCommandSuite) TestBucketEncryption(c *C) {
	// 设置桶加密配置
	_, err := client.PutBucketEncryption(&s3.PutBucketEncryptionInput{
		Bucket: aws.String(bucket),
		ServerSideEncryptionConfiguration: &s3.ServerSideEncryptionConfiguration{
			Rule: &s3.BucketEncryptionRule{
				ApplyServerSideEncryptionByDefault: &s3.ApplyServerSideEncryptionByDefault{
					SSEAlgorithm: aws.String(s3.AlgorithmAES256),
				},
			},
		},
	})
	c.Assert(err, IsNil)

	// 获取桶加密配置
	resp, err := client.GetBucketEncryption(&s3.GetBucketEncryptionInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, IsNil)
	c.Assert(*resp.ServerSideEncryptionConfiguration.Rule.ApplyServerSideEncryptionByDefault.SSEAlgorithm, Equals, s3.AlgorithmAES256)

	// 删除桶加密配置
	_, err = client.DeleteBucketEncryption(&s3.DeleteBucketEncryptionInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, IsNil)
}

func (s *Ks3utilCommandSuite) TestBucketTransferAcceleration(c *C) {
	// 设置桶传输加速配置
	_, err := client.PutBucketTransferAcceleration(&s3.PutBucketTransferAccelerationInput{
		Bucket: aws.String(bucket),
		TransferAccelerationConfiguration: &s3.TransferAccelerationConfiguration{
			Enabled: aws.Boolean(true),
		},
	})
	c.Assert(err, IsNil)

	// 获取桶传输加速配置
	resp, err := client.GetBucketTransferAcceleration(&s3.GetBucketTransferAccelerationInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, IsNil)
	c.Assert(*resp.TransferAccelerationConfiguration.Enabled, Equals, true)
}
