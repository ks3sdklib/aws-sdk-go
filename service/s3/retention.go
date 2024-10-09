package s3

import (
	"github.com/ks3sdklib/aws-sdk-go/aws"
	"time"
)

// PutBucketRetentionRequest generates a request for the PutBucketRetention operation.
func (c *S3) PutBucketRetentionRequest(input *PutBucketRetentionInput) (req *aws.Request, output *PutBucketRetentionOutput) {
	oprw.Lock()
	defer oprw.Unlock()
	if opPutBucketRetention == nil {
		opPutBucketRetention = &aws.Operation{
			Name:       "PutBucketRetention",
			HTTPMethod: "PUT",
			HTTPPath:   "/{Bucket}?retention",
		}
	}
	if input == nil {
		input = &PutBucketRetentionInput{}
	}
	input.AutoFillMD5 = true
	req = c.newRequest(opPutBucketRetention, input, output)
	output = &PutBucketRetentionOutput{}
	req.Data = output
	return
}

// PutBucketRetention sets the retention configuration on a bucket.
func (c *S3) PutBucketRetention(input *PutBucketRetentionInput) (*PutBucketRetentionOutput, error) {
	req, out := c.PutBucketRetentionRequest(input)
	err := req.Send()
	return out, err
}

func (c *S3) PutBucketRetentionWithContext(ctx aws.Context, input *PutBucketRetentionInput) (*PutBucketRetentionOutput, error) {
	req, out := c.PutBucketRetentionRequest(input)
	req.SetContext(ctx)
	err := req.Send()
	return out, err
}

var opPutBucketRetention *aws.Operation

type PutBucketRetentionInput struct {
	Bucket *string `location:"uri" locationName:"Bucket" type:"string" required:"true"`

	RetentionConfiguration *BucketRetentionConfiguration `locationName:"RetentionConfiguration" type:"structure"`

	ContentType *string `location:"header" locationName:"Content-Type" type:"string"`

	metadataPutBucketRetentionInput `json:"-" xml:"-"`
}

type metadataPutBucketRetentionInput struct {
	SDKShapeTraits bool `type:"structure" payload:"RetentionConfiguration"`

	AutoFillMD5 bool
}

type BucketRetentionConfiguration struct {
	// A container that contains a specific rule for the recycle bin.
	Rule *RetentionRule `locationName:"Rule" type:"structure" required:"true"`
}

type RetentionRule struct {
	// The open status of the recycle bin is not case-sensitive.
	// Valid values: Enabled, Disabled. Enabled indicates enabling the recycle bin, Disabled indicates disabling the recycle bin.
	Status *string `locationName:"Status" type:"string" required:"true"`

	// Specify how many days after the object enters the recycle bin to be completely deleted.
	// When Days is not set, the object will be permanently retained in the recycle bin after deletion.
	// Value range: 1-365
	Days *int64 `locationName:"Days" type:"integer"`
}

type PutBucketRetentionOutput struct {
	Metadata map[string]*string `location:"headers"  type:"map"`

	StatusCode *int64 `location:"statusCode" type:"integer"`
}

// GetBucketRetentionRequest generates a request for the GetBucketRetention operation.
func (c *S3) GetBucketRetentionRequest(input *GetBucketRetentionInput) (req *aws.Request, output *GetBucketRetentionOutput) {
	oprw.Lock()
	defer oprw.Unlock()
	if opGetBucketRetention == nil {
		opGetBucketRetention = &aws.Operation{
			Name:       "GetBucketRetention",
			HTTPMethod: "GET",
			HTTPPath:   "/{Bucket}?retention",
		}
	}
	if input == nil {
		input = &GetBucketRetentionInput{}
	}
	req = c.newRequest(opGetBucketRetention, input, output)
	output = &GetBucketRetentionOutput{
		RetentionConfiguration: &BucketRetentionConfiguration{},
	}
	req.Data = output
	return
}

// GetBucketRetention gets the retention configuration for the bucket.
func (c *S3) GetBucketRetention(input *GetBucketRetentionInput) (*GetBucketRetentionOutput, error) {
	req, out := c.GetBucketRetentionRequest(input)
	err := req.Send()
	return out, err
}

func (c *S3) GetBucketRetentionWithContext(ctx aws.Context, input *GetBucketRetentionInput) (*GetBucketRetentionOutput, error) {
	req, out := c.GetBucketRetentionRequest(input)
	req.SetContext(ctx)
	err := req.Send()
	return out, err
}

var opGetBucketRetention *aws.Operation

type GetBucketRetentionInput struct {
	Bucket *string `location:"uri" locationName:"Bucket" type:"string" required:"true"`
}

type GetBucketRetentionOutput struct {
	RetentionConfiguration *BucketRetentionConfiguration `locationName:"RetentionConfiguration" type:"structure"`

	Metadata map[string]*string `location:"headers"  type:"map"`

	StatusCode *int64 `location:"statusCode" type:"integer"`

	metadataGetBucketRetentionInput `json:"-" xml:"-"`
}

type metadataGetBucketRetentionInput struct {
	SDKShapeTraits bool `type:"structure" payload:"RetentionConfiguration"`
}

// ListRetentionRequest generates a request for the ListRetention operation.
func (c *S3) ListRetentionRequest(input *ListRetentionInput) (req *aws.Request, output *ListRetentionOutput) {
	oprw.Lock()
	defer oprw.Unlock()

	if opListRetention == nil {
		opListRetention = &aws.Operation{
			Name:       "ListRetention",
			HTTPMethod: "GET",
			HTTPPath:   "/{Bucket}?recycle",
		}
	}

	if input == nil {
		input = &ListRetentionInput{}
	}

	req = c.newRequest(opListRetention, input, output)
	output = &ListRetentionOutput{}
	req.Data = output
	return
}

