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

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) {
	TestingT(t)
}

type Ks3utilCommandSuite struct {
	startT time.Time
}

var _ = Suite(&Ks3utilCommandSuite{})

var (
	endpoint        = os.Getenv("KS3_TEST_ENDPOINT")
	accessKeyID     = os.Getenv("KS3_TEST_ACCESS_KEY_ID")
	accessKeySecret = os.Getenv("KS3_TEST_ACCESS_KEY_SECRET")
	bucket          = os.Getenv("KS3_TEST_BUCKET")
	region          = os.Getenv("KS3_TEST_REGION")
	bucketEndpoint  = os.Getenv("KS3_TEST_BUCKET_ENDPOINT")

	logPath          = "report/ks3go-sdk-test_" + time.Now().Format("20060102_150405") + ".log"
	content          = "abc"
	client    *s3.S3 = nil
	out              = os.Stdout
	errout           = os.Stderr
	sleepTime        = time.Second

	//accessKeyID     = "p0C0lcSAet4Bdfk8"
	//accessKeySecret = "e40keB3spMJ7Z65rttEPZiCZqz7Vos5s"
	//bucket          = "aaaab"
	//endpoint        = "aaaab.cqc.cool:9000"
	//region          = "us-east-1"
)

var (
	commonNamePrefix = "go-sdk-test-"
)

// Run once when the suite starts running
func (s *Ks3utilCommandSuite) SetUpSuite(c *C) {

	fmt.Printf("set up Ks3utilCommandSuite\n")
	var cre = credentials.NewStaticCredentials(accessKeyID, accessKeySecret, "") //online
	client = s3.New(&aws.Config{
		//Region 可参考 https://docs.ksyun.com/documents/6761
		Region:      region,
		Credentials: cre,
		//Endpoint 可参考 https://docs.ksyun.com/documents/6761
		Endpoint:         endpoint,
		DisableSSL:       true,  //是否禁用https
		LogLevel:         0,     //是否开启日志,0为关闭日志，1为开启日志
		LogHTTPBody:      false, //是否把HTTP请求body打入日志
		S3ForcePathStyle: false,
		Logger:           nil,   //打日志的位置
		DomainMode:       false, //是否开启自定义bucket绑定域名，当开启时 S3ForcePathStyle 参数不生效。
		//可选值有 ： V2 OR V4 OR V4_UNSIGNED_PAYLOAD_SIGNER
		SignerVersion: "V4",
		MaxRetries:    1,
	})

	//创建测试文件
	//s.createFile(key, content, c)
	//fd, _ := os.Open(content)
	//input := s3.PutObjectInput{
	//	Bucket: aws.String(bucket),
	//	Key:    aws.String(key),
	//	ACL:    aws.String("public-read"),
	//	Body:   fd,
	//}
	//client.PutObject(&input)
	s.SetUpBucketEnv(c)
}

func (s *Ks3utilCommandSuite) SetUpBucketEnv(c *C) {
	os.Remove(key)
}

// Run before each test or benchmark starts running
func (s *Ks3utilCommandSuite) TearDownSuite(c *C) {
	fmt.Printf("tear down Ks3utilCommandSuite\n")
	//os.Stdout = out
	//os.Stderr = errout
	//os.Remove(key)
}

var a int = 1

// Run after each test or benchmark runs
func (s *Ks3utilCommandSuite) SetUpTest(c *C) {
	fmt.Printf("set up test:%s\n", c.TestName())
	s.startT = time.Now()
}

var letters = []rune("0123456789abcdefghijklmnopqrstuvwxyz")

// Run once after all tests or benchmarks have finished running
func (s *Ks3utilCommandSuite) TearDownTest(c *C) {
	endT := time.Now()
	cost := endT.UnixNano()/1000/1000 - s.startT.UnixNano()/1000/1000
	fmt.Printf("tear down test:%s,cost:%d(ms)\n", c.TestName(), cost)
	a = a + 1
}

func randStr(n int) string {
	b := make([]rune, n)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := range b {
		b[i] = letters[r.Intn(len(letters))]
	}
	return string(b)
}

func randLowStr(n int) string {
	return strings.ToLower(randStr(n))
}

func (s *Ks3utilCommandSuite) createFile(fileName, content string, c *C) {
	fout, err := os.Create(fileName)
	defer fout.Close()
	c.Assert(err, IsNil)
	_, err = fout.WriteString(content)
	c.Assert(err, IsNil)
}
