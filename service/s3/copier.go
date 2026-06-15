package s3

import (
	"context"
	"errors"
	"fmt"
	"github.com/ks3sdklib/aws-sdk-go/aws"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type CopyFileInput struct {
	// The name of the bucket.
	Bucket *string `location:"uri" locationName:"Bucket" type:"string" required:"true"`

	// Object key of the object.
	Key *string `location:"uri" locationName:"Key" type:"string" required:"true"`

	// The name of the source bucket.
	SourceBucket *string `location:"uri" locationName:"SourceBucket" type:"string" required:"true"`

	// Object key of the source object.
	SourceKey *string `location:"uri" locationName:"SourceKey" type:"string" required:"true"`

	// The size of each part.
	PartSize *int64 `type:"integer"`

	// The number of tasks to upload the file.
	TaskNum *int64 `type:"integer"`

	// Whether to enable checkpoint.
	EnableCheckpoint *bool `type:"boolean"`

	// The directory to store the checkpoint file.
	CheckpointDir *string `type:"string"`

	// The checkpoint file path.
	CheckpointFile *string `type:"string"`

	// The canned ACL to apply to the object.
	ACL *string `location:"header" locationName:"x-amz-acl" type:"string"`

	// Specifies caching behavior along the request/reply chain.
	CacheControl *string `location:"header" locationName:"Cache-Control" type:"string"`

	// Specifies presentational information for the object.
	ContentDisposition *string `location:"header" locationName:"Content-Disposition" type:"string"`

	// Specifies what content encodings have been applied to the object and thus
	// what decoding mechanisms must be applied to obtain the media-type referenced
	// by the Content-Type header field.
	ContentEncoding *string `location:"header" locationName:"Content-Encoding" type:"string"`

	// A standard MIME type describing the format of the object data.
	ContentType *string `location:"header" locationName:"Content-Type" type:"string"`

	// The date and time at which the object is no longer cacheable.
	Expires *time.Time `location:"header" locationName:"Expires" type:"timestamp" timestampFormat:"rfc822"`

	// A map of metadata to store with the object in S3.
	Metadata map[string]*string `location:"headers" locationName:"x-amz-meta-" type:"map"`

	// Specifies whether the metadata is copied from the source object or replaced
	// with metadata provided in the request.
	MetadataDirective *string `location:"header" locationName:"x-amz-metadata-directive" type:"string"`

	// The type of storage to use for the object. Defaults to 'STANDARD'.
	StorageClass *string `location:"header" locationName:"x-amz-storage-class" type:"string"`

	// Specifies the object tag of the object. Multiple tags can be set at the same time, such as: TagA=A&TagB=B.
	// Note: Key and Value need to be URL-encoded first. If an item does not have "=", the Value is considered to be an empty string.
	Tagging *string `location:"header" locationName:"x-amz-tagging" type:"string"`

	// Specifies how to set the object tag of the target object.
	// Default value: COPY
	// Valid values:
	// COPY (default value): Copies the object tag of the source object to the target object.
	// REPLACE: Ignores the object tag of the source object and directly uses the object tag specified in the request.
	TaggingDirective *string `location:"header" locationName:"x-amz-tagging-directive" type:"string"`

	// Specifies whether the object is forbidden to overwrite.
	ForbidOverwrite *bool `location:"header" locationName:"x-amz-forbid-overwrite" type:"boolean"`

	// Allows grantee to read the object data and its metadata.
	GrantRead *string `location:"header" locationName:"x-amz-grant-read" type:"string"`

	// Gives the grantee READ, READ_ACP, and WRITE_ACP permissions on the object.
	GrantFullControl *string `location:"header" locationName:"x-amz-grant-full-control" type:"string"`

	// Copies the object if its entity tag (ETag) matches the specified tag.
	CopySourceIfMatch *string `location:"header" locationName:"x-amz-copy-source-if-match" type:"string"`

	// Copies the object if it has been modified since the specified time.
	CopySourceIfModifiedSince *time.Time `location:"header" locationName:"x-amz-copy-source-if-modified-since" type:"timestamp" timestampFormat:"rfc822"`

	// Copies the object if its entity tag (ETag) is different from the specified ETag.
	CopySourceIfNoneMatch *string `location:"header" locationName:"x-amz-copy-source-if-none-match" type:"string"`

	// Copies the object if it hasn't been modified since the specified time.
	CopySourceIfUnmodifiedSince *time.Time `location:"header" locationName:"x-amz-copy-source-if-unmodified-since" type:"timestamp" timestampFormat:"rfc822"`

	// Specifies the decryption algorithm used to decrypt the data source object. Valid value: AES256.
	CopySourceSSECustomerAlgorithm *string `location:"header" locationName:"x-amz-copy-source-server-side-encryption-customer-algorithm" type:"string"`

	// The base64-encoded encryption key used for KS3 decryption specified by the user.
	// Its value must be the same as the key used when the data source object was created.
	CopySourceSSECustomerKey *string `location:"header" locationName:"x-amz-copy-source-server-side-encryption-customer-key" type:"string"`

	// Specifies the 128-bit MD5 digest of the encryption key according to RFC 1321.
	// If the server encrypted with a user-provided encryption key, when decryption is requested,
	// the response will include this header to provide data consistency verification information
	// for the user-provided encryption key.
	CopySourceSSECustomerKeyMD5 *string `location:"header" locationName:"x-amz-copy-source-server-side-encryption-customer-key-MD5" type:"string"`

	// The Server-side encryption algorithm used when storing this object in KS3, eg: AES256.
	ServerSideEncryption *string `location:"header" locationName:"x-amz-server-side-encryption" type:"string"`

	// Specifies the algorithm to use to when encrypting the object, eg: AES256.
	SSECustomerAlgorithm *string `location:"header" locationName:"x-amz-server-side-encryption-customer-algorithm" type:"string"`

	// Specifies the customer-provided encryption key for KS3 to use in encrypting data.
	SSECustomerKey *string `location:"header" locationName:"x-amz-server-side-encryption-customer-key" type:"string"`

	// Specifies the 128-bit MD5 digest of the encryption key according to RFC 1321.
	SSECustomerKeyMD5 *string `location:"header" locationName:"x-amz-server-side-encryption-customer-key-MD5" type:"string"`

	// Progress callback function
	ProgressFn aws.ProgressFunc `location:"function"`
}

type CopyFileOutput struct {
	Bucket *string

	Key *string

	ETag *string

	ChecksumCRC64ECMA *string
}

func (c *S3) CopyFile(request *CopyFileInput) (*CopyFileOutput, error) {
	return c.CopyFileWithContext(context.Background(), request)
}

func (c *S3) CopyFileWithContext(ctx context.Context, request *CopyFileInput) (*CopyFileOutput, error) {
	return newCopier(c, ctx, request).copyFile()
}

func (c *S3) CopyFileAcrossRegion(request *CopyFileInput, dstClient *S3) (*UploadFileOutput, error) {
	return c.CopyFileAcrossRegionWithContext(context.Background(), request, dstClient)
}

func (c *S3) CopyFileAcrossRegionWithContext(ctx context.Context, request *CopyFileInput, dstClient *S3) (*UploadFileOutput, error) {
	uploadFileRequest, err := c.buildUploadFileRequest(ctx, request)
	if err != nil {
		return nil, err
	}

	return dstClient.UploadFileWithContext(ctx, uploadFileRequest)
}

func (c *S3) buildUploadFileRequest(ctx context.Context, request *CopyFileInput) (*UploadFileInput, error) {
	if request == nil {
		return nil, errors.New("copyFileRequest is required")
	}

	if aws.ToString(request.Bucket) == "" {
		return nil, errors.New("bucket is required")
	}

	if aws.ToString(request.Key) == "" {
		return nil, errors.New("key is required")
	}

	if aws.ToString(request.SourceBucket) == "" {
		return nil, errors.New("source bucket is required")
	}

	if aws.ToString(request.SourceKey) == "" {
		return nil, errors.New("source key is required")
	}

	input := &UploadFileInput{
		Bucket:               request.Bucket,
		Key:                  request.Key,
		PartSize:             request.PartSize,
		TaskNum:              request.TaskNum,
		EnableCheckpoint:     request.EnableCheckpoint,
		CheckpointDir:        request.CheckpointDir,
		CheckpointFile:       request.CheckpointFile,
		ACL:                  request.ACL,
		CacheControl:         request.CacheControl,
		ContentDisposition:   request.ContentDisposition,
		ContentEncoding:      request.ContentEncoding,
		ContentType:          request.ContentType,
		Expires:              request.Expires,
		Metadata:             request.Metadata,
		StorageClass:         request.StorageClass,
		Tagging:              request.Tagging,
		ForbidOverwrite:      request.ForbidOverwrite,
		GrantRead:            request.GrantRead,
		GrantFullControl:     request.GrantFullControl,
		ServerSideEncryption: request.ServerSideEncryption,
		SSECustomerAlgorithm: request.SSECustomerAlgorithm,
		SSECustomerKey:       request.SSECustomerKey,
		SSECustomerKeyMD5:    request.SSECustomerKeyMD5,
		ProgressFn:           request.ProgressFn,
	}

	fetcher := &Fetcher{
		client:  c,
		request: request,
	}

	var filePartFetcher FilePartFetcher = fetcher
	input.FilePartFetcher = &filePartFetcher

	resp, err := c.HeadObjectWithContext(ctx, &HeadObjectInput{
		Bucket:               request.SourceBucket,
		Key:                  request.SourceKey,
		IfModifiedSince:      request.CopySourceIfModifiedSince,
		IfUnmodifiedSince:    request.CopySourceIfUnmodifiedSince,
		IfMatch:              request.CopySourceIfMatch,
		IfNoneMatch:          request.CopySourceIfNoneMatch,
		SSECustomerAlgorithm: request.CopySourceSSECustomerAlgorithm,
		SSECustomerKey:       request.CopySourceSSECustomerKey,
		SSECustomerKeyMD5:    request.CopySourceSSECustomerKeyMD5,
	})
	if err != nil {
		return nil, err
	}
	input.ObjectMeta = resp.Metadata

	if !strings.EqualFold(aws.ToString(request.MetadataDirective), "REPLACE") {
		input.CacheControl = resp.Metadata[HTTPHeaderCacheControl]
		input.ContentDisposition = resp.Metadata[HTTPHeaderContentDisposition]
		input.ContentEncoding = resp.Metadata[HTTPHeaderContentEncoding]
		input.ContentType = resp.Metadata[HTTPHeaderContentType]
		expires, err := time.Parse("Mon, 02 Jan 2006 15:04:05 GMT", aws.ToString(resp.Metadata[HTTPHeaderExpires]))
		if err == nil {
			input.Expires = aws.Time(expires)
		}

		metaData := map[string]*string{}
		for k, v := range resp.Metadata {
			if strings.HasPrefix(strings.ToLower(k), MetaPrefix) {
				metaData[k] = v
			}
		}
		input.Metadata = metaData
	}
	if !strings.EqualFold(aws.ToString(request.TaggingDirective), "REPLACE") {
		taggingResp, err := c.GetObjectTaggingWithContext(ctx, &GetObjectTaggingInput{
			Bucket: request.SourceBucket,
			Key:    request.SourceKey,
		})
		if err != nil {
			return nil, err
		}

		tagStr := taggingResp.Tagging.ToString()
		if tagStr != "" {
			input.Tagging = aws.String(tagStr)
		}
	}
	if input.ACL == nil {
		aclResp, err := c.GetObjectACLWithContext(ctx, &GetObjectACLInput{
			Bucket: request.SourceBucket,
			Key:    request.SourceKey,
		})
		if err != nil {
			return nil, err
		}

		input.ACL = aws.String(GetCannedACL(aclResp.Grants))
	}
	if input.StorageClass == nil {
		input.StorageClass = resp.Metadata[HTTPHeaderAmzStorageClass]
	}

	return input, nil
}

type Fetcher struct {
	client  *S3
	request *CopyFileInput
}

type Body struct {
	io.ReadCloser
}

func (b *Body) Seek(offset int64, whence int) (int64, error) {
	return 0, os.ErrInvalid
}

func (f *Fetcher) Fetch(ctx context.Context, objectRange []int64) (io.ReadSeeker, error) {
	resp, err := f.client.GetObjectWithContext(ctx, &GetObjectInput{
		Bucket:               f.request.SourceBucket,
		Key:                  f.request.SourceKey,
		Range:                aws.String(fmt.Sprintf("bytes=%d-%d", objectRange[0], objectRange[1])),
		SSECustomerAlgorithm: f.request.CopySourceSSECustomerAlgorithm,
		SSECustomerKey:       f.request.CopySourceSSECustomerKey,
		SSECustomerKeyMD5:    f.request.CopySourceSSECustomerKeyMD5,
	})
	if err != nil {
		return nil, err
	}

	body := &Body{resp.Body}

	return body, nil
}

type Copier struct {
	client *S3

	context context.Context

	copyFileRequest *CopyFileInput

	copyCheckpoint *CopyCheckpoint

	CompletedSize int64

	copyObjectMeta map[string]*string

	mu sync.Mutex

	error error
}

func newCopier(s3 *S3, ctx context.Context, request *CopyFileInput) *Copier {
	return &Copier{
		client:          s3,
		context:         ctx,
		copyFileRequest: request,
	}
}

func (c *Copier) copyFile() (*CopyFileOutput, error) {
	err := c.validate()
	if err != nil {
		return nil, err
	}

	c.copyObjectMeta, err = c.headObject()
	if err != nil {
		return nil, err
	}

	fileSize, _ := strconv.ParseInt(aws.ToString(c.copyObjectMeta[HTTPHeaderContentLength]), 10, 64)

	var resp *CopyFileOutput
	if fileSize <= aws.ToLong(c.copyFileRequest.PartSize) {
		resp, err = c.copyObject()
	} else {
		resp, err = c.multipartCopy()
	}
	if err != nil {
		return nil, err
	}

	if c.client.Config.CrcCheckEnabled {
		clientCrc64, _ := strconv.ParseUint(aws.ToString(c.copyObjectMeta[HTTPHeaderAmzChecksumCrc64ecma]), 10, 64)
		serverCrc64, _ := strconv.ParseUint(aws.ToString(resp.ChecksumCRC64ECMA), 10, 64)
		c.client.Config.LogDebug("check file crc64, client crc64:%d, server crc64:%d", clientCrc64, serverCrc64)
		if serverCrc64 != 0 && clientCrc64 != serverCrc64 {
			return nil, errors.New(fmt.Sprintf("crc64 check failed, client crc64:%d, server crc64:%d", clientCrc64, serverCrc64))
		}
	}

	return resp, err
}

func (c *Copier) validate() error {
	request := c.copyFileRequest
	if request == nil {
		return errors.New("copyFileRequest is required")
	}

	if aws.ToString(request.Bucket) == "" {
		return errors.New("bucket is required")
	}

	if aws.ToString(request.Key) == "" {
		return errors.New("key is required")
	}

	if aws.ToString(request.SourceBucket) == "" {
		return errors.New("source bucket is required")
	}

	if aws.ToString(request.SourceKey) == "" {
		return errors.New("source key is required")
	}

	if request.PartSize == nil {
		request.PartSize = aws.Long(DefaultPartSize)
	} else if aws.ToLong(request.PartSize) < MinPartSize {
		request.PartSize = aws.Long(MinPartSize)
	} else if aws.ToLong(request.PartSize) > MaxPartSize {
		request.PartSize = aws.Long(MaxPartSize)
	}

	if aws.ToLong(request.TaskNum) <= 0 {
		request.TaskNum = aws.Long(DefaultTaskNum)
	}

	return nil
}

func (c *Copier) copyObject() (*CopyFileOutput, error) {
	request := c.copyFileRequest
	input := &CopyObjectInput{
		Bucket:                         request.Bucket,
		Key:                            request.Key,
		SourceBucket:                   request.SourceBucket,
		SourceKey:                      request.SourceKey,
		ACL:                            request.ACL,
		CacheControl:                   request.CacheControl,
		ContentDisposition:             request.ContentDisposition,
		ContentEncoding:                request.ContentEncoding,
		ContentType:                    request.ContentType,
		Expires:                        request.Expires,
		Metadata:                       request.Metadata,
		MetadataDirective:              request.MetadataDirective,
		StorageClass:                   request.StorageClass,
		Tagging:                        request.Tagging,
		TaggingDirective:               request.TaggingDirective,
		ForbidOverwrite:                request.ForbidOverwrite,
		GrantRead:                      request.GrantRead,
		GrantFullControl:               request.GrantFullControl,
		CopySourceIfMatch:              request.CopySourceIfMatch,
		CopySourceIfNoneMatch:          request.CopySourceIfNoneMatch,
		CopySourceIfModifiedSince:      request.CopySourceIfModifiedSince,
		CopySourceIfUnmodifiedSince:    request.CopySourceIfUnmodifiedSince,
		CopySourceSSECustomerAlgorithm: request.CopySourceSSECustomerAlgorithm,
		CopySourceSSECustomerKey:       request.CopySourceSSECustomerKey,
		CopySourceSSECustomerKeyMD5:    request.CopySourceSSECustomerKeyMD5,
		ServerSideEncryption:           request.ServerSideEncryption,
		SSECustomerAlgorithm:           request.SSECustomerAlgorithm,
		SSECustomerKey:                 request.SSECustomerKey,
		SSECustomerKeyMD5:              request.SSECustomerKeyMD5,
	}
	if c.copyFileRequest.StorageClass == nil {
		input.StorageClass = c.copyObjectMeta[HTTPHeaderAmzStorageClass]
	}

	resp, err := c.client.CopyObjectWithContext(c.context, input)
	if err != nil {
		return nil, err
	}

	return &CopyFileOutput{
		Bucket:            request.Bucket,
		Key:               request.Key,
		ETag:              resp.CopyObjectResult.ETag,
		ChecksumCRC64ECMA: resp.CopyObjectResult.ChecksumCRC64ECMA,
	}, nil
}

func (c *Copier) multipartCopy() (*CopyFileOutput, error) {
	ccp, err := newCopyCheckpoint(c)
	if err != nil {
		return nil, err
	}
	c.copyCheckpoint = ccp

	if aws.ToBoolean(c.copyFileRequest.EnableCheckpoint) {
		cpFilePath := aws.ToString(c.copyFileRequest.CheckpointFile)
		if cpFilePath == "" {
			cpFilePath, err = generateCopyCpFilePath(c.copyFileRequest)
			if err != nil {
				return nil, err
			}
		}
		ccp.CpFilePath = cpFilePath

		err = c.copyCheckpoint.load()
		if err != nil {
			return nil, err
		}

		if ccp.UploadId != "" && !c.isUploadIdValid() {
			ccp.UploadId = ""
			ccp.PartETagList = make([]*CompletedPart, 0)
			ccp.remove()
		}
	}

	if ccp.UploadId == "" {
		ccp.UploadId, err = c.initUploadId()
		if err != nil {
			return nil, err
		}
		ccp.dump()
	}

	fileSize := ccp.SrcObjectSize
	partSize := ccp.PartSize
	totalPartNum := (fileSize-1)/partSize + 1
	tasks := make(chan CopyPartTask, totalPartNum)

	var i int64
	for i = 0; i < totalPartNum; i++ {
		partNum := i + 1
		offset := (partNum - 1) * partSize
		actualPartSize := c.getActualPartSize(fileSize, partSize, partNum)
		partETag := c.getPartETag(partNum)
		if partETag != nil {
			c.publishProgress(actualPartSize)
		} else {
			uploadPartTask := CopyPartTask{
				partNumber:     partNum,
				offset:         offset,
				actualPartSize: actualPartSize,
			}
			tasks <- uploadPartTask
		}
	}
	close(tasks)

	var wg sync.WaitGroup
	for i = 0; i < aws.ToLong(c.copyFileRequest.TaskNum); i++ {
		wg.Add(1)
		go c.runTask(tasks, &wg)
	}
	wg.Wait()

	if c.error != nil {
		return nil, c.error
	}

	completedMultipartUpload := c.getMultipartUploadParts()
	resp, err := c.completeMultipartUpload(completedMultipartUpload)
	if err != nil {
		return nil, err
	}

	return c.getCopyFileOutput(resp), nil
}

func (c *Copier) getCopyFileOutput(resp *CompleteMultipartUploadOutput) *CopyFileOutput {
	return &CopyFileOutput{
		Bucket:            resp.Bucket,
		Key:               resp.Key,
		ETag:              resp.ETag,
		ChecksumCRC64ECMA: resp.ChecksumCRC64ECMA,
	}
}

func (c *Copier) getPartSize(fileSize int64, originPartSize int64) int64 {
	partSize := originPartSize
	totalPartNum := (fileSize-1)/partSize + 1
	for totalPartNum > MaxPartNum {
		partSize += originPartSize
		totalPartNum = (fileSize-1)/partSize + 1
	}
	return partSize
}

func (c *Copier) getActualPartSize(fileSize int64, partSize int64, partNum int64) int64 {
	offset := (partNum - 1) * partSize
	actualPartSize := partSize
	if offset+partSize >= fileSize {
		actualPartSize = fileSize - offset
	}
	return actualPartSize
}

func (c *Copier) getPartETag(partNumber int64) *CompletedPart {
	for _, partETag := range c.copyCheckpoint.PartETagList {
		if *partETag.PartNumber == partNumber {
			return partETag
		}
	}
	return nil
}

type CopyPartTask struct {
	partNumber int64

	offset int64

	actualPartSize int64
}

func (c *Copier) runTask(tasks <-chan CopyPartTask, wg *sync.WaitGroup) {
	defer wg.Done()
	for task := range tasks {
		if c.error != nil {
			return
		}

		partETag, err := c.copyPart(task)
		if err != nil {
			c.setError(err)
			return
		}

		c.updatePart(partETag)
	}
}

func (c *Copier) copyPart(task CopyPartTask) (CompletedPart, error) {
	request := c.copyFileRequest
	ccp := c.copyCheckpoint
	var partETag CompletedPart

	start := task.offset
	end := task.offset + task.actualPartSize - 1
	resp, err := c.client.UploadPartCopyWithContext(c.context, &UploadPartCopyInput{
		Bucket:                         aws.String(ccp.BucketName),
		Key:                            aws.String(ccp.ObjectKey),
		SourceBucket:                   aws.String(ccp.SrcBucketName),
		SourceKey:                      aws.String(ccp.SrcObjectKey),
		UploadID:                       aws.String(ccp.UploadId),
		PartNumber:                     aws.Long(task.partNumber),
		CopySourceRange:                aws.String(fmt.Sprintf("bytes=%d-%d", start, end)),
		CopySourceIfMatch:              request.CopySourceIfMatch,
		CopySourceIfNoneMatch:          request.CopySourceIfNoneMatch,
		CopySourceIfModifiedSince:      request.CopySourceIfModifiedSince,
		CopySourceIfUnmodifiedSince:    request.CopySourceIfUnmodifiedSince,
		CopySourceSSECustomerAlgorithm: request.CopySourceSSECustomerAlgorithm,
		CopySourceSSECustomerKey:       request.CopySourceSSECustomerKey,
		CopySourceSSECustomerKeyMD5:    request.CopySourceSSECustomerKeyMD5,
		SSECustomerAlgorithm:           request.SSECustomerAlgorithm,
		SSECustomerKey:                 request.SSECustomerKey,
		SSECustomerKeyMD5:              request.SSECustomerKeyMD5,
	})

	if err != nil {
		return partETag, err
	}

	partETag.PartNumber = aws.Long(task.partNumber)
	partETag.ETag = resp.CopyPartResult.ETag
	partETag.ChecksumCRC64ECMA = resp.CopyPartResult.ChecksumCRC64ECMA
	c.publishProgress(task.actualPartSize)

	return partETag, nil
}

func (c *Copier) updatePart(partETag CompletedPart) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.copyCheckpoint.PartETagList = append(c.copyCheckpoint.PartETagList, &partETag)
	c.copyCheckpoint.dump()
}

