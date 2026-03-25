package s3

import (
	"github.com/ks3sdklib/aws-sdk-go/aws"
)

// InitiateBucketWormInput 新建合规保留策略的输入参数
type InitiateBucketWormInput struct {
	// 存储桶名称。
	Bucket *string `location:"uri" locationName:"Bucket" type:"string" required:"true"`

	// 合规保留策略配置的容器。
	InitiateWormConfiguration *InitiateWormConfiguration `locationName:"InitiateWormConfiguration" type:"structure" required:"true"`

	// 设置扩展请求头。如果现有字段不支持设置所需的请求头，您可以通过此字段进行设置。
	ExtendHeaders map[string]*string `location:"extendHeaders" type:"map"`

	// 设置扩展查询参数。如果现有字段不支持设置所需的查询参数，您可以通过此字段进行设置。
	ExtendQueryParams map[string]*string `location:"extendQueryParams" type:"map"`

	metadataInitiateBucketWormInput `json:"-" xml:"-"`
}

type metadataInitiateBucketWormInput struct {
	SDKShapeTraits bool `type:"structure" payload:"InitiateWormConfiguration"`
}

// InitiateWormConfiguration 合规保留策略配置的容器
type InitiateWormConfiguration struct {
	// 指定保留天数，取值范围为[1, 36500]。
	RetentionPeriodInDays *int64 `locationName:"RetentionPeriodInDays" type:"integer" required:"true"`
}

// InitiateBucketWormOutput 新建合规保留策略的输出参数
type InitiateBucketWormOutput struct {
	// http响应头。
	Metadata map[string]*string `location:"headers" type:"map"`

	// http响应状态码。
	StatusCode *int64 `location:"statusCode" type:"integer"`

	metadataInitiateBucketWormOutput `json:"-" xml:"-"`
}

type metadataInitiateBucketWormOutput struct {
	SDKShapeTraits bool `type:"structure"`
}

// InitiateBucketWormRequest generates a request for the InitiateBucketWorm operation.
func (c *S3) InitiateBucketWormRequest(input *InitiateBucketWormInput) (req *aws.Request, output *InitiateBucketWormOutput) {
	op := &aws.Operation{
		Name:       "InitiateBucketWorm",
		HTTPMethod: "POST",
		HTTPPath:   "/{Bucket}?worm",
	}

	if input == nil {
		input = &InitiateBucketWormInput{}
	}

	req = c.newRequest(op, input, output)
	output = &InitiateBucketWormOutput{}
	req.Data = output
	return
}

// InitiateBucketWorm 新建合规保留策略。
func (c *S3) InitiateBucketWorm(input *InitiateBucketWormInput) (*InitiateBucketWormOutput, error) {
	req, out := c.InitiateBucketWormRequest(input)
	err := req.Send()
	return out, err
}

// InitiateBucketWormWithContext 新建合规保留策略，支持传入上下文。
func (c *S3) InitiateBucketWormWithContext(ctx aws.Context, input *InitiateBucketWormInput) (*InitiateBucketWormOutput, error) {
	req, out := c.InitiateBucketWormRequest(input)
	req.SetContext(ctx)
	err := req.Send()
	return out, err
}

// AbortBucketWormInput 删除未锁定的合规保留策略的输入参数
type AbortBucketWormInput struct {
	// 存储桶名称。
	Bucket *string `location:"uri" locationName:"Bucket" type:"string" required:"true"`

	// 设置扩展请求头。如果现有字段不支持设置所需的请求头，您可以通过此字段进行设置。
	ExtendHeaders map[string]*string `location:"extendHeaders" type:"map"`

	// 设置扩展查询参数。如果现有字段不支持设置所需的查询参数，您可以通过此字段进行设置。
	ExtendQueryParams map[string]*string `location:"extendQueryParams" type:"map"`
}

// AbortBucketWormOutput 删除未锁定的合规保留策略的输出参数
type AbortBucketWormOutput struct {
	// http响应头。
	Metadata map[string]*string `location:"headers" type:"map"`

	// http响应状态码。
	StatusCode *int64 `location:"statusCode" type:"integer"`
}

// AbortBucketWormRequest generates a request for the AbortBucketWorm operation.
func (c *S3) AbortBucketWormRequest(input *AbortBucketWormInput) (req *aws.Request, output *AbortBucketWormOutput) {
	op := &aws.Operation{
		Name:       "AbortBucketWorm",
		HTTPMethod: "DELETE",
		HTTPPath:   "/{Bucket}?worm",
	}

	if input == nil {
		input = &AbortBucketWormInput{}
	}

	req = c.newRequest(op, input, output)
	output = &AbortBucketWormOutput{}
	req.Data = output
	return
}

