package rest

import (
	"net/url"
	"strings"
	"testing"

	"github.com/ks3sdklib/aws-sdk-go/aws"
)

// oldUpdatePath 为 updatePath 的 Opaque 版本实现。
func oldUpdatePath(u *url.URL, cfg *aws.Config) {
	urlPath := u.Path
	scheme, query := u.Scheme, u.RawQuery
	urlPath = strings.Replace(urlPath, "//", "/%2F", -1)
	if !cfg.DisableRestProtocolURICleaning {
		urlPath = cleanPath(urlPath)
	}
	u.Scheme, u.Path, u.RawQuery = "", "", ""
	s := u.String()
	u.Scheme = scheme
	u.RawQuery = query
	u.Opaque = s + urlPath
}

func parseTestURL(path string) *url.URL {
	u, _ := url.Parse("https://s3.example.com")
	u.Path = path
	return u
}

// signPathFromOpaque 从 Opaque 拆 [3:] 取 path（Opaque 版本签名器的取法）。
func signPathFromOpaque(u *url.URL) string {
	uri := "/" + strings.Join(strings.Split(u.Opaque, "/")[3:], "/")
	if uri == "" {
		uri = "/"
	}
	return uri
}

// signPathFromEscaped 从 EscapedPath 取 path（EscapedPath 版本签名器的取法）。
func signPathFromEscaped(u *url.URL) string {
	uri := u.EscapedPath()
	if uri == "" {
		uri = "/"
	}
	return uri
}

// TestUpdatePathEquivalence 对比两种 updatePath 实现产出的签名 path 与发送 path 是否字节一致。
func TestUpdatePathEquivalence(t *testing.T) {
	cases := []string{
		"/bucket/key",
		"/bucket/demo/hello.txt",
		"/bucket//leading-key",                          // 触发 // -> /%2F
		"/bucket/a%20b/c%21d%40e",                       // 特殊字符
		"/bucket/%E7%9B%AE%E5%BD%95/%E6%96%87%E4%BB%B6", // 中文
		"/bucket",            // bucket-only
		"/",                  // 根
		"/bucket/100%25done", // 含 %
	}
	for _, in := range cases {
		cfg := &aws.Config{}

		uOld := parseTestURL(in)
		oldUpdatePath(uOld, cfg)
		oldSignPath := signPathFromOpaque(uOld)

		uNew := parseTestURL(in)
		updatePath(uNew, cfg)
		newSignPath := signPathFromEscaped(uNew)
		newWirePath := uNew.RequestURI()

		t.Logf("in=%-44q | Opaque版=%q EscapedPath签名=%q EscapedPath发送=%q", in, oldSignPath, newSignPath, newWirePath)
		if oldSignPath != newSignPath {
			t.Errorf("签名 path 不一致: in=%q Opaque版=%q EscapedPath版=%q", in, oldSignPath, newSignPath)
		}
		if newSignPath != newWirePath {
			t.Errorf("签名与发送 path 不一致: in=%q 签名=%q 发送=%q", in, newSignPath, newWirePath)
		}
	}
}
