package s3

import (
	"github.com/ks3sdklib/aws-sdk-go/aws"
)

type PutPublicNetworkBlockInput struct {
	// 公网访问控制规则的容器。
	PublicNetworkBlockConfiguration *PublicNetworkBlockConfiguration `locationName:"PublicNetworkBlockConfiguration" type:"structure" required:"true"`

	// 设置扩展请求头。如果现有字段不支持设置所需的请求头，您可以通过此字段进行设置。
	ExtendHeaders map[string]*string `location:"extendHeaders" type:"map"`

	// 设置扩展查询参数。如果现有字段不支持设置所需的查询参数，您可以通过此字段进行设置。
	ExtendQueryParams map[string]*string `location:"extendQueryParams" type:"map"`

	metadataPutPublicNetworkBlockInput `json:"-" xml:"-"`
}

type metadataPutPublicNetworkBlockInput struct {
	SDKShapeTraits bool `type:"structure" payload:"PublicNetworkBlockConfiguration"`
}

type PublicNetworkBlockConfiguration struct {
	// 设置阻止公网访问类型。
	// All：阻止所有公网访问
	// ExcludeAuthorization：阻止公网访问，除有效鉴权
	// ExcludeConsole：阻止公网访问，控制台除外
	BlockType *string `locationName:"BlockType" type:"string" required:"true"`
}

type PutPublicNetworkBlockOutput struct {
	// http响应头。
	Metadata map[string]*string `location:"headers" type:"map"`

	// http响应状态码。
	StatusCode *int64 `location:"statusCode" type:"integer"`
}

// PutPublicNetworkBlockRequest generates a request for the PutPublicNetworkBlock operation.
func (c *S3) PutPublicNetworkBlockRequest(input *PutPublicNetworkBlockInput) (req *aws.Request, output *PutPublicNetworkBlockOutput) {
	op := &aws.Operation{
		Name:       "PutPublicNetworkBlock",
		HTTPMethod: "PUT",
		HTTPPath:   "/?PublicNetworkBlock",
	}

	if input == nil {
		input = &PutPublicNetworkBlockInput{}
	}

	req = c.newRequest(op, input, output)
	output = &PutPublicNetworkBlockOutput{}
	req.Data = output
	return
}

// PutPublicNetworkBlock sets public network block configuration.
func (c *S3) PutPublicNetworkBlock(input *PutPublicNetworkBlockInput) (*PutPublicNetworkBlockOutput, error) {
	req, out := c.PutPublicNetworkBlockRequest(input)
	err := req.Send()
	return out, err
}

func (c *S3) PutPublicNetworkBlockWithContext(ctx aws.Context, input *PutPublicNetworkBlockInput) (*PutPublicNetworkBlockOutput, error) {
	req, out := c.PutPublicNetworkBlockRequest(input)
	req.SetContext(ctx)
	err := req.Send()
	return out, err
}

type GetPublicNetworkBlockInput struct {
	// 设置扩展请求头。如果现有字段不支持设置所需的请求头，您可以通过此字段进行设置。
	ExtendHeaders map[string]*string `location:"extendHeaders" type:"map"`

	// 设置扩展查询参数。如果现有字段不支持设置所需的查询参数，您可以通过此字段进行设置。
	ExtendQueryParams map[string]*string `location:"extendQueryParams" type:"map"`
}

type GetPublicNetworkBlockOutput struct {
	// 公网访问控制规则的容器。
	PublicNetworkBlockConfiguration *PublicNetworkBlockConfiguration `locationName:"PublicNetworkBlockConfiguration" type:"structure"`

	// http响应头。
	Metadata map[string]*string `location:"headers" type:"map"`

	// http响应状态码。
	StatusCode *int64 `location:"statusCode" type:"integer"`

	metadataGetPublicNetworkBlockOutput `json:"-" xml:"-"`
}

type metadataGetPublicNetworkBlockOutput struct {
	SDKShapeTraits bool `type:"structure" payload:"PublicNetworkBlockConfiguration"`
}

// GetPublicNetworkBlockRequest generates a request for the GetPublicNetworkBlock operation.
func (c *S3) GetPublicNetworkBlockRequest(input *GetPublicNetworkBlockInput) (req *aws.Request, output *GetPublicNetworkBlockOutput) {
	op := &aws.Operation{
		Name:       "GetPublicNetworkBlock",
		HTTPMethod: "GET",
		HTTPPath:   "/?PublicNetworkBlock",
	}

	if input == nil {
		input = &GetPublicNetworkBlockInput{}
	}

	req = c.newRequest(op, input, output)
	output = &GetPublicNetworkBlockOutput{}
	req.Data = output
	return
}

// GetPublicNetworkBlock gets public network block configuration.
func (c *S3) GetPublicNetworkBlock(input *GetPublicNetworkBlockInput) (*GetPublicNetworkBlockOutput, error) {
	req, out := c.GetPublicNetworkBlockRequest(input)
	err := req.Send()
	return out, err
}

func (c *S3) GetPublicNetworkBlockWithContext(ctx aws.Context, input *GetPublicNetworkBlockInput) (*GetPublicNetworkBlockOutput, error) {
	req, out := c.GetPublicNetworkBlockRequest(input)
	req.SetContext(ctx)
	err := req.Send()
	return out, err
}

type DeletePublicNetworkBlockInput struct {
	// 设置扩展请求头。如果现有字段不支持设置所需的请求头，您可以通过此字段进行设置。
	ExtendHeaders map[string]*string `location:"extendHeaders" type:"map"`

	// 设置扩展查询参数。如果现有字段不支持设置所需的查询参数，您可以通过此字段进行设置。
	ExtendQueryParams map[string]*string `location:"extendQueryParams" type:"map"`
}

type DeletePublicNetworkBlockOutput struct {
	// http响应头。
	Metadata map[string]*string `location:"headers" type:"map"`

	// http响应状态码。
	StatusCode *int64 `location:"statusCode" type:"integer"`
}

// DeletePublicNetworkBlockRequest generates a request for the DeletePublicNetworkBlock operation.
func (c *S3) DeletePublicNetworkBlockRequest(input *DeletePublicNetworkBlockInput) (req *aws.Request, output *DeletePublicNetworkBlockOutput) {
	op := &aws.Operation{
		Name:       "DeletePublicNetworkBlock",
		HTTPMethod: "DELETE",
		HTTPPath:   "/?PublicNetworkBlock",
	}

	if input == nil {
		input = &DeletePublicNetworkBlockInput{}
	}

	req = c.newRequest(op, input, output)
	output = &DeletePublicNetworkBlockOutput{}
	req.Data = output
	return
}

// DeletePublicNetworkBlock deletes public network block configuration.
func (c *S3) DeletePublicNetworkBlock(input *DeletePublicNetworkBlockInput) (*DeletePublicNetworkBlockOutput, error) {
	req, out := c.DeletePublicNetworkBlockRequest(input)
	err := req.Send()
	return out, err
}

func (c *S3) DeletePublicNetworkBlockWithContext(ctx aws.Context, input *DeletePublicNetworkBlockInput) (*DeletePublicNetworkBlockOutput, error) {
	req, out := c.DeletePublicNetworkBlockRequest(input)
	req.SetContext(ctx)
	err := req.Send()
	return out, err
}