func (c *Copier) setError(err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.error == nil {
		c.error = err
	}
}

func (c *Copier) getMultipartUploadParts() *CompletedMultipartUpload {
	partETags := c.copyCheckpoint.PartETagList
	// 按照PartNumber排序
	sort.Sort(CompletedParts(partETags))
	return &CompletedMultipartUpload{
		Parts: partETags,
	}
}

func (c *Copier) completeMultipartUpload(completedMultipartUpload *CompletedMultipartUpload) (*CompleteMultipartUploadOutput, error) {
	resp, err := c.client.CompleteMultipartUploadWithContext(c.context, &CompleteMultipartUploadInput{
		Bucket:          c.copyFileRequest.Bucket,
		Key:             c.copyFileRequest.Key,
		UploadID:        aws.String(c.copyCheckpoint.UploadId),
		MultipartUpload: completedMultipartUpload,
		ForbidOverwrite: c.copyFileRequest.ForbidOverwrite,
	})
	if err != nil {
		return nil, err
	}
	c.copyCheckpoint.remove()
	return resp, err
}

func (c *Copier) publishProgress(actualPartSize int64) {
	if c.copyFileRequest.ProgressFn != nil {
		atomic.AddInt64(&c.CompletedSize, actualPartSize)
		c.copyFileRequest.ProgressFn(actualPartSize, c.CompletedSize, c.copyCheckpoint.SrcObjectSize)
	}
}

