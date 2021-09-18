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
	"github.com/ks3sdklib/aws-sdk-go/service/utils"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"strconv"
	"sync"
)

var (
	region     = "BEIJING"
	ak         = "AK"
	sk         = "SK"
	endpoint   = "ks3-cn-beijing.ksyuncs.com"
	bucketname = ""
	key        = "1.ipa"
	prefix     = "test/" //目录名称
	objname    = "file1.mp4"
)

func main() {
	credentials := credentials.NewStaticCredentials(ak, sk, "")
	client := s3.New(&aws.Config{
		Region:           region,
		Credentials:      credentials,
		Endpoint:         endpoint, //ks3地址
		DisableSSL:       true,     //是否禁用https
		LogLevel:         0,        //是否开启日志,0为关闭日志，1为开启日志
		S3ForcePathStyle: true,     //是否强制使用path style方式访问
		LogHTTPBody:      false,    //是否把HTTP请求body打入日志
		Logger:           nil,      //打日志的位置
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
	initNewUploader(client)
	//PutBucketMirror(client)

	//CreateBucket(client)
}

var (
	uploadkey  = "src.zipold"
	uploadID   = "cc0e00789d724b5d81dca5f40a1a1dd2"
	uploadpath = "C:\\Users\\yangzhen1\\Desktop\\es10000Console_old\\src.zipold"
	etag       = "504eb4064ceaf14ea157aff50cfb3c15"
)

func multipartUploadAllTagging(svc *s3.S3) {

	var bucket = "cqc-test-b"                                                      //欲上传的桶的名字
	var name = "test012506.zip"                                                    //上传的对象的新的名字
	var container = int64(4096 * 4096)                                             //每次上传的块的大小
	var filepath = "C:\\Users\\yangzhen1\\Desktop\\es10000Console_old\\src.zipold" //本地需要上传的文件

	var uploadID = ""
	var etags = []string{}  //每次上传块返回的etag
	var numbers = []int64{} //每次上传块号

	v := url.Values{}
	v.Add("name", "yz")
	v.Add("sex", "female")
	tag := v.Encode()
	input := s3.CreateMultipartUploadInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(name),
		XAmzTagging: aws.String(tag),
	}
	resp, err := svc.CreateMultipartUpload(&input)
	if err != nil {
		panic(errors.New("阶段1fail"))
	}
	uploadID = *(resp.UploadID)
	fmt.Println("阶段1\n", resp)

	fd, err := os.Open(filepath)
	defer fd.Close()
	var i int64 = 0
	for {
		offset := i * container
		buffer := make([]byte, container)
		len, err := fd.ReadAt(buffer, offset)
		if err != nil && err != io.EOF {
			panic(err)
		}
		fmt.Println("读取了：", len)
		i++
		//上传之
		input2 := s3.UploadPartInput{
			Bucket:     aws.String(bucket),
			Key:        aws.String(name),
			UploadID:   aws.String(uploadID),
			PartNumber: aws.Long(i),
			Body:       bytes.NewReader(buffer),
		}
		resp2, err := svc.UploadPart(&input2)
		if err != nil {
			panic(errors.New("阶段2失败：" + strconv.Itoa(int(i))))
		}
		etags = append(etags, *(resp2.ETag))
		numbers = append(numbers, i)

		//上传完最后一块，打印出来etags  numbers
		if len < int(container) {
			fmt.Println("阶段2\n", resp)
			fmt.Println("etags:", etags)
			fmt.Println("numbers:", numbers)
			break
		}
	}

	comparts := []*s3.CompletedPart{}
	for i, etag := range etags {
		part := s3.CompletedPart{
			ETag: aws.String(etag), PartNumber: aws.Long(numbers[i]),
		}
		comparts = append(comparts, &part)
	}
	input3 := s3.CompleteMultipartUploadInput{
		Bucket:          aws.String(bucket),
		Key:             aws.String(name),
		UploadID:        aws.String(uploadID),
		MultipartUpload: &s3.CompletedMultipartUpload{Parts: comparts},
	}
	resp3, err := svc.CompleteMultipartUpload(&input3)
	if err != nil {
		panic(errors.New("阶段3fail"))
	}
	fmt.Println("阶段3：\n", awsutil.StringValue(resp3))
}

func multipartUploadAll(svc *s3.S3) {

	var container = int64(4096 * 4096)                                             //每次上传的块的大小
	var filepath = "C:\\Users\\yangzhen1\\Desktop\\es10000Console_old\\src.zipold" //本地需要上传的文件

	var uploadID = ""
	var etags = []string{}  //每次上传块返回的etag
	var numbers = []int64{} //每次上传块号

	input := s3.CreateMultipartUploadInput{
		Bucket: aws.String(bucketname),
		Key:    aws.String(key),
	}
	resp, err := svc.CreateMultipartUpload(&input)
	if err != nil {
		panic(errors.New("阶段1fail"))
	}
	uploadID = *(resp.UploadID)
	fmt.Println("阶段1\n", resp)

	fd, err := os.Open(filepath)
	defer fd.Close()
	var i int64 = 0
	for {
		offset := i * container
		buffer := make([]byte, container)
		len, err := fd.ReadAt(buffer, offset)
		if err != nil && err != io.EOF {
			panic(err)
		}
		fmt.Println("读取了：", len)
		i++
		//上传之
		input2 := s3.UploadPartInput{
			Bucket:     aws.String(bucketname),
			Key:        aws.String(key),
			UploadID:   aws.String(uploadID),
			PartNumber: aws.Long(i),
			Body:       bytes.NewReader(buffer),
		}
		resp2, err := svc.UploadPart(&input2)
		if err != nil {
			panic(errors.New("阶段2失败：" + strconv.Itoa(int(i))))
		}
		etags = append(etags, *(resp2.ETag))
		numbers = append(numbers, i)

		//上传完最后一块，打印出来etags  numbers
		if len < int(container) {
			fmt.Println("阶段2\n", resp)
			fmt.Println("etags:", etags)
			fmt.Println("numbers:", numbers)
			break
		}
	}

	comparts := []*s3.CompletedPart{}
	for i, etag := range etags {
		part := s3.CompletedPart{
			ETag: aws.String(etag), PartNumber: aws.Long(numbers[i]),
		}
		comparts = append(comparts, &part)
	}
	input3 := s3.CompleteMultipartUploadInput{
		Bucket:          aws.String(bucketname),
		Key:             aws.String(key),
		UploadID:        aws.String(uploadID),
		MultipartUpload: &s3.CompletedMultipartUpload{Parts: comparts},
	}
	resp3, err := svc.CompleteMultipartUpload(&input3)
	if err != nil {
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
	fd, err := os.Open(uploadpath)
	if err != nil {
		panic(err)
	}
	input := s3.UploadPartInput{
		Bucket:     aws.String(bucketname),
		Key:        aws.String(uploadkey),
		UploadID:   aws.String(uploadID),
		PartNumber: aws.Long(1),
		Body:       fd,
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
		Bucket:   aws.String(bucketname),
		Key:      aws.String(uploadkey),
		UploadID: aws.String(uploadID),
		MultipartUpload: &s3.CompletedMultipartUpload{Parts: []*s3.CompletedPart{
			{ETag: aws.String(etag), PartNumber: aws.Long(1)},
		}},
	}
	resp, err := svc.CompleteMultipartUpload(&input)

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
		ACL:         aws.String("public-read"),
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
		Bucket:               aws.String(bucketname),
		Key:                  aws.String(key),
		CopySource:           aws.String("/cqc-test-b/yztestfile1"),
		XAmzTagging:          aws.String(XAmzTagging),
		XAmzTaggingDirective: aws.String("REPLACE"),
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

func initNewUploader(client *s3.S3) {

	wg := sync.WaitGroup{}
	mgr := s3manager.NewUploader(&s3manager.UploadOptions{
		S3:          client,
		PartSize:    1024 * 1024 * 7,
		Concurrency: 1,
	})
	dir := "/Users/cqc/data/临时文件/tmp"
	files, _ := utils.WalkDir(dir, "")
	for i := 0; i < len(files); i++ {
		path := files[i]
		wg.Add(1)
		go testUpload(mgr, path, &wg)
		//fmt.Println("Sleep 3")
		//time.Sleep(3*time.Second)
		//fmt.Println("Sleep 3 over")
	}
	wg.Wait()
	fmt.Println("over")
}

func testUpload(mgr *s3manager.Uploader, fileName string, wg *sync.WaitGroup) {

	defer wg.Done()
	file, _ := ioutil.ReadFile(fileName)
	filenameall := path.Base(fileName)
	fmt.Println("upload path :" + fileName)
	_, err := mgr.Upload(&s3manager.UploadInput{
		Bucket: aws.String("12312vbb"),
		Key:    aws.String(filenameall),
		Body:   bytes.NewReader(file),
	})
	if err == nil {
		fmt.Println(fileName + " upload ok")
	}
}

