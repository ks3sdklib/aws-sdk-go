package s3

import (
	"github.com/google/uuid"
	"github.com/ks3sdklib/aws-sdk-go/aws"
)

type CreateJobInput struct {
	// 批量处理配置规则的容器。
	CreateJobRequest *CreateJobRequest `locationName:"CreateJobRequest" type:"structure" required:"true"`

	// 设置扩展请求头。如果现有字段不支持设置所需的请求头，您可以通过此字段进行设置。
	ExtendHeaders map[string]*string `location:"extendHeaders" type:"map"`

	// 设置扩展查询参数。如果现有字段不支持设置所需的查询参数，您可以通过此字段进行设置。
	ExtendQueryParams map[string]*string `location:"extendQueryParams" type:"map"`

	metadataCreateJobInput `json:"-" xml:"-"`
}

type metadataCreateJobInput struct {
	SDKShapeTraits bool `type:"structure" payload:"CreateJobRequest"`

	AutoFillMD5 bool
}

type CreateJobRequest struct {
	// 任务描述。描述长度范围为0 - 256字节。
	// 仅支持字母、数字、中文、下划线（_）和横线（-）的组合。
	Description *string `locationName:"Description" type:"string"`

	// 待处理的文件信息。
	Manifest *JobManifest `locationName:"Manifest" type:"structure" required:"true"`

	// 选择要执行的具体操作。支持批量解冻、批量修改ACL、批量删除操作。
	// 单个任务中Operation只能设置一种操作类型（批量解冻/批量修改ACL/批量删除）。
	Operation *JobOperation `locationName:"Operation" type:"structure" required:"true"`

	// 任务优先级。取值越大表示任务执行的优先级越高。
	// 取值范围：0-2147483647
	Priority *int64 `locationName:"Priority" type:"integer" required:"true"`

	// 任务完成报告。仅支持导出操作失败列表报告。
	Report *JobReport `locationName:"Report" type:"structure" required:"true"`

	// 每个请求唯一的 token，用于避免前端重复发起同一批处理任务。长度为1 - 64字节，建议使用uuid。
	// 1. 仅支持数字、字母、横线（-）。
	// 2. 同一个UID下的ClientRequestToken必须唯一。
	// 3. 创建任务时设置的ClientRequestToken不能与近48小时内已删除的规则ClientRequestToken重复。
	ClientRequestToken *string `locationName:"ClientRequestToken" type:"string" required:"true"`
}

type JobOperation struct {
	// 批量设置ACL的具体参数。
	KS3PutObjectAcl *KS3PutObjectAcl `locationName:"KS3PutObjectAcl" type:"structure"`

	// 对归档类型文件批量执行解冻操作的具体参数。
	KS3RestoreObject *KS3RestoreObject `locationName:"KS3RestoreObject" type:"structure"`

	// 对文件批量执行删除操作的具体参数。对文件批量执行删除操作时，该参数取值设置为空即可。
	KS3DeleteObject *KS3DeleteObject `locationName:"KS3DeleteObject" useEmpty:"true" type:"structure"`
}

type KS3RestoreObject struct {
	// 表示需要解冻的存储类型。
	StorageClass *string `locationName:"StorageClass" type:"string"`

	// 设置解冻优先级
	Tier *string `locationName:"Tier" type:"string"`

	// 设置解冻持续时间。
	Days *int64 `locationName:"Days" type:"integer"`
}

type KS3PutObjectAcl struct {
	// 预定义ACL，针对所有用户生效。
	// 取值：private、public-read
	// 1. 设置为private表示只有文件的拥有者可以对该文件进行读写操作，其他人无法访问该文件。
	// 2. 设置为public-read表示任何人（包括匿名访问者）都可以对该文件进行读操作。
	CannedAccessControlList *string `locationName:"CannedAccessControlList" type:"string"`

	// 针对指定用户设置ACL权限。
	AccessControlList *JobAccessControlList `locationName:"AccessControlList" type:"structure"`
}

