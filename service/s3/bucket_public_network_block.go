package s3

import (
	"github.com/ks3sdklib/aws-sdk-go/aws"
)

type PutBucketPublicNetworkBlockInput struct {
	// 存储桶名称。
	Bucket *string `location:"uri" locationName:"Bucket" type:"string" required:"true"`

	// 公网访问控制规则的容器。
	PublicNetworkBlockConfiguration *PublicNetworkBlockConfiguration `locationName:"PublicNetworkBlockConfiguration" type:"structure" required:"true"`

	// 设置扩展请求头。如果现有字段不支持设置所需的请求头，您可以通过此字段进行设置。
	ExtendHeaders map[string]*string `location:"extendHeaders" type:"map"`

	// 设置扩展查询参数。如果现有字段不支持设置所需的查询参数，您可以通过此字段进行设置。
	ExtendQueryParams map[string]*string `location:"extendQueryParams" type:"map"`

	metadataPutBucketPublicNetworkBlockInput `json:"-" xml:"-"`
}

type metadataPutBucketPublicNetworkBlockInput struct {
	SDKShapeTraits bool `type:"structure" payload:"BucketPublicNetworkBlockConfiguration"`
}

type PutBucketPublicNetworkBlockOutput struct {
	// http响应头。
	Metadata map[string]*string `location:"headers" type:"map"`

	// http响应状态码。
	StatusCode *int64 `location:"statusCode" type:"integer"`
}

// PutBucketPublicNetworkBlockRequest generates a request for the PutBucketPublicNetworkBlock operation.
func (c *S3) PutBucketPublicNetworkBlockRequest(input *PutBucketPublicNetworkBlockInput) (req *aws.Request, output *PutBucketPublicNetworkBlockOutput) {
	op := &aws.Operation{
		Name:       "PutBucketPublicNetworkBlock",
		HTTPMethod: "PUT",
		HTTPPath:   "/{Bucket}?BucketPublicNetworkBlock",
	}

	if input == nil {
		input = &PutBucketPublicNetworkBlockInput{}
	}

	req = c.newRequest(op, input, output)
	output = &PutBucketPublicNetworkBlockOutput{}
	req.Data = output
	return
}

// PutBucketPublicNetworkBlock sets bucket public network block configuration.
func (c *S3) PutBucketPublicNetworkBlock(input *PutBucketPublicNetworkBlockInput) (*PutBucketPublicNetworkBlockOutput, error) {
	req, out := c.PutBucketPublicNetworkBlockRequest(input)
	err := req.Send()
	return out, err
}

func (c *S3) PutBucketPublicNetworkBlockWithContext(ctx aws.Context, input *PutBucketPublicNetworkBlockInput) (*PutBucketPublicNetworkBlockOutput, error) {
	req, out := c.PutBucketPublicNetworkBlockRequest(input)
	req.SetContext(ctx)
	err := req.Send()
	return out, err
}

type GetBucketPublicNetworkBlockInput struct {
	// 存储桶名称。
	Bucket *string `location:"uri" locationName:"Bucket" type:"string" required:"true"`

	// 设置扩展请求头。如果现有字段不支持设置所需的请求头，您可以通过此字段进行设置。
	ExtendHeaders map[string]*string `location:"extendHeaders" type:"map"`

	// 设置扩展查询参数。如果现有字段不支持设置所需的查询参数，您可以通过此字段进行设置。
	ExtendQueryParams map[string]*string `location:"extendQueryParams" type:"map"`
}

type GetBucketPublicNetworkBlockOutput struct {
	// 公网访问控制规则的容器。
	PublicNetworkBlockConfiguration *PublicNetworkBlockConfiguration `locationName:"PublicNetworkBlockConfiguration" type:"structure"`

	// http响应头。
	Metadata map[string]*string `location:"headers" type:"map"`

	// http响应状态码。
	StatusCode *int64 `location:"statusCode" type:"integer"`

	metadataGetBucketPublicNetworkBlockOutput `json:"-" xml:"-"`
}

