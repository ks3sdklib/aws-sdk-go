package s3

import (
	"github.com/ks3sdklib/aws-sdk-go/aws"
)

// PutBucketArchiveDirectReadInput 设置桶归档直读配置的输入参数
type PutBucketArchiveDirectReadInput struct {
	// 存储桶名称。
	Bucket *string `location:"uri" locationName:"Bucket" type:"string" required:"true"`

	// 归档直读配置的容器。
	ArchiveDirectReadConfiguration *ArchiveDirectReadConfiguration `locationName:"ArchiveDirectReadConfiguration" type:"structure" required:"true"`

	// 设置扩展请求头。如果现有字段不支持设置所需的请求头，您可以通过此字段进行设置。
	ExtendHeaders map[string]*string `location:"extendHeaders" type:"map"`

	// 设置扩展查询参数。如果现有字段不支持设置所需的查询参数，您可以通过此字段进行设置。
	ExtendQueryParams map[string]*string `location:"extendQueryParams" type:"map"`

	metadataPutBucketArchiveDirectReadInput `json:"-" xml:"-"`
}

type metadataPutBucketArchiveDirectReadInput struct {
	SDKShapeTraits bool `type:"structure" payload:"ArchiveDirectReadConfiguration"`

	AutoFillMD5 bool
}

// ArchiveDirectReadConfiguration 归档直读配置的容器
type ArchiveDirectReadConfiguration struct {
	// 是否开启归档直读。true：开启，false：关闭。
	Enabled *bool `locationName:"Enabled" type:"boolean" required:"true"`
}

// PutBucketArchiveDirectReadOutput 设置桶归档直读配置的输出参数
type PutBucketArchiveDirectReadOutput struct {
	// http响应头。
	Metadata map[string]*string `location:"headers" type:"map"`

	// http响应状态码。
	StatusCode *int64 `location:"statusCode" type:"integer"`

	metadataPutBucketArchiveDirectReadOutput `json:"-" xml:"-"`
}

type metadataPutBucketArchiveDirectReadOutput struct {
	SDKShapeTraits bool `type:"structure"`
}

// PutBucketArchiveDirectReadRequest 设置桶归档直读配置操作的请求。
func (c *S3) PutBucketArchiveDirectReadRequest(input *PutBucketArchiveDirectReadInput) (req *aws.Request, output *PutBucketArchiveDirectReadOutput) {
	op := &aws.Operation{
		Name:       "PutBucketArchiveDirectRead",
		HTTPMethod: "PUT",
		HTTPPath:   "/{Bucket}?archiveDirectRead",
	}

	if input == nil {
		input = &PutBucketArchiveDirectReadInput{}
	}

	input.AutoFillMD5 = true
	req = c.newRequest(op, input, output)
	output = &PutBucketArchiveDirectReadOutput{}
	req.Data = output
	return
}

// PutBucketArchiveDirectRead 设置桶归档直读配置。
func (c *S3) PutBucketArchiveDirectRead(input *PutBucketArchiveDirectReadInput) (*PutBucketArchiveDirectReadOutput, error) {
	req, out := c.PutBucketArchiveDirectReadRequest(input)
	err := req.Send()
	return out, err
}

// PutBucketArchiveDirectReadWithContext 设置桶归档直读配置，支持传入上下文。
func (c *S3) PutBucketArchiveDirectReadWithContext(ctx aws.Context, input *PutBucketArchiveDirectReadInput) (*PutBucketArchiveDirectReadOutput, error) {
	req, out := c.PutBucketArchiveDirectReadRequest(input)
	req.SetContext(ctx)
	err := req.Send()
	return out, err
}

// GetBucketArchiveDirectReadInput 获取桶归档直读配置的输入参数
type GetBucketArchiveDirectReadInput struct {
	// 存储桶名称。
	Bucket *string `location:"uri" locationName:"Bucket" type:"string" required:"true"`

	// 设置扩展请求头。如果现有字段不支持设置所需的请求头，您可以通过此字段进行设置。
	ExtendHeaders map[string]*string `location:"extendHeaders" type:"map"`

	// 设置扩展查询参数。如果现有字段不支持设置所需的查询参数，您可以通过此字段进行设置。
	ExtendQueryParams map[string]*string `location:"extendQueryParams" type:"map"`
}

// GetBucketArchiveDirectReadOutput 获取桶归档直读配置的输出参数
type GetBucketArchiveDirectReadOutput struct {
	// 归档直读配置的容器。
	ArchiveDirectReadConfiguration *ArchiveDirectReadConfiguration `locationName:"ArchiveDirectReadConfiguration" type:"structure"`

	// http响应头。
	Metadata map[string]*string `location:"headers" type:"map"`

	// http响应状态码。
	StatusCode *int64 `location:"statusCode" type:"integer"`

	metadataGetBucketArchiveDirectReadOutput `json:"-" xml:"-"`
}

type metadataGetBucketArchiveDirectReadOutput struct {
	SDKShapeTraits bool `type:"structure" payload:"ArchiveDirectReadConfiguration"`
}

// GetBucketArchiveDirectReadRequest 获取桶归档直读配置操作的请求。
func (c *S3) GetBucketArchiveDirectReadRequest(input *GetBucketArchiveDirectReadInput) (req *aws.Request, output *GetBucketArchiveDirectReadOutput) {
	op := &aws.Operation{
		Name:       "GetBucketArchiveDirectRead",
		HTTPMethod: "GET",
		HTTPPath:   "/{Bucket}?archiveDirectRead",
	}

	if input == nil {
		input = &GetBucketArchiveDirectReadInput{}
	}

	req = c.newRequest(op, input, output)
	output = &GetBucketArchiveDirectReadOutput{}
	req.Data = output
	return
}

// GetBucketArchiveDirectRead 获取桶归档直读配置。
func (c *S3) GetBucketArchiveDirectRead(input *GetBucketArchiveDirectReadInput) (*GetBucketArchiveDirectReadOutput, error) {
	req, out := c.GetBucketArchiveDirectReadRequest(input)
	err := req.Send()
	return out, err
}

// GetBucketArchiveDirectReadWithContext 获取桶归档直读配置，支持传入上下文。
func (c *S3) GetBucketArchiveDirectReadWithContext(ctx aws.Context, input *GetBucketArchiveDirectReadInput) (*GetBucketArchiveDirectReadOutput, error) {
	req, out := c.GetBucketArchiveDirectReadRequest(input)
	req.SetContext(ctx)
	err := req.Send()
	return out, err
}