type JobAccessControlList struct {
	// 包含被授权者和其ACL信息。
	// 单条规则最多支持设置100条Grant。
	Grants []*JobGrant `locationName:"Grant" type:"list" flattened:"true"`
}

type JobGrant struct {
	// 被授权者的账号（UID）信息。
	// 1. 单个Grantee参数仅支持传入一个账号信息。
	// 2. 仅支持针对UID授予ACL权限。
	Grantee *string `locationName:"Grantee" type:"string"`

	// 指明授予被授权者的权限信息。
	// 取值：FULL_CONTROL、READ
	// 1. FULL_CONTROL表示被授权者具有对文件的读写权限。
	// 2. READ表示被授权者具有对文件的只读权限。
	Permission *string `locationName:"Permission" type:"string"`
}

type KS3DeleteObject struct {
}

type JobManifest struct {
	// 待处理的文件位置信息。
	Location *ManifestLocation `locationName:"Location" type:"structure"`

	// 描述待处理文件列表的格式信息。
	Spec *ManifestSpec `locationName:"Spec" type:"structure"`
}

type ManifestLocation struct {
	// 指定需要进行批量操作的桶或前缀。
	// 单个规则可以同时设置多个Filter来实现对不同桶内的不同前缀进行批量处理。单个规则最多支持设置100个Filter。
	Filters []*LocationFilter `locationName:"Filter" type:"list" flattened:"true"`
}

type LocationFilter struct {
	// 指定需要进行批量操作桶的资源标识符。
	// 格式：krn:ksc:ks3:::bucketname
	// 单个Filter内仅支持填写一个桶名称，如果需要同时对多个桶设置批量处理规则，可通过设置多个Filter实现。
	Bucket *string `locationName:"Bucket" type:"string"`

	// 指定需要进行批量操作的前缀。
	// 1. 单个Filter内填写的Prefix不支持重叠。
	// 2. 单个Filter内支持填写多个前缀，表示对桶内多个指定前缀的文件进行批量处理。
	// 3. Prefix参数值设置为空，表示对桶内的全部文件进行批量操作。
	// 4. 单个Filter内最多支持填写1000个Prefix。
	Prefixes []string `locationName:"Prefix" type:"list" flattened:"true"`
}

type ManifestSpec struct {
	// 指定待处理文件列表的格式信息。固定取值：KS3BatchOperations_Bucket_V1。
	Format *string `locationName:"Format" type:"string"`
}

type JobReport struct {
	// 任务完成报告的投递存储桶。
	Bucket *string `locationName:"Bucket" type:"string"`

	// 任务完成报告投递的前缀信息。
	Prefix *string `locationName:"Prefix" type:"string"`

	// 任务完成报告内容类型。固定取值：FailedTasksOnly。
	ReportScope *string `locationName:"ReportScope" type:"string"`

	// 是否输出任务完成报告。
	// 取值为true表示输出任务完成报告，取值为false表示不输出任务完成报告。
	Enabled *bool `locationName:"Enabled" type:"boolean"`
}

type CreateJobOutput struct {
	// 包含任务ID的容器。
	CreateJobResult *CreateJobResult `locationName:"CreateJobResult" type:"structure"`

	// http响应头。
	Metadata map[string]*string `location:"headers" type:"map"`

	// http响应状态码。
	StatusCode *int64 `location:"statusCode" type:"integer"`

	metadataCreateJobOutput `json:"-" xml:"-"`
}

type metadataCreateJobOutput struct {
	SDKShapeTraits bool `type:"structure" payload:"CreateJobResult"`
}

type CreateJobResult struct {
	// 任务ID。规则创建成功后，KS3会自动返回该参数，每个规则对应唯一的任务ID。
	// JobId的值与请求参数ClientRequestToken值一致。
	JobId *string `locationName:"JobId" type:"string"`
}