type metadataGetBucketPublicNetworkBlockOutput struct {
	SDKShapeTraits bool `type:"structure" payload:"PublicNetworkBlockConfiguration"`
}

// GetBucketPublicNetworkBlockRequest generates a request for the GetBucketPublicNetworkBlock operation.
func (c *S3) GetBucketPublicNetworkBlockRequest(input *GetBucketPublicNetworkBlockInput) (req *aws.Request, output *GetBucketPublicNetworkBlockOutput) {
	op := &aws.Operation{
		Name:       "GetBucketPublicNetworkBlock",
		HTTPMethod: "GET",
		HTTPPath:   "/{Bucket}?BucketPublicNetworkBlock",
	}

	if input == nil {
		input = &GetBucketPublicNetworkBlockInput{}
	}

	req = c.newRequest(op, input, output)
	output = &GetBucketPublicNetworkBlockOutput{}
	req.Data = output
	return
}

// GetBucketPublicNetworkBlock gets bucket public network block configuration.
func (c *S3) GetBucketPublicNetworkBlock(input *GetBucketPublicNetworkBlockInput) (*GetBucketPublicNetworkBlockOutput, error) {
	req, out := c.GetBucketPublicNetworkBlockRequest(input)
	err := req.Send()
	return out, err
}

func (c *S3) GetBucketPublicNetworkBlockWithContext(ctx aws.Context, input *GetBucketPublicNetworkBlockInput) (*GetBucketPublicNetworkBlockOutput, error) {
	req, out := c.GetBucketPublicNetworkBlockRequest(input)
	req.SetContext(ctx)
	err := req.Send()
	return out, err
}

type DeleteBucketPublicNetworkBlockInput struct {
	// 存储桶名称。
	Bucket *string `location:"uri" locationName:"Bucket" type:"string" required:"true"`

	// 设置扩展请求头。如果现有字段不支持设置所需的请求头，您可以通过此字段进行设置。
	ExtendHeaders map[string]*string `location:"extendHeaders" type:"map"`

	// 设置扩展查询参数。如果现有字段不支持设置所需的查询参数，您可以通过此字段进行设置。
	ExtendQueryParams map[string]*string `location:"extendQueryParams" type:"map"`
}

type DeleteBucketPublicNetworkBlockOutput struct {
	// http响应头。
	Metadata map[string]*string `location:"headers" type:"map"`

	// http响应状态码。
	StatusCode *int64 `location:"statusCode" type:"integer"`
}

// DeleteBucketPublicNetworkBlockRequest generates a request for the DeleteBucketPublicNetworkBlock operation.
func (c *S3) DeleteBucketPublicNetworkBlockRequest(input *DeleteBucketPublicNetworkBlockInput) (req *aws.Request, output *DeleteBucketPublicNetworkBlockOutput) {
	op := &aws.Operation{
		Name:       "DeleteBucketPublicNetworkBlock",
		HTTPMethod: "DELETE",
		HTTPPath:   "/{Bucket}?BucketPublicNetworkBlock",
	}

	if input == nil {
		input = &DeleteBucketPublicNetworkBlockInput{}
	}

	req = c.newRequest(op, input, output)
	output = &DeleteBucketPublicNetworkBlockOutput{}
	req.Data = output
	return
}

// DeleteBucketPublicNetworkBlock deletes bucket public network block configuration.
func (c *S3) DeleteBucketPublicNetworkBlock(input *DeleteBucketPublicNetworkBlockInput) (*DeleteBucketPublicNetworkBlockOutput, error) {
	req, out := c.DeleteBucketPublicNetworkBlockRequest(input)
	err := req.Send()
	return out, err
}

func (c *S3) DeleteBucketPublicNetworkBlockWithContext(ctx aws.Context, input *DeleteBucketPublicNetworkBlockInput) (*DeleteBucketPublicNetworkBlockOutput, error) {
	req, out := c.DeleteBucketPublicNetworkBlockRequest(input)
	req.SetContext(ctx)
	err := req.Send()
	return out, err
}