func (c *Copier) headObject() (map[string]*string, error) {
	request := c.copyFileRequest
	resp, err := c.client.HeadObjectWithContext(c.context, &HeadObjectInput{
		Bucket:               request.SourceBucket,
		Key:                  request.SourceKey,
		IfModifiedSince:      request.CopySourceIfModifiedSince,
		IfUnmodifiedSince:    request.CopySourceIfUnmodifiedSince,
		IfMatch:              request.CopySourceIfMatch,
		IfNoneMatch:          request.CopySourceIfNoneMatch,
		SSECustomerAlgorithm: request.CopySourceSSECustomerAlgorithm,
		SSECustomerKey:       request.CopySourceSSECustomerKey,
		SSECustomerKeyMD5:    request.CopySourceSSECustomerKeyMD5,
	})
	if err != nil {
		return nil, err
	}
	return resp.Metadata, err
}

func (c *Copier) initUploadId() (string, error) {
	request := c.copyFileRequest
	input := &CreateMultipartUploadInput{
		Bucket:               request.Bucket,
		Key:                  request.Key,
		ACL:                  request.ACL,
		CacheControl:         request.CacheControl,
		ContentDisposition:   request.ContentDisposition,
		ContentEncoding:      request.ContentEncoding,
		ContentType:          request.ContentType,
		Expires:              request.Expires,
		Metadata:             request.Metadata,
		StorageClass:         request.StorageClass,
		Tagging:              request.Tagging,
		ForbidOverwrite:      request.ForbidOverwrite,
		GrantRead:            request.GrantRead,
		GrantFullControl:     request.GrantFullControl,
		ServerSideEncryption: request.ServerSideEncryption,
		SSECustomerAlgorithm: request.SSECustomerAlgorithm,
		SSECustomerKey:       request.SSECustomerKey,
		SSECustomerKeyMD5:    request.SSECustomerKeyMD5,
	}
	if !strings.EqualFold(aws.ToString(request.MetadataDirective), "REPLACE") {
		input.CacheControl = c.copyObjectMeta[HTTPHeaderCacheControl]
		input.ContentDisposition = c.copyObjectMeta[HTTPHeaderContentDisposition]
		input.ContentEncoding = c.copyObjectMeta[HTTPHeaderContentEncoding]
		input.ContentType = c.copyObjectMeta[HTTPHeaderContentType]
		expires, err := time.Parse("Mon, 02 Jan 2006 15:04:05 GMT", aws.ToString(c.copyObjectMeta[HTTPHeaderExpires]))
		if err == nil {
			input.Expires = aws.Time(expires)
		}

		metaData := map[string]*string{}
		for k, v := range c.copyObjectMeta {
			if strings.HasPrefix(strings.ToLower(k), MetaPrefix) {
				metaData[k] = v
			}
		}
		input.Metadata = metaData
	}
	if !strings.EqualFold(aws.ToString(request.TaggingDirective), "REPLACE") {
		taggingResp, err := c.client.GetObjectTaggingWithContext(c.context, &GetObjectTaggingInput{
			Bucket: request.SourceBucket,
			Key:    request.SourceKey,
		})
		if err != nil {
			return "", err
		}

		tagStr := taggingResp.Tagging.ToString()
		if tagStr != "" {
			input.Tagging = aws.String(tagStr)
		}
	}
	if c.copyFileRequest.ACL == nil {
		aclResp, err := c.client.GetObjectACLWithContext(c.context, &GetObjectACLInput{
			Bucket: request.SourceBucket,
			Key:    request.SourceKey,
		})
		if err != nil {
			return "", err
		}

		input.ACL = aws.String(GetCannedACL(aclResp.Grants))
	}
	if c.copyFileRequest.StorageClass == nil {
		input.StorageClass = c.copyObjectMeta[HTTPHeaderAmzStorageClass]
	}

	resp, err := c.client.CreateMultipartUploadWithContext(c.context, input)
	if err != nil {
		return "", err
	}
	return aws.ToString(resp.UploadID), nil
}