// CreateJobRequest generates a request for the CreateJob operation.
func (c *S3) CreateJobRequest(input *CreateJobInput) (req *aws.Request, output *CreateJobOutput) {
	op := &aws.Operation{
		Name:       "CreateJob",
		HTTPMethod: "PUT",
		HTTPPath:   "/?jobs",
	}

	if input == nil {
		input = &CreateJobInput{}
	}

	if IsEmpty(input.CreateJobRequest.ClientRequestToken) {
		input.CreateJobRequest.ClientRequestToken = aws.String(uuid.NewString())
	}

	input.AutoFillMD5 = true
	req = c.newRequest(op, input, output)
	output = &CreateJobOutput{}
	req.Data = output
	return
}

// CreateJob creates a batch job.
func (c *S3) CreateJob(input *CreateJobInput) (*CreateJobOutput, error) {
	req, out := c.CreateJobRequest(input)
	err := req.Send()
	return out, err
}

func (c *S3) CreateJobWithContext(ctx aws.Context, input *CreateJobInput) (*CreateJobOutput, error) {
	req, out := c.CreateJobRequest(input)
	req.SetContext(ctx)
	err := req.Send()
	return out, err
}

type DescribeJobInput struct {
	// 任务ID。个任务对应唯一的任务ID，创建批量处理任务（CreateJob）成功后，KS3会返回任务ID。
	JobId *string `location:"querystring" locationName:"jobId" type:"string" required:"true"`

	// 设置扩展请求头。如果现有字段不支持设置所需的请求头，您可以通过此字段进行设置。
	ExtendHeaders map[string]*string `location:"extendHeaders" type:"map"`

	// 设置扩展查询参数。如果现有字段不支持设置所需的查询参数，您可以通过此字段进行设置。
	ExtendQueryParams map[string]*string `location:"extendQueryParams" type:"map"`
}

type DescribeJobOutput struct {
	// 包含批量处理规则信息的容器。
	DescribeJobResult *DescribeJobResult `locationName:"DescribeJobResult" type:"structure"`

	// http响应头。
	Metadata map[string]*string `location:"headers" type:"map"`

	// http响应状态码。
	StatusCode *int64 `location:"statusCode" type:"integer"`

	metadataDescribeJobOutput `json:"-" xml:"-"`
}

type metadataDescribeJobOutput struct {
	SDKShapeTraits bool `type:"structure" payload:"DescribeJobResult"`
}

type DescribeJobResult struct {
	// 任务ID。规则创建成功后，KS3会自动返回该参数，每个规则对应唯一的任务ID。
	JobId *string `locationName:"JobId" type:"string"`

	// 任务创建时间。
	CreationTime *string `locationName:"CreationTime" type:"timestamp" timestampFormat:"iso8601"`

	// 任务描述。
	Description *string `locationName:"Description" type:"string"`

	// 任务执行的当前状态。合法参数值包括：New、Active、Complete。
	// New：批量处理任务刚被创建，任务正在解析中。
	// Active：批量操作任务进行中。
	// Complete：批量操作任务已完成，处于最终状态。
	Status *string `locationName:"Status" type:"string"`

	// 待处理的文件信息。
	Manifest *JobManifest `locationName:"Manifest" type:"structure"`

	// 具体操作。支持批量解冻、批量修改ACL、批量删除操作。
	Operation *JobOperation `locationName:"Operation" type:"structure"`

	// 任务优先级。取值越大表示任务执行的优先级越高。
	Priority *int64 `locationName:"Priority" type:"integer"`

	// 任务执行状况概述。描述该批量处理任务中所执行的操作总数，成功的操作数量以及失败的操作数量。
	ProgressSummary *JobProgressSummary `locationName:"ProgressSummary" type:"structure"`

	// 任务完成报告。仅支持导出操作失败列表报告。
	Report *JobReport `locationName:"Report" type:"structure"`

	// 任务终止的时间。
	TerminationDate *string `locationName:"TerminationDate" type:"timestamp" timestampFormat:"iso8601"`
}

