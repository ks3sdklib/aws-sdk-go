package s3

import (
	"github.com/ks3sdklib/aws-sdk-go/aws"
)

type PutBucketNotificationInput struct {
	// 存储桶名称。
	Bucket *string `location:"uri" locationName:"Bucket" type:"string" required:"true"`

	// 存储桶事件通知规则的容器。
	BucketNotification *BucketNotification `locationName:"Notification" type:"structure" required:"true"`

	// 设置扩展请求头。如果现有字段不支持设置所需的请求头，您可以通过此字段进行设置。
	ExtendHeaders map[string]*string `location:"extendHeaders" type:"map"`

	// 设置扩展查询参数。如果现有字段不支持设置所需的查询参数，您可以通过此字段进行设置。
	ExtendQueryParams map[string]*string `location:"extendQueryParams" type:"map"`

	metadataPutBucketNotificationInput `json:"-" xml:"-"`
}

type metadataPutBucketNotificationInput struct {
	SDKShapeTraits bool `type:"structure" payload:"BucketNotification"`
}

type BucketNotification struct {
	// 事件通知规则信息。
	// 每个Bucket最多可同时配置10个事件通知规则。
	Notifications []*Notification `locationName:"Notifications" type:"list"`
}

type Notification struct {
	// 设置事件通知规则名称（单个UID内唯一），命名规范如下：
	// 1.单个UID下创建的规则名称不能重复，重复设置就会覆盖。
	// 2.长度不能超过32个字符。
	// 3.必须以小写字母开头 ([a-z])，后面可以跟字母、数字或下划线 ([a-zA-Z0-9_]{0,31})。
	RuleId *string `locationName:"RuleId" type:"string"`

	// 事件通知内容推送至客户端地址的方式，取值如下：
	// POST（默认）
	// PUT
	Method *string `locationName:"Method" type:"string"`

	// 设置需要触发消息通知的事件类型，单个规则支持选择多个事件类型。
	Events []string `locationName:"Events" type:"list"`

	// 设置需要触发事件通知的对象过滤规则。支持设置前缀（Prefix）、后缀（Suffix），即符合前后缀要求的文件才会触发事件通知规则。
	// 1.如果同时设置了前缀与后缀规则，则事件需要同时满足二者，才会触发事件通知。
	// 2.如果前后缀均未设置，则会匹配存储桶内所有对象。
	// 3.单个规则（Rule）的路径不允许有重叠。
	// 4.单条规则（Rule）支持最多同时设置5个触发路径。
	// 5.两个规则（Rule）如果事件类型一致，则前缀或后缀不允许存在重叠。
	Resources []*NotificationResource `locationName:"Resources" type:"list"`

	// 包含回调地址信息的容器。
	// 单个规则（Rule）内，最多支持填写5个回调地址。
	Destinations []*NotificationDestination `locationName:"Destinations" type:"list"`

	// 导出推送失败列表报告。
	// 1.当触发事件通知规则时，KS3会将事件通知内容推送至客户回调地址内，推送成功后KS3接口会返回响应头x-kss-eventBridge-status ，该响应头的Value值为Base64编码，解码后的内容为Code: Success。
	// 2.当触发事件通知规则时，由于网络抖动或其他异常场景导致事件通知内容推送至客户回调地址失败，KS3接口会返回响应头x-kss-eventBridge-status ，该响应头的Value值为Base64编码，解码后的内容为Code: Fail。
	// 3.针对异常场景导致事件通知内容推送至客户回调地址失败，KS3会将该通知内容以失败列表的方式导出并每天定时投递至客户的指定桶内。
	Report *NotificationReport `locationName:"Report" type:"structure"`
}

