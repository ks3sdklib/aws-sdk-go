package lib

import (
	"fmt"
	"github.com/ks3sdklib/aws-sdk-go/aws"
	"github.com/ks3sdklib/aws-sdk-go/aws/credentials"
	"github.com/ks3sdklib/aws-sdk-go/service/s3"
	. "gopkg.in/check.v1"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"
)

// Hook up go check into the "go test" runner.
func Test(t *testing.T) {
	TestingT(t)
}

type Ks3utilCommandSuite struct {
	startT time.Time
}

var _ = Suite(&Ks3utilCommandSuite{})

var (
	endpoint                = os.Getenv("KS3_TEST_ENDPOINT")
	accessKeyID             = os.Getenv("KS3_TEST_ACCESS_KEY_ID")
	accessKeySecret         = os.Getenv("KS3_TEST_ACCESS_KEY_SECRET")
	bucket                  = os.Getenv("KS3_TEST_BUCKET")
	region                  = os.Getenv("KS3_TEST_REGION")
	bucketEndpoint          = os.Getenv("KS3_TEST_BUCKET_ENDPOINT")
	key                     = randLowStr(10)
	logPath                 = "report/ks3go-sdk-test_" + time.Now().Format("20060102_150405") + ".log"
	content                 = "abc"
	client           *s3.S3 = nil
	out                     = os.Stdout
	errout                  = os.Stderr
	sleepTime               = time.Second
	commonNamePrefix        = "go-sdk-test-"
	testFileDir             = "go-sdk-test-file/"
	timeout                 = 1 * time.Microsecond
)

// SetUpSuite 在测试套件启动前执行一次
func (s *Ks3utilCommandSuite) SetUpSuite(c *C) {
	fmt.Printf("set up Ks3utilCommandSuite\n")
	var cre = credentials.NewStaticCredentials(accessKeyID, accessKeySecret, "") //online
	client = s3.New(&aws.Config{
		Credentials:      cre,      // 访问凭证
		Region:           region,   // 填写您的Region
		Endpoint:         endpoint, // 填写您的Endpoint
		DisableSSL:       false,    // 禁用HTTPS，默认值为false
		LogLevel:         aws.Off,  // 日志等级，默认关闭日志，可选值：Off, Error, Warn, Info, Debug
		LogHTTPBody:      false,    // 把HTTP请求body打入日志，默认值为false
		Logger:           nil,      // 日志输出位置，可设置指定文件
		S3ForcePathStyle: false,    // 使用二级域名，默认值为false
		DomainMode:       false,    // 开启自定义Bucket绑定域名，当开启时S3ForcePathStyle参数不生效，默认值为false
		SignerVersion:    "V2",     // 签名方式可选值有：V2 OR V4 OR V4_UNSIGNED_PAYLOAD_SIGNER，默认值为V2
		MaxRetries:       3,        // 请求失败时最大重试次数，值小于0时不重试
		CrcCheckEnabled:  true,     // 开启CRC64校验，默认值为false
		HTTPClient:       nil,      // HTTP请求的Client对象，若为空则使用默认值
	})

	s.SetUpBucketEnv(c)
	err := os.MkdirAll(testFileDir, os.FileMode(0775))
	c.Assert(err, IsNil)
}

// SetUpBucketEnv 测试开始时，创建测试用bucket
func (s *Ks3utilCommandSuite) SetUpBucketEnv(c *C) {
	bucket = commonNamePrefix + randLowStr(10)
	_, err := client.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(bucket),
	})
	c.Assert(err, IsNil)
	fmt.Printf("create bucket:%s\n", bucket)
}

// TearDownSuite 在测试套件用例都执行完成后执行一次
func (s *Ks3utilCommandSuite) TearDownSuite(c *C) {
	fmt.Printf("tear down Ks3utilCommandSuite\n")
	// 删除测试bucket
	RemoveBuckets(commonNamePrefix, c)
	// 删除测试文件夹
	os.RemoveAll(testFileDir)
}

// RemoveBuckets 删除以prefix开头的bucket
func RemoveBuckets(prefix string, c *C) {
	resp, err := client.ListBuckets(&s3.ListBucketsInput{})
	c.Assert(err, IsNil)
	for _, bucket := range resp.Buckets {
		bucketName := *bucket.Name
		if strings.Contains(bucketName, prefix) {
			fmt.Printf("remove bucket begin:%s\n", bucketName)
			// 1. 删除bucket中的全部对象
			RemoveObjects(bucketName, c)
			// 2.删除bucket中未完成的分块上传任务
			RemoveMultipartUploads(bucketName, c)
			// 3. 删除bucket
			RemoveBucket(bucketName, c)
			fmt.Printf("remove bucket end:%s\n", bucketName)
		}
	}
}

