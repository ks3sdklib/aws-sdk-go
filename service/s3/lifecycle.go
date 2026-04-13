package s3

import (
	"github.com/ks3sdklib/aws-sdk-go/aws"
	"time"
)

type PutBucketLifecycleInput struct {
	Bucket *string `location:"uri" locationName:"Bucket" type:"string" required:"true"`

	LifecycleConfiguration *LifecycleConfiguration `locationName:"LifecycleConfiguration" type:"structure"`

	ContentType *string `location:"header" locationName:"Content-Type" type:"string"`

	// Specifies whether lifecycle rules allow prefix overlap.
	AllowSameActionOverlap *bool `location:"header" locationName:"x-amz-allow-same-action-overlap" type:"boolean"`

	// Set extend request headers. If the existing fields do not support setting the request header you need, you can set it through this field.
	ExtendHeaders map[string]*string `location:"extendHeaders" type:"map"`

	// Set extend query params. If the existing fields do not support setting the query param you need, you can set it through this field.
	ExtendQueryParams map[string]*string `location:"extendQueryParams" type:"map"`

	metadataPutBucketLifecycleInput `json:"-" xml:"-"`
}

type metadataPutBucketLifecycleInput struct {
	SDKShapeTraits bool `type:"structure" payload:"LifecycleConfiguration"`

	AutoFillMD5 bool
}

type LifecycleConfiguration struct {
	Rules []*LifecycleRule `locationName:"Rule" type:"list" flattened:"true" required:"true"`
}

type LifecycleRule struct {
	// Unique identifier for the rule. The value cannot be longer than 255 characters.
	ID *string `type:"string"`

	// If 'Enabled', the rule is currently being applied. If 'Disabled', the rule
	// is not currently being applied.
	Status *string `type:"string" required:"true"`

	// Specifies the prefix, each Rule can only have one Filter, and the prefixes of different
	// Rules cannot conflict.
	Filter *LifecycleFilter `type:"structure"`

	// Specifies the time when the object is deleted
	Expiration *LifecycleExpiration `type:"structure"`

	// Specifies when an object transitions to a specified storage class.
	Transitions []*Transition `locationName:"Transition" type:"list" flattened:"true"`

	// Specifies the expiration time for multipart uploads.
	AbortIncompleteMultipartUpload *AbortIncompleteMultipartUpload `type:"structure"`
}

// LifecycleFilter 生命周期规则过滤器。
type LifecycleFilter struct {
	// 对象前缀。
	Prefix *string `type:"string"`

	// 逻辑与条件，当存在多个筛选条件时使用。
	And *And `locationName:"And" type:"structure"`

	// 筛选条件，指定规则所适用的对象的大小必须大于某值（不包括等于），单位：Bytes。
	ObjectSizeGreaterThan *int64 `type:"integer"`

	// 筛选条件，指定规则所适用的对象的大小必须小于某值（不包括等于），单位：Bytes。
	ObjectSizeLessThan *int64 `type:"integer"`

	// 包含多个Not取反规则的容器，单个Filter下最多一个Nots节点。
	Nots *Nots `locationName:"Nots" type:"structure"`
}

// And 逻辑与条件，用于组合多个筛选条件。
type And struct {
	// 对象前缀。
	Prefix *string `type:"string"`

	// 标签列表。
	Tags []*Tag `locationName:"Tag" type:"list" flattened:"true"`
}

// Nots 包含多个Not取反规则的容器。
type Nots struct {
	// 取反规则列表，单个Nots下至少存在1个Not，不同Not之间为或的关系。
	NotList []*Not `locationName:"Not" type:"list" flattened:"true"`
}

// Not 取反规则，如果Object满足该条件，则跳过本条规则的匹配。
type Not struct {
	// 本条排除规则所适用的Object前缀。
	Prefix *string `type:"string"`

	// 本条排除规则所适用的Object标签，至多一个。
	Tag *Tag `locationName:"Tag" type:"structure"`
}

