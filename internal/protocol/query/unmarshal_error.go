package query

import (
	"encoding/xml"
	"io"
	"io/ioutil"
	"log"

	"github.com/ks3sdklib/aws-sdk-go/aws"
	"github.com/ks3sdklib/aws-sdk-go/internal/apierr"
)

type XmlErrorResponse struct {
	XMLName    xml.Name `xml:"Error"`
	Code       string   `xml:"Code"`
	StatusCode int      `"StatusCode"`
	Message    string   `xml:"Message"`
	Resource   string   `xml:"Resource"`
	RequestID  string   `xml:"RequestId"`
}

// UnmarshalError unmarshals an error response for an AWS Query service.
func UnmarshalError(r *aws.Request) {
	defer r.HTTPResponse.Body.Close()

	resp := &XmlErrorResponse{}
	body, err := ioutil.ReadAll(r.HTTPResponse.Body)
	if err != nil {
		log.Printf("read body err, %v\n", err)
		return
	}
	err = xml.Unmarshal(body, &resp)
	resp.StatusCode = r.HTTPResponse.StatusCode

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
