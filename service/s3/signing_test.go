package s3

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sort"
	"strings"
	"testing"

	"github.com/ks3sdklib/aws-sdk-go/aws"
	"github.com/ks3sdklib/aws-sdk-go/aws/credentials"
)

// 端到端验证真实发出的请求行应为 origin-form（不含 scheme://host）。
func TestRequestLineOriginFormE2E(t *testing.T) {
	var seenURI string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenURI = r.RequestURI
		w.WriteHeader(200)
		_, _ = io.Copy(io.Discard, r.Body)
		_ = r.Body.Close()
	}))
	defer srv.Close()

	cases := []struct {
		name      string
		pathStyle bool
		domain    bool
		do        func(c *S3) error
		expect    string // 期望的服务端 RequestURI
	}{
		{"PutObject path-style", true, false,
			func(c *S3) error {
				_, err := c.PutObject(&PutObjectInput{Bucket: aws.String("b"), Key: aws.String("demo/hello.txt"), Body: strings.NewReader("hello")})
				return err
			},
			"/b/demo/hello.txt"},
		{"GetObject path-style", true, false,
			func(c *S3) error {
				_, err := c.GetObject(&GetObjectInput{Bucket: aws.String("b"), Key: aws.String("demo/hello.txt")})
				return err
			},
			"/b/demo/hello.txt"},
		{"ListBuckets path-style", true, false,
			func(c *S3) error { _, err := c.ListBuckets(&ListBucketsInput{}); return err },
			"/"},
		{"PutObject DomainMode", false, true,
			func(c *S3) error {
				_, err := c.PutObject(&PutObjectInput{Bucket: aws.String("b"), Key: aws.String("k"), Body: strings.NewReader("x")})
				return err
			},
			"/k"},
	}
	for _, tc := range cases {
		seenURI = ""
		c := newTestClient(srv.URL, "AK", "SK", tc.pathStyle, tc.domain, "")
		_ = tc.do(c)
		t.Logf("%-26s | RequestURI=%q", tc.name, seenURI)
		if seenURI != tc.expect {
			t.Errorf("%s: got %q, want %q", tc.name, seenURI, tc.expect)
		}
	}
}

// TestV2SignatureMatchesRequestTarget 模拟服务端用 request-target 做 V2 验签，
// 确认客户端签名与服务端按请求行重算的签名一致。
func TestV2SignatureMatchesRequestTarget(t *testing.T) {
	var (
		gotAuth, gotURI, gotDate, gotCT, gotMD5 string
		gotHeaders                              http.Header
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotURI = r.RequestURI
		gotDate = r.Header.Get("Date")
		gotCT = r.Header.Get("Content-Type")
		gotMD5 = r.Header.Get("Content-Md5")
		gotHeaders = r.Header
		w.WriteHeader(200)
		_, _ = io.Copy(io.Discard, r.Body)
		_ = r.Body.Close()
	}))
	defer srv.Close()

	c := newTestClient(srv.URL, "AKEXAMPLE", "SKEXAMPLE", true, false, "")
	if _, err := c.PutObject(&PutObjectInput{
		Bucket: aws.String("example-bucket"),
		Key:    aws.String("demo/hello.txt"),
		Body:   strings.NewReader("hello ks3 v2 upload demo"),
	}); err != nil {
		t.Fatalf("PutObject: %v", err)
	}

	// 服务端按 SigV2 用 request-target 重算签名
	canonicalHeaders := buildCanonicalHeaders(gotHeaders)
	signItems := []string{"PUT", gotMD5, gotCT, gotDate}
	if canonicalHeaders != "" {
		signItems = append(signItems, canonicalHeaders)
	}
	signItems = append(signItems, gotURI)
	stringToSign := strings.Join(signItems, "\n")

	mac := hmac.New(sha1.New, []byte("SKEXAMPLE"))
	mac.Write([]byte(stringToSign))
	serverSig := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	expectedAuth := "AWS AKEXAMPLE:" + serverSig

	if gotAuth != expectedAuth {
		t.Errorf("签名不匹配: 服务端按 request-target 验签失败\n服务端=%q\n客户端=%q", expectedAuth, gotAuth)
	}
}

// newTestClient 构造指向指定 endpoint 的测试 client，Region 固定 us-east-1。
func newTestClient(endpoint, ak, sk string, pathStyle, domainMode bool, signerVersion string) *S3 {
	return New(&aws.Config{
		Credentials:      credentials.NewStaticCredentials(ak, sk, ""),
		Endpoint:         endpoint,
		S3ForcePathStyle: pathStyle,
		DomainMode:       domainMode,
		Region:           "us-east-1",
		SignerVersion:    signerVersion,
	})
}

// buildCanonicalHeaders 构造 SigV2 的 CanonicalizedAmzHeaders。
func buildCanonicalHeaders(h http.Header) string {
	var keys []string
	for k := range h {
		if strings.HasPrefix(strings.ToLower(k), "x-amz-") {
			keys = append(keys, strings.ToLower(k))
		}
	}
	sort.Strings(keys)
	var parts []string
	for _, k := range keys {
		parts = append(parts, k+":"+strings.Join(h[http.CanonicalHeaderKey(k)], ","))
	}
	return strings.Join(parts, "\n")
}

