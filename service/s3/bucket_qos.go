package s3

import (
	"github.com/ks3sdklib/aws-sdk-go/aws"
)

var opPutBucketQos *aws.Operation

type PutBucketQosInput struct {
	// The name of the bucket.
	Bucket *string `location:"uri" locationName:"Bucket" type:"string" required:"true"`

	// Bucket flow control configuration container.
	BucketQosConfiguration *BucketQosConfiguration `locationName:"BucketQosConfiguration" type:"structure" required:"true"`

	metadataPutBucketQosInput `json:"-" xml:"-"`
}

type metadataPutBucketQosInput struct {
	SDKShapeTraits bool `type:"structure" payload:"BucketQosConfiguration"`
}

type BucketQosConfiguration struct {
	// Set the bucket flow control quota.
	Quotas []*BucketQosQuota `locationName:"Quota" type:"list" flattened:"true" required:"true"`
}

type BucketQosQuota struct {
	// Specify the storage medium type that needs flow control. Options: Extreme/Normal (default)
	// Extreme: SSD type storage medium
	// Normal (default): HDD type storage medium
	StorageMedium *string `locationName:"StorageMedium" type:"string"`
	// External network upload bandwidth, in Gbps, the value must be a positive integer.
	ExtranetUploadBandwidth *int64 `locationName:"ExtranetUploadBandwidth" type:"integer"`
	// Intranet network upload bandwidth, in Gbps, the value must be a positive integer.
	IntranetUploadBandwidth *int64 `locationName:"IntranetUploadBandwidth" type:"integer"`
	// External network download bandwidth, in Gbps, the value must be a positive integer.
	ExtranetDownloadBandwidth *int64 `locationName:"ExtranetDownloadBandwidth" type:"integer"`
	// Intranet network download bandwidth, in Gbps, the value must be a positive integer.
	IntranetDownloadBandwidth *int64 `locationName:"IntranetDownloadBandwidth" type:"integer"`
}

type PutBucketQosOutput struct {
	// The HTTP headers of the response.
	Metadata map[string]*string `location:"headers" type:"map"`
	// The HTTP status code of the response.
	StatusCode *int64 `location:"statusCode" type:"integer"`
}

// PutBucketQosRequest generates a request for the PutBucketQos operation.
func (c *S3) PutBucketQosRequest(input *PutBucketQosInput) (req *aws.Request, output *PutBucketQosOutput) {
	oprw.Lock()
	defer oprw.Unlock()
	if opPutBucketQos == nil {
		opPutBucketQos = &aws.Operation{
			Name:       "PutBucketQos",
			HTTPMethod: "PUT",
			HTTPPath:   "/{Bucket}?bucketqos",
		}
	}
	if input == nil {
		input = &PutBucketQosInput{}
	}
	req = c.newRequest(opPutBucketQos, input, output)
	output = &PutBucketQosOutput{}
	req.Data = output
	return
}

// PutBucketQos sets bucket flow control configuration.
func (c *S3) PutBucketQos(input *PutBucketQosInput) (*PutBucketQosOutput, error) {
	req, out := c.PutBucketQosRequest(input)
	err := req.Send()
	return out, err
}

func (c *S3) PutBucketQosWithContext(ctx aws.Context, input *PutBucketQosInput) (*PutBucketQosOutput, error) {
	req, out := c.PutBucketQosRequest(input)
	req.SetContext(ctx)
	err := req.Send()
	return out, err
}

var opGetBucketQos *aws.Operation

type GetBucketQosInput struct {
	// The name of the bucket.
	Bucket *string `location:"uri" locationName:"Bucket" type:"string" required:"true"`
}

type GetBucketQosOutput struct {
	// Bucket flow control configuration container.
	BucketQosConfiguration *BucketQosConfiguration `locationName:"BucketQosConfiguration" type:"structure"`
	// The HTTP headers of the response.
	Metadata map[string]*string `location:"headers" type:"map"`
	// The HTTP status code of the response.
	StatusCode *int64 `location:"statusCode" type:"integer"`

	metadataGetBucketQosOutput `json:"-" xml:"-"`
}

type metadataGetBucketQosOutput struct {
	SDKShapeTraits bool `type:"structure" payload:"BucketQosConfiguration"`
}

// GetBucketQosRequest generates a request for the GetBucketQos operation.
func (c *S3) GetBucketQosRequest(input *GetBucketQosInput) (req *aws.Request, output *GetBucketQosOutput) {
	oprw.Lock()
	defer oprw.Unlock()
	if opGetBucketQos == nil {
		opGetBucketQos = &aws.Operation{
			Name:       "GetBucketQos",
			HTTPMethod: "GET",
			HTTPPath:   "/{Bucket}?bucketqos",
		}
	}
	if input == nil {
		input = &GetBucketQosInput{}
	}
	req = c.newRequest(opGetBucketQos, input, output)
	output = &GetBucketQosOutput{}
	req.Data = output
	return
}

// GetBucketQos gets bucket flow control configuration.
func (c *S3) GetBucketQos(input *GetBucketQosInput) (*GetBucketQosOutput, error) {
	req, out := c.GetBucketQosRequest(input)
	err := req.Send()
	return out, err
}

func (c *S3) GetBucketQosWithContext(ctx aws.Context, input *GetBucketQosInput) (*GetBucketQosOutput, error) {
	req, out := c.GetBucketQosRequest(input)
	req.SetContext(ctx)
	err := req.Send()
	return out, err
}

var opDeleteBucketQos *aws.Operation

type DeleteBucketQosInput struct {
	// The name of the bucket.
	Bucket *string `location:"uri" locationName:"Bucket" type:"string" required:"true"`
}

type DeleteBucketQosOutput struct {
	// The HTTP headers of the response.
	Metadata map[string]*string `location:"headers" type:"map"`
	// The HTTP status code of the response.
	StatusCode *int64 `location:"statusCode" type:"integer"`
}

// DeleteBucketQosRequest generates a request for the DeleteBucketQos operation.
func (c *S3) DeleteBucketQosRequest(input *DeleteBucketQosInput) (req *aws.Request, output *DeleteBucketQosOutput) {
	oprw.Lock()
	defer oprw.Unlock()
	if opDeleteBucketQos == nil {
		opDeleteBucketQos = &aws.Operation{
			Name:       "DeleteBucketQos",
			HTTPMethod: "DELETE",
			HTTPPath:   "/{Bucket}?bucketqos",
		}
	}
	if input == nil {
		input = &DeleteBucketQosInput{}
	}
	req = c.newRequest(opDeleteBucketQos, input, output)
	output = &DeleteBucketQosOutput{}
	req.Data = output
	return
}

// DeleteBucketQos deletes bucket flow control configuration.
func (c *S3) DeleteBucketQos(input *DeleteBucketQosInput) (*DeleteBucketQosOutput, error) {
	req, out := c.DeleteBucketQosRequest(input)
	err := req.Send()
	return out, err
}

func (c *S3) DeleteBucketQosWithContext(ctx aws.Context, input *DeleteBucketQosInput) (*DeleteBucketQosOutput, error) {
	req, out := c.DeleteBucketQosRequest(input)
	req.SetContext(ctx)
	err := req.Send()
	return out, err
}
