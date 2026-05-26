package s3

import (
	"github.com/ks3sdklib/aws-sdk-go/aws"
)

// PutBucketHttp2Input 设置桶HTTP/2配置的输入参数
type PutBucketHttp2Input struct {
	// 存储桶名称。
	Bucket *string `location:"uri" locationName:"Bucket" type:"string" required:"true"`

	// HTTP/2配置的容器。
	Http2Configuration *Http2Configuration `locationName:"Http2Configuration" type:"structure" required:"true"`

	// 设置扩展请求头。如果现有字段不支持设置所需的请求头，您可以通过此字段进行设置。
	ExtendHeaders map[string]*string `location:"extendHeaders" type:"map"`

	// 设置扩展查询参数。如果现有字段不支持设置所需的查询参数，您可以通过此字段进行设置。
	ExtendQueryParams map[string]*string `location:"extendQueryParams" type:"map"`

	metadataPutBucketHttp2Input `json:"-" xml:"-"`
}

type metadataPutBucketHttp2Input struct {
	SDKShapeTraits bool `type:"structure" payload:"Http2Configuration"`

	AutoFillMD5 bool
}

// Http2Configuration HTTP/2配置的容器
type Http2Configuration struct {
	// HTTP/2协议的开启状态。Enabled：开启HTTP/2，Disabled：关闭HTTP/2。
	Status *string `locationName:"Status" type:"string" required:"true"`
}

// PutBucketHttp2Output 设置桶HTTP/2配置的输出参数
type PutBucketHttp2Output struct {
	// http响应头。
	Metadata map[string]*string `location:"headers" type:"map"`

	// http响应状态码。
	StatusCode *int64 `location:"statusCode" type:"integer"`

	metadataPutBucketHttp2Output `json:"-" xml:"-"`
}

type metadataPutBucketHttp2Output struct {
	SDKShapeTraits bool `type:"structure"`
}

// PutBucketHttp2Request 设置桶HTTP/2配置操作的请求。
func (c *S3) PutBucketHttp2Request(input *PutBucketHttp2Input) (req *aws.Request, output *PutBucketHttp2Output) {
	op := &aws.Operation{
		Name:       "PutBucketHttp2",
		HTTPMethod: "PUT",
		HTTPPath:   "/{Bucket}?http2",
	}

	if input == nil {
		input = &PutBucketHttp2Input{}
	}

	input.AutoFillMD5 = true
	req = c.newRequest(op, input, output)
	output = &PutBucketHttp2Output{}
	req.Data = output
	return
}

// PutBucketHttp2 设置桶HTTP/2配置。
func (c *S3) PutBucketHttp2(input *PutBucketHttp2Input) (*PutBucketHttp2Output, error) {
	req, out := c.PutBucketHttp2Request(input)
	err := req.Send()
	return out, err
}

// PutBucketHttp2WithContext 设置桶HTTP/2配置，支持传入上下文。
func (c *S3) PutBucketHttp2WithContext(ctx aws.Context, input *PutBucketHttp2Input) (*PutBucketHttp2Output, error) {
	req, out := c.PutBucketHttp2Request(input)
	req.SetContext(ctx)
	err := req.Send()
	return out, err
}

// GetBucketHttp2Input 获取桶HTTP/2配置的输入参数
type GetBucketHttp2Input struct {
	// 存储桶名称。
	Bucket *string `location:"uri" locationName:"Bucket" type:"string" required:"true"`

	// 设置扩展请求头。如果现有字段不支持设置所需的请求头，您可以通过此字段进行设置。
	ExtendHeaders map[string]*string `location:"extendHeaders" type:"map"`

	// 设置扩展查询参数。如果现有字段不支持设置所需的查询参数，您可以通过此字段进行设置。
	ExtendQueryParams map[string]*string `location:"extendQueryParams" type:"map"`
}

// GetBucketHttp2Output 获取桶HTTP/2配置的输出参数
type GetBucketHttp2Output struct {
	// HTTP/2配置的容器。
	Http2Configuration *Http2Configuration `locationName:"Http2Configuration" type:"structure"`

	// http响应头。
	Metadata map[string]*string `location:"headers" type:"map"`

	// http响应状态码。
	StatusCode *int64 `location:"statusCode" type:"integer"`

	metadataGetBucketHttp2Output `json:"-" xml:"-"`
}

type metadataGetBucketHttp2Output struct {
	SDKShapeTraits bool `type:"structure" payload:"Http2Configuration"`
}

// GetBucketHttp2Request 获取桶HTTP/2配置操作的请求。
func (c *S3) GetBucketHttp2Request(input *GetBucketHttp2Input) (req *aws.Request, output *GetBucketHttp2Output) {
	op := &aws.Operation{
		Name:       "GetBucketHttp2",
		HTTPMethod: "GET",
		HTTPPath:   "/{Bucket}?http2",
	}

	if input == nil {
		input = &GetBucketHttp2Input{}
	}

	req = c.newRequest(op, input, output)
	output = &GetBucketHttp2Output{}
	req.Data = output
	return
}

// GetBucketHttp2 获取桶HTTP/2配置。
func (c *S3) GetBucketHttp2(input *GetBucketHttp2Input) (*GetBucketHttp2Output, error) {
	req, out := c.GetBucketHttp2Request(input)
	err := req.Send()
	return out, err
}

// GetBucketHttp2WithContext 获取桶HTTP/2配置，支持传入上下文。
func (c *S3) GetBucketHttp2WithContext(ctx aws.Context, input *GetBucketHttp2Input) (*GetBucketHttp2Output, error) {
	req, out := c.GetBucketHttp2Request(input)
	req.SetContext(ctx)
	err := req.Send()
	return out, err
}
