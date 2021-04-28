package main

import (
	"aws-sdk-go/service/s3/s3manager"
	"bytes"
	"errors"
	"fmt"
	"github.com/ks3sdklib/aws-sdk-go/aws"
	"github.com/ks3sdklib/aws-sdk-go/aws/awserr"
	"github.com/ks3sdklib/aws-sdk-go/aws/awsutil"
	"github.com/ks3sdklib/aws-sdk-go/aws/credentials"
	"github.com/ks3sdklib/aws-sdk-go/service/s3"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"strconv"
)

var (
	region     = ""
	ak         = ""
	sk         = ""
	endpoint   = "ks3-cn-shanghai.ksyun.com"
	bucketname = ""
	key        = "1.ipa"
	prefix  = "test/" //目录名称
	objname = "file1.mp4"
)

func main() {
	credentials := credentials.NewStaticCredentials(ak, sk, "")
	client := s3.New(&aws.Config{
		Region:           region,
		Credentials:      credentials,
		Endpoint:         endpoint,  //ks3地址
		DisableSSL:       true,      //是否禁用https
		LogLevel:         1,         //是否开启日志,0为关闭日志，1为开启日志
		S3ForcePathStyle: true,      //是否强制使用path style方式访问
		LogHTTPBody:      true,      //是否把HTTP请求body打入日志
		Logger:           os.Stdout, //打日志的位置
	})
	//multipartUploadAllTagging(client)

	//分块上传1/3 初始化分块上传
	//initiateUpload(client)
	//分块上传2/3 上传文件快
	//partUpload(client)
	//分块上传3/3 完成分块上传
	//completeUpload(client)

	//fetchObj(client) //返回成功，但是未查看到fetch效果 也无法查看tag效果
	//copyObj(client) //成功，能查看到copy结果，但是无法查看到tag,需要ks3协助排查
	 // putObj(client)
	//getObj(client)
	//putTag(client)
	//headObj(client)

	//putLifecycle(client)
	//getLifecycle(client)


	//getTag(client)
	//delTag(client)
	//testUpload(client)
	PutBucketMirror(client)
}

var (
	uploadkey  = "src.zipold"
	uploadID   = "cc0e00789d724b5d81dca5f40a1a1dd2"
	uploadpath = "C:\\Users\\yangzhen1\\Desktop\\es10000Console_old\\src.zipold"
	etag = "504eb4064ceaf14ea157aff50cfb3c15"
)

func multipartUploadAllTagging(svc *s3.S3){

	var bucket  = "cqc-test-b"//欲上传的桶的名字
	var name  = "test012506.zip"//上传的对象的新的名字
	var container  = int64(4096*4096)//每次上传的块的大小
	var filepath = "C:\\Users\\yangzhen1\\Desktop\\es10000Console_old\\src.zipold" //本地需要上传的文件

	var uploadID = ""
	var etags  = []string{}//每次上传块返回的etag
	var numbers  = []int64{}//每次上传块号

	v := url.Values{}
	v.Add("name","yz")
	v.Add("sex","female")
	tag := v.Encode()
	input := s3.CreateMultipartUploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(name),
		XAmzTagging:aws.String(tag),
	}
	resp, err := svc.CreateMultipartUpload(&input)
	if err != nil {
		panic(errors.New("阶段1fail"))
	}
	uploadID  = *(resp.UploadID)
	fmt.Println("阶段1\n",resp)



	fd,err := os.Open(filepath)
	defer fd.Close()
	var i int64 = 0
	for {
		offset := i*container
		buffer := make([]byte,container)
		len,err := fd.ReadAt(buffer,offset)
		if err != nil && err != io.EOF  {
			panic(err)
		}
		fmt.Println("读取了：",len)
		i++
		//上传之
		input2 := s3.UploadPartInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(name),
			UploadID:aws.String(uploadID),
			PartNumber:aws.Long(i),
			Body:bytes.NewReader(buffer),
		}
		resp2, err := svc.UploadPart(&input2)
		if err!= nil {
			panic(errors.New("阶段2失败："+strconv.Itoa(int(i))))
		}
		etags = append(etags,*(resp2.ETag))
		numbers = append(numbers,i)

		//上传完最后一块，打印出来etags  numbers
		if len<int(container) {
			fmt.Println("阶段2\n",resp)
			fmt.Println("etags:",etags)
			fmt.Println("numbers:",numbers)
			break
		}
	}


	comparts := []*s3.CompletedPart{}
	for i,etag := range etags {
		part := s3.CompletedPart{
			ETag: aws.String(etag), PartNumber: aws.Long(numbers[i]),
		}
		comparts = append(comparts,&part)
	}
	input3 := s3.CompleteMultipartUploadInput{
		Bucket:aws.String(bucket),
		Key:aws.String(name),
		UploadID:aws.String(uploadID),
		MultipartUpload: &s3.CompletedMultipartUpload{Parts: comparts},
	}
	resp3,err := svc.CompleteMultipartUpload(&input3)
	if err!=nil {
		panic(errors.New("阶段3fail"))
	}
	fmt.Println("阶段3：\n", awsutil.StringValue(resp3))
}


