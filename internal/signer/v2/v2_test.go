package v2

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/ks3sdklib/aws-sdk-go/aws"
	"github.com/ks3sdklib/aws-sdk-go/aws/awsutil"
	"github.com/ks3sdklib/aws-sdk-go/aws/credentials"
	"github.com/ks3sdklib/aws-sdk-go/internal/protocol/rest"
)

// newTestSigner 构造一个可运行 buildCanonicalResource 的最小 signer。
func newTestSigner(endpoint, host, urlPath string, pathStyle, domainMode bool, params interface{}) *signer {
	cfg := &aws.Config{
		Endpoint:         endpoint,
		S3ForcePathStyle: pathStyle,
		DomainMode:       domainMode,
	}
	cfg = aws.DefaultConfig.Merge(cfg)
	cfg.Credentials = credentials.NewStaticCredentials("AK", "SK", "")
	svc := &aws.Service{Config: cfg, Endpoint: endpoint, ServiceName: "s3"}
	req := &aws.Request{Service: svc, Params: params}

	decoded, _ := url.PathUnescape(urlPath)
	u := &url.URL{Scheme: "https", Host: host, Path: decoded, RawPath: urlPath}
	httpReq := &http.Request{URL: u}

	return &signer{
		Service:     svc,
		Request:     httpReq,
		Query:       url.Values{},
		ServiceName: "s3",
		awsRequest:  req,
	}
}

type bucketParams struct {
	Bucket *string
}

func strPtr(s string) *string { return &s }

// TestBuildCanonicalResource 覆盖三种寻址方式下 object 与 bucket-only 的 canonicalResource。
func TestBuildCanonicalResource(t *testing.T) {
	cases := []struct {
		name      string
		endpoint  string
		host      string
		urlPath   string
		pathStyle bool
		domain    bool
		params    interface{}
		want      string
	}{
		{"path-style object", "https://s3.example.com", "s3.example.com",
			"/bucket/key", true, false, &bucketParams{strPtr("bucket")}, "/bucket/key"},
		{"path-style bucket-only", "https://s3.example.com", "s3.example.com",
			"/bucket", true, false, &bucketParams{strPtr("bucket")}, "/bucket/"},
		{"vhost object", "https://s3.example.com", "bucket.s3.example.com",
			"/key", false, false, &bucketParams{strPtr("bucket")}, "/bucket/key"},
		{"vhost bucket-only", "https://s3.example.com", "bucket.s3.example.com",
			"/", false, false, &bucketParams{strPtr("bucket")}, "/bucket/"},
		{"domain object", "https://s3.example.com", "s3.example.com",
			"/key", false, true, &bucketParams{strPtr("bucket")}, "/bucket/key"},
		{"domain bucket-only", "https://s3.example.com", "s3.example.com",
			"/", false, true, &bucketParams{strPtr("bucket")}, "/bucket/"},
		{"无 bucket", "https://s3.example.com", "s3.example.com",
			"/", true, false, &bucketParams{strPtr("")}, "/"},
	}
	for _, tc := range cases {
		s := newTestSigner(tc.endpoint, tc.host, tc.urlPath, tc.pathStyle, tc.domain, tc.params)
		s.buildCanonicalResource()
		if s.canonicalResource != tc.want {
			t.Errorf("%s: got %q, want %q", tc.name, s.canonicalResource, tc.want)
		} else {
			t.Logf("%-26s -> %q", tc.name, s.canonicalResource)
		}
	}
}

// TestBuildCanonicalResourceWithSubresource 验证 signQuerys 白名单参数参与签名。
func TestBuildCanonicalResourceWithSubresource(t *testing.T) {
	s := newTestSigner("https://s3.example.com", "s3.example.com", "/bucket", true, false, &bucketParams{strPtr("bucket")})
	s.Query.Set("acl", "")
	s.buildCanonicalResource()
	if s.canonicalResource != "/bucket/?acl" {
		t.Errorf("got %q, want /bucket/?acl", s.canonicalResource)
	}
}

