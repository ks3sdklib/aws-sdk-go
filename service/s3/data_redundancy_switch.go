package s3

import (
	"github.com/ks3sdklib/aws-sdk-go/aws"
	"time"
)

var opPutBucketDataRedundancySwitch *aws.Operation

type PutBucketDataRedundancySwitchInput struct {
	// The name of the bucket.
	Bucket *string `location:"uri" locationName:"Bucket" type:"string" required:"true"`
	// The bucket data redundancy type.
	// Valid value: LRS丨ZRS
	// LRS: local redundancy storage
	// ZRS: zone redundancy storage
	DataRedundancyType *string `location:"header" locationName:"x-amz-data-redundancy-type" type:"string" required:"true"`
}

type PutBucketDataRedundancySwitchOutput struct {
	// The HTTP headers of the response.
	Metadata map[string]*string `location:"headers" type:"map"`
	// The HTTP status code of the response.
	StatusCode *int64 `location:"statusCode" type:"integer"`
}

// PutBucketDataRedundancySwitchRequest generates a request for the PutBucketDataRedundancySwitch operation.
func (c *S3) PutBucketDataRedundancySwitchRequest(input *PutBucketDataRedundancySwitchInput) (req *aws.Request, output *PutBucketDataRedundancySwitchOutput) {
	oprw.Lock()
	defer oprw.Unlock()
	if opPutBucketDataRedundancySwitch == nil {
		opPutBucketDataRedundancySwitch = &aws.Operation{
			Name:       "PutBucketDataRedundancySwitch",
			HTTPMethod: "PUT",
			HTTPPath:   "/{Bucket}?dataRedundancySwitch",
		}
	}
	if input == nil {
		input = &PutBucketDataRedundancySwitchInput{}
	}
	req = c.newRequest(opPutBucketDataRedundancySwitch, input, output)
	output = &PutBucketDataRedundancySwitchOutput{}
	req.Data = output
	return
}

// PutBucketDataRedundancySwitch sets the data redundancy type for the bucket.
func (c *S3) PutBucketDataRedundancySwitch(input *PutBucketDataRedundancySwitchInput) (*PutBucketDataRedundancySwitchOutput, error) {
	req, out := c.PutBucketDataRedundancySwitchRequest(input)
	err := req.Send()
	return out, err
}

func (c *S3) PutBucketDataRedundancySwitchWithContext(ctx aws.Context, input *PutBucketDataRedundancySwitchInput) (*PutBucketDataRedundancySwitchOutput, error) {
	req, out := c.PutBucketDataRedundancySwitchRequest(input)
	req.SetContext(ctx)
	err := req.Send()
	return out, err
}

var opGetBucketDataRedundancySwitch *aws.Operation

type GetBucketDataRedundancySwitchInput struct {
	// The name of the bucket.
	Bucket *string `location:"uri" locationName:"Bucket" type:"string" required:"true"`
}

type GetBucketDataRedundancySwitchOutput struct {
	// The bucket data redundancy switch configuration.
	DataRedundancySwitch *DataRedundancySwitch `locationName:"DataRedundancySwitch" type:"structure"`
	// The HTTP headers of the response.
	Metadata map[string]*string `location:"headers" type:"map"`
	// The HTTP status code of the response.
	StatusCode *int64 `location:"statusCode" type:"integer"`

	metadataGetGetBucketDataRedundancySwitchOutput `json:"-" xml:"-"`
}

type DataRedundancySwitch struct {
	// The bucket data redundancy type.
	DataRedundancyType *string `locationName:"DataRedundancyType" type:"string"`
	// Time when zone redundancy is enabled.
	SwitchTime *time.Time `locationName:"SwitchTime" type:"timestamp" timestampFormat:"iso8601"`
}

type metadataGetGetBucketDataRedundancySwitchOutput struct {
	SDKShapeTraits bool `type:"structure" payload:"DataRedundancySwitch"`
}

// GetBucketDataRedundancySwitchRequest generates a request for the GetBucketDataRedundancySwitch operation.
func (c *S3) GetBucketDataRedundancySwitchRequest(input *GetBucketDataRedundancySwitchInput) (req *aws.Request, output *GetBucketDataRedundancySwitchOutput) {
	oprw.Lock()
	defer oprw.Unlock()
	if opGetBucketDataRedundancySwitch == nil {
		opGetBucketDataRedundancySwitch = &aws.Operation{
			Name:       "GetBucketDataRedundancySwitch",
			HTTPMethod: "GET",
			HTTPPath:   "/{Bucket}?dataRedundancySwitch",
		}
	}
	if input == nil {
		input = &GetBucketDataRedundancySwitchInput{}
	}
	req = c.newRequest(opGetBucketDataRedundancySwitch, input, output)
	output = &GetBucketDataRedundancySwitchOutput{}
	req.Data = output
	return
}

// GetBucketDataRedundancySwitch gets the data redundancy switch configuration for the bucket.
func (c *S3) GetBucketDataRedundancySwitch(input *GetBucketDataRedundancySwitchInput) (*GetBucketDataRedundancySwitchOutput, error) {
	req, out := c.GetBucketDataRedundancySwitchRequest(input)
	err := req.Send()
	return out, err
}

func (c *S3) GetBucketDataRedundancySwitchWithContext(ctx aws.Context, input *GetBucketDataRedundancySwitchInput) (*GetBucketDataRedundancySwitchOutput, error) {
	req, out := c.GetBucketDataRedundancySwitchRequest(input)
	req.SetContext(ctx)
	err := req.Send()
	return out, err
}
