package v4

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ks3sdklib/aws-sdk-go/aws"
	"github.com/ks3sdklib/aws-sdk-go/aws/credentials"
	"github.com/ks3sdklib/aws-sdk-go/internal/protocol/rest"
	"github.com/stretchr/testify/assert"
)

func buildSigner(serviceName string, region string, signTime time.Time, expireTime time.Duration, body string) signer {
	endpoint := "https://" + serviceName + "." + region + ".amazonaws.com"
	reader := strings.NewReader(body)
	req, _ := http.NewRequest("POST", endpoint, reader)
	rawPath := "/bucket/key-._~%2C%21%40%23%24%25%5E%26%2A%28%29"
	decoded, _ := url.PathUnescape(rawPath)
	req.URL.Path = decoded
	req.URL.RawPath = rawPath
	req.Header.Add("X-Amz-Target", "prefix.Operation")
	req.Header.Add("Content-Type", "application/x-amz-json-1.0")
	req.Header.Add("Content-Length", strconv.Itoa(len(body)))
	req.Header.Add("X-Amz-Meta-Other-Header", "some-value=!@#$%^&* (+)")

	return signer{
		Request:     req,
		Time:        signTime,
		ExpireTime:  int64(expireTime),
		Query:       req.URL.Query(),
		Body:        reader,
		ServiceName: serviceName,
		Region:      region,
		awsRequest:  &aws.Request{},
		Service:     &aws.Service{Config: aws.DefaultConfig.Merge(&aws.Config{Credentials: credentials.NewStaticCredentials("AKID", "SECRET", "SESSION")})},
		Credentials: credentials.NewStaticCredentials("AKID", "SECRET", "SESSION"),
		isSignBody:  true,
	}
}

func TestSignEmptyBody(t *testing.T) {
	signer := buildSigner("dynamodb", "us-east-1", time.Now(), 0, "")
	signer.Body = nil
	signer.sign()
	hash := signer.Request.Header.Get("X-Amz-Content-Sha256")
	assert.Equal(t, "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", hash)
}

func TestSignBody(t *testing.T) {
	signer := buildSigner("dynamodb", "us-east-1", time.Now(), 0, "hello")
	signer.sign()
	hash := signer.Request.Header.Get("X-Amz-Content-Sha256")
	assert.Equal(t, "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824", hash)
}

func TestSignSeekedBody(t *testing.T) {
	signer := buildSigner("dynamodb", "us-east-1", time.Now(), 0, "   hello")
	signer.Body.Read(make([]byte, 3)) // consume first 3 bytes so body is now "hello"
	signer.sign()
	hash := signer.Request.Header.Get("X-Amz-Content-Sha256")
	assert.Equal(t, "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824", hash)

	start, _ := signer.Body.Seek(0, 1)
	assert.Equal(t, int64(3), start)
}

func TestPresignEmptyBodyS3(t *testing.T) {
	signer := buildSigner("s3", "us-east-1", time.Now(), 5*time.Minute, "hello")
	signer.sign()
	hash := signer.Request.Header.Get("X-Amz-Content-Sha256")
	assert.Equal(t, "UNSIGNED-PAYLOAD", hash)
}

func TestSignPrecomputedBodyChecksum(t *testing.T) {
	signer := buildSigner("dynamodb", "us-east-1", time.Now(), 0, "hello")
	signer.Request.Header.Set("X-Amz-Content-Sha256", "PRECOMPUTED")
	signer.sign()
	hash := signer.Request.Header.Get("X-Amz-Content-Sha256")
	assert.Equal(t, "PRECOMPUTED", hash)
}

func TestAnonymousCredentials(t *testing.T) {
	r := aws.NewRequest(
		aws.NewService(&aws.Config{Credentials: credentials.AnonymousCredentials}),
		&aws.Operation{
			Name:       "BatchGetItem",
			HTTPMethod: "POST",
			HTTPPath:   "/",
		},
		nil,
		nil,
	)
	Sign(r)

	urlQ := r.HTTPRequest.URL.Query()
	assert.Empty(t, urlQ.Get("X-Amz-Signature"))
	assert.Empty(t, urlQ.Get("X-Amz-Credential"))
	assert.Empty(t, urlQ.Get("X-Amz-SignedHeaders"))
	assert.Empty(t, urlQ.Get("X-Amz-Date"))

	hQ := r.HTTPRequest.Header
	assert.Empty(t, hQ.Get("Authorization"))
	assert.Empty(t, hQ.Get("X-Amz-Date"))
}

