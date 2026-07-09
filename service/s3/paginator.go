package s3

import (
	"fmt"

	"github.com/ks3sdklib/aws-sdk-go/aws"
)

// ListObjectsPaginator ListObjects分页器。
type ListObjectsPaginator struct {
	client      *S3
	input       *ListObjectsInput
	marker      *string
	isTruncated *bool
}

// NewListObjectsPaginator 创建ListObjects分页器。
func (c *S3) NewListObjectsPaginator(input *ListObjectsInput) *ListObjectsPaginator {
	if input == nil {
		input = &ListObjectsInput{}
	}

	return &ListObjectsPaginator{
		client:      c,
		input:       input,
		marker:      input.Marker,
		isTruncated: aws.Boolean(true),
	}
}

// HasNext 是否有下一页。
func (p *ListObjectsPaginator) HasNext() bool {
	return *p.isTruncated
}

// NextPage 获取下一页。
func (p *ListObjectsPaginator) NextPage() (*ListObjectsOutput, error) {
	return p.NextPageWithContext(aws.BackgroundContext())
}

// NextPageWithContext 获取下一页，支持传入上下文。
func (p *ListObjectsPaginator) NextPageWithContext(ctx aws.Context) (*ListObjectsOutput, error) {
	if !p.HasNext() {
		return nil, fmt.Errorf("no more pages available")
	}

	input := *p.input
	input.Marker = p.marker

	result, err := p.client.ListObjectsWithContext(ctx, &input)
	if err != nil {
		return nil, err
	}

	p.isTruncated = result.IsTruncated
	if result.IsTruncated == nil {
		p.isTruncated = aws.Boolean(false)
	}

	if *p.isTruncated {
		if result.NextMarker != nil {
			p.marker = result.NextMarker
		} else if len(result.Contents) > 0 && result.Contents[len(result.Contents)-1].Key != nil {
			p.marker = result.Contents[len(result.Contents)-1].Key
		}
	}

	return result, nil
}

// ListObjectsV2Paginator ListObjectsV2分页器。
type ListObjectsV2Paginator struct {
	client        *S3
	input         *ListObjectsV2Input
	continueToken *string
	isTruncated   *bool
}

// NewListObjectsV2Paginator 创建ListObjectsV2分页器。
func (c *S3) NewListObjectsV2Paginator(input *ListObjectsV2Input) *ListObjectsV2Paginator {
	if input == nil {
		input = &ListObjectsV2Input{}
	}

	return &ListObjectsV2Paginator{
		client:        c,
		input:         input,
		continueToken: input.ContinuationToken,
		isTruncated:   aws.Boolean(true),
	}
}

// HasNext 是否有下一页。
func (p *ListObjectsV2Paginator) HasNext() bool {
	return *p.isTruncated
}

// NextPage 获取下一页。
func (p *ListObjectsV2Paginator) NextPage() (*ListObjectsV2Output, error) {
	return p.NextPageWithContext(aws.BackgroundContext())
}

// NextPageWithContext 获取下一页，支持传入上下文。
func (p *ListObjectsV2Paginator) NextPageWithContext(ctx aws.Context) (*ListObjectsV2Output, error) {
	if !p.HasNext() {
		return nil, fmt.Errorf("no more pages available")
	}

	input := *p.input
	input.ContinuationToken = p.continueToken

	result, err := p.client.ListObjectsV2WithContext(ctx, &input)
	if err != nil {
		return nil, err
	}

	p.isTruncated = result.IsTruncated
	if result.IsTruncated == nil {
		p.isTruncated = aws.Boolean(false)
	}
	p.continueToken = result.NextContinuationToken

	return result, nil
}

// ListMultipartUploadsPaginator ListMultipartUploads分页器。
type ListMultipartUploadsPaginator struct {
	client         *S3
	input          *ListMultipartUploadsInput
	keyMarker      *string
	uploadIDMarker *string
	isTruncated    *bool
}