func (c *Copier) isUploadIdValid() bool {
	_, err := c.client.ListPartsWithContext(c.context, &ListPartsInput{
		Bucket:   c.copyFileRequest.Bucket,
		Key:      c.copyFileRequest.Key,
		UploadID: aws.String(c.copyCheckpoint.UploadId),
	})
	if err != nil && strings.Contains(err.Error(), "NoSuchUpload") {
		return false
	}
	return true
}

// CopyDirInput 目录复制输入参数。
type CopyDirInput struct {
	// 源存储桶名称，必填。
	SourceBucket *string `type:"string" required:"true"`

	// 源对象名称前缀，列举源对象以此开头。默认为""。
	SourcePrefix *string `type:"string"`

	// 目标存储桶名称，必填。
	Bucket *string `location:"uri" locationName:"Bucket" type:"string" required:"true"`

	// 目标对象名称前缀，目标Key为 Prefix + 源Key去除SourcePrefix后的部分。默认为""。
	Prefix *string `type:"string"`

	// 分块大小，默认5MB。
	PartSize *int64 `type:"integer"`

	// 文件并发数，即同时复制的文件数，默认3。
	Jobs *int64 `type:"integer"`

	// 块并发数，即单文件分块复制并发数，默认3。
	Parallel *int64 `type:"integer"`

	// 目录复制跳过策略，默认Never不跳过。可选值：IfExists/IfSizeEquals/IfNewer/IfNewerAndSizeEquals/IfCrc64Equals。
	SkipRule *string `type:"string"`

	// 是否不记录复制成功文件详情，默认记录。大目录复制时设为true可节省内存。
	IgnoreSuccessFiles *bool `type:"boolean"`

	// 是否启用断点续传，默认不启用。
	EnableCheckpoint *bool `type:"boolean"`

	// 断点续传记录文件的存放目录。
	CheckpointDir *string `type:"string"`

	// 对象的预设ACL。
	ACL *string `location:"header" locationName:"x-amz-acl" type:"string"`

	// 指定请求/响应链上的缓存行为。
	CacheControl *string `location:"header" locationName:"Cache-Control" type:"string"`

	// 指定对象的展示信息。
	ContentDisposition *string `location:"header" locationName:"Content-Disposition" type:"string"`

	// 指定已应用于对象的内容编码。
	ContentEncoding *string `location:"header" locationName:"Content-Encoding" type:"string"`

	// 描述对象数据格式的标准MIME类型。
	ContentType *string `location:"header" locationName:"Content-Type" type:"string"`

	// 存储在KS3中的对象元数据。
	Metadata map[string]*string `location:"headers" locationName:"x-amz-meta-" type:"map"`

	// 指定元数据是复制源对象的还是用请求中提供的元数据替换。
	MetadataDirective *string `location:"header" locationName:"x-amz-metadata-directive" type:"string"`

	// 对象的存储类型，默认为STANDARD。
	StorageClass *string `location:"header" locationName:"x-amz-storage-class" type:"string"`

	// 指定对象标签。
	Tagging *string `location:"header" locationName:"x-amz-tagging" type:"string"`

	// 指定目标对象标签的设置方式，默认COPY。
	TaggingDirective *string `location:"header" locationName:"x-amz-tagging-directive" type:"string"`

	// 服务端加密算法，如AES256。
	ServerSideEncryption *string `location:"header" locationName:"x-amz-server-side-encryption" type:"string"`

	// 目标对象加密算法。
	SSECustomerAlgorithm *string `location:"header" locationName:"x-amz-server-side-encryption-customer-algorithm" type:"string"`

	// 目标对象加密密钥。
	SSECustomerKey *string `location:"header" locationName:"x-amz-server-side-encryption-customer-key" type:"string"`

	// 目标对象加密密钥的MD5摘要。
	SSECustomerKeyMD5 *string `location:"header" locationName:"x-amz-server-side-encryption-customer-key-MD5" type:"string"`

	// 源对象解密算法。
	CopySourceSSECustomerAlgorithm *string `location:"header" locationName:"x-amz-copy-source-server-side-encryption-customer-algorithm" type:"string"`

	// 源对象解密密钥。
	CopySourceSSECustomerKey *string `location:"header" locationName:"x-amz-copy-source-server-side-encryption-customer-key" type:"string"`

	// 源对象解密密钥的MD5摘要。
	CopySourceSSECustomerKeyMD5 *string `location:"header" locationName:"x-amz-copy-source-server-side-encryption-customer-key-MD5" type:"string"`

	// 目录复制进度回调，每次文件完成时调用。
	ProgressFn DirProgressFunc `location:"function"`
}