// AbortBucketWorm 删除未锁定的合规保留策略。
func (c *S3) AbortBucketWorm(input *AbortBucketWormInput) (*AbortBucketWormOutput, error) {
	req, out := c.AbortBucketWormRequest(input)
	err := req.Send()
	return out, err
}

// AbortBucketWormWithContext 删除未锁定的合规保留策略，支持传入上下文。
func (c *S3) AbortBucketWormWithContext(ctx aws.Context, input *AbortBucketWormInput) (*AbortBucketWormOutput, error) {
	req, out := c.AbortBucketWormRequest(input)
	req.SetContext(ctx)
	err := req.Send()
	return out, err
}

// CompleteBucketWormInput 锁定合规保留策略的输入参数
type CompleteBucketWormInput struct {
	// 存储桶名称。
	Bucket *string `location:"uri" locationName:"Bucket" type:"string" required:"true"`

	// 保留策略ID。
	WormId *string `location:"querystring" locationName:"wormId" type:"string" required:"true"`

	// 设置扩展请求头。如果现有字段不支持设置所需的请求头，您可以通过此字段进行设置。
	ExtendHeaders map[string]*string `location:"extendHeaders" type:"map"`

	// 设置扩展查询参数。如果现有字段不支持设置所需的查询参数，您可以通过此字段进行设置。
	ExtendQueryParams map[string]*string `location:"extendQueryParams" type:"map"`
}

// CompleteBucketWormOutput 锁定合规保留策略的输出参数
type CompleteBucketWormOutput struct {
	// http响应头。
	Metadata map[string]*string `location:"headers" type:"map"`

	// http响应状态码。
	StatusCode *int64 `location:"statusCode" type:"integer"`
}

// CompleteBucketWormRequest generates a request for the CompleteBucketWorm operation.
func (c *S3) CompleteBucketWormRequest(input *CompleteBucketWormInput) (req *aws.Request, output *CompleteBucketWormOutput) {
	op := &aws.Operation{
		Name:       "CompleteBucketWorm",
		HTTPMethod: "POST",
		HTTPPath:   "/{Bucket}?wormId",
	}

	if input == nil {
		input = &CompleteBucketWormInput{}
	}

	req = c.newRequest(op, input, output)
	output = &CompleteBucketWormOutput{}
	req.Data = output
	return
}

// CompleteBucketWorm 锁定合规保留策略。
func (c *S3) CompleteBucketWorm(input *CompleteBucketWormInput) (*CompleteBucketWormOutput, error) {
	req, out := c.CompleteBucketWormRequest(input)
	err := req.Send()
	return out, err
}

// CompleteBucketWormWithContext 锁定合规保留策略，支持传入上下文。
func (c *S3) CompleteBucketWormWithContext(ctx aws.Context, input *CompleteBucketWormInput) (*CompleteBucketWormOutput, error) {
	req, out := c.CompleteBucketWormRequest(input)
	req.SetContext(ctx)
	err := req.Send()
	return out, err
}

// ExtendBucketWormInput 延长已锁定的合规保留策略的输入参数
type ExtendBucketWormInput struct {
	// 存储桶名称。
	Bucket *string `location:"uri" locationName:"Bucket" type:"string" required:"true"`

	// 保留策略ID。
	WormId *string `location:"querystring" locationName:"wormId" type:"string" required:"true"`

	// 延长保留策略配置的容器。
	ExtendWormConfiguration *ExtendWormConfiguration `locationName:"ExtendWormConfiguration" type:"structure" required:"true"`

	// 设置扩展请求头。如果现有字段不支持设置所需的请求头，您可以通过此字段进行设置。
	ExtendHeaders map[string]*string `location:"extendHeaders" type:"map"`

	// 设置扩展查询参数。如果现有字段不支持设置所需的查询参数，您可以通过此字段进行设置。
	ExtendQueryParams map[string]*string `location:"extendQueryParams" type:"map"`

	metadataExtendBucketWormInput `json:"-" xml:"-"`
}

type metadataExtendBucketWormInput struct {
	SDKShapeTraits bool `type:"structure" payload:"ExtendWormConfiguration"`
}

// ExtendWormConfiguration 延长保留策略配置的容器
type ExtendWormConfiguration struct {
	// 指定保留天数，取值范围为[1, 36500]，延长后的保留天数必须大于当前保留天数。
	RetentionPeriodInDays *int64 `locationName:"RetentionPeriodInDays" type:"integer" required:"true"`
}