type LifecycleExpiration struct {
	// Indicates at what date the object is to be moved or deleted. Should be in
	// GMT ISO 8601 Format.
	Date *time.Time `type:"timestamp" timestampFormat:"iso8601"`

	// Indicates the lifetime, in days, of the objects that are subject to the rule.
	// The value must be a non-zero positive integer.
	Days *int64 `type:"integer"`
}

type Transition struct {
	// Indicates at what date the object is to be moved or deleted. Should be in
	// GMT ISO 8601 Format.
	Date *time.Time `type:"timestamp" timestampFormat:"iso8601"`

	// Specifies the number of days after the object is last modified or accessed that the lifecycle rule takes effect.
	// When the value of IsAccessTime in the request is true, this parameter indicates that the lifecycle rule takes
	// effect based on the last access time of the object. When IsAccessTime is not set in the request or is set to false,
	// this parameter indicates that the lifecycle rule takes effect based on the last modification time of the object.
	// This parameter is mutually exclusive with Date.
	Days *int64 `type:"integer"`

	// The class of storage used to store the object.
	StorageClass *string `type:"string"`

	// Specifies whether to use the last access time matching rule.
	// true: indicates that the last access time of the object is used for matching.
	// false: indicates that the last modification time of the object is used for matching.
	IsAccessTime *bool `type:"boolean"`

	// Specifies whether to convert the object to the source storage type when accessed again after the object is
	// converted to another storage type. This is only valid when IsAccessTime is set to true.
	// true: Indicates that the object is converted to the source storage type when accessed again.
	// false: Indicates that the object is still the target storage type when accessed again.
	ReturnToStdWhenVisit *bool `type:"boolean"`
}

type AbortIncompleteMultipartUpload struct {
	// Relative expiration time: The expiration time in days after the last modified time
	DaysAfterInitiation *int64 `type:"integer"`
	// objects created before the date will be expired
	Date *string `type:"string"`
}

type PutBucketLifecycleOutput struct {
	// The HTTP headers of the response.
	Metadata map[string]*string `location:"headers" type:"map"`

	// The HTTP status code of the response.
	StatusCode *int64 `location:"statusCode" type:"integer"`
}

// PutBucketLifecycleRequest generates a request for the PutBucketLifecycle operation.
func (c *S3) PutBucketLifecycleRequest(input *PutBucketLifecycleInput) (req *aws.Request, output *PutBucketLifecycleOutput) {
	op := &aws.Operation{
		Name:       "PutBucketLifecycle",
		HTTPMethod: "PUT",
		HTTPPath:   "/{Bucket}?lifecycle",
	}

	if input == nil {
		input = &PutBucketLifecycleInput{}
	}

	input.AutoFillMD5 = true
	req = c.newRequest(op, input, output)
	output = &PutBucketLifecycleOutput{}
	req.Data = output
	return
}

// PutBucketLifecycle Sets lifecycle configuration for your bucket. If a lifecycle configuration
// exists, it replaces it.
func (c *S3) PutBucketLifecycle(input *PutBucketLifecycleInput) (*PutBucketLifecycleOutput, error) {
	req, out := c.PutBucketLifecycleRequest(input)
	err := req.Send()
	return out, err
}

func (c *S3) PutBucketLifecycleWithContext(ctx aws.Context, input *PutBucketLifecycleInput) (*PutBucketLifecycleOutput, error) {
	req, out := c.PutBucketLifecycleRequest(input)
	req.SetContext(ctx)
	err := req.Send()
	return out, err
}

type GetBucketLifecycleInput struct {
	Bucket *string `location:"uri" locationName:"Bucket" type:"string" required:"true"`

	ContentType *string `location:"header" locationName:"Content-Type" type:"string"`

	// Set extend request headers. If the existing fields do not support setting the request header you need, you can set it through this field.
	ExtendHeaders map[string]*string `location:"extendHeaders" type:"map"`

	// Set extend query params. If the existing fields do not support setting the query param you need, you can set it through this field.
	ExtendQueryParams map[string]*string `location:"extendQueryParams" type:"map"`
}

