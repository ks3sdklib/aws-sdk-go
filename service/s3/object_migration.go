package s3

import (
	"github.com/ks3sdklib/aws-sdk-go/aws"
	"time"
)

type PutObjectMigrationInput struct {
	// 目标桶名。
	Bucket *string `location:"uri" locationName:"Bucket" type:"string" required:"true"`

	// 目标对象的Key。
	Key *string `location:"uri" locationName:"Key" type:"string" required:"true"`

	// 源桶名。
	SourceBucket *string `location:"uri" locationName:"SourceBucket" type:"string" required:"true"`

	// 源对象的Key。
	SourceKey *string `location:"uri" locationName:"SourceKey" type:"string" required:"true"`

	// 迁移的源文件路径，无需指定，由SourceBucket与SourceKey自动生成。
	MigrationSource *string `location:"header" locationName:"x-amz-migration-source" type:"string"`

	// 若源文件为极速类型文件，目标文件必须设置为非极速类型文件，若源文件为非极速类型文件，目标文件必须设置为极速类型文件。
	// 若目标桶为非极速类型桶，不支持将目标文件设置为极速类型。不设置时默认与存储桶一致。
	StorageClass *string `location:"header" locationName:"x-amz-storage-class" type:"string"`

	// KS3解密时对数据源对象使用的解密算法。可选值：AES256、SM4。如果源对象使用客户提供的密钥加密，则需要提供。
	SourceSSECustomerAlgorithm *string `location:"header" locationName:"x-amz-migration-source-server-side-encryption-customer-algorithm" type:"string"`

	// KS3解密时使用的Base64编码后加密秘钥，其值必须与源Object创建时使用的秘钥一致。如果源对象使用客户提供的密钥加密，则需要提供。
	SourceSSECustomerKey *string `location:"header" locationName:"x-amz-migration-source-server-side-encryption-customer-key" type:"string"`

	// KS3解密时使用的对加密秘钥Base64编码后的MD5值，如果源对象使用客户提供的密钥加密，则需要提供。
	SourceSSECustomerKeyMD5 *string `location:"header" locationName:"x-amz-migration-source-server-side-encryption-customer-key-MD5" type:"string"`

	// 客户端提供的加密算法。可选值：AES256、SM4。如果目标对象使用客户提供密钥加密，则需要提供。
	SSECustomerAlgorithm *string `location:"header" locationName:"x-amz-server-side-encryption-customer-algorithm" type:"string"`

	// 客户端提供的Base64编码后加密秘钥。如果目标对象使用客户提供密钥加密，则需要提供。
	SSECustomerKey *string `location:"header" locationName:"x-amz-server-side-encryption-customer-key" type:"string"`

	// 客户端提供的对加密秘钥Base64编码后的MD5值。如果目标对象使用客户提供密钥加密，则需要提供。
	SSECustomerKeyMD5 *string `location:"header" locationName:"x-amz-server-side-encryption-customer-key-MD5" type:"string"`

	// 如果目标对象使用KS3托管的服务端加密，则需要提供，服务端将对数据进行加密处理。可选值：AES256、SM4。
	ServerSideEncryption *string `location:"header" locationName:"x-amz-server-side-encryption" type:"string"`

	// 设置扩展请求头。如果现有字段不支持设置所需的请求头，您可以通过此字段进行设置。
	ExtendHeaders map[string]*string `location:"extendHeaders" type:"map"`

	// 设置扩展查询参数。如果现有字段不支持设置所需的查询参数，您可以通过此字段进行设置。
	ExtendQueryParams map[string]*string `location:"extendQueryParams" type:"map"`
}

type PutObjectMigrationOutput struct {
	// http响应头。
	Metadata map[string]*string `location:"headers" type:"map"`

	// http响应状态码。
	StatusCode *int64 `location:"statusCode" type:"integer"`
}

// PutObjectMigrationRequest generates a request for the PutObjectMigration operation.
func (c *S3) PutObjectMigrationRequest(input *PutObjectMigrationInput) (req *aws.Request, output *PutObjectMigrationOutput) {
	op := &aws.Operation{
		Name:       "PutObjectMigration",
		HTTPMethod: "PUT",
		HTTPPath:   "/{Bucket}/{Key+}?migration",
	}

	if input == nil {
		input = &PutObjectMigrationInput{}
	}

	if input.MigrationSource == nil {
		input.MigrationSource = aws.String(BuildCopySource(input.SourceBucket, input.SourceKey))
	}
	req = c.newRequest(op, input, output)
	output = &PutObjectMigrationOutput{}
	req.Data = output
	return
}