func TestIgnoreResignRequestWithValidCreds(t *testing.T) {
	r := aws.NewRequest(
		aws.NewService(&aws.Config{
			Credentials: credentials.NewStaticCredentials("AKID", "SECRET", "SESSION"),
			Region:      "us-west-2",
		}),
		&aws.Operation{
			Name:       "BatchGetItem",
			HTTPMethod: "POST",
			HTTPPath:   "/",
		},
		nil,
		nil,
	)

	Sign(r)
	sig := r.HTTPRequest.Header.Get("Authorization")

	Sign(r)
	assert.Equal(t, sig, r.HTTPRequest.Header.Get("Authorization"))
}

func TestIgnorePreResignRequestWithValidCreds(t *testing.T) {
	r := aws.NewRequest(
		aws.NewService(&aws.Config{
			Credentials: credentials.NewStaticCredentials("AKID", "SECRET", "SESSION"),
			Region:      "us-west-2",
		}),
		&aws.Operation{
			Name:       "BatchGetItem",
			HTTPMethod: "POST",
			HTTPPath:   "/",
		},
		nil,
		nil,
	)
	r.ExpireTime = 10

	Sign(r)
	sig := r.HTTPRequest.Header.Get("X-Amz-Signature")

	Sign(r)
	assert.Equal(t, sig, r.HTTPRequest.Header.Get("X-Amz-Signature"))
}

func TestResignRequestExpiredCreds(t *testing.T) {
	creds := credentials.NewStaticCredentials("AKID", "SECRET", "SESSION")
	r := aws.NewRequest(
		aws.NewService(&aws.Config{Credentials: creds}),
		&aws.Operation{
			Name:       "BatchGetItem",
			HTTPMethod: "POST",
			HTTPPath:   "/",
		},
		nil,
		nil,
	)
	Sign(r)
	querySig := r.HTTPRequest.Header.Get("Authorization")

	creds.Expire()

	Sign(r)
	assert.NotEqual(t, querySig, r.HTTPRequest.Header.Get("Authorization"))
}

// oldV4URI 为 uri 计算的 Opaque 版本实现，newV4URI 为 EscapedPath 版本实现，用于等价对比。
func oldV4URI(s *signer) string {
	uri := strings.Replace(s.Request.URL.Opaque, "%2F", "/", -1)
	if uri != "" {
		uri = "/" + strings.Join(strings.Split(uri, "/")[3:], "/")
	} else {
		uri = s.Request.URL.Path
	}
	return finalizeV4URI(uri, s.ServiceName)
}

func newV4URI(s *signer) string {
	return finalizeV4URI(strings.Replace(s.Request.URL.EscapedPath(), "%2F", "/", -1), s.ServiceName)
}

func finalizeV4URI(uri, serviceName string) string {
	if uri == "" {
		uri = "/"
	}
	if serviceName != "s3" {
		uri = rest.EscapePath(uri, false)
	}
	return uri
}

func TestBuildCanonicalStringURIEquivalence(t *testing.T) {
	cases := []struct {
		name    string
		host    string
		opaque  string // Opaque 版本输入
		rawPath string // EscapedPath 版本输入（Amazon 编码）
		path    string // EscapedPath 版本输入（解码 Path）
	}{
		{"path-style object", "s3.example.com", "//s3.example.com/bucket/key", "/bucket/key", "/bucket/key"},
		{"vhost object", "bucket.s3.example.com", "//bucket.s3.example.com/key", "/key", "/key"},
		{"bucket-only path-style", "s3.example.com", "//s3.example.com/bucket", "/bucket", "/bucket"},
		{"以/开头 key(%2F)", "s3.example.com", "//s3.example.com/bucket/%2Fkey", "/bucket/%2Fkey", "/bucket//key"},
		{"根路径", "s3.example.com", "//s3.example.com/", "/", "/"},
	}
	for _, tc := range cases {
		// Opaque 版本
		uOld, _ := url.Parse("https://" + tc.host)
		uOld.Opaque = tc.opaque
		sOld := signer{Request: &http.Request{URL: uOld}, ServiceName: "s3"}
		// EscapedPath 版本
		uNew, _ := url.Parse("https://" + tc.host)
		uNew.Path = tc.path
		uNew.RawPath = tc.rawPath
		sNew := signer{Request: &http.Request{URL: uNew}, ServiceName: "s3"}

		gotOld := oldV4URI(&sOld)
		gotNew := newV4URI(&sNew)
		if gotOld != gotNew {
			t.Errorf("%s: EscapedPath版=%q Opaque版=%q", tc.name, gotNew, gotOld)
		} else {
			t.Logf("%-24s -> %q", tc.name, gotNew)
		}
	}
}