func multipartUploadAll(svc *s3.S3){

	var container  = int64(4096*4096)//每次上传的块的大小
	var filepath = "C:\\Users\\yangzhen1\\Desktop\\es10000Console_old\\src.zipold" //本地需要上传的文件

	var uploadID = ""
	var etags  = []string{}//每次上传块返回的etag
	var numbers  = []int64{}//每次上传块号

	input := s3.CreateMultipartUploadInput{
		Bucket: aws.String(bucketname),
		Key:    aws.String(key),
	}
	resp, err := svc.CreateMultipartUpload(&input)
	if err != nil {
		panic(errors.New("阶段1fail"))
	}
	uploadID  = *(resp.UploadID)
	fmt.Println("阶段1\n",resp)



	fd,err := os.Open(filepath)
	defer fd.Close()
	var i int64 = 0
	for {
		offset := i*container
		buffer := make([]byte,container)
		len,err := fd.ReadAt(buffer,offset)
		if err != nil && err != io.EOF  {
			panic(err)
		}
		fmt.Println("读取了：",len)
		i++
		//上传之
		input2 := s3.UploadPartInput{
			Bucket: aws.String(bucketname),
			Key:    aws.String(key),
			UploadID:aws.String(uploadID),
			PartNumber:aws.Long(i),
			Body:bytes.NewReader(buffer),
		}
		resp2, err := svc.UploadPart(&input2)
		if err!= nil {
			panic(errors.New("阶段2失败："+strconv.Itoa(int(i))))
		}
		etags = append(etags,*(resp2.ETag))
		numbers = append(numbers,i)

		//上传完最后一块，打印出来etags  numbers
		if len<int(container) {
			fmt.Println("阶段2\n",resp)
			fmt.Println("etags:",etags)
			fmt.Println("numbers:",numbers)
			break
		}
	}


	comparts := []*s3.CompletedPart{}
	for i,etag := range etags {
		part := s3.CompletedPart{
			ETag: aws.String(etag), PartNumber: aws.Long(numbers[i]),
		}
		comparts = append(comparts,&part)
	}
	input3 := s3.CompleteMultipartUploadInput{
		Bucket:aws.String(bucketname),
		Key:aws.String(key),
		UploadID:aws.String(uploadID),
		MultipartUpload: &s3.CompletedMultipartUpload{Parts: comparts},
	}
	resp3,err := svc.CompleteMultipartUpload(&input3)
	if err!=nil {
		panic(errors.New("阶段3fail"))
	}
	fmt.Println("阶段3：\n", awsutil.StringValue(resp3))
}

//分块长串-1 获取上传id
func initiateUpload(svc *s3.S3) {
	input := s3.CreateMultipartUploadInput{
		Bucket: aws.String(bucketname),
		Key:    aws.String(uploadkey),
	}
	resp, err := svc.CreateMultipartUpload(&input)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			fmt.Println(awsErr.Code(), awsErr.Message(), awsErr.OrigErr())
			if reqErr, ok := err.(awserr.RequestFailure); ok {
				fmt.Println(reqErr.Code(), reqErr.Message(), reqErr.StatusCode(), reqErr.RequestID())
			}
		} else {
			fmt.Println(err.Error())
		}
	}

	fmt.Println("结果：\n", awsutil.StringValue(resp))
	uploadID = *(resp.UploadID)

}