type JobProgressSummary struct {
	// 当前失败的操作数。
	NumberOfTasksFailed *int64 `locationName:"NumberOfTasksFailed" type:"integer"`

	// 当前成功的操作数。
	NumberOfTasksSucceeded *int64 `locationName:"NumberOfTasksSucceeded" type:"integer"`

	// 总操作数。
	TotalNumberOfTasks *int64 `locationName:"TotalNumberOfTasks" type:"integer"`
}

// DescribeJobRequest generates a request for the DescribeJob operation.
func (c *S3) DescribeJobRequest(input *DescribeJobInput) (req *aws.Request, output *DescribeJobOutput) {
	op := &aws.Operation{
		Name:       "DescribeJob",
		HTTPMethod: "GET",
		HTTPPath:   "/?jobs",
	}

	if input == nil {
		input = &DescribeJobInput{}
	}

	req = c.newRequest(op, input, output)
	output = &DescribeJobOutput{
		DescribeJobResult: &DescribeJobResult{},
	}
	req.Data = output
	return
}

// DescribeJob gets a batch job's details.
func (c *S3) DescribeJob(input *DescribeJobInput) (*DescribeJobOutput, error) {
	req, out := c.DescribeJobRequest(input)
	err := req.Send()
	return out, err
}

func (c *S3) DescribeJobWithContext(ctx aws.Context, input *DescribeJobInput) (*DescribeJobOutput, error) {
	req, out := c.DescribeJobRequest(input)
	req.SetContext(ctx)
	err := req.Send()
	return out, err
}

type ListJobsInput struct {
	// 返回任务数量最大值。可配合nextToken参数实现分页返回。
	// 取值范围：1-1000
	// 默认值：1000
	// 1. 如果配置了该参数，单次返回的任务数量最多不会超过该值，如果没有配置该参数，默认最多返回1000条任务。
	// 2. 最终将按照JobId顺序进行返回。
	MaxResults *int64 `location:"querystring" locationName:"maxResults" type:"integer"`

	// 所需查询的任务状态信息。可选的任务状态包括：
	// New：表示批量处理任务刚被创建，任务正在解析中。
	// Active：表示批量操作任务正在进行中。
	// Complete：表示批量操作任务已完成，处于最终状态。
	// 如果未指定任务状态，KS3将返回所有状态的任务。如果指定了任务状态，KS3仅返回指定状态的任务。
	JobStatuses []string `location:"querystrings" locationName:"jobStatuses" type:"list"`

	// 分页符。List操作结束后将返回本次任务列表的最后一个JobId作为nextToken，在下一次List操作时传入该nextToken值，即可接续上一次List的内容进行List，用于分页。
	NextToken *string `location:"querystring" locationName:"nextToken" type:"string"`

	// 设置扩展请求头。如果现有字段不支持设置所需的请求头，您可以通过此字段进行设置。
	ExtendHeaders map[string]*string `location:"extendHeaders" type:"map"`

	// 设置扩展查询参数。如果现有字段不支持设置所需的查询参数，您可以通过此字段进行设置。
	ExtendQueryParams map[string]*string `location:"extendQueryParams" type:"map"`
}

type ListJobsOutput struct {
	// 包含所列举批量处理规则信息的容器。
	ListJobsResult *ListJobsResult `locationName:"ListJobsResult" type:"structure"`

	// http响应头。
	Metadata map[string]*string `location:"headers"  type:"map"`

	// http响应状态码。
	StatusCode *int64 `location:"statusCode" type:"integer"`

	metadataListJobsInput `json:"-" xml:"-"`
}

type metadataListJobsInput struct {
	SDKShapeTraits bool `type:"structure" payload:"ListJobsResult"`
}

