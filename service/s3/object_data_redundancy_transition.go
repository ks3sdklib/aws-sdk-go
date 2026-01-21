package s3

import "github.com/ks3sdklib/aws-sdk-go/aws"

type PutObjectDataRedundancyTransitionInput struct {
	// 目标桶名。
	Bucket *string `location:"uri" locationName:"Bucket" type:"string" required:"true"`

	// 目标对象的Key。
	Key *string `location:"uri" locationName:"Key" type:"string" required:"true"`

	// 文件的冗余类型。
	DataRedundancyType *string `location:"header" locationName:"x-amz-data-redundancy-type" type:"string"`

	// 设置扩展请求头。如果现有字段不支持设置所需的请求头，您可以通过此字段进行设置。
	ExtendHeaders map[string]*string `location:"extendHeaders" type:"map"`

	// 设置扩展查询参数。如果现有字段不支持设置所需的查询参数，您可以通过此字段进行设置。
	ExtendQueryParams map[string]*string `location:"extendQueryParams" type:"map"`
}

type PutObjectDataRedundancyTransitionOutput struct {
	// http响应头。
	Metadata map[string]*string `location:"headers" type:"map"`

	// http响应状态码。
	StatusCode *int64 `location:"statusCode" type:"integer"`
}

// PutObjectDataRedundancyTransitionRequest generates a request for the PutObjectDataRedundancyTransition operation.
func (c *S3) PutObjectDataRedundancyTransitionRequest(input *PutObjectDataRedundancyTransitionInput) (req *aws.Request, output *PutObjectDataRedundancyTransitionOutput) {
	op := &aws.Operation{
		Name:       "PutObjectDataRedundancyTransition",
		HTTPMethod: "PUT",
		HTTPPath:   "/{Bucket}/{Key+}?dataRedundancyTransition",
	}

	if input == nil {
		input = &PutObjectDataRedundancyTransitionInput{}
	}

	if IsEmpty(input.DataRedundancyType) {
		input.DataRedundancyType = aws.String(DataRedundancyTypeZRS)
	}

	req = c.newRequest(op, input, output)
	output = &PutObjectDataRedundancyTransitionOutput{}
	req.Data = output
	return
}

// PutObjectDataRedundancyTransition sets object data redundancy transition.
func (c *S3) PutObjectDataRedundancyTransition(input *PutObjectDataRedundancyTransitionInput) (*PutObjectDataRedundancyTransitionOutput, error) {
	req, out := c.PutObjectDataRedundancyTransitionRequest(input)
	err := req.Send()
	return out, err
}

func (c *S3) PutObjectDataRedundancyTransitionWithContext(ctx aws.Context, input *PutObjectDataRedundancyTransitionInput) (*PutObjectDataRedundancyTransitionOutput, error) {
	req, out := c.PutObjectDataRedundancyTransitionRequest(input)
	req.SetContext(ctx)
	err := req.Send()
	return out, err
}
