package s3

import (
	"github.com/ks3sdklib/aws-sdk-go/aws"
	"time"
)

type LifecycleConfiguration struct {
	Rules []*LifecycleRule `locationName:"Rule" type:"list" flattened:"true" required:"true"`

	metadataLifecycleConfiguration `json:"-" xml:"-"`
}

type metadataLifecycleConfiguration struct {
	SDKShapeTraits bool `type:"structure"`
}

type LifecycleExpiration struct {
	// Indicates at what date the object is to be moved or deleted. Should be in
	// GMT ISO 8601 Format.
	Date *time.Time `type:"timestamp" timestampFormat:"iso8601"`

	// Indicates the lifetime, in days, of the objects that are subject to the rule.
	// The value must be a non-zero positive integer.
	Days *int64 `type:"integer"`

	metadataLifecycleExpiration `json:"-" xml:"-"`
}

type metadataLifecycleExpiration struct {
	SDKShapeTraits bool `type:"structure"`
}

type LifecycleRule struct {
	Expiration *LifecycleExpiration `type:"structure"`

	// Unique identifier for the rule. The value cannot be longer than 255 characters.
	ID *string `type:"string"`

	// Specifies when noncurrent object versions expire. Upon expiration, Amazon
	// S3 permanently deletes the noncurrent object versions. You set this lifecycle
	// configuration action on a bucket that has versioning enabled (or suspended)
	// to request that Amazon S3 delete noncurrent object versions at a specific
	// period in the object's lifetime.
	NoncurrentVersionExpiration *NoncurrentVersionExpiration `type:"structure"`

	// Container for the transition rule that describes when noncurrent objects
	// transition to the GLACIER storage class. If your bucket is versioning-enabled
	// (or versioning is suspended), you can set this action to request that Amazon
	// S3 transition noncurrent object versions to the GLACIER storage class at
	// a specific period in the object's lifetime.
	NoncurrentVersionTransition *NoncurrentVersionTransition `type:"structure"`

	// Prefix identifying one or more objects to which the rule applies.
	Filter *LifecycleFilter `type:"structure"`

	// If 'Enabled', the rule is currently being applied. If 'Disabled', the rule
	// is not currently being applied.
	Status *string `type:"string" required:"true"`

	Transitions []*Transition `locationName:"Transition" type:"list" flattened:"true"`

	metadataLifecycleRule `json:"-" xml:"-"`
}

type PutBucketLifecycleInput struct {
	Bucket *string `location:"uri" locationName:"Bucket" type:"string" required:"true"`

	LifecycleConfiguration *LifecycleConfiguration `locationName:"LifecycleConfiguration" type:"structure"`

	ContentType *string `location:"header" locationName:"Content-Type" type:"string"`

	metadataPutBucketLifecycleInput `json:"-" xml:"-"`
}

type metadataPutBucketLifecycleInput struct {
	SDKShapeTraits bool `type:"structure" payload:"LifecycleConfiguration"`
	AutoFillMD5    bool
}

type PutBucketLifecycleOutput struct {
	metadataPutBucketLifecycleOutput `json:"-" xml:"-"`

	Metadata map[string]*string `location:"headers"  type:"map"`

	StatusCode *int64 `location:"statusCode" type:"integer"`
}

type metadataPutBucketLifecycleOutput struct {
	SDKShapeTraits bool `type:"structure"`
}

type DeleteBucketLifecycleInput struct {
	Bucket *string `location:"uri" locationName:"Bucket" type:"string" required:"true"`

	ContentType *string `location:"header" locationName:"Content-Type" type:"string"`

	metadataDeleteBucketLifecycleInput `json:"-" xml:"-"`
}

type metadataDeleteBucketLifecycleInput struct {
	SDKShapeTraits bool `type:"structure"`
}

type DeleteBucketLifecycleOutput struct {
	metadataDeleteBucketLifecycleOutput `json:"-" xml:"-"`

	Metadata map[string]*string `location:"headers"  type:"map"`

	StatusCode *int64 `location:"statusCode" type:"integer"`
}

type metadataDeleteBucketLifecycleOutput struct {
	SDKShapeTraits bool `type:"structure"`
}
type GetBucketLifecycleInput struct {
	Bucket *string `location:"uri" locationName:"Bucket" type:"string" required:"true"`

	ContentType *string `location:"header" locationName:"Content-Type" type:"string"`

	metadataGetBucketLifecycleInput `json:"-" xml:"-"`
}