type ListJobsResult struct {
	// 包含KS3返回的多个批量处理任务信息。
	Jobs *JobList `locationName:"Jobs" type:"structure"`

	// 分页符。
	NextToken *string `locationName:"NextToken" type:"string"`
}

type JobList struct {
	// 包含KS3返回的单个批量处理任务信息。
	Members []*JobMember `locationName:"Member" type:"list" flattened:"true"`
}

type JobMember struct {
	// 任务ID。规则创建成功后，KS3会自动返回该参数，每个规则对应唯一的任务ID。
	JobId *string `locationName:"JobId" type:"string"`

	// 任务描述。
	Description *string `locationName:"Description" type:"string"`

	// 具体操作。支持批量解冻、批量修改ACL、批量删除操作。
	Operation *string `locationName:"Operation" type:"string"`

	// 任务创建时间。
	CreationTime *string `locationName:"CreationTime" type:"timestamp" timestampFormat:"iso8601"`

	// 任务优先级。取值越大表示任务执行的优先级越高。
	Priority *int64 `locationName:"Priority" type:"integer"`

	// 任务执行状况概述。描述该批量处理任务中所执行的操作总数，成功的操作数量以及失败的操作数量。
	ProgressSummary *JobProgressSummary `locationName:"ProgressSummary" type:"structure"`

	// 任务执行的当前状态。合法参数值包括：New、Active、Complete。
	// New：批量处理任务刚被创建，任务正在解析中。
	// Active：批量操作任务进行中。
	// Complete：批量操作任务已完成，处于最终状态。
	Status *string `locationName:"Status" type:"string"`

	// 任务终止的时间。
	TerminationDate *string `locationName:"TerminationDate" type:"timestamp" timestampFormat:"iso8601"`
}

// ListJobsRequest generates a request for the ListJobs operation.
func (c *S3) ListJobsRequest(input *ListJobsInput) (req *aws.Request, output *ListJobsOutput) {
	op := &aws.Operation{
		Name:       "ListJobs",
		HTTPMethod: "GET",
		HTTPPath:   "/?jobs",
	}

	if input == nil {
		input = &ListJobsInput{}
	}

	req = c.newRequest(op, input, output)
	output = &ListJobsOutput{
		ListJobsResult: &ListJobsResult{},
	}
	req.Data = output
	return
}

// ListJobs lists the jobs.
func (c *S3) ListJobs(input *ListJobsInput) (*ListJobsOutput, error) {
	req, out := c.ListJobsRequest(input)
	err := req.Send()
	return out, err
}

func (c *S3) ListJobsWithContext(ctx aws.Context, input *ListJobsInput) (*ListJobsOutput, error) {
	req, out := c.ListJobsRequest(input)
	req.SetContext(ctx)
	err := req.Send()
	return out, err
}

type DeleteJobInput struct {
	// 需要删除的任务ID。每个任务对应唯一的任务ID，创建批量处理任务（CreateJob）成功后，KS3会返回任务ID。
	JobId *string `location:"querystring" locationName:"jobId" type:"string" required:"true"`

	// 设置扩展请求头。如果现有字段不支持设置所需的请求头，您可以通过此字段进行设置。
	ExtendHeaders map[string]*string `location:"extendHeaders" type:"map"`

	// 设置扩展查询参数。如果现有字段不支持设置所需的查询参数，您可以通过此字段进行设置。
	ExtendQueryParams map[string]*string `location:"extendQueryParams" type:"map"`
}

type DeleteJobOutput struct {
	// http响应头。
	Metadata map[string]*string `location:"headers" type:"map"`

	// http响应状态码。
	StatusCode *int64 `location:"statusCode" type:"integer"`
}

// DeleteJobRequest generates a request for the DeleteJob operation.
func (c *S3) DeleteJobRequest(input *DeleteJobInput) (req *aws.Request, output *DeleteJobOutput) {
	op := &aws.Operation{
		Name:       "DeleteJob",
		HTTPMethod: "DELETE",
		HTTPPath:   "/?jobs",
	}

	if input == nil {
		input = &DeleteJobInput{}
	}

	req = c.newRequest(op, input, output)
	output = &DeleteJobOutput{}
	req.Data = output
	return
}

