package s3

import (
	"github.com/ks3sdklib/aws-sdk-go/aws"
	"io"
)

var opAppendObject *aws.Operation

type AppendObjectInput struct {
	// The name of the bucket.
	Bucket *string `location:"uri" locationName:"Bucket" type:"string" required:"true"`

	// The name of the object.
	Key *string `location:"uri" locationName:"Key" type:"string" required:"true"`

	// The starting position of the AppendObject operation.
	// When the AppendObject operation is successful, the x-kss-next-append-position header describes the starting position of the next operation.
	Position *int64 `location:"querystring" locationName:"position" type:"integer"`

	// The readable body payload to send to KS3.
	Body io.Reader `type:"blob"`

	// The canned ACL to apply to the object.
	ACL *string `location:"header" locationName:"x-amz-acl" type:"string"`

	// The type of storage to use for the object. Defaults to 'STANDARD'.
	StorageClass *string `location:"header" locationName:"x-amz-storage-class" type:"string"`

	metadataAppendObjectInput `json:"-" xml:"-"`
}

type metadataAppendObjectInput struct {
	SDKShapeTraits bool `type:"structure" payload:"Body"`
}

type AppendObjectOutput struct {
	// Entity tag for the uploaded object.
	ETag *string `location:"header" locationName:"ETag" type:"string"`

	// If the object expiration is configured, this will contain the expiration
	// date (expiry-date) and rule ID (rule-id). The value of rule-id is URL encoded.
	Expiration *string `location:"header" locationName:"x-amz-expiration" type:"string"`

	// If present, indicates that the requester was successfully charged for the
	// request.
	RequestCharged *string `location:"header" locationName:"x-amz-request-charged" type:"string"`

	// If server-side encryption with a customer-provided encryption key was requested,
	// the response will include this header confirming the encryption algorithm
	// used.
	SSECustomerAlgorithm *string `location:"header" locationName:"x-amz-server-side-encryption-customer-algorithm" type:"string"`

	// If server-side encryption with a customer-provided encryption key was requested,
	// the response will include this header to provide round trip message integrity
	// verification of the customer-provided encryption key.
	SSECustomerKeyMD5 *string `location:"header" locationName:"x-amz-server-side-encryption-customer-key-MD5" type:"string"`

	// If present, specifies the ID of the AWS Key Management Service (KMS) master
	// encryption key that was used for the object.
	SSEKMSKeyID *string `location:"header" locationName:"x-amz-server-side-encryption-aws-kms-key-id" type:"string"`

	// The Server-side encryption algorithm used when storing this object in S3
	// (e.g., AES256, aws:kms).
	ServerSideEncryption *string `location:"header" locationName:"x-amz-server-side-encryption" type:"string"`

	// Version of the object.
	VersionID *string `location:"header" locationName:"x-amz-version-id" type:"string"`

	NewFileName *string `location:"header" locationName:"newfilename" type:"string"`

	metadataPutObjectOutput `json:"-" xml:"-"`

	Metadata map[string]*string `location:"headers"  type:"map"`

	StatusCode *int64 `location:"statusCode" type:"integer"`
}

func (c *S3) AppendObjectRequest(input *AppendObjectInput) (req *aws.Request, output *AppendObjectOutput) {
	oprw.Lock()
	defer oprw.Unlock()

	if opPutObject == nil {
		opPutObject = &aws.Operation{
			Name:       "AppendObject",
			HTTPMethod: "POST",
			HTTPPath:   "/{Bucket}/{Key+}?append",
		}
	}

	if input == nil {
		input = &AppendObjectInput{}
	}

	req = c.newRequest(opPutObject, input, output)
	output = &AppendObjectOutput{}
	req.Data = output
	return
}

func (c *S3) AppendObject(input *AppendObjectInput) (*AppendObjectOutput, error) {
	req, out := c.AppendObjectRequest(input)
	err := req.Send()
	return out, err
}

func (c *S3) AppendObjectWithContext(ctx aws.Context, input *AppendObjectInput) (*AppendObjectOutput, error) {
	req, out := c.AppendObjectRequest(input)
	req.SetContext(ctx)
	err := req.Send()
	return out, err
}
