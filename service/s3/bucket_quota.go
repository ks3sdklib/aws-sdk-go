package s3

import (
	"github.com/ks3sdklib/aws-sdk-go/aws"
)

type PutBucketQuotaInput struct {
	// 存储桶名称。
	Bucket *string `location:"uri" locationName:"Bucket" type:"string" required:"true"`

	// 存储桶配额规则的容器。
	BucketQuota *BucketQuota `locationName:"Quota" type:"structure" required:"true"`

	// 设置扩展请求头。如果现有字段不支持设置所需的请求头，您可以通过此字段进行设置。
	ExtendHeaders map[string]*string `location:"extendHeaders" type:"map"`

	// 设置扩展查询参数。如果现有字段不支持设置所需的查询参数，您可以通过此字段进行设置。
	ExtendQueryParams map[string]*string `location:"extendQueryParams" type:"map"`

	metadataPutBucketQuotaInput `json:"-" xml:"-"`
}

type metadataPutBucketQuotaInput struct {
	SDKShapeTraits bool `type:"structure" payload:"BucketQuota"`
}

// BucketQuota 桶配额的容器
type BucketQuota struct {
	// 指定桶空间配额值，单位为字节，取值必须为正整数。取值范围为1~9223372036854775807。
	StorageQuota *int64 `locationName:"StorageQuota" type:"integer"`

	// 指定桶每天各维度可用配额。
	Day *TransferQuota `locationName:"Day" type:"structure"`

	// 指定桶每月各维度可用配额。
	Month *TransferQuota `locationName:"Month" type:"structure"`
}

// TransferQuota 流量配额配置的容器
type TransferQuota struct {
	// 指定桶内网上行流量配额值，单位为字节，取值必须为正整数。取值范围为1~9223372036854775807。
	IntranetFlowUp *int64 `locationName:"IntranetFlowUp" type:"integer"`

	// 指定桶内网下行流量配额值，单位为字节，取值必须为正整数。取值范围为1~9223372036854775807。
	IntranetFlowDown *int64 `locationName:"IntranetFlowDown" type:"integer"`

	// 指定桶外网上行流量配额值，单位为字节，取值必须为正整数。取值范围为1~9223372036854775807。
	ExtranetFlowUp *int64 `locationName:"ExtranetFlowUp" type:"integer"`

	// 指定桶外网下行流量配额值，单位为字节，取值必须为正整数。取值范围为1~9223372036854775807。
	ExtranetFlowDown *int64 `locationName:"ExtranetFlowDown" type:"integer"`

	// 指定桶CDN上行流量配额值，单位为字节，取值必须为正整数。取值范围为1~9223372036854775807。
	CDNFlowUp *int64 `locationName:"CDNFlowUp" type:"integer"`

	// 指定桶CDN下行流量配额值，单位为字节，取值必须为正整数。取值范围为1~9223372036854775807。
	CDNFlowDown *int64 `locationName:"CDNFlowDown" type:"integer"`

	// 指定桶PUT类请求配额值，单位为个，取值必须为正整数。取值范围为1~9223372036854775807。
	Put *int64 `locationName:"Put" type:"integer"`

	// 指定桶GET类请求配额值，单位为个，取值必须为正整数。取值范围为1~9223372036854775807。
	Get *int64 `locationName:"Get" type:"integer"`

	// 指定桶List请求配额值，单位为个，取值必须为正整数。取值范围为1~9223372036854775807。
	List *int64 `locationName:"List" type:"integer"`
}

type PutBucketQuotaOutput struct {
	// http响应头。
	Metadata map[string]*string `location:"headers" type:"map"`

	// http响应状态码。
	StatusCode *int64 `location:"statusCode" type:"integer"`
}

// PutBucketQuotaRequest 生成设置存储桶配额的请求。
func (c *S3) PutBucketQuotaRequest(input *PutBucketQuotaInput) (req *aws.Request, output *PutBucketQuotaOutput) {
	op := &aws.Operation{
		Name:       "PutBucketQuota",
		HTTPMethod: "PUT",
		HTTPPath:   "/{Bucket}?quota",
	}

	if input == nil {
		input = &PutBucketQuotaInput{}
	}

	req = c.newRequest(op, input, output)
	output = &PutBucketQuotaOutput{}
	req.Data = output
	return
}

// PutBucketQuota 设置存储桶配额。
func (c *S3) PutBucketQuota(input *PutBucketQuotaInput) (*PutBucketQuotaOutput, error) {
	req, out := c.PutBucketQuotaRequest(input)
	err := req.Send()
	return out, err
}