// CopyDir 复制KS3目录。
func (c *S3) CopyDir(request *CopyDirInput) (*DirResult, error) {
	return c.CopyDirWithContext(context.Background(), request)
}

// CopyDirWithContext 复制KS3目录，支持上下文取消。
func (c *S3) CopyDirWithContext(ctx context.Context, request *CopyDirInput) (*DirResult, error) {
	return newDirCopier(c, ctx, request).copyDir()
}

// CopyDirAcrossRegion 跨区域复制KS3目录。
func (c *S3) CopyDirAcrossRegion(request *CopyDirInput, dstClient *S3) (*DirResult, error) {
	return c.CopyDirAcrossRegionWithContext(context.Background(), request, dstClient)
}

// CopyDirAcrossRegionWithContext 跨区域复制KS3目录，支持上下文取消。
func (c *S3) CopyDirAcrossRegionWithContext(ctx context.Context, request *CopyDirInput, dstClient *S3) (*DirResult, error) {
	return newDirCopier(c, ctx, request, dstClient).copyDir()
}

// DirCopier 目录复制实现。
type DirCopier struct {
	srcClient   *S3           // 源端KS3客户端。
	dstClient   *S3           // 目标端KS3客户端，跨区域时使用。
	context     context.Context // 上下文，用于取消。
	request     *CopyDirInput // 复制输入参数。
	producerErr error         // 生产者错误。
	done        chan struct{} // 生产者完成信号。

	*TransferManager
}