//分块长串-2 上传一个分块
func partUpload(svc *s3.S3) {
	fd,err := os.Open(uploadpath)
	if err!= nil {
		panic(err)
	}
	input := s3.UploadPartInput{
		Bucket: aws.String(bucketname),
		Key:    aws.String(uploadkey),
		UploadID:aws.String(uploadID),
		PartNumber:aws.Long(1),
		Body:fd,
	}
	resp, err := svc.UploadPart(&input)
	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			fmt.Println(awsErr.Code(), awsErr.Message(), awsErr.OrigErr())
			if reqErr, ok := err.(awserr.RequestFailure); ok {
				fmt.Println(reqErr.Code(), reqErr.Message(), reqErr.StatusCode(), reqErr.RequestID())
			}
		} else {
			fmt.Println(err.Error())
		}
	}

	fmt.Println("结果：\n", awsutil.StringValue(resp))
}

//分块长串-3 组装分块
func completeUpload(svc *s3.S3) {
	input := s3.CompleteMultipartUploadInput{
		Bucket:aws.String(bucketname),
		Key:aws.String(uploadkey),
		UploadID:aws.String(uploadID),
		MultipartUpload: &s3.CompletedMultipartUpload{Parts: []*s3.CompletedPart{
			{ETag: aws.String(etag), PartNumber: aws.Long(1)},
		}},
	}
	resp,err := svc.CompleteMultipartUpload(&input)

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			fmt.Println(awsErr.Code(), awsErr.Message(), awsErr.OrigErr())
			if reqErr, ok := err.(awserr.RequestFailure); ok {
				fmt.Println(reqErr.Code(), reqErr.Message(), reqErr.StatusCode(), reqErr.RequestID())
			}
		} else {
			fmt.Println(err.Error())
		}
	}

	fmt.Println("结果：\n", awsutil.StringValue(resp))
}

//fetch obj
func fetchObj(svc *s3.S3) {
	v := url.Values{}
	v.Add("schoole", "bbvvvvvv")
	v.Add("class", "123123123123")
	XAmzTagging := v.Encode()

	sourceUrl := "https://img0.pconline.com.cn/pconline/1111/04/2483449_20061139501.jpg"

	input := s3.FetchObjectInput{
		Bucket:      aws.String(bucketname),
		Key:         aws.String("dst/testa"),
		SourceUrl:   aws.String(sourceUrl),
		XAmzTagging: aws.String(XAmzTagging),
		ACL:aws.String("public-read"),
	}
	resp, err := svc.FetchObject(&input)

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			fmt.Println(awsErr.Code(), awsErr.Message(), awsErr.OrigErr())
			if reqErr, ok := err.(awserr.RequestFailure); ok {
				fmt.Println(reqErr.Code(), reqErr.Message(), reqErr.StatusCode(), reqErr.RequestID())
			}
		} else {
			fmt.Println(err.Error())
		}
	}

	fmt.Println("结果：\n", awsutil.StringValue(resp))
}

//copy obj
func copyObj(svc *s3.S3) {
	//fd,err := os.Open("D:\\suiyi.jpg")
	v := url.Values{}
	v.Add("schoole", "yz")
	v.Add("class", "11")
	XAmzTagging := v.Encode()

	input := s3.CopyObjectInput{
		Bucket:      aws.String(bucketname),
		Key:         aws.String(key),
		CopySource:  aws.String("/cqc-test-b/yztestfile1"),
		XAmzTagging: aws.String(XAmzTagging),
		XAmzTaggingDirective:aws.String("REPLACE"),
	}
	resp, err := svc.CopyObject(&input)

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			fmt.Println(awsErr.Code(), awsErr.Message(), awsErr.OrigErr())
			if reqErr, ok := err.(awserr.RequestFailure); ok {
				fmt.Println(reqErr.Code(), reqErr.Message(), reqErr.StatusCode(), reqErr.RequestID())
			}
		} else {
			fmt.Println(err.Error())
		}
	}

	fmt.Println("结果：\n", awsutil.StringValue(resp))
}

//put obj
func putObj(svc *s3.S3) {

	fd, err := os.Open("D:\\suiyi.jpg")
	v := url.Values{}
	v.Add("name", "yz")
	v.Add("age", "11")
	XAmzTagging := v.Encode()

	input := s3.PutObjectInput{
		Bucket:      aws.String(bucketname),
		Key:         aws.String(key),
		ACL:         aws.String("public-read"),
		Body:        fd,
		XAmzTagging: aws.String(XAmzTagging),
	}
	resp, err := svc.PutObject(&input)

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			fmt.Println(awsErr.Code(), awsErr.Message(), awsErr.OrigErr())
			if reqErr, ok := err.(awserr.RequestFailure); ok {
				fmt.Println(reqErr.Code(), reqErr.Message(), reqErr.StatusCode(), reqErr.RequestID())
			}
		} else {
			fmt.Println(err.Error())
		}
	}

	fmt.Println("结果：\n", awsutil.StringValue(resp))
}

