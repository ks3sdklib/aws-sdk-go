package body

import (
	"github.com/ks3sdklib/aws-sdk-go/aws"
	"github.com/ks3sdklib/aws-sdk-go/internal/protocol/rest"
	"github.com/ks3sdklib/aws-sdk-go/internal/protocol/restjson"
	"github.com/ks3sdklib/aws-sdk-go/internal/protocol/restxml"
)

// Build builds the REST component of a service request.
func Build(r *aws.Request) {
	index := IndexOf(jsonRequestApiName, r.Operation.Name)
	if index != -1 {
		restjson.Build(r)
	} else {
		restxml.Build(r)
	}
}

// UnmarshalBody unmarshal a response body for the REST protocol.
func UnmarshalBody(r *aws.Request) {
	rest.Unmarshal(r)
	index := IndexOf(jsonResponseApiName, r.Operation.Name)
	if index != -1 {
		restjson.Unmarshal(r)
	} else {
		restxml.Unmarshal(r)
	}
}

// UnmarshalMeta unmarshal response headers for the REST protocol.
func UnmarshalMeta(r *aws.Request) {
	rest.UnmarshalMeta(r)
}

// UnmarshalError unmarshal a response error for the REST protocol.
func UnmarshalError(r *aws.Request) {
	restxml.UnmarshalError(r)
}

var jsonRequestApiName = []string{
	"PutBucketMirror",
}

var jsonResponseApiName = []string{
	"GetBucketMirror",
}

func IndexOf(apiNames []string, apiName string) int {
	for index, value := range apiNames {
		if value == apiName {
			return index
		}
	}
	return -1
}