func newDirCopier(srcClient *S3, ctx context.Context, request *CopyDirInput, dstClient ...*S3) *DirCopier {
	var target *S3
	if len(dstClient) > 0 {
		target = dstClient[0]
	}
	return &DirCopier{
		srcClient:       srcClient,
		dstClient:       target,
		context:         ctx,
		request:         request,
		done:            make(chan struct{}),
		TransferManager: newTransferManager(request.ProgressFn),
	}
}

func (d *DirCopier) copyDir() (*DirResult, error) {
	if err := d.validate(); err != nil {
		return nil, err
	}

	jobs := aws.ToLong(d.request.Jobs)
	fileCh := make(chan dirFileInfo, DefaultFileChanSize)

	go d.produceObjects(fileCh)

	var workerWg sync.WaitGroup
	var i int64
	for i = 0; i < jobs; i++ {
		workerWg.Add(1)
		go d.runWorker(fileCh, &workerWg)
	}

	workerWg.Wait()
	<-d.done

	if d.producerErr != nil {
		return nil, d.producerErr
	}
	return d.setResult()
}

func (d *DirCopier) validate() error {
	request := d.request
	if request == nil {
		return errors.New("copy dir request is required")
	}

	if aws.ToString(request.SourceBucket) == "" {
		return errors.New("source bucket is required")
	}

	if aws.ToString(request.Bucket) == "" {
		return errors.New("bucket is required")
	}

	if request.SourcePrefix == nil {
		request.SourcePrefix = aws.String("")
	}

	if request.Prefix == nil {
		request.Prefix = aws.String("")
	}

	if request.PartSize == nil {
		request.PartSize = aws.Long(DefaultPartSize)
	} else if aws.ToLong(request.PartSize) < MinPartSize {
		request.PartSize = aws.Long(MinPartSize)
	} else if aws.ToLong(request.PartSize) > MaxPartSize {
		request.PartSize = aws.Long(MaxPartSize)
	}

	if aws.ToLong(request.Parallel) <= 0 {
		request.Parallel = aws.Long(DefaultTaskNum)
	}

	if aws.ToLong(request.Jobs) <= 0 {
		request.Jobs = aws.Long(DefaultJobs)
	}

	if request.SkipRule == nil {
		request.SkipRule = aws.String(SkipNever)
	}

	return nil
}