// RemoveObjects 删除bucket中的全部对象
func RemoveObjects(bucketName string, c *C) {
	resp, err := client.DeleteBucketPrefix(&s3.DeleteBucketPrefixInput{
		Bucket:          aws.String(bucketName),
		IsReTurnResults: aws.Boolean(true),
	})
	c.Assert(err, IsNil)
	c.Assert(len(resp.Errors), Equals, 0)
}

// RemoveMultipartUploads 删除bucket中未完成的分块上传任务
func RemoveMultipartUploads(bucketName string, c *C) {
	resp, err := client.ListMultipartUploads(&s3.ListMultipartUploadsInput{
		Bucket: aws.String(bucketName),
	})
	c.Assert(err, IsNil)
	for _, upload := range resp.Uploads {
		_, err := client.AbortMultipartUpload(&s3.AbortMultipartUploadInput{
			Bucket:   aws.String(bucketName),
			Key:      upload.Key,
			UploadID: upload.UploadID,
		})
		c.Assert(err, IsNil)
	}
}

// RemoveBucket 删除bucket
func RemoveBucket(bucketName string, c *C) {
	_, err := client.DeleteBucket(&s3.DeleteBucketInput{
		Bucket: aws.String(bucketName),
	})
	c.Assert(err, IsNil)
}

// SetUpTest 在每个用例执行前执行一次
func (s *Ks3utilCommandSuite) SetUpTest(c *C) {
	fmt.Printf("set up test:%s\n", c.TestName())
	s.startT = time.Now()
}

var letters = []rune("0123456789abcdefghijklmnopqrstuvwxyz")

// TearDownTest 在每个用例执行后执行一次
func (s *Ks3utilCommandSuite) TearDownTest(c *C) {
	endT := time.Now()
	cost := endT.UnixNano()/1000/1000 - s.startT.UnixNano()/1000/1000
	fmt.Printf("tear down test:%s,cost:%d(ms)\n", c.TestName(), cost)
}

// randStr 生成随机字符串
func randStr(n int) string {
	b := make([]rune, n)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := range b {
		b[i] = letters[r.Intn(len(letters))]
	}
	return string(b)
}

// randLowStr 生成随机小写字符串
func randLowStr(n int) string {
	return strings.ToLower(randStr(n))
}

// createFile 创建文件
func createFile(filePath string, size int64) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Set the file size
	err = file.Truncate(size)
	if err != nil {
		return err
	}

	fmt.Printf("File created: %s (size: %d bytes)\n", filePath, size)
	return nil
}

// createFileWithContent 创建文件并写入内容
func createFileWithContent(filePath, content string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		return err
	}
	return nil
}

// PutObject 上传单个文件
func (s *Ks3utilCommandSuite) PutObject(key string, c *C) {
	_, err := client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   strings.NewReader(content),
	})
	c.Assert(err, IsNil)
}

// CopyObject 拷贝单个文件
func (s *Ks3utilCommandSuite) CopyObject(srcKey string, dstKey string, c *C) {
	_, err := client.CopyObject(&s3.CopyObjectInput{
		Bucket:       aws.String(bucket),
		Key:          aws.String(dstKey),
		SourceBucket: aws.String(bucket),
		SourceKey:    aws.String(srcKey),
	})
	c.Assert(err, IsNil)
}

// HeadObject 判断文件是否存在
func (s *Ks3utilCommandSuite) HeadObject(key string, c *C) {
	resp, err := client.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	c.Assert(err, IsNil)
	c.Assert(*resp.StatusCode, Equals, int64(200))
}

// DeleteObject 删除单个文件
func (s *Ks3utilCommandSuite) DeleteObject(key string, c *C) {
	_, err := client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	c.Assert(err, IsNil)
}

func (s *Ks3utilCommandSuite) DeleteObjectWithClient(client *s3.S3, bucket string, key string, c *C) {
	_, err := client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	c.Assert(err, IsNil)
}

func (s *Ks3utilCommandSuite) CreateBucket(bucketName string, c *C) {
	_, err := client.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	})
	c.Assert(err, IsNil)
}

func (s *Ks3utilCommandSuite) DeleteBucket(bucketName string, c *C) {
	_, err := client.DeleteBucket(&s3.DeleteBucketInput{
		Bucket: aws.String(bucketName),
	})
	c.Assert(err, IsNil)
}