// ListRetention lists the objects in the recycle bin.
func (c *S3) ListRetention(input *ListRetentionInput) (*ListRetentionOutput, error) {
	req, out := c.ListRetentionRequest(input)
	err := req.Send()
	return out, err
}

func (c *S3) ListRetentionWithContext(ctx aws.Context, input *ListRetentionInput) (*ListRetentionOutput, error) {
	req, out := c.ListRetentionRequest(input)
	req.SetContext(ctx)
	err := req.Send()
	return out, err
}

var opListRetention *aws.Operation

type ListRetentionInput struct {
	// The name of the bucket.
	Bucket *string `location:"uri" locationName:"Bucket" type:"string" required:"true"`

	// Specifies the key to start with when listing objects in a bucket.
	Marker *string `location:"querystring" locationName:"marker" type:"string"`

	// Sets the maximum number of keys returned in the response. The response might
	// contain fewer keys but will never contain more.
	MaxKeys *int64 `location:"querystring" locationName:"max-keys" type:"integer"`

	// Limits the response to keys that begin with the specified prefix.
	Prefix *string `location:"querystring" locationName:"prefix" type:"string"`

	ContentType *string `location:"header" locationName:"Content-Type" type:"string"`

	metadataListRetentionInput `json:"-" xml:"-"`
}

type metadataListRetentionInput struct {
	SDKShapeTraits bool `type:"structure"`
}

type ListRetentionOutput struct {
	// A container that lists information about the list of objects in the recycle bin.
	ListRetentionResult *ListRetentionResult `locationName:"ListRetentionResult" type:"structure"`

	Metadata map[string]*string `location:"headers"  type:"map"`

	StatusCode *int64 `location:"statusCode" type:"integer"`

	metadataListRetentionOutput `json:"-" xml:"-"`
}

type metadataListRetentionOutput struct {
	SDKShapeTraits bool `type:"structure" payload:"ListRetentionResult"`
}

type ListRetentionResult struct {
	// The name of the bucket.
	Name *string `type:"string"`

	// Specify the prefix of the Key when requesting this List.
	Prefix *string `type:"string"`

	// The maximum number of objects returned is 1000 by default.
	MaxKeys *int64 `type:"integer"`

	// Specify the starting position of the object in the target bucket.
	Marker *string `type:"string"`

	// The starting point for the next listed file. Users can use this value as a marker parameter
	// for the next List Retention.
	NextMarker *string `type:"string"`

	// Whether it has been truncated. If the number of records in the Object list exceeds the set
	// maximum value, it will be truncated.
	IsTruncated *bool `type:"boolean"`

	// The encoding method for Object names.
	EncodingType *string `type:"string"`

	// List of Objects Listed.
	Contents []*RetentionObject `type:"list" flattened:"true"`
}

type RetentionObject struct {
	// The key of the object.
	Key *string `type:"string"`

	// The size of the object is measured in bytes.
	Size *int64 `type:"integer"`

	// The entity label of an object, ETag, is generated when uploading an object to identify its content.
	ETag *string `type:"string"`

	// The last time the object was modified.
	LastModified *time.Time `type:"timestamp" timestampFormat:"iso8601"`

	// The owner information of this bucket.
	Owner *Owner `type:"structure"`

	// The class of storage used to store the object.
	StorageClass *string `type:"string"`

	// The version ID of the object.
	RetentionId *string `type:"string"`

	// The time when the object was moved to the recycle bin.
	RecycleTime *time.Time `type:"timestamp" timestampFormat:"iso8601"`

	// The time when an object is completely deleted from the recycle bin.
	EstimatedClearTime *time.Time `type:"timestamp" timestampFormat:"iso8601"`
}

// RecoverObjectRequest generates a request for the RecoverObject operation.`
func (c *S3) RecoverObjectRequest(input *ListRetentionInput) (req *aws.Request, output *ListRetentionOutput) {
	oprw.Lock()
	defer oprw.Unlock()

	if opListRetention == nil {
		opListRetention = &aws.Operation{
			Name:       "ListRetention",
			HTTPMethod: "GET",
			HTTPPath:   "/{Bucket}/{Key+}?recover",
		}
	}

	if input == nil {
		input = &ListRetentionInput{}
	}

	req = c.newRequest(opListRetention, input, output)
	output = &ListRetentionOutput{}
	req.Data = output
	return
}

// RecoverObject recovers the object from the recycle bin.
func (c *S3) RecoverObject(input *ListRetentionInput) (*ListRetentionOutput, error) {
	req, out := c.ListRetentionRequest(input)
	err := req.Send()
	return out, err
}

func (c *S3) RecoverObjectWithContext(ctx aws.Context, input *ListRetentionInput) (*ListRetentionOutput, error) {
	req, out := c.ListRetentionRequest(input)
	req.SetContext(ctx)
	err := req.Send()
	return out, err
}

var opRecoverObject *aws.Operation

type RecoverObjectInput struct {
	// The name of the bucket.
	Bucket *string `location:"uri" locationName:"Bucket" type:"string" required:"true"`

	// The key of the object.
	Key *string `location:"uri" locationName:"Key" type:"string" required:"true"`
}