func (d *DirCopier) produceObjects(fileCh chan<- dirFileInfo) {
	defer close(d.done)
	d.producerErr = d.listSourceObjects(fileCh)
	close(fileCh)
}

func (d *DirCopier) listSourceObjects(fileCh chan<- dirFileInfo) error {
	sourcePrefix := aws.ToString(d.request.SourcePrefix)
	keyPrefix := aws.ToString(d.request.Prefix)

	paginator := d.srcClient.NewListObjectsPaginator(&ListObjectsInput{
		Bucket: d.request.SourceBucket,
		Prefix: aws.String(sourcePrefix),
	})

	for paginator.HasNext() {
		select {
		case <-d.context.Done():
			return d.context.Err()
		default:
		}

		resp, err := paginator.NextPageWithContext(d.context)
		if err != nil {
			return err
		}

		for _, obj := range resp.Contents {
			srcKey := aws.ToString(obj.Key)
			// 跳过目录标记对象
			if strings.HasSuffix(srcKey, "/") {
				continue
			}
			// 计算目标Key
			dstKey := srcKey
			if sourcePrefix != "" && strings.HasPrefix(srcKey, sourcePrefix) {
				dstKey = keyPrefix + srcKey[len(sourcePrefix):]
			} else {
				dstKey = keyPrefix + srcKey
			}

			select {
			case fileCh <- dirFileInfo{
				filePath:     dstKey,
				objectKey:    srcKey,
				objectSize:   aws.ToLong(obj.Size),
				lastModified: getTimeValue(obj.LastModified),
			}:
			case <-d.context.Done():
				return d.context.Err()
			}
		}
	}
	return nil
}

