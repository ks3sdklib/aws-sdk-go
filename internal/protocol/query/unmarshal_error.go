package query

import (
	"bytes"
	"encoding/xml"
	"golang.org/x/net/html"
	"io"
	"strings"

	"github.com/ks3sdklib/aws-sdk-go/aws"
	"github.com/ks3sdklib/aws-sdk-go/internal/apierr"
)

type XmlErrorResponse struct {
	XMLName    xml.Name `xml:"Error"`
	Code       string   `xml:"Code"`
	StatusCode int      `xml:"StatusCode"`
	Message    string   `xml:"Message"`
	Resource   string   `xml:"Resource"`
	RequestID  string   `xml:"RequestId"`
}

// UnmarshalError unmarshal an error response for an AWS Query service.
func UnmarshalError(r *aws.Request) {
	defer r.HTTPResponse.Body.Close()

	resp := &XmlErrorResponse{}
	body, err := io.ReadAll(r.HTTPResponse.Body)
	if err != nil {
		r.Error = apierr.New("Unmarshal", "failed to read body", err)
		return
	}

	// 如果响应类型是html，则解析html文本
	if strings.Contains(r.HTTPResponse.Header.Get("Content-Type"), "html") {
		// 解析HTML文本
		doc, err := html.Parse(bytes.NewReader(body))
		if err != nil {
			r.Error = apierr.New("Unmarshal", "failed to parse html", err)
			return
		}
		title := findTitle(doc)
		r.Error = apierr.NewRequestError(
			apierr.New(title, "", nil),
			r.HTTPResponse.StatusCode,
			"",
		)
		return
	}

	err = xml.Unmarshal(body, &resp)
	resp.StatusCode = r.HTTPResponse.StatusCode

	// head请求无法从body中获取request id，如果是head请求，则从header中获取
	if resp.RequestID == "" && r.HTTPRequest.Method == "HEAD" {
		resp.RequestID = r.HTTPResponse.Header.Get("X-KSS-Request-Id")
	}

	if err != nil && err != io.EOF {
		r.Error = apierr.New("Unmarshal", "failed to decode query XML error response", err)
	} else {
		r.Error = apierr.NewRequestError(
			apierr.New(resp.Code, resp.Message, nil),
			r.HTTPResponse.StatusCode,
			resp.RequestID,
		)
	}
}

// findTitle 提取HTML文档中<title>标签的内容
func findTitle(doc *html.Node) string {
	var title string

	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "title" && n.FirstChild != nil {
			title = n.FirstChild.Data
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}

	traverse(doc)
	return title
}
