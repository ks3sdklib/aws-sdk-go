package s3

import (
	"github.com/ks3sdklib/aws-sdk-go/aws"
)

var opPutRequesterQos *aws.Operation

type PutRequesterQosInput struct {
	// The name of the bucket.
	Bucket *string `location:"uri" locationName:"Bucket" type:"string" required:"true"`

	// Requester flow control configuration container.
	RequesterQosConfiguration *RequesterQosConfiguration `locationName:"RequesterQosConfiguration" type:"structure" required:"true"`

	metadataPutRequesterQosInput `json:"-" xml:"-"`
}

type metadataPutRequesterQosInput struct {
	SDKShapeTraits bool `type:"structure" payload:"RequesterQosConfiguration"`
}

type RequesterQosConfiguration struct {
	// Set the requester flow control rules.
	Rules []*RequesterQosRule `locationName:"Rule" type:"list" flattened:"true" required:"true"`
}

type RequesterQosRule struct {
	// Specify the account type that needs flow control.
	// Optional values: User/Role.
	UserType *string `locationName:"UserType" type:"string" required:"true"`

	// Specify the account that needs flow control.
	// Format: accountId/userName、accountId/roleName.
	Krn *string `locationName:"Krn" type:"string" required:"true"`

	// Set access account flow control quota.
	Quotas []*BucketQosQuota `locationName:"Quota" type:"list" flattened:"true" required:"true"`
}

type PutRequesterQosOutput struct {
	// The HTTP headers of the response.
	Metadata map[string]*string `location:"headers" type:"map"`
	// The HTTP status code of the response.
	StatusCode *int64 `location:"statusCode" type:"integer"`
}

// PutRequesterQosRequest generates a request for the PutRequesterQos operation.
func (c *S3) PutRequesterQosRequest(input *PutRequesterQosInput) (req *aws.Request, output *PutRequesterQosOutput) {
	oprw.Lock()
	defer oprw.Unlock()
	if opPutRequesterQos == nil {
		opPutRequesterQos = &aws.Operation{
			Name:       "PutRequesterQos",
			HTTPMethod: "PUT",
			HTTPPath:   "/{Bucket}?requesterqos",
		}
	}
	if input == nil {
		input = &PutRequesterQosInput{}
	}
	req = c.newRequest(opPutRequesterQos, input, output)
	output = &PutRequesterQosOutput{}
	req.Data = output
	return
}

// PutRequesterQos sets requester flow control configuration.
func (c *S3) PutRequesterQos(input *PutRequesterQosInput) (*PutRequesterQosOutput, error) {
	req, out := c.PutRequesterQosRequest(input)
	err := req.Send()
	return out, err
}

func (c *S3) PutRequesterQosWithContext(ctx aws.Context, input *PutRequesterQosInput) (*PutRequesterQosOutput, error) {
	req, out := c.PutRequesterQosRequest(input)
	req.SetContext(ctx)
	err := req.Send()
	return out, err
}

var opGetRequesterQos *aws.Operation

type GetRequesterQosInput struct {
	// The name of the bucket.
	Bucket *string `location:"uri" locationName:"Bucket" type:"string" required:"true"`
}

type GetRequesterQosOutput struct {
	// Requester flow control configuration container.
	RequesterQosConfiguration *RequesterQosConfiguration `locationName:"RequesterQosConfiguration" type:"structure"`
	// The HTTP headers of the response.
	Metadata map[string]*string `location:"headers" type:"map"`
	// The HTTP status code of the response.
	StatusCode *int64 `location:"statusCode" type:"integer"`

	metadataGetRequesterQosOutput `json:"-" xml:"-"`
}

type metadataGetRequesterQosOutput struct {
	SDKShapeTraits bool `type:"structure" payload:"RequesterQosConfiguration"`
}

// GetRequesterQosRequest generates a request for the GetRequesterQos operation.
func (c *S3) GetRequesterQosRequest(input *GetRequesterQosInput) (req *aws.Request, output *GetRequesterQosOutput) {
	oprw.Lock()
	defer oprw.Unlock()
	if opGetRequesterQos == nil {
		opGetRequesterQos = &aws.Operation{
			Name:       "GetRequesterQos",
			HTTPMethod: "GET",
			HTTPPath:   "/{Bucket}?requesterqos",
		}
	}
	if input == nil {
		input = &GetRequesterQosInput{}
	}
	req = c.newRequest(opGetRequesterQos, input, output)
	output = &GetRequesterQosOutput{}
	req.Data = output
	return
}

// GetRequesterQos gets requester flow control configuration.
func (c *S3) GetRequesterQos(input *GetRequesterQosInput) (*GetRequesterQosOutput, error) {
	req, out := c.GetRequesterQosRequest(input)
	err := req.Send()
	return out, err
}

func (c *S3) GetRequesterQosWithContext(ctx aws.Context, input *GetRequesterQosInput) (*GetRequesterQosOutput, error) {
	req, out := c.GetRequesterQosRequest(input)
	req.SetContext(ctx)
	err := req.Send()
	return out, err
}

var opDeleteRequesterQos *aws.Operation

type DeleteRequesterQosInput struct {
	// The name of the bucket.
	Bucket *string `location:"uri" locationName:"Bucket" type:"string" required:"true"`
}

type DeleteRequesterQosOutput struct {
	// The HTTP headers of the response.
	Metadata map[string]*string `location:"headers" type:"map"`
	// The HTTP status code of the response.
	StatusCode *int64 `location:"statusCode" type:"integer"`
}

// DeleteRequesterQosRequest generates a request for the DeleteRequesterQos operation.
func (c *S3) DeleteRequesterQosRequest(input *DeleteRequesterQosInput) (req *aws.Request, output *DeleteRequesterQosOutput) {
	oprw.Lock()
	defer oprw.Unlock()
	if opDeleteRequesterQos == nil {
		opDeleteRequesterQos = &aws.Operation{
			Name:       "DeleteRequesterQos",
			HTTPMethod: "DELETE",
			HTTPPath:   "/{Bucket}?requesterqos",
		}
	}
	if input == nil {
		input = &DeleteRequesterQosInput{}
	}
	req = c.newRequest(opDeleteRequesterQos, input, output)
	output = &DeleteRequesterQosOutput{}
	req.Data = output
	return
}

// DeleteRequesterQos deletes requester flow control configuration.
func (c *S3) DeleteRequesterQos(input *DeleteRequesterQosInput) (*DeleteRequesterQosOutput, error) {
	req, out := c.DeleteRequesterQosRequest(input)
	err := req.Send()
	return out, err
}

func (c *S3) DeleteRequesterQosWithContext(ctx aws.Context, input *DeleteRequesterQosInput) (*DeleteRequesterQosOutput, error) {
	req, out := c.DeleteRequesterQosRequest(input)
	req.SetContext(ctx)
	err := req.Send()
	return out, err
}