// PutObjectMigration sets object migration task.
func (c *S3) PutObjectMigration(input *PutObjectMigrationInput) (*PutObjectMigrationOutput, error) {
	req, out := c.PutObjectMigrationRequest(input)
	err := req.Send()
	return out, err
}

func (c *S3) PutObjectMigrationWithContext(ctx aws.Context, input *PutObjectMigrationInput) (*PutObjectMigrationOutput, error) {
	req, out := c.PutObjectMigrationRequest(input)
	req.SetContext(ctx)
	err := req.Send()
	return out, err
}

type GetObjectMigrationInput struct {
	// 指定存储桶名。
	Bucket *string `location:"uri" locationName:"Bucket" type:"string" required:"true"`

	// 指定对象的Key。
	Key *string `location:"uri" locationName:"Key" type:"string" required:"true"`

	// 设置扩展请求头。如果现有字段不支持设置所需的请求头，您可以通过此字段进行设置。
	ExtendHeaders map[string]*string `location:"extendHeaders" type:"map"`

	// 设置扩展查询参数。如果现有字段不支持设置所需的查询参数，您可以通过此字段进行设置。
	ExtendQueryParams map[string]*string `location:"extendQueryParams" type:"map"`
}

type GetObjectMigrationOutput struct {
	// 存放单个迁移任务参数的容器。
	MigrationConfiguration *MigrationConfiguration `locationName:"Migration" type:"structure"`

	// http响应头。
	Metadata map[string]*string `location:"headers" type:"map"`

	// http响应状态码。
	StatusCode *int64 `location:"statusCode" type:"integer"`

	metadataGetObjectMigrationOutput `json:"-" xml:"-"`
}

type metadataGetObjectMigrationOutput struct {
	SDKShapeTraits bool `type:"structure" payload:"MigrationConfiguration"`
}

type MigrationConfiguration struct {
	// 任务状态。
	Status *string `locationName:"Status" type:"string"`

	// 包含迁移失败原因的容器，只有任务状态为Failed才会返回该参数。
	MigrationFailure *MigrationFailure `locationName:"MigrationFailure" type:"structure"`

	// 包含操作的容器。
	Operation *MigrationOperation `locationName:"Operation" type:"structure"`

	// 任务创建时间。采用ISO 8601日期和时间表示法，示例：2024-08-17T17:04:52Z，加8小时表示中国北京时间。
	CreationTime *time.Time `locationName:"CreationTime" type:"timestamp" timestampFormat:"iso8601"`

	// 任务终止的时间。采用ISO 8601日期和时间表示法，示例：2024-08-17T17:04:52Z，加8小时表示中国北京时间。
	TerminationTime *time.Time `locationName:"TerminationTime" type:"timestamp" timestampFormat:"iso8601"`
}

type MigrationFailure struct {
	// 失败的响应码，只有任务状态为Failed才会返回该参数。
	Code *string `locationName:"Code" type:"string"`

	// 失败的详细信息，只有任务状态为Failed才会返回该参数。
	Reason *string `locationName:"Reason" type:"string"`
}

type MigrationOperation struct {
	// 源文件路径。
	MigrationSource *string `locationName:"MigrationSource" type:"string"`

	// 目标文件路径。
	MigrationDest *string `locationName:"MigrationDest" type:"string"`

	// 设置的转换存储类型信息。
	StorageClass *string `locationName:"StorageClass" type:"string"`
}

// GetObjectMigrationRequest generates a request for the GetObjectMigration operation.
func (c *S3) GetObjectMigrationRequest(input *GetObjectMigrationInput) (req *aws.Request, output *GetObjectMigrationOutput) {
	op := &aws.Operation{
		Name:       "GetObjectMigration",
		HTTPMethod: "GET",
		HTTPPath:   "/{Bucket}/{Key+}?migration",
	}

	if input == nil {
		input = &GetObjectMigrationInput{}
	}

	req = c.newRequest(op, input, output)
	output = &GetObjectMigrationOutput{}
	req.Data = output
	return
}

// GetObjectMigration gets object migration task.
func (c *S3) GetObjectMigration(input *GetObjectMigrationInput) (*GetObjectMigrationOutput, error) {
	req, out := c.GetObjectMigrationRequest(input)
	err := req.Send()
	return out, err
}

func (c *S3) GetObjectMigrationWithContext(ctx aws.Context, input *GetObjectMigrationInput) (*GetObjectMigrationOutput, error) {
	req, out := c.GetObjectMigrationRequest(input)
	req.SetContext(ctx)
	err := req.Send()
	return out, err
}