// PutBucketQuotaWithContext 设置存储桶配额，支持传入上下文。
func (c *S3) PutBucketQuotaWithContext(ctx aws.Context, input *PutBucketQuotaInput) (*PutBucketQuotaOutput, error) {
	req, out := c.PutBucketQuotaRequest(input)
	req.SetContext(ctx)
	err := req.Send()
	return out, err
}

type GetBucketQuotaInput struct {
	// 存储桶名称。
	Bucket *string `location:"uri" locationName:"Bucket" type:"string" required:"true"`

	// 设置扩展请求头。如果现有字段不支持设置所需的请求头，您可以通过此字段进行设置。
	ExtendHeaders map[string]*string `location:"extendHeaders" type:"map"`

	// 设置扩展查询参数。如果现有字段不支持设置所需的查询参数，您可以通过此字段进行设置。
	ExtendQueryParams map[string]*string `location:"extendQueryParams" type:"map"`
}

type GetBucketQuotaOutput struct {
	// 存储桶配额规则的容器。
	BucketQuota *BucketQuota `locationName:"Quota" type:"structure"`

	// http响应头。
	Metadata map[string]*string `location:"headers" type:"map"`

	// http响应状态码。
	StatusCode *int64 `location:"statusCode" type:"integer"`

	metadataGetBucketQuotaOutput `json:"-" xml:"-"`
}

type metadataGetBucketQuotaOutput struct {
	SDKShapeTraits bool `type:"structure" payload:"BucketQuota"`
}

// GetBucketQuotaRequest 生成获取存储桶配额的请求。
func (c *S3) GetBucketQuotaRequest(input *GetBucketQuotaInput) (req *aws.Request, output *GetBucketQuotaOutput) {
	op := &aws.Operation{
		Name:       "GetBucketQuota",
		HTTPMethod: "GET",
		HTTPPath:   "/{Bucket}?quota",
	}

	if input == nil {
		input = &GetBucketQuotaInput{}
	}

	req = c.newRequest(op, input, output)
	output = &GetBucketQuotaOutput{}
	req.Data = output
	return
}

// GetBucketQuota 获取存储桶配额。
func (c *S3) GetBucketQuota(input *GetBucketQuotaInput) (*GetBucketQuotaOutput, error) {
	req, out := c.GetBucketQuotaRequest(input)
	err := req.Send()
	return out, err
}

// GetBucketQuotaWithContext 获取存储桶配额，支持传入上下文。
func (c *S3) GetBucketQuotaWithContext(ctx aws.Context, input *GetBucketQuotaInput) (*GetBucketQuotaOutput, error) {
	req, out := c.GetBucketQuotaRequest(input)
	req.SetContext(ctx)
	err := req.Send()
	return out, err
}

type DeleteBucketQuotaInput struct {
	// 存储桶名称。
	Bucket *string `location:"uri" locationName:"Bucket" type:"string" required:"true"`

	// 设置扩展请求头。如果现有字段不支持设置所需的请求头，您可以通过此字段进行设置。
	ExtendHeaders map[string]*string `location:"extendHeaders" type:"map"`

	// 设置扩展查询参数。如果现有字段不支持设置所需的查询参数，您可以通过此字段进行设置。
	ExtendQueryParams map[string]*string `location:"extendQueryParams" type:"map"`
}

type DeleteBucketQuotaOutput struct {
	// http响应头。
	Metadata map[string]*string `location:"headers" type:"map"`

	// http响应状态码。
	StatusCode *int64 `location:"statusCode" type:"integer"`
}

// DeleteBucketQuotaRequest 生成删除存储桶配额的请求。
func (c *S3) DeleteBucketQuotaRequest(input *DeleteBucketQuotaInput) (req *aws.Request, output *DeleteBucketQuotaOutput) {
	op := &aws.Operation{
		Name:       "DeleteBucketQuota",
		HTTPMethod: "DELETE",
		HTTPPath:   "/{Bucket}?quota",
	}

	if input == nil {
		input = &DeleteBucketQuotaInput{}
	}

	req = c.newRequest(op, input, output)
	output = &DeleteBucketQuotaOutput{}
	req.Data = output
	return
}

// DeleteBucketQuota 删除存储桶配额。
func (c *S3) DeleteBucketQuota(input *DeleteBucketQuotaInput) (*DeleteBucketQuotaOutput, error) {
	req, out := c.DeleteBucketQuotaRequest(input)
	err := req.Send()
	return out, err
}

// DeleteBucketQuotaWithContext 删除存储桶配额，支持传入上下文。
func (c *S3) DeleteBucketQuotaWithContext(ctx aws.Context, input *DeleteBucketQuotaInput) (*DeleteBucketQuotaOutput, error) {
	req, out := c.DeleteBucketQuotaRequest(input)
	req.SetContext(ctx)
	err := req.Send()
	return out, err
}
