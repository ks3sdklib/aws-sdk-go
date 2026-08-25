package v2

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"github.com/ks3sdklib/aws-sdk-go/aws/awsutil"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ks3sdklib/aws-sdk-go/aws/credentials"
	"github.com/ks3sdklib/aws-sdk-go/internal/protocol/rest"

	"github.com/ks3sdklib/aws-sdk-go/aws"
)

const (
	authHeaderPrefix = "AWS"
	timeFormat       = "Mon, 02 Jan 2006 15:04:05 GMT"
)

var signQuerys = map[string]bool{
	"acl":                          true,
	"lifecycle":                    true,
	"location":                     true,
	"logging":                      true,
	"notification":                 true,
	"policy":                       true,
	"requestPayment":               true,
	"torrent":                      true,
	"uploadId":                     true,
	"uploads":                      true,
	"versionId":                    true,
	"versioning":                   true,
	"versions":                     true,
	"website":                      true,
	"delete":                       true,
	"thumbnail":                    true,
	"cors":                         true,
	"pfop":                         true,
	"querypfop":                    true,
	"partNumber":                   true,
	"response-content-type":        true,
	"response-content-language":    true,
	"response-expires":             true,
	"response-cache-control":       true,
	"response-content-disposition": true,
	"response-content-encoding":    true,
	"tagging":                      true,
	"fetch":                        true,
	"copy":                         true,
	"mirror":                       true,
	"restore":                      true,
	"append":                       true,
	"position":                     true,
	"decompresspolicy":             true,
	"retention":                    true,
	"crr":                          true,
	"inventory":                    true,
	"recycle":                      true,
	"recover":                      true,
	"clear":                        true,
	"id":                           true,
	"dataRedundancySwitch":         true,
	"bucketqos":                    true,
	"requesterqos":                 true,
	"encryption":                   true,
	"accessmonitor":                true,
	"transferAcceleration":         true,
	"VpcAccessBlock":               true,
	"quota":                        true,
	"migration":                    true,
	"jobs":                         true,
	"jobId":                        true,
	"X-Amz-Policy":                 true,
	"action":                       true,
	"priority":                     true,
	"PublicNetworkBlock":           true,
	"BucketPublicNetworkBlock":     true,
	"dataRedundancyTransition":     true,
	"worm":                         true,
	"wormId":                       true,
	"wormExtend":                   true,
	"dataAccelerator":              true,
	"archiveDirectRead":            true,
	"http2":                        true,
}

type signer struct {
	Service     *aws.Service
	Request     *http.Request
	Time        time.Time
	ExpireTime  int64
	ServiceName string
	Region      string
	CredValues  credentials.Value
	Credentials *credentials.Credentials
	Query       url.Values
	Body        io.ReadSeeker
	Debug       uint
	Logger      io.Writer

	isPresign     bool
	formattedTime string

	canonicalHeaders  string
	canonicalResource string
	stringToSign      string
	signature         string
	authorization     string
	awsRequest        *aws.Request
}

func Sign(req *aws.Request) {
	if req.Service.Config.Credentials == credentials.AnonymousCredentials {
		return
	}

	region := req.Service.SigningRegion
	if region == "" {
		region = req.Service.Config.Region
	}

	name := req.Service.SigningName
	if name == "" {
		name = req.Service.ServiceName
	}

	s := signer{
		Service:     req.Service,
		Request:     req.HTTPRequest,
		Time:        req.Time,
		ExpireTime:  req.ExpireTime,
		Query:       req.HTTPRequest.URL.Query(),
		Body:        req.Body,
		ServiceName: name,
		Region:      region,
		Credentials: req.Service.Config.Credentials,
		Debug:       req.Service.Config.LogLevel,
		Logger:      req.Service.Config.Logger,
		awsRequest:  req,
	}

	req.Error = s.sign()
}

