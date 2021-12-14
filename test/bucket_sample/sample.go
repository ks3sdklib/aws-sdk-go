package bucket_sample

import (
	"fmt"
	"github.com/ks3sdklib/aws-sdk-go/aws"
	"github.com/ks3sdklib/aws-sdk-go/aws/awsutil"
	"github.com/ks3sdklib/aws-sdk-go/aws/credentials"
	"github.com/ks3sdklib/aws-sdk-go/service/s3"
)

var bucket = string("yourbucket")
var key = string("yourkey")
var key_encode = string("yourkey")
var key_copy = string("yourkey")
var content = string("content")

// 金山云主账号 AccessKey 拥有所有API的访问权限，风险很高。
// 强烈建议您创建并使用子账号账号进行 API 访问或日常运维，请登录 https://uc.console.ksyun.com/pro/iam/#/user/list 创建子账号。
// 通过指定 host(Endpoint)，您可以在指定的地域创建新的存储空间。
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

//创建bucket并关联项目
func TestCreateBucket(svc *s3.S3) {
	resp, _ := svc.CreateBucket(&s3.CreateBucketInput{
		ACL:       aws.String("public-read"),
		Bucket:    aws.String(bucket),
		ProjectId: aws.String("1232"), //项目ID
	})
	fmt.Println("结果：\n", awsutil.StringValue(resp))
}

//设置bucketAcl
func TestPutBucketAcl(svc *s3.S3) {

	resp, _ := svc.PutBucketACL(&s3.PutBucketACLInput{
		Bucket: aws.String(bucket),
		ACL:    aws.String("public-read"),
	})
	fmt.Println("结果：\n", awsutil.StringValue(resp))
}

//设置bucketAcl
func TestGetBucketAcl(svc *s3.S3) {
	resp, _ := svc.GetBucketACL(&s3.GetBucketACLInput{
		Bucket: aws.String(bucket),
	})
	fmt.Println("结果：\n", awsutil.StringValue(resp))

}

//遍历bucket
func TestListBuckets(svc *s3.S3) {
	resp, _ := svc.ListBuckets(nil)
	//bucket列表
	buckets := resp.Buckets
	for i := 0; i < len(buckets); i++ {
		fmt.Println(*buckets[i].Name)
		fmt.Println(*buckets[i].Region)
	}
	fmt.Println("结果：\n", awsutil.StringValue(resp))
}

//判断bucket是否存在
func TestHeadBucket(svc *s3.S3) {
	resp, _ := svc.HeadBucket(&s3.HeadBucketInput{
		Bucket: aws.String(bucket),
	})
	fmt.Println("结果：\n", awsutil.StringValue(resp))
}

//删除bucket
func TestDeleteBucket(svc *s3.S3) {
	resp, _ := svc.DeleteBucket(&s3.DeleteBucketInput{
		Bucket: aws.String(bucket),
	})
	fmt.Println("结果：\n", awsutil.StringValue(resp))
}

//设置镜像回源规则
//详情见API(https://docs.ksyun.com/documents/39134)
func PutBucketMirrorRules(client *s3.S3) {

	params := &s3.PutBucketMirrorInput{
		Bucket: aws.String(bucket), // Required
		BucketMirror: &s3.BucketMirror{
			Version:          aws.String("V3"), //回源类型
			UseDefaultRobots: aws.Boolean(false),//是否使用默认的robots.txt，如果为true则会在bucket下生成一个robots.txt。
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
						Follow3Xx:       aws.Boolean(false),
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
		},
	}
	resp, _ := client.PutBucketMirror(params)
	fmt.Println("结果：\n", awsutil.StringValue(resp))

}

//获取镜像回源规则
func GetBucketMirrorRules(client *s3.S3) {

	params := &s3.GetBucketMirrorInput{
		Bucket: aws.String(bucket),
	}
	resp, _ := client.GetBucketMirror(params)
	fmt.Println("结果：\n", awsutil.StringValue(resp))
}

//删除镜像回源规则
func DeleteBucketMirrorRules(client *s3.S3) {

	params := &s3.DeleteBucketMirrorInput{
		Bucket: aws.String(bucket),
	}
	resp, _ := client.DeleteBucketMirror(params)
	fmt.Println("结果：\n", awsutil.StringValue(resp))
}
