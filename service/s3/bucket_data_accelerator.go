package s3

import (
	"github.com/ks3sdklib/aws-sdk-go/aws"
)

// PutBucketDataAcceleratorInput 创建或修改加速器的输入参数
type PutBucketDataAcceleratorInput struct {
	// 存储桶名称。
	Bucket *string `location:"uri" locationName:"Bucket" type:"string" required:"true"`

	// 加速器配置。
	DataAcceleratorConfiguration *DataAcceleratorConfiguration `locationName:"DataAcceleratorConfiguration" type:"structure" required:"true"`

	// 设置扩展请求头。如果现有字段不支持设置所需的请求头，您可以通过此字段进行设置。
	ExtendHeaders map[string]*string `location:"extendHeaders" type:"map"`

	// 设置扩展查询参数。如果现有字段不支持设置所需的查询参数，您可以通过此字段进行设置。
	ExtendQueryParams map[string]*string `location:"extendQueryParams" type:"map"`

	metadataPutBucketDataAcceleratorInput `json:"-" xml:"-"`
}

type metadataPutBucketDataAcceleratorInput struct {
	SDKShapeTraits bool `type:"structure" payload:"DataAcceleratorConfiguration"`

	AutoFillMD5 bool
}

// DataAcceleratorConfiguration 加速器配置的容器
type DataAcceleratorConfiguration struct {
	// 加速器的可用区。
	AvailableZone *string `locationName:"AvailableZone" type:"string"`

	// 加速器容量，单位GB。取值范围：[50, 204800]。
	Quota *int64 `locationName:"Quota" type:"integer"`

	// 加速策略配置的容器。
	AcceleratePaths *AcceleratePaths `locationName:"AcceleratePaths" type:"structure"`
}

// AcceleratePaths 加速策略配置的容器
type AcceleratePaths struct {
	// 存放加速前缀的容器。单个加速器规则最多支持填写10个Path，且前缀之间不能重叠。
	Path []*Path `locationName:"Path" type:"list" flattened:"true"`
}

// Path 存放加速前缀的容器
type Path struct {
	// 加速前缀。取值范围：1-1024。单个规则内的所有Prefix不允许重叠。
	Prefix *string `locationName:"Prefix" type:"string"`

	// 前缀是否开启同步预热。默认值：false。
	SyncWarmup *bool `locationName:"SyncWarmup" type:"boolean"`
}

// PutBucketDataAcceleratorOutput 创建或修改加速器的输出参数
type PutBucketDataAcceleratorOutput struct {
	// http响应头。
	Metadata map[string]*string `location:"headers" type:"map"`

	// http响应状态码。
	StatusCode *int64 `location:"statusCode" type:"integer"`
}

// PutBucketDataAcceleratorRequest 创建或修改加速器操作的请求。
func (c *S3) PutBucketDataAcceleratorRequest(input *PutBucketDataAcceleratorInput) (req *aws.Request, output *PutBucketDataAcceleratorOutput) {
	op := &aws.Operation{
		Name:       "PutBucketDataAccelerator",
		HTTPMethod: "PUT",
		HTTPPath:   "/{Bucket}?dataAccelerator",
	}

	if input == nil {
		input = &PutBucketDataAcceleratorInput{}
	}

	input.AutoFillMD5 = true
	req = c.newRequest(op, input, output)
	output = &PutBucketDataAcceleratorOutput{}
	req.Data = output
	return
}

// PutBucketDataAccelerator 创建或修改加速器。
func (c *S3) PutBucketDataAccelerator(input *PutBucketDataAcceleratorInput) (*PutBucketDataAcceleratorOutput, error) {
	req, out := c.PutBucketDataAcceleratorRequest(input)
	err := req.Send()
	return out, err
}

// PutBucketDataAcceleratorWithContext 创建或修改加速器，支持传入上下文。
func (c *S3) PutBucketDataAcceleratorWithContext(ctx aws.Context, input *PutBucketDataAcceleratorInput) (*PutBucketDataAcceleratorOutput, error) {
	req, out := c.PutBucketDataAcceleratorRequest(input)
	req.SetContext(ctx)
	err := req.Send()
	return out, err
}

// GetBucketDataAcceleratorInput 获取加速器配置的输入参数
type GetBucketDataAcceleratorInput struct {
	// 存储桶名称。
	Bucket *string `location:"uri" locationName:"Bucket" type:"string" required:"true"`

	// 设置扩展请求头。如果现有字段不支持设置所需的请求头，您可以通过此字段进行设置。
	ExtendHeaders map[string]*string `location:"extendHeaders" type:"map"`

	// 设置扩展查询参数。如果现有字段不支持设置所需的查询参数，您可以通过此字段进行设置。
	ExtendQueryParams map[string]*string `location:"extendQueryParams" type:"map"`
}