//get obj
func getObj(svc *s3.S3) {
	input := s3.GetObjectInput{
		Bucket: aws.String(bucketname),
		Key:    aws.String(key),
	}
	resp, err := svc.GetObject(&input)

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			fmt.Println(awsErr.Code(), awsErr.Message(), awsErr.OrigErr())
			if reqErr, ok := err.(awserr.RequestFailure); ok {
				fmt.Println(reqErr.Code(), reqErr.Message(), reqErr.StatusCode(), reqErr.RequestID())
			}
		} else {
			fmt.Println(err.Error())
		}
	}

	fmt.Println("结果：\n", awsutil.StringValue(resp))
}

//head obj
func headObj(svc *s3.S3) {
	input := s3.HeadObjectInput{
		Bucket: aws.String(bucketname),
		Key:    aws.String(key),
	}
	resp, err := svc.HeadObject(&input)

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			fmt.Println(awsErr.Code(), awsErr.Message(), awsErr.OrigErr())
			if reqErr, ok := err.(awserr.RequestFailure); ok {
				fmt.Println(reqErr.Code(), reqErr.Message(), reqErr.StatusCode(), reqErr.RequestID())
			}
		} else {
			fmt.Println(err.Error())
		}
	}

	fmt.Println("结果：\n", awsutil.StringValue(resp))
}
//delObjectTagging
func delTag(svc *s3.S3) {

	params := &s3.DeleteObjectTaggingInput{
		Bucket: aws.String(bucketname), // Required
		Key:    aws.String(key),
	}
	resp, err := svc.DeleteObjectTagging(params)

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			fmt.Println(awsErr.Code(), awsErr.Message(), awsErr.OrigErr())
			if reqErr, ok := err.(awserr.RequestFailure); ok {
				fmt.Println(reqErr.Code(), reqErr.Message(), reqErr.StatusCode(), reqErr.RequestID())
			}
		} else {
			fmt.Println(err.Error())
		}
	}

	fmt.Println("结果：\n", awsutil.StringValue(resp))
}

//getObjectTagging
func getTag(svc *s3.S3) {

	params := &s3.GetObjectTaggingInput{
		Bucket: aws.String(bucketname), // Required
		Key:    aws.String(key),
	}
	resp, err := svc.GetObjectTagging(params)

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			fmt.Println(awsErr.Code(), awsErr.Message(), awsErr.OrigErr())
			if reqErr, ok := err.(awserr.RequestFailure); ok {
				fmt.Println(reqErr.Code(), reqErr.Message(), reqErr.StatusCode(), reqErr.RequestID())
			}
		} else {
			fmt.Println(err.Error())
		}
	}

	fmt.Println("结果：\n", awsutil.StringValue(resp))
}

//putObjectTagging
func putTag(svc *s3.S3) {

	tagkey := "name"
	tagval := "yz"
	tagkey2 := "sex"
	tagval2 := "female"
	objTagging := s3.Tagging{
		TagSet: []*s3.Tag{&s3.Tag{
			Key:   aws.String(tagkey),
			Value: aws.String(tagval),
		}, &s3.Tag{
			Key:   aws.String(tagkey2),
			Value: aws.String(tagval2),
		},
		},
	}

	params := &s3.PutObjectTaggingInput{
		Bucket:  aws.String(bucketname), // Required
		Key:     aws.String(key),
		Tagging: &objTagging,
	}
	resp, err := svc.PutObjectTagging(params)

	if err != nil {
		if awsErr, ok := err.(awserr.Error); ok {
			fmt.Println(awsErr.Code(), awsErr.Message(), awsErr.OrigErr())
			if reqErr, ok := err.(awserr.RequestFailure); ok {
				fmt.Println(reqErr.Code(), reqErr.Message(), reqErr.StatusCode(), reqErr.RequestID())
			}
		} else {
			fmt.Println(err.Error())
		}
	}

	fmt.Println("结果：\n", awsutil.StringValue(resp))
}