type NotificationResource struct {
	// 设置符合规则的对象前缀。取值：0-1024字节
	// 1.当不填写Prefix和Suffix时，表示对整个桶内的文件均设置事件通知规则。
	// 2.如果要匹配Bucket内下名称为examplefolder目录中的全部对象，则前缀填写为examplefolder/，后缀（Suffix）置空即可。
	// 3.两个规则（Rule）如果事件类型一致，则前缀或后缀不允许存在重叠。
	Prefix *string `locationName:"Prefix" type:"string"`

	// 设置符合规则的对象后缀。取值：0-1024字节
	// 1.如果要匹配Bucket内所有名称以.jpg结尾的对象，则前缀（Prefix）置空，后缀（Suffix）填写为.jpg即可。
	// 2.两个规则（Rule）如果事件类型一致，则前缀或后缀不允许存在重叠。
	Suffix *string `locationName:"Suffix" type:"string"`
}

type NotificationDestination struct {
	// 回调地址信息的类型，固定取值：EndPoint。
	DestType *string `locationName:"DestType" type:"string"`

	// 回调地址。当触发事件通知时，KS3会以回调方式将事件通知消息体以JSON格式推送至该地址。
	// 1.支持HTTP或HTTPS。
	// 2.支持填写IP+端口号。示例：http://198.51.100.1:8080
	// 3.支持填写域名地址。示例：http://test.com
	// 4.支持填写带参数地址。示例：https://ip:port/oss/sync/{id}?ak=xxxx
	// 5.仅支持公网推送，不支持内网推送。
	CallbackUrl *string `locationName:"CallbackUrl" type:"string"`
}

type NotificationReport struct {
	// 推送的报告类型，仅支持导出推送失败的列表报告，固定取值：failed。
	ReportType *string `locationName:"ReportType" type:"string"`

	// 是否导出并投递失败列表报告，固定取值：true。
	Enabled *bool `locationName:"Enabled" type:"boolean"`

	// 推送失败列表报告投递的桶名称。
	Bucket *string `locationName:"Bucket" type:"string"`

	// 推送失败列表报告投递的目录名称。取值范围：0-1024字节
	// 如果Prefix不存在，KS3将自动创建该名称的Prefix。
	Prefix *string `locationName:"Prefix" type:"string"`
}

type PutBucketNotificationOutput struct {
	// http响应头。
	Metadata map[string]*string `location:"headers" type:"map"`

	// http响应状态码。
	StatusCode *int64 `location:"statusCode" type:"integer"`
}

// PutBucketNotificationRequest generates a request for the PutBucketNotification operation.
func (c *S3) PutBucketNotificationRequest(input *PutBucketNotificationInput) (req *aws.Request, output *PutBucketNotificationOutput) {
	op := &aws.Operation{
		Name:       "PutBucketNotification",
		HTTPMethod: "PUT",
		HTTPPath:   "/{Bucket}?notification",
	}

	if input == nil {
		input = &PutBucketNotificationInput{}
	}

	req = c.newRequest(op, input, output)
	req.ContentType = "application/json"
	output = &PutBucketNotificationOutput{}
	req.Data = output
	return
}

// PutBucketNotification sets bucket notification configuration.
func (c *S3) PutBucketNotification(input *PutBucketNotificationInput) (*PutBucketNotificationOutput, error) {
	req, out := c.PutBucketNotificationRequest(input)
	err := req.Send()
	return out, err
}

func (c *S3) PutBucketNotificationWithContext(ctx aws.Context, input *PutBucketNotificationInput) (*PutBucketNotificationOutput, error) {
	req, out := c.PutBucketNotificationRequest(input)
	req.SetContext(ctx)
	err := req.Send()
	return out, err
}

type GetBucketNotificationInput struct {
	// 存储桶名称。
	Bucket *string `location:"uri" locationName:"Bucket" type:"string" required:"true"`

	// 设置扩展请求头。如果现有字段不支持设置所需的请求头，您可以通过此字段进行设置。
	ExtendHeaders map[string]*string `location:"extendHeaders" type:"map"`

	// 设置扩展查询参数。如果现有字段不支持设置所需的查询参数，您可以通过此字段进行设置。
	ExtendQueryParams map[string]*string `location:"extendQueryParams" type:"map"`
}