func (v2 *signer) sign() error {
	if v2.ExpireTime != 0 {
		v2.isPresign = true
	}

	if v2.isRequestSigned() {
		if !v2.Credentials.IsExpired() {
			// If the request is already signed, and the credentials have not
			// expired yet ignore the signing request.
			return nil
		}

		// The credentials have expired for this request. The current signing
		// is invalid, and needs to be request because the request will fail.
		if v2.isPresign {
			v2.removePresign()
			// Update the request's query string to ensure the values stays in
			// sync in the case retrieving the new credentials fails.
			v2.Request.URL.RawQuery = v2.Query.Encode()
		}
	}

	var err error
	v2.CredValues, err = v2.Credentials.Get()
	if err != nil {
		return err
	}

	v2.build()

	v2.logSigningInfo()

	return nil
}

func (v2 *signer) logSigningInfo() {
	cfg := v2.Service.Config
	cfg.LogDebug("%s", "---[ STRING TO SIGN ]--------------------------------")
	cfg.LogDebug("%s", v2.stringToSign)
	if v2.isPresign {
		cfg.LogDebug("---[ SIGNED URL ]--------------------------------")
		cfg.LogDebug("%s", v2.Request.URL)
	}
	cfg.LogDebug("-----------------------------------------------------")
}

func (v2 *signer) build() {
	v2.buildTime()              // no depends
	v2.buildCanonicalHeaders()  // depends on cred string
	v2.buildCanonicalResource() // depends on canon headers / signed headers
	v2.buildStringToSign()      // depends on canon string
	v2.buildSignature()         // depends on string to sign

	if v2.isPresign {
		var querys url.Values
		querys = make(map[string][]string)
		querys.Add("Signature", v2.signature)
		querys.Add("AWSAccessKeyId", v2.CredValues.AccessKeyID)
		v2.Request.URL.RawQuery += "&" + querys.Encode()
	} else {
		v2.Request.Header.Set("Authorization", "AWS "+v2.CredValues.AccessKeyID+":"+v2.signature)
	}
}

func (v2 *signer) buildTime() {
	v2.formattedTime = v2.Time.UTC().Format(timeFormat)

	if v2.isPresign {
		duration := int64(v2.ExpireTime)
		v2.Query.Set("Expires", strconv.FormatInt(duration, 10))
	} else {
		v2.Request.Header.Set("Date", v2.formattedTime)
	}
}

func (v2 *signer) buildCanonicalHeaders() {
	var headers []string
	for k, v := range v2.Request.Header {
		if strings.HasPrefix(k, "X-Amz-") {
			key := strings.ToLower(http.CanonicalHeaderKey(k))
			v2.Request.Header[key] = v
			v2.Request.Header.Del(http.CanonicalHeaderKey(k))
			headers = append(headers, key)
		}
	}
	sort.Strings(headers)

	headerValues := make([]string, len(headers))
	for i, k := range headers {
		key := strings.ToLower(http.CanonicalHeaderKey(k))
		headerValues[i] = key + ":" + strings.Join(v2.Request.Header[key], ",")
	}

	v2.canonicalHeaders = strings.Join(headerValues, "\n")
}