type GetBucketLifecycleOutput struct {
	Rules []*LifecycleRule `locationName:"Rule" type:"list" flattened:"true"`

	metadataGetBucketLifecycleOutput `json:"-" xml:"-"`

	Metadata map[string]*string `location:"headers"  type:"map"`

	StatusCode *int64 `location:"statusCode" type:"integer"`
}

type metadataGetBucketLifecycleOutput struct {
	SDKShapeTraits bool `type:"structure"`
}

// GetBucketLifecycleRequest generates a request for the GetBucketLifecycle operation.
func (c *S3) GetBucketLifecycleRequest(input *GetBucketLifecycleInput) (req *aws.Request, output *GetBucketLifecycleOutput) {
	op := &aws.Operation{
		Name:       "GetBucketLifecycle",
		HTTPMethod: "GET",
		HTTPPath:   "/{Bucket}?lifecycle",
	}

	if input == nil {
		input = &GetBucketLifecycleInput{}
	}

	req = c.newRequest(op, input, output)
	output = &GetBucketLifecycleOutput{}
	req.Data = output
	return
}

// GetBucketLifecycle Returns the lifecycle configuration information set on the bucket.
func (c *S3) GetBucketLifecycle(input *GetBucketLifecycleInput) (*GetBucketLifecycleOutput, error) {
	req, out := c.GetBucketLifecycleRequest(input)
	err := req.Send()
	return out, err
}

func (c *S3) GetBucketLifecycleWithContext(ctx aws.Context, input *GetBucketLifecycleInput) (*GetBucketLifecycleOutput, error) {
	req, out := c.GetBucketLifecycleRequest(input)
	req.SetContext(ctx)
	err := req.Send()
	return out, err
}

type DeleteBucketLifecycleInput struct {
	Bucket *string `location:"uri" locationName:"Bucket" type:"string" required:"true"`

	ContentType *string `location:"header" locationName:"Content-Type" type:"string"`

	// Set extend request headers. If the existing fields do not support setting the request header you need, you can set it through this field.
	ExtendHeaders map[string]*string `location:"extendHeaders" type:"map"`

	// Set extend query params. If the existing fields do not support setting the query param you need, you can set it through this field.
	ExtendQueryParams map[string]*string `location:"extendQueryParams" type:"map"`
}

type DeleteBucketLifecycleOutput struct {
	Metadata map[string]*string `location:"headers"  type:"map"`

	StatusCode *int64 `location:"statusCode" type:"integer"`
}

// DeleteBucketLifecycleRequest generates a request for the DeleteBucketLifecycle operation.
func (c *S3) DeleteBucketLifecycleRequest(input *DeleteBucketLifecycleInput) (req *aws.Request, output *DeleteBucketLifecycleOutput) {
	op := &aws.Operation{
		Name:       "DeleteBucketLifecycle",
		HTTPMethod: "DELETE",
		HTTPPath:   "/{Bucket}?lifecycle",
	}

	if input == nil {
		input = &DeleteBucketLifecycleInput{}
	}

	req = c.newRequest(op, input, output)
	output = &DeleteBucketLifecycleOutput{}
	req.Data = output
	return
}

// DeleteBucketLifecycle Deletes the lifecycle configuration from the bucket.
func (c *S3) DeleteBucketLifecycle(input *DeleteBucketLifecycleInput) (*DeleteBucketLifecycleOutput, error) {
	req, out := c.DeleteBucketLifecycleRequest(input)
	err := req.Send()
	return out, err
}

func (c *S3) DeleteBucketLifecycleWithContext(ctx aws.Context, input *DeleteBucketLifecycleInput) (*DeleteBucketLifecycleOutput, error) {
	req, out := c.DeleteBucketLifecycleRequest(input)
	req.SetContext(ctx)
	err := req.Send()
	return out, err
}