// NewListMultipartUploadsPaginator 创建ListMultipartUploads分页器。
func (c *S3) NewListMultipartUploadsPaginator(input *ListMultipartUploadsInput) *ListMultipartUploadsPaginator {
	if input == nil {
		input = &ListMultipartUploadsInput{}
	}

	return &ListMultipartUploadsPaginator{
		client:         c,
		input:          input,
		keyMarker:      input.KeyMarker,
		uploadIDMarker: input.UploadIDMarker,
		isTruncated:    aws.Boolean(true),
	}
}

// HasNext 是否有下一页。
func (p *ListMultipartUploadsPaginator) HasNext() bool {
	return *p.isTruncated
}

// NextPage 获取下一页。
func (p *ListMultipartUploadsPaginator) NextPage() (*ListMultipartUploadsOutput, error) {
	return p.NextPageWithContext(aws.BackgroundContext())
}

// NextPageWithContext 获取下一页，支持传入上下文。
func (p *ListMultipartUploadsPaginator) NextPageWithContext(ctx aws.Context) (*ListMultipartUploadsOutput, error) {
	if !p.HasNext() {
		return nil, fmt.Errorf("no more pages available")
	}

	input := *p.input
	input.KeyMarker = p.keyMarker
	input.UploadIDMarker = p.uploadIDMarker

	result, err := p.client.ListMultipartUploadsWithContext(ctx, &input)
	if err != nil {
		return nil, err
	}

	p.isTruncated = result.IsTruncated
	if result.IsTruncated == nil {
		p.isTruncated = aws.Boolean(false)
	}
	p.keyMarker = result.NextKeyMarker
	p.uploadIDMarker = result.NextUploadIDMarker

	return result, nil
}

// ListPartsPaginator ListParts分页器。
type ListPartsPaginator struct {
	client           *S3
	input            *ListPartsInput
	partNumberMarker *int64
	isTruncated      *bool
}

// NewListPartsPaginator 创建ListParts分页器。
func (c *S3) NewListPartsPaginator(input *ListPartsInput) *ListPartsPaginator {
	if input == nil {
		input = &ListPartsInput{}
	}

	return &ListPartsPaginator{
		client:           c,
		input:            input,
		partNumberMarker: input.PartNumberMarker,
		isTruncated:      aws.Boolean(true),
	}
}

// HasNext 是否有下一页。
func (p *ListPartsPaginator) HasNext() bool {
	return *p.isTruncated
}

// NextPage 获取下一页。
func (p *ListPartsPaginator) NextPage() (*ListPartsOutput, error) {
	return p.NextPageWithContext(aws.BackgroundContext())
}

// NextPageWithContext 获取下一页，支持传入上下文。
func (p *ListPartsPaginator) NextPageWithContext(ctx aws.Context) (*ListPartsOutput, error) {
	if !p.HasNext() {
		return nil, fmt.Errorf("no more pages available")
	}

	input := *p.input
	input.PartNumberMarker = p.partNumberMarker

	result, err := p.client.ListPartsWithContext(ctx, &input)
	if err != nil {
		return nil, err
	}

	p.isTruncated = result.IsTruncated
	if result.IsTruncated == nil {
		p.isTruncated = aws.Boolean(false)
	}
	p.partNumberMarker = result.NextPartNumberMarker

	return result, nil
}

// ListRetentionPaginator ListRetention分页器。
type ListRetentionPaginator struct {
	client      *S3
	input       *ListRetentionInput
	marker      *string
	isTruncated *bool
}

// NewListRetentionPaginator 创建ListRetention分页器。
func (c *S3) NewListRetentionPaginator(input *ListRetentionInput) *ListRetentionPaginator {
	if input == nil {
		input = &ListRetentionInput{}
	}

	return &ListRetentionPaginator{
		client:      c,
		input:       input,
		marker:      input.Marker,
		isTruncated: aws.Boolean(true),
	}
}

// HasNext 是否有下一页。
func (p *ListRetentionPaginator) HasNext() bool {
	return *p.isTruncated
}

// NextPage 获取下一页。
func (p *ListRetentionPaginator) NextPage() (*ListRetentionOutput, error) {
	return p.NextPageWithContext(aws.BackgroundContext())
}

