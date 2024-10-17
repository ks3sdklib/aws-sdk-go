package query

import (
	"encoding/xml"
	"github.com/ks3sdklib/aws-sdk-go/aws"
	"github.com/ks3sdklib/aws-sdk-go/internal/apierr"
	"io"
	"regexp"
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
	if r.HTTPResponse.Header.Get("Content-Type") == "text/html" {
		// 获取HTML文本中title标签的内容
		re := regexp.MustCompile(`<title>(.*?)</title>`)
		matches := re.FindStringSubmatch(string(body))

		title := ""
		if len(matches) > 1 {
			title = matches[1]
		}

		r.Error = apierr.NewRequestError(apierr.New(title, "", nil), r.HTTPResponse.StatusCode, "")
		return
	}

	err = xml.Unmarshal(body, &resp)
	resp.StatusCode = r.HTTPResponse.StatusCode

	// head请求无法从body中获取request id，如果是head请求，则从header中获取
	if resp.RequestID == "" && r.HTTPRequest.Method == "HEAD" {
		resp.RequestID = r.HTTPResponse.Header.Get("X-Kss-Request-Id")
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