// TestPresignedURLForm 验证预签名/分享 URL 形态正常（scheme://host/path?query）。
func TestPresignedURLForm(t *testing.T) {
	cases := []struct {
		name string
		sign string
		gen  func(c *S3) (string, error)
	}{
		{"V2 预签名", "",
			func(c *S3) (string, error) {
				return c.GeneratePresignedUrl(&GeneratePresignedUrlInput{
					Bucket:     aws.String("b"),
					Key:        aws.String("demo/hello.txt"),
					HTTPMethod: "GET",
					Expires:    900,
				})
			}},
		{"V4 预签名", "V4",
			func(c *S3) (string, error) {
				return c.GeneratePresignedUrl(&GeneratePresignedUrlInput{
					Bucket:     aws.String("b"),
					Key:        aws.String("demo/hello.txt"),
					HTTPMethod: "GET",
					Expires:    900,
				})
			}},
		{"V2 分享", "",
			func(c *S3) (string, error) {
				return c.GenerateShareUrl(&GenerateShareUrlInput{Bucket: aws.String("b")})
			}},
		{"V4 分享", "V4",
			func(c *S3) (string, error) {
				return c.GenerateShareUrl(&GenerateShareUrlInput{Bucket: aws.String("b")})
			}},
	}
	for _, tc := range cases {
		c := newTestClient("https://s3.example.com", "AK", "SK", true, false, tc.sign)
		u, err := tc.gen(c)
		if err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		parsed, err := url.Parse(u)
		if err != nil {
			t.Errorf("%s: URL 无法解析 %q: %v", tc.name, u, err)
			continue
		}
		// 形态必须含 scheme 和 host
		if parsed.Scheme == "" || parsed.Host == "" {
			t.Errorf("%s: URL 形态异常 %q (scheme=%q host=%q)", tc.name, u, parsed.Scheme, parsed.Host)
		}
		// 预签名 path 应含 bucket
		if strings.HasPrefix(tc.name, "V") && strings.Contains(tc.name, "预签名") {
			if !strings.HasPrefix(parsed.Path, "/b") {
				t.Errorf("%s: path 异常 %q", tc.name, parsed.Path)
			}
		}
		if parsed.RawQuery == "" {
			t.Errorf("%s: 缺 query %q", tc.name, u)
		}
		t.Logf("%-10s -> %s", tc.name, u)
	}
}

// TestBuildProducesEncodedPath 验证真实 Build 产出的 path 是 Amazon 编码且 RawPath 有效，
// 覆盖特殊字符、空格、中文、以/开头 key。
func TestBuildProducesEncodedPath(t *testing.T) {
	c := newTestClient("https://s3.example.com", "AK", "SK", true, false, "")
	cases := []string{"demo/hello.txt", "a b/c!d@e", "目录/文件", "/leading/key", "100%done"}
	for _, key := range cases {
		req, _ := c.PutObjectRequest(&PutObjectInput{
			Bucket: aws.String("b"),
			Key:    aws.String(key),
			Body:   strings.NewReader("x"),
		})
		req.Build()
		u := req.HTTPRequest.URL
		if u.EscapedPath() != u.RawPath {
			t.Errorf("key=%q: EscapedPath(%q) != RawPath(%q)", key, u.EscapedPath(), u.RawPath)
		}
		if u.Opaque != "" {
			t.Errorf("key=%q: Opaque 应为空，got %q", key, u.Opaque)
		}
	}
}

// TestV2SignatureSpecialChars 对各种特殊字符 key 做端到端 V2 验签，
// 确认编码不导致签名与请求行分叉。覆盖空格、特殊符号、中文、以/开头、含%、Unicode 等。
func TestV2SignatureSpecialChars(t *testing.T) {
	var (
		gotAuth, gotURI, gotMethod, gotDate, gotCT, gotMD5 string
		gotHeaders                                         http.Header
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotURI = r.RequestURI
		gotMethod = r.Method
		gotDate = r.Header.Get("Date")
		gotCT = r.Header.Get("Content-Type")
		gotMD5 = r.Header.Get("Content-Md5")
		gotHeaders = r.Header
		w.WriteHeader(200)
		_, _ = io.Copy(io.Discard, r.Body)
		_ = r.Body.Close()
	}))
	defer srv.Close()

	c := newTestClient(srv.URL, "AKEXAMPLE", "SKEXAMPLE", true, false, "")
	cases := []struct {
		name string
		key  string
	}{
		{"普通", "demo/hello.txt"},
		{"空格", "a b/c d.txt"},
		{"特殊符号", "a!b@c#d$e%f^g&h*i(j)k.txt"},
		{"中文", "目录/文件.txt"},
		{"以/开头", "/leading/key.txt"},
		{"含百分号", "100%done.txt"},
		{"含加号", "a+b/c.txt"},
		{"Unicode", "café/naïve.txt"},
		{"混合", "a b/中文!@#/100%done.txt"},
	}
	for _, tc := range cases {
		if _, err := c.PutObject(&PutObjectInput{
			Bucket: aws.String("test-bucket"),
			Key:    aws.String(tc.key),
			Body:   strings.NewReader("content"),
		}); err != nil {
			t.Fatalf("%s: PutObject: %v", tc.name, err)
		}
		// 服务端按 SigV2 用 request-target 重算签名
		canonicalHeaders := buildCanonicalHeaders(gotHeaders)
		signItems := []string{gotMethod, gotMD5, gotCT, gotDate}
		if canonicalHeaders != "" {
			signItems = append(signItems, canonicalHeaders)
		}
		signItems = append(signItems, gotURI)
		stringToSign := strings.Join(signItems, "\n")

		mac := hmac.New(sha1.New, []byte("SKEXAMPLE"))
		mac.Write([]byte(stringToSign))
		expectedAuth := "AWS AKEXAMPLE:" + base64.StdEncoding.EncodeToString(mac.Sum(nil))

		if gotAuth != expectedAuth {
			t.Errorf("[%s] 签名不匹配 key=%q RequestURI=%q\n服务端=%q\n客户端=%q",
				tc.name, tc.key, gotURI, expectedAuth, gotAuth)
		} else {
			t.Logf("[%s] OK key=%q RequestURI=%q", tc.name, tc.key, gotURI)
		}
	}
}