//
func listBucket(client *s3.S3) {
	fmt.Println("==list all bucket in account==")
	out, err := client.ListBuckets(nil)
	if err != nil {
		panic(err)
	}
	fmt.Println(*(out.Buckets[0].Name))
}
func listObj(client *s3.S3) {
	fmt.Println("==list all obj / prefix dir/some bucketname==")
	input := &s3.ListObjectsInput{
		Bucket:    aws.String(bucketname),
		Delimiter: aws.String("/"),
		MaxKeys:   aws.Long(int64(30)),
		Prefix:    aws.String(prefix),
		Marker:    aws.String(""),
	}
	out, err := client.ListObjects(input)
	if err != nil {
		panic(err)
	}
	fmt.Println(*(out.Name))
}

func testIinitMultipartUpload(client *s3.S3) {
	params := &s3.CreateMultipartUploadInput{
		Bucket: aws.String("jiangrantest"), // bucket名称
		Key:    aws.String("qc_m.mp4"),     // object key
		//ACL:         aws.String("ObjectCannedACL"),         //权限，支持private(私有)，public-read(公开读)
		//ContentType: aws.String("application/ocet-stream"), //设置content-type
		//Metadata: map[string]*string{
		//	//"Key": aws.String("MetadataValue"), // 设置用户元数据
		//	// More values...
		//},
	}

	resp, err := client.CreateMultipartUpload(params)

	if err != nil {
		panic(err)
	}

	//获取这次初始化的uploadid
	fmt.Println(*resp.UploadID)
}

func testUploadPart(client *s3.S3, uploadId string) {

	filename := "/Users/qichao/Downloads/seapark.mp4"

	file, _ := ioutil.ReadFile(filename)

	params := &s3.UploadPartInput{
		Bucket:     aws.String("jiangrantest"), //bucket名称
		Key:        aws.String("qc_m.mp4"),     //文件名
		PartNumber: aws.Long(1),                //当前块的序号
		UploadID:   aws.String(uploadId),       //由初始化获取到得uploadid
		Body:       bytes.NewReader(file),      //当前块的内容
		//ContentLength: aws.Long(int64(len(content))), //内容长度
	}

	resp, err := client.UploadPart(params)

	if err != nil {
		panic(err)
	}

	fmt.Println(resp)
}

//设置镜像回源规则
func PutBucketMirrorRules() {

	params := &s3.PutBucketMirrorInput{
		Bucket: aws.String(BucketName), // Required
		BucketMirror: &s3.BucketMirror{
			Version:          "V3",
			UseDefaultRobots: false,
			AsyncMirrorRule: s3.AsyncMirrorRule{
				MirrorUrls: []string{
					"http://abc.om",
					"http://www.wps.cn",
				},
				SavingSetting: s3.SavingSetting{
					ACL: "private",
				},
			},
			SyncMirrorRules: []s3.SyncMirrorRules{
				{
					MatchCondition: s3.MatchCondition{
						HTTPCodes: []string{
							"404",
						},
						KeyPrefixes: []string{
							"abc",
						},
					},
					MirrorURL: "http://v-ks-a-i.originalvod.com",
					MirrorRequestSetting: s3.MirrorRequestSetting{
						PassQueryString: false,
						Follow3Xx:       false,
						HeaderSetting: s3.HeaderSetting{
							SetHeaders: []s3.SetHeaders{
								{
									Key:   "d",
									Value: "b",
								},
							},
							RemoveHeaders: []s3.RemoveHeaders{
								{
									Key: "d",
								},
								{
									Key: "d",
								},
							},
							PassAll: false,
							PassHeaders: []s3.PassHeaders{
								{
									Key: "asdb",
								},
							},
						},
					},
					SavingSetting: s3.SavingSetting{
						ACL: "private",
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
func GetBucketMirrorRules() {

	params := &s3.GetBucketMirrorInput{
		Bucket: aws.String(BucketName),
	}
	resp, _ := client.GetBucketMirror(params)
	fmt.Println("resp.code is:", resp.HttpCode)
	fmt.Println("resp.Header is:", resp.Header)
	// Pretty-print the response data.
	var bodyStr = string(resp.Body[:])
	fmt.Println("resp.Body is:", bodyStr)

}

//删除镜像回源规则
func DeleteBucketMirrorRules() {

	params := &s3.DeleteBucketMirrorInput{
		Bucket: aws.String(BucketName),
	}
	resp, _ := client.DeleteBucketMirror(params)
	fmt.Println("resp.code is:", resp.HttpCode)
	fmt.Println("resp.Header is:", resp.Header)
	// Pretty-print the response data.
	var bodyStr = string(resp.Body[:])
	fmt.Println("resp.Body is:", bodyStr)

}