// NextPageWithContext 获取下一页，支持传入上下文。
func (p *ListRetentionPaginator) NextPageWithContext(ctx aws.Context) (*ListRetentionOutput, error) {
	if !p.HasNext() {
		return nil, fmt.Errorf("no more pages available")
	}

	input := *p.input
	input.Marker = p.marker

	result, err := p.client.ListRetentionWithContext(ctx, &input)
	if err != nil {
		return nil, err
	}

	if result.ListRetentionResult != nil {
		p.isTruncated = result.ListRetentionResult.IsTruncated
		if result.ListRetentionResult.IsTruncated == nil {
			p.isTruncated = aws.Boolean(false)
		}
		p.marker = result.ListRetentionResult.NextMarker
	} else {
		p.isTruncated = aws.Boolean(false)
	}

	return result, nil
}

// ListBucketInventoryPaginator ListBucketInventory分页器。
type ListBucketInventoryPaginator struct {
	client        *S3
	input         *ListBucketInventoryInput
	continueToken *string
	isTruncated   *bool
}

// NewListBucketInventoryPaginator 创建ListBucketInventory分页器。
func (c *S3) NewListBucketInventoryPaginator(input *ListBucketInventoryInput) *ListBucketInventoryPaginator {
	if input == nil {
		input = &ListBucketInventoryInput{}
	}

	return &ListBucketInventoryPaginator{
		client:        c,
		input:         input,
		continueToken: input.ContinuationToken,
		isTruncated:   aws.Boolean(true),
	}
}

// HasNext 是否有下一页。
func (p *ListBucketInventoryPaginator) HasNext() bool {
	return *p.isTruncated
}

// NextPage 获取下一页。
func (p *ListBucketInventoryPaginator) NextPage() (*ListBucketInventoryOutput, error) {
	return p.NextPageWithContext(aws.BackgroundContext())
}

// NextPageWithContext 获取下一页，支持传入上下文。
func (p *ListBucketInventoryPaginator) NextPageWithContext(ctx aws.Context) (*ListBucketInventoryOutput, error) {
	if !p.HasNext() {
		return nil, fmt.Errorf("no more pages available")
	}

	input := *p.input
	input.ContinuationToken = p.continueToken

	result, err := p.client.ListBucketInventoryWithContext(ctx, &input)
	if err != nil {
		return nil, err
	}

	if result.InventoryConfigurationsResult != nil {
		p.isTruncated = result.InventoryConfigurationsResult.IsTruncated
		if result.InventoryConfigurationsResult.IsTruncated == nil {
			p.isTruncated = aws.Boolean(false)
		}
		p.continueToken = result.InventoryConfigurationsResult.NextContinuationToken
	} else {
		p.isTruncated = aws.Boolean(false)
	}

	return result, nil
}

// ListJobsPaginator ListJobs分页器。
type ListJobsPaginator struct {
	client    *S3
	input     *ListJobsInput
	nextToken *string
	hasNext   *bool
}

// NewListJobsPaginator 创建ListJobs分页器。
func (c *S3) NewListJobsPaginator(input *ListJobsInput) *ListJobsPaginator {
	if input == nil {
		input = &ListJobsInput{}
	}

	return &ListJobsPaginator{
		client:    c,
		input:     input,
		nextToken: input.NextToken,
		hasNext:   aws.Boolean(true),
	}
}

// HasNext 是否有下一页。
func (p *ListJobsPaginator) HasNext() bool {
	return *p.hasNext
}

// NextPage 获取下一页。
func (p *ListJobsPaginator) NextPage() (*ListJobsOutput, error) {
	return p.NextPageWithContext(aws.BackgroundContext())
}

// NextPageWithContext 获取下一页，支持传入上下文。
func (p *ListJobsPaginator) NextPageWithContext(ctx aws.Context) (*ListJobsOutput, error) {
	if !p.HasNext() {
		return nil, fmt.Errorf("no more pages available")
	}

	input := *p.input
	input.NextToken = p.nextToken

	result, err := p.client.ListJobsWithContext(ctx, &input)
	if err != nil {
		return nil, err
	}

	if result.ListJobsResult != nil {
		p.nextToken = result.ListJobsResult.NextToken
		p.hasNext = aws.Boolean(result.ListJobsResult.NextToken != nil)
	} else {
		p.hasNext = aws.Boolean(false)
	}

	return result, nil
}
