package s3

import (
	"strings"

	"github.com/ks3sdklib/aws-sdk-go/aws"
)

// ListBucketsInput 列举桶的输入参数
type ListBucketsInput struct {
	// 项目ID过滤规则，多个值用逗号分隔。
	ProjectIds []int64 `location:"querystrings" locationName:"projectIds" type:"list"`

	// 桶名前缀过滤规则，多个值为"或"的关系。
	Prefixes []string `type:"list"`

	// 桶所在区域过滤规则，多个值为"或"的关系。
	Regions []string `type:"list"`

	// 桶存储类型过滤规则，多个值为"或"的关系。
	BucketTypes []string `type:"list"`

	// 桶访问类型过滤规则，多个值为"或"的关系。
	VisitTypes []string `type:"list"`

	// 桶数据冗余类型过滤规则，多个值为"或"的关系。
	DataRedundancyTypes []string `type:"list"`

	// 设置扩展请求头。如果现有字段不支持设置所需的请求头，您可以通过此字段进行设置。
	ExtendHeaders map[string]*string `location:"extendHeaders" type:"map"`

	// 设置扩展查询参数。如果现有字段不支持设置所需的查询参数，您可以通过此字段进行设置。
	ExtendQueryParams map[string]*string `location:"extendQueryParams" type:"map"`
}

// ListBucketsOutput 列举桶的输出参数
type ListBucketsOutput struct {
	// 桶列表。
	Buckets []*Bucket `locationNameList:"Bucket" type:"list"`

	// 桶拥有者信息。
	Owner *Owner `type:"structure"`

	metadataListBucketsOutput `json:"-" xml:"-"`

	// http响应头。
	Metadata map[string]*string `location:"headers" type:"map"`

	// http响应状态码。
	StatusCode *int64 `location:"statusCode" type:"integer"`
}

type metadataListBucketsOutput struct {
	SDKShapeTraits bool `type:"structure"`
}

// ListBucketsRequest 列举桶操作的请求。
func (c *S3) ListBucketsRequest(input *ListBucketsInput) (req *aws.Request, output *ListBucketsOutput) {
	op := &aws.Operation{
		Name:       "ListBuckets",
		HTTPMethod: "GET",
		HTTPPath:   "/",
	}

	if input == nil {
		input = &ListBucketsInput{}
	}

	req = c.newRequest(op, input, output)
	output = &ListBucketsOutput{}
	req.Data = output
	return
}

// ListBuckets 列举桶。
func (c *S3) ListBuckets(input *ListBucketsInput) (*ListBucketsOutput, error) {
	req, out := c.ListBucketsRequest(input)
	err := req.Send()
	if err != nil {
		return out, err
	}
	out.Buckets = input.filterBuckets(out.Buckets)
	return out, nil
}

// ListBucketsWithContext 列举桶，支持传入上下文。
func (c *S3) ListBucketsWithContext(ctx aws.Context, input *ListBucketsInput) (*ListBucketsOutput, error) {
	req, out := c.ListBucketsRequest(input)
	req.SetContext(ctx)
	err := req.Send()
	if err != nil {
		return out, err
	}
	out.Buckets = input.filterBuckets(out.Buckets)
	return out, nil
}

// filterBuckets 根据输入参数中的过滤规则过滤桶列表。
// 同一个条件的多个值为"或"的关系，不同条件为"且"的关系。
func (input *ListBucketsInput) filterBuckets(buckets []*Bucket) []*Bucket {
	if input == nil {
		return buckets
	}
	if len(input.Prefixes) == 0 && len(input.Regions) == 0 &&
		len(input.BucketTypes) == 0 && len(input.VisitTypes) == 0 &&
		len(input.DataRedundancyTypes) == 0 {
		return buckets
	}

	result := make([]*Bucket, 0, len(buckets))
	for _, bucket := range buckets {
		if !matchPrefix(bucket.Name, input.Prefixes) ||
			!matchValue(bucket.Region, input.Regions) ||
			!matchValue(bucket.Type, input.BucketTypes) ||
			!matchValue(bucket.VisitType, input.VisitTypes) ||
			!matchValue(bucket.DataRedundancyType, input.DataRedundancyTypes) {
			continue
		}
		result = append(result, bucket)
	}
	return result
}

// matchPrefix 判断桶名是否匹配任一前缀，无过滤规则时返回true。
func matchPrefix(name *string, prefixes []string) bool {
	if len(prefixes) == 0 {
		return true
	}
	if name == nil {
		return false
	}
	for _, prefix := range prefixes {
		if strings.HasPrefix(*name, prefix) {
			return true
		}
	}
	return false
}

// matchValue 判断字段值是否匹配任一过滤值，无过滤规则时返回true。
func matchValue(value *string, filters []string) bool {
	if len(filters) == 0 {
		return true
	}
	if value == nil {
		return false
	}
	for _, filter := range filters {
		if *value == filter {
			return true
		}
	}
	return false
}