func (v2 *signer) buildCanonicalResource() {
	v2.Request.URL.RawQuery = strings.Replace(v2.Query.Encode(), "+", "%20", -1)

	// 从 EscapedPath() 取待签名 path。updatePath 已改设 Path/RawPath（不再设 Opaque），
	// 读 EscapedPath() 得到 Amazon 编码的 path，与请求行发送字节一致，重签时输入稳定。
	uri := v2.Request.URL.EscapedPath()
	pathStyle := v2.awsRequest.Config.S3ForcePathStyle

	bucketInHost := ""
	if !pathStyle && !v2.awsRequest.Config.DomainMode {
		bucketInHost = v2.bucketInHost()
	}

	// path 中只有桶没有 key 时，补尾斜杠成 /bucket/
	appendSlash := false
	if bucketInHost == "" && !v2.awsRequest.Config.DomainMode {
		seg := strings.Split(strings.Trim(uri, "/"), "/")
		if len(seg) == 1 && seg[0] != "" {
			appendSlash = true
		}
	}

	// uri 来自 EscapedPath，virtual-host/DomainMode 下不含桶，需拼回桶得到签名用的 /bucket/key：
	// virtual-host 从 Host 拆桶，DomainMode 从入参取桶，path-style 桶已在 uri 中。
	if bucketInHost != "" {
		if uri == "/" {
			uri = "/" + bucketInHost + "/"
		} else {
			uri = "/" + bucketInHost + uri
		}
	} else if v2.awsRequest.Config.DomainMode {
		if b := awsutil.ValuesAtPath(v2.awsRequest.Params, "Bucket"); len(b) > 0 {
			bucket := b[0].(string)
			if uri == "/" {
				uri = "/" + bucket + "/"
			} else {
				uri = "/" + bucket + uri
			}
		}
	}

	if appendSlash && !strings.HasSuffix(uri, "/") {
		uri += "/"
	}
	if uri == "" {
		uri = "/"
	}

	if v2.ServiceName != "s3" {
		uri = rest.EscapePath(uri, false)
	}

	var querys []string
	for k := range v2.Query {
		if _, ok := signQuerys[k]; ok {
			querys = append(querys, k)
		}
	}
	sort.Strings(querys)

	queryValues := make([]string, len(querys))
	for i, k := range querys {
		v := v2.Query[k]
		vString := strings.Join(v, ",")
		if vString != "" {
			queryValues[i] = k + "=" + vString
		} else {
			queryValues[i] = k
		}
	}
	queryString := strings.Join(queryValues, "&")
	if queryString == "" {
		v2.canonicalResource = uri
	} else {
		v2.canonicalResource = uri + "?" + queryString
	}
}

// bucketInHost 从 URL.Host 拆出 virtual-host 风格的桶名（bucket.endpoint）。
func (v2 *signer) bucketInHost() string {
	host := v2.Request.URL.Host
	endpoint := v2.Service.Endpoint
	endpointHost := strings.TrimPrefix(endpoint, "http://")
	endpointHost = strings.TrimPrefix(endpointHost, "https://")
	if host == endpointHost {
		return ""
	}
	if idx := strings.Index(host, "."+endpointHost); idx > 0 {
		return host[:idx]
	}
	return ""
}

func (v2 *signer) buildStringToSign() {
	md5list := v2.Request.Header["Content-Md5"]
	md5 := ""
	if len(md5list) > 0 {
		md5 = v2.Request.Header["Content-Md5"][0]
	}

	typelist := v2.Request.Header["Content-Type"]
	contenttype := ""
	if len(typelist) > 0 {
		contenttype = v2.Request.Header["Content-Type"][0]
	}
	signItems := []string{v2.Request.Method, md5, contenttype}
	if v2.isPresign {
		signItems = append(signItems, v2.Query["Expires"][0])
	} else {
		signItems = append(signItems, v2.formattedTime)
	}
	if v2.canonicalHeaders != "" {
		signItems = append(signItems, v2.canonicalHeaders)
	}
	signItems = append(signItems, v2.canonicalResource)

	v2.stringToSign = strings.Join(signItems, "\n")
	if v2.awsRequest.SignType == "policy" {
		v2.stringToSign = strings.Join([]string{v2.Query["Expires"][0], v2.canonicalResource}, "\n")
	}
}

func (v2 *signer) buildSignature() {
	secret := v2.CredValues.SecretAccessKey
	signature := string(base64Encode(makeHmac([]byte(secret), []byte(v2.stringToSign))))
	v2.signature = signature
}

// isRequestSigned returns if the request is currently signed or presigned
func (v2 *signer) isRequestSigned() bool {
	if v2.isPresign && v2.Query.Get("Signature") != "" {
		return true
	}
	if v2.Request.Header.Get("Authorization") != "" {
		return true
	}

	return false
}

// unsign removes signing flags for both signed and presigned requests.
func (v2 *signer) removePresign() {
	v2.Query.Del("AWSAccessKeyId")
	v2.Query.Del("Signature")
	v2.Query.Del("Expires")

}

func makeHmac(key []byte, data []byte) []byte {
	hash := hmac.New(sha1.New, key)
	hash.Write(data)
	return hash.Sum(nil)
}
func base64Encode(src []byte) []byte {
	return []byte(base64.StdEncoding.EncodeToString(src))
}