// oldCanonical 为 buildCanonicalResource 的 Opaque 版本实现，用于等价对比。
func oldCanonical(s *signer) string {
	endpoint := s.Service.Endpoint
	s.Request.URL.RawQuery = strings.Replace(s.Query.Encode(), "+", "%20", -1)
	urlStr := s.Request.URL.String()
	pathStyle := strings.HasPrefix(urlStr, endpoint)
	uri := s.Request.URL.Opaque
	bucketInHost := ""
	if !pathStyle {
		if strings.HasPrefix(urlStr, "http://") {
			urlStr = urlStr[7:]
			endpoint = endpoint[7:]
		} else if strings.HasPrefix(urlStr, "https://") {
			urlStr = urlStr[8:]
			endpoint = endpoint[8:]
		}
		if idx := strings.Index(urlStr, endpoint); idx > 0 {
			bucketInHost = urlStr[0 : idx-1]
		}
	}
	if uri != "" {
		uris := strings.Split(uri, "/")[3:]
		appendSlash := false
		if len(uris) == 1 && uris[0] != "" && bucketInHost == "" {
			appendSlash = true
		} else if len(uris) == 0 && bucketInHost != "" {
			appendSlash = true
		}
		uri = "/" + strings.Join(strings.Split(uri, "/")[3:], "/")
		if bucketInHost != "" {
			uri = "/" + bucketInHost + uri
		}
		if s.awsRequest.Config.DomainMode {
			b := awsutil.ValuesAtPath(s.awsRequest.Params, "Bucket")
			bucket := b[0].(string)
			uri = "/" + bucket + uri
			appendSlash = false
		}
		if appendSlash {
			uri += "/"
		}
	} else {
		uri = s.Request.URL.Path
	}
	if uri == "" {
		uri = "/"
	}
	if s.ServiceName != "s3" {
		uri = rest.EscapePath(uri, false)
	}
	return uri
}

// TestCanonicalResourceEquivalence 对比两种实现在各种寻址下是否算出相同 canonicalResource。
func TestCanonicalResourceEquivalence(t *testing.T) {
	cases := []struct {
		name      string
		endpoint  string
		host      string
		urlPath   string // Amazon 编码 path（如 /bucket/key）
		pathStyle bool   // S3ForcePathStyle
		domain    bool
		bucket    string
	}{
		// 标准三寻址
		{"path-style object", "https://s3.example.com", "s3.example.com", "/bucket/key", true, false, "bucket"},
		{"path-style bucket-only", "https://s3.example.com", "s3.example.com", "/bucket", true, false, "bucket"},
		{"vhost object", "https://s3.example.com", "bucket.s3.example.com", "/key", false, false, "bucket"},
		{"vhost bucket-only", "https://s3.example.com", "bucket.s3.example.com", "/", false, false, "bucket"},
		{"domain object", "https://s3.example.com", "s3.example.com", "/key", false, true, "bucket"},
		{"domain bucket-only", "https://s3.example.com", "s3.example.com", "/", false, true, "bucket"},
		// bucket 非 DNS 兼容且未强制 path-style
		{"非DNS兼容 object", "https://s3.example.com", "s3.example.com", "/my_bucket/key", false, false, "my_bucket"},
		{"非DNS兼容 bucket-only", "https://s3.example.com", "s3.example.com", "/my_bucket", false, false, "my_bucket"},
		// endpoint 带端口
		{"带端口 path-style object", "https://s3.example.com:8443", "s3.example.com:8443", "/bucket/key", true, false, "bucket"},
		{"带端口 vhost object", "https://s3.example.com:8443", "bucket.s3.example.com:8443", "/key", false, false, "bucket"},
		// 以 / 开头的 key（V2 保留 %2F）
		{"path-style %2F key", "https://s3.example.com", "s3.example.com", "/bucket/%2Fleading/key", true, false, "bucket"},
		{"vhost %2F key", "https://s3.example.com", "bucket.s3.example.com", "/%2Fkey", false, false, "bucket"},
		// 特殊字符 key
		{"path-style 特殊字符", "https://s3.example.com", "s3.example.com", "/bucket/a%20b/c%21d", true, false, "bucket"},
		// 中文 key
		{"path-style 中文", "https://s3.example.com", "s3.example.com", "/bucket/%E7%9B%AE%E5%BD%95/%E6%96%87%E4%BB%B6", true, false, "bucket"},
	}
	for _, tc := range cases {
		params := &bucketParams{strPtr(tc.bucket)}

		sNew := newTestSigner(tc.endpoint, tc.host, tc.urlPath, tc.pathStyle, tc.domain, params)
		sNew.buildCanonicalResource()
		gotNew := sNew.canonicalResource

		sOld := newTestSigner(tc.endpoint, tc.host, tc.urlPath, tc.pathStyle, tc.domain, params)
		sOld.Request.URL.Opaque = "//" + tc.host + tc.urlPath
		sOld.Request.URL.Path = ""
		sOld.Request.URL.RawPath = ""
		gotOld := oldCanonical(sOld)

		if gotNew != gotOld {
			t.Errorf("%s: EscapedPath版=%q Opaque版=%q (不一致)", tc.name, gotNew, gotOld)
		} else {
			t.Logf("%-30s -> %q", tc.name, gotNew)
		}
	}
}