// DeleteJob deletes a batch job.
func (c *S3) DeleteJob(input *DeleteJobInput) (*DeleteJobOutput, error) {
	req, out := c.DeleteJobRequest(input)
	err := req.Send()
	return out, err
}

func (c *S3) DeleteJobWithContext(ctx aws.Context, input *DeleteJobInput) (*DeleteJobOutput, error) {
	req, out := c.DeleteJobRequest(input)
	req.SetContext(ctx)
	err := req.Send()
	return out, err
}

type UpdateJobPriorityInput struct {
	// 任务ID。每个任务对应唯一的任务ID，创建批量处理任务（CreateJob）成功后，KS3会返回任务ID。
	JobId *string `location:"querystring" locationName:"jobId" type:"string" required:"true"`

	// 任务优先级，取值越大表示任务执行的优先级越高。
	// 取值范围：0-2147483647
	Priority *int64 `location:"querystring" locationName:"priority" type:"integer" required:"true"`

	// 设置扩展请求头。如果现有字段不支持设置所需的请求头，您可以通过此字段进行设置。
	ExtendHeaders map[string]*string `location:"extendHeaders" type:"map"`

	// 设置扩展查询参数。如果现有字段不支持设置所需的查询参数，您可以通过此字段进行设置。
	ExtendQueryParams map[string]*string `location:"extendQueryParams" type:"map"`
}

type UpdateJobPriorityOutput struct {
	// 包含任务当前优先级信息的容器。
	UpdateJobPriorityResult *UpdateJobPriorityResult `locationName:"UpdateJobPriorityResult" type:"structure"`

	// http响应头。
	Metadata map[string]*string `location:"headers" type:"map"`

	// http响应状态码。
	StatusCode *int64 `location:"statusCode" type:"integer"`

	metadataUpdateJobPriorityOutput `json:"-" xml:"-"`
}

type metadataUpdateJobPriorityOutput struct {
	SDKShapeTraits bool `type:"structure" payload:"UpdateJobPriorityResult"`
}

type UpdateJobPriorityResult struct {
	// 任务ID。每个任务对应唯一的任务ID，创建批量处理任务（CreateJob）成功后，KS3会返回任务ID。
	JobId *string `locationName:"JobId" type:"string"`

	// 任务的当前优先级。任务优先级数值越大，优先级越高，高优先级的任务会被优先执行。
	Priority *int64 `locationName:"Priority" type:"integer"`
}

// UpdateJobPriorityRequest generates a request for the UpdateJobPriority operation.
func (c *S3) UpdateJobPriorityRequest(input *UpdateJobPriorityInput) (req *aws.Request, output *UpdateJobPriorityOutput) {
	op := &aws.Operation{
		Name:       "UpdateJobPriority",
		HTTPMethod: "PUT",
		HTTPPath:   "/?jobs&action=updateJobPriority",
	}

	if input == nil {
		input = &UpdateJobPriorityInput{}
	}

	req = c.newRequest(op, input, output)
	output = &UpdateJobPriorityOutput{}
	req.Data = output
	return
}

// UpdateJobPriority updates a job's priority.
func (c *S3) UpdateJobPriority(input *UpdateJobPriorityInput) (*UpdateJobPriorityOutput, error) {
	req, out := c.UpdateJobPriorityRequest(input)
	err := req.Send()
	return out, err
}

func (c *S3) UpdateJobPriorityWithContext(ctx aws.Context, input *UpdateJobPriorityInput) (*UpdateJobPriorityOutput, error) {
	req, out := c.UpdateJobPriorityRequest(input)
	req.SetContext(ctx)
	err := req.Send()
	return out, err
}