// GetBucketDataAcceleratorOutput 获取加速器配置的输出参数
type GetBucketDataAcceleratorOutput struct {
	// 加速器配置。
	DataAcceleratorConfiguration *DataAcceleratorConfiguration `locationName:"DataAcceleratorConfiguration" type:"structure"`

	// http响应头。
	Metadata map[string]*string `location:"headers" type:"map"`

	// http响应状态码。
	StatusCode *int64 `location:"statusCode" type:"integer"`

	metadataGetBucketDataAcceleratorOutput `json:"-" xml:"-"`
}

type metadataGetBucketDataAcceleratorOutput struct {
	SDKShapeTraits bool `type:"structure" payload:"DataAcceleratorConfiguration"`
}

// GetBucketDataAcceleratorRequest 获取加速器配置操作的请求。
func (c *S3) GetBucketDataAcceleratorRequest(input *GetBucketDataAcceleratorInput) (req *aws.Request, output *GetBucketDataAcceleratorOutput) {
	op := &aws.Operation{
		Name:       "GetBucketDataAccelerator",
		HTTPMethod: "GET",
		HTTPPath:   "/{Bucket}?dataAccelerator",
	}

	if input == nil {
		input = &GetBucketDataAcceleratorInput{}
	}

	req = c.newRequest(op, input, output)
	output = &GetBucketDataAcceleratorOutput{}
	req.Data = output
	return
}

// GetBucketDataAccelerator 获取加速器配置。
func (c *S3) GetBucketDataAccelerator(input *GetBucketDataAcceleratorInput) (*GetBucketDataAcceleratorOutput, error) {
	req, out := c.GetBucketDataAcceleratorRequest(input)
	err := req.Send()
	return out, err
}

// GetBucketDataAcceleratorWithContext 获取加速器配置，支持传入上下文。
func (c *S3) GetBucketDataAcceleratorWithContext(ctx aws.Context, input *GetBucketDataAcceleratorInput) (*GetBucketDataAcceleratorOutput, error) {
	req, out := c.GetBucketDataAcceleratorRequest(input)
	req.SetContext(ctx)
	err := req.Send()
	return out, err
}

// DeleteBucketDataAcceleratorInput 删除加速器配置的输入参数
type DeleteBucketDataAcceleratorInput struct {
	// 存储桶名称。
	Bucket *string `location:"uri" locationName:"Bucket" type:"string" required:"true"`

	// 设置扩展请求头。如果现有字段不支持设置所需的请求头，您可以通过此字段进行设置。
	ExtendHeaders map[string]*string `location:"extendHeaders" type:"map"`

	// 设置扩展查询参数。如果现有字段不支持设置所需的查询参数，您可以通过此字段进行设置。
	ExtendQueryParams map[string]*string `location:"extendQueryParams" type:"map"`
}

// DeleteBucketDataAcceleratorOutput 删除加速器配置的输出参数
type DeleteBucketDataAcceleratorOutput struct {
	// http响应头。
	Metadata map[string]*string `location:"headers" type:"map"`

	// http响应状态码。
	StatusCode *int64 `location:"statusCode" type:"integer"`
}

// DeleteBucketDataAcceleratorRequest 删除加速器配置操作的请求。
func (c *S3) DeleteBucketDataAcceleratorRequest(input *DeleteBucketDataAcceleratorInput) (req *aws.Request, output *DeleteBucketDataAcceleratorOutput) {
	op := &aws.Operation{
		Name:       "DeleteBucketDataAccelerator",
		HTTPMethod: "DELETE",
		HTTPPath:   "/{Bucket}?dataAccelerator",
	}

	if input == nil {
		input = &DeleteBucketDataAcceleratorInput{}
	}

	req = c.newRequest(op, input, output)
	output = &DeleteBucketDataAcceleratorOutput{}
	req.Data = output
	return
}

// DeleteBucketDataAccelerator 删除加速器配置。
func (c *S3) DeleteBucketDataAccelerator(input *DeleteBucketDataAcceleratorInput) (*DeleteBucketDataAcceleratorOutput, error) {
	req, out := c.DeleteBucketDataAcceleratorRequest(input)
	err := req.Send()
	return out, err
}

// DeleteBucketDataAcceleratorWithContext 删除加速器配置，支持传入上下文。
func (c *S3) DeleteBucketDataAcceleratorWithContext(ctx aws.Context, input *DeleteBucketDataAcceleratorInput) (*DeleteBucketDataAcceleratorOutput, error) {
	req, out := c.DeleteBucketDataAcceleratorRequest(input)
	req.SetContext(ctx)
	err := req.Send()
	return out, err
}
