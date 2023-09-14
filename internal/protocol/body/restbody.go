package body

import (
	"encoding/json"
	"github.com/ks3sdklib/aws-sdk-go/aws"
	"github.com/ks3sdklib/aws-sdk-go/internal/apierr"
	"github.com/ks3sdklib/aws-sdk-go/internal/protocol/rest"
	"github.com/ks3sdklib/aws-sdk-go/internal/protocol/restjson"
	"github.com/ks3sdklib/aws-sdk-go/internal/protocol/restxml"
	"reflect"
)

func Build(r *aws.Request) {
	if r.Operation.Name == "PutBucketMirror" {
		rest.Build(r)
		bucketMirror := reflect.ValueOf(r.Params).Elem().FieldByName("BucketMirror").Interface()
		data, err := json.Marshal(bucketMirror)
		if err != nil {
			r.Error = apierr.New("Marshal", "failed to enode rest JSON request", err)
			return
		}
		r.SetBufferBody(data)
	} else {
		restxml.Build(r)
	}
}

func Unmarshal(r *aws.Request) {
	if r.Operation.Name == "GetBucketMirror" {
		restjson.Unmarshal(r)
	} else {
		restxml.Unmarshal(r)
	}
}

func UnmarshalMeta(r *aws.Request) {
	rest.Unmarshal(r)
}

func UnmarshalError(r *aws.Request) {
	restxml.UnmarshalError(r)
}