type metadataGetBucketLifecycleInput struct {
	SDKShapeTraits bool `type:"structure"`
}
type metadataLifecycleRule struct {
	SDKShapeTraits bool `type:"structure"`
}
type LifecycleFilter struct {
	And                     *And `locationName:"And" type :"structure"`
	metadataLifecycleFilter `json:"-" xml:"-"`
}
type metadataLifecycleFilter struct {
	SDKShapeTraits bool `type:"structure"`
}
type And struct {
	Prefix      *string `type:"string" required:"true"`
	Tag         []*Tag  `locationNameList:"Tag" type:"list" flattened:"true"`
	metadataAnd `json:"-" xml:"-"`
}
type metadataAnd struct {
	SDKShapeTraits bool `type:"structure"`
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

// DeleteBucketLifecycleRequest generates a request for the DeleteBucketLifecycle operation.
func (c *S3) DeleteBucketLifecycleRequest(input *DeleteBucketLifecycleInput) (req *aws.Request, output *DeleteBucketLifecycleOutput) {
	oprw.Lock()
	defer oprw.Unlock()

	if opDeleteBucketLifecycle == nil {
		opDeleteBucketLifecycle = &aws.Operation{
			Name:       "DeleteBucketLifecycle",
			HTTPMethod: "DELETE",
			HTTPPath:   "/{Bucket}?lifecycle",
		}
	}

	if input == nil {
		input = &DeleteBucketLifecycleInput{}
	}

	req = c.newRequest(opDeleteBucketLifecycle, input, output)
	output = &DeleteBucketLifecycleOutput{}
	req.Data = output
	return
}

// Deletes the lifecycle configuration from the bucket.
func (c *S3) DeleteBucketLifecycle(input *DeleteBucketLifecycleInput) (*DeleteBucketLifecycleOutput, error) {
	req, out := c.DeleteBucketLifecycleRequest(input)
	err := req.Send()
	return out, err
}

var opDeleteBucketLifecycle *aws.Operation

// GetBucketLifecycleRequest generates a request for the GetBucketLifecycle operation.
func (c *S3) GetBucketLifecycleRequest(input *GetBucketLifecycleInput) (req *aws.Request, output *GetBucketLifecycleOutput) {
	oprw.Lock()
	defer oprw.Unlock()

	if opGetBucketLifecycle == nil {
		opGetBucketLifecycle = &aws.Operation{
			Name:       "GetBucketLifecycle",
			HTTPMethod: "GET",
			HTTPPath:   "/{Bucket}?lifecycle",
		}
	}

	if input == nil {
		input = &GetBucketLifecycleInput{}
	}

	req = c.newRequest(opGetBucketLifecycle, input, output)
	output = &GetBucketLifecycleOutput{}
	req.Data = output
	return
}

// Returns the lifecycle configuration information set on the bucket.
func (c *S3) GetBucketLifecycle(input *GetBucketLifecycleInput) (*GetBucketLifecycleOutput, error) {
	req, out := c.GetBucketLifecycleRequest(input)
	err := req.Send()
	return out, err
}

var opGetBucketLifecycle *aws.Operation

// PutBucketLifecycleRequest generates a request for the PutBucketLifecycle operation.
func (c *S3) PutBucketLifecycleRequest(input *PutBucketLifecycleInput) (req *aws.Request, output *PutBucketLifecycleOutput) {
	oprw.Lock()
	defer oprw.Unlock()

	if opPutBucketLifecycle == nil {
		opPutBucketLifecycle = &aws.Operation{
			Name:       "PutBucketLifecycle",
			HTTPMethod: "PUT",
			HTTPPath:   "/{Bucket}?lifecycle",
		}
	}

	if input == nil {
		input = &PutBucketLifecycleInput{}
	}
	input.AutoFillMD5 = true
	req = c.newRequest(opPutBucketLifecycle, input, output)
	output = &PutBucketLifecycleOutput{}
	req.Data = output
	return
}

// Sets lifecycle configuration for your bucket. If a lifecycle configuration
// exists, it replaces it.
func (c *S3) PutBucketLifecycle(input *PutBucketLifecycleInput) (*PutBucketLifecycleOutput, error) {
	req, out := c.PutBucketLifecycleRequest(input)
	err := req.Send()
	return out, err
}

var opPutBucketLifecycle *aws.Operation