type GetBucketNotificationOutput struct {
	// 存储桶事件通知规则的容器。
	BucketNotification *BucketNotification `locationName:"Notification" type:"structure"`

	// http响应头。
	Metadata map[string]*string `location:"headers" type:"map"`

	// http响应状态码。
	StatusCode *int64 `location:"statusCode" type:"integer"`

	metadataGetBucketNotificationOutput `json:"-" xml:"-"`
}

type metadataGetBucketNotificationOutput struct {
	SDKShapeTraits bool `type:"structure" payload:"BucketNotification"`
}

// GetBucketNotificationRequest generates a request for the GetBucketNotification operation.
func (c *S3) GetBucketNotificationRequest(input *GetBucketNotificationInput) (req *aws.Request, output *GetBucketNotificationOutput) {
	op := &aws.Operation{
		Name:       "GetBucketNotification",
		HTTPMethod: "GET",
		HTTPPath:   "/{Bucket}?notification",
	}

	if input == nil {
		input = &GetBucketNotificationInput{}
	}

	req = c.newRequest(op, input, output)
	req.ContentType = "application/json"
	output = &GetBucketNotificationOutput{
		BucketNotification: &BucketNotification{},
	}
	req.Data = output
	return
}

// GetBucketNotification gets bucket notification configuration.
func (c *S3) GetBucketNotification(input *GetBucketNotificationInput) (*GetBucketNotificationOutput, error) {
	req, out := c.GetBucketNotificationRequest(input)
	err := req.Send()
	return out, err
}

func (c *S3) GetBucketNotificationWithContext(ctx aws.Context, input *GetBucketNotificationInput) (*GetBucketNotificationOutput, error) {
	req, out := c.GetBucketNotificationRequest(input)
	req.SetContext(ctx)
	err := req.Send()
	return out, err
}

type DeleteBucketNotificationInput struct {
	// 存储桶名称。
	Bucket *string `location:"uri" locationName:"Bucket" type:"string" required:"true"`

	// 设置扩展请求头。如果现有字段不支持设置所需的请求头，您可以通过此字段进行设置。
	ExtendHeaders map[string]*string `location:"extendHeaders" type:"map"`

	// 设置扩展查询参数。如果现有字段不支持设置所需的查询参数，您可以通过此字段进行设置。
	ExtendQueryParams map[string]*string `location:"extendQueryParams" type:"map"`
}

type DeleteBucketNotificationOutput struct {
	// http响应头。
	Metadata map[string]*string `location:"headers" type:"map"`

	// http响应状态码。
	StatusCode *int64 `location:"statusCode" type:"integer"`
}

// DeleteBucketNotificationRequest generates a request for the DeleteBucketNotification operation.
func (c *S3) DeleteBucketNotificationRequest(input *DeleteBucketNotificationInput) (req *aws.Request, output *DeleteBucketNotificationOutput) {
	op := &aws.Operation{
		Name:       "DeleteBucketNotification",
		HTTPMethod: "DELETE",
		HTTPPath:   "/{Bucket}?notification",
	}

	if input == nil {
		input = &DeleteBucketNotificationInput{}
	}

	req = c.newRequest(op, input, output)
	output = &DeleteBucketNotificationOutput{}
	req.Data = output
	return
}

// DeleteBucketNotification deletes bucket notification configuration.
func (c *S3) DeleteBucketNotification(input *DeleteBucketNotificationInput) (*DeleteBucketNotificationOutput, error) {
	req, out := c.DeleteBucketNotificationRequest(input)
	err := req.Send()
	return out, err
}

func (c *S3) DeleteBucketNotificationWithContext(ctx aws.Context, input *DeleteBucketNotificationInput) (*DeleteBucketNotificationOutput, error) {
	req, out := c.DeleteBucketNotificationRequest(input)
	req.SetContext(ctx)
	err := req.Send()
	return out, err
}