// ExtendBucketWormOutput 延长已锁定的合规保留策略的输出参数
type ExtendBucketWormOutput struct {
	// http响应头。
	Metadata map[string]*string `location:"headers" type:"map"`

	// http响应状态码。
	StatusCode *int64 `location:"statusCode" type:"integer"`

	metadataExtendBucketWormOutput `json:"-" xml:"-"`
}

type metadataExtendBucketWormOutput struct {
	SDKShapeTraits bool `type:"structure"`
}

// ExtendBucketWormRequest generates a request for the ExtendBucketWorm operation.
func (c *S3) ExtendBucketWormRequest(input *ExtendBucketWormInput) (req *aws.Request, output *ExtendBucketWormOutput) {
	op := &aws.Operation{
		Name:       "ExtendBucketWorm",
		HTTPMethod: "POST",
		HTTPPath:   "/{Bucket}?wormId&wormExtend",
	}

	if input == nil {
		input = &ExtendBucketWormInput{}
	}

	req = c.newRequest(op, input, output)
	output = &ExtendBucketWormOutput{}
	req.Data = output
	return
}

// ExtendBucketWorm 延长已锁定的合规保留策略对应Bucket中Object的保留天数。
func (c *S3) ExtendBucketWorm(input *ExtendBucketWormInput) (*ExtendBucketWormOutput, error) {
	req, out := c.ExtendBucketWormRequest(input)
	err := req.Send()
	return out, err
}

// ExtendBucketWormWithContext 延长已锁定的合规保留策略，支持传入上下文。
func (c *S3) ExtendBucketWormWithContext(ctx aws.Context, input *ExtendBucketWormInput) (*ExtendBucketWormOutput, error) {
	req, out := c.ExtendBucketWormRequest(input)
	req.SetContext(ctx)
	err := req.Send()
	return out, err
}

// GetBucketWormInput 获取合规保留策略信息的输入参数
type GetBucketWormInput struct {
	// 存储桶名称。
	Bucket *string `location:"uri" locationName:"Bucket" type:"string" required:"true"`

	// 设置扩展请求头。如果现有字段不支持设置所需的请求头，您可以通过此字段进行设置。
	ExtendHeaders map[string]*string `location:"extendHeaders" type:"map"`

	// 设置扩展查询参数。如果现有字段不支持设置所需的查询参数，您可以通过此字段进行设置。
	ExtendQueryParams map[string]*string `location:"extendQueryParams" type:"map"`
}

// GetBucketWormOutput 获取合规保留策略信息的输出参数
type GetBucketWormOutput struct {
	// 合规保留策略配置信息的容器。
	WormConfiguration *WormConfiguration `locationName:"WormConfiguration" type:"structure"`

	// http响应头。
	Metadata map[string]*string `location:"headers" type:"map"`

	// http响应状态码。
	StatusCode *int64 `location:"statusCode" type:"integer"`

	metadataGetBucketWormOutput `json:"-" xml:"-"`
}

type metadataGetBucketWormOutput struct {
	SDKShapeTraits bool `type:"structure" payload:"WormConfiguration"`
}

// WormConfiguration 合规保留策略配置信息的容器
type WormConfiguration struct {
	// 保留策略ID。
	WormId *string `locationName:"WormId" type:"string"`

	// 保留策略状态，可选值为：InProgress（配置中）、Locked（已锁定）、Expired（已失效）。
	State *string `locationName:"State" type:"string"`

	// 保留天数。
	RetentionPeriodInDays *int64 `locationName:"RetentionPeriodInDays" type:"integer"`

	// 保留策略创建时间，格式为ISO8601。
	CreationDate *string `locationName:"CreationDate" type:"string"`
}

// GetBucketWormRequest generates a request for the GetBucketWorm operation.
func (c *S3) GetBucketWormRequest(input *GetBucketWormInput) (req *aws.Request, output *GetBucketWormOutput) {
	op := &aws.Operation{
		Name:       "GetBucketWorm",
		HTTPMethod: "GET",
		HTTPPath:   "/{Bucket}?worm",
	}

	if input == nil {
		input = &GetBucketWormInput{}
	}

	req = c.newRequest(op, input, output)
	output = &GetBucketWormOutput{}
	req.Data = output
	return
}

// GetBucketWorm 获取合规保留策略信息。
func (c *S3) GetBucketWorm(input *GetBucketWormInput) (*GetBucketWormOutput, error) {
	req, out := c.GetBucketWormRequest(input)
	err := req.Send()
	return out, err
}

// GetBucketWormWithContext 获取合规保留策略信息，支持传入上下文。
func (c *S3) GetBucketWormWithContext(ctx aws.Context, input *GetBucketWormInput) (*GetBucketWormOutput, error) {
	req, out := c.GetBucketWormRequest(input)
	req.SetContext(ctx)
	err := req.Send()
	return out, err
}