func (d *DirCopier) runWorker(fileCh <-chan dirFileInfo, wg *sync.WaitGroup) {
	defer wg.Done()
	for fi := range fileCh {
		select {
		case <-d.context.Done():
			return
		default:
		}
		if err := d.copySingleFile(fi); err != nil {
			d.TransferManager.addFailure(fi.filePath, fi.objectKey, fi.objectSize, err)
		}
	}
}

func (d *DirCopier) copySingleFile(fi dirFileInfo) error {
	fileSize := fi.objectSize

	if skip := d.shouldSkipCopy(fi); skip {
		d.TransferManager.addSkip(fi.filePath, fi.objectKey, fileSize)
		return nil
	}

	input := &CopyFileInput{
		Bucket:                         d.request.Bucket,
		Key:                            aws.String(fi.filePath),
		SourceBucket:                   d.request.SourceBucket,
		SourceKey:                      aws.String(fi.objectKey),
		PartSize:                       d.request.PartSize,
		TaskNum:                        d.request.Parallel,
		EnableCheckpoint:               d.request.EnableCheckpoint,
		CheckpointDir:                  d.request.CheckpointDir,
		ACL:                            d.request.ACL,
		CacheControl:                   d.request.CacheControl,
		ContentDisposition:             d.request.ContentDisposition,
		ContentEncoding:                d.request.ContentEncoding,
		ContentType:                    d.request.ContentType,
		Metadata:                       d.request.Metadata,
		MetadataDirective:              d.request.MetadataDirective,
		StorageClass:                   d.request.StorageClass,
		Tagging:                        d.request.Tagging,
		TaggingDirective:               d.request.TaggingDirective,
		ForbidOverwrite:                aws.Boolean(false),
		ServerSideEncryption:           d.request.ServerSideEncryption,
		SSECustomerAlgorithm:           d.request.SSECustomerAlgorithm,
		SSECustomerKey:                 d.request.SSECustomerKey,
		SSECustomerKeyMD5:              d.request.SSECustomerKeyMD5,
		CopySourceSSECustomerAlgorithm: d.request.CopySourceSSECustomerAlgorithm,
		CopySourceSSECustomerKey:       d.request.CopySourceSSECustomerKey,
		CopySourceSSECustomerKeyMD5:    d.request.CopySourceSSECustomerKeyMD5,
	}

	var err error
	if d.dstClient != nil {
		_, err = d.srcClient.CopyFileAcrossRegionWithContext(d.context, input, d.dstClient)
	} else {
		_, err = d.srcClient.CopyFileWithContext(d.context, input)
	}
	if err != nil {
		return err
	}
	d.TransferManager.addSuccess(fi.filePath, fi.objectKey, fileSize, aws.ToBoolean(d.request.IgnoreSuccessFiles))
	return nil
}

func (d *DirCopier) shouldSkipCopy(fi dirFileInfo) bool {
	rule := aws.ToString(d.request.SkipRule)
	if rule == "" || rule == SkipNever {
		return false
	}

	// Head目标对象判断是否已存在
	dstClient := d.srcClient
	if d.dstClient != nil {
		dstClient = d.dstClient
	}
	resp, err := dstClient.HeadObjectWithContext(d.context, &HeadObjectInput{
		Bucket: d.request.Bucket,
		Key:    aws.String(fi.filePath),
	})
	if err != nil || resp == nil {
		return false
	}

	switch rule {
	case SkipIfExists:
		return true
	case SkipIfSizeEquals:
		return aws.ToLong(resp.ContentLength) == fi.objectSize
	case SkipIfNewer:
		return resp.LastModified != nil && !fi.lastModified.After(*resp.LastModified)
	case SkipIfNewerAndSizeEquals:
		return resp.LastModified != nil && !fi.lastModified.After(*resp.LastModified) && aws.ToLong(resp.ContentLength) == fi.objectSize
	case SkipIfCrc64Equals:
		if resp.Metadata == nil {
			return false
		}
		dstCrc := aws.ToString(resp.Metadata[HTTPHeaderAmzChecksumCrc64ecma])
		if dstCrc == "" {
			return false
		}
		// Head源对象取CRC64
		srcResp, srcErr := d.srcClient.HeadObjectWithContext(d.context, &HeadObjectInput{
			Bucket: d.request.SourceBucket,
			Key:    aws.String(fi.objectKey),
		})
		if srcErr != nil || srcResp.Metadata == nil {
			return false
		}
		srcCrc := aws.ToString(srcResp.Metadata[HTTPHeaderAmzChecksumCrc64ecma])
		return srcCrc != "" && srcCrc == dstCrc
	}
	return false
}
