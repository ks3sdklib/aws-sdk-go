package s3

import (
	"context"
	"errors"
	"fmt"
	"github.com/ks3sdklib/aws-sdk-go/aws"
	"github.com/ks3sdklib/aws-sdk-go/internal/crc"
	"hash"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type DownloadFileInput struct {
	// The name of the bucket.
	Bucket *string `location:"uri" locationName:"Bucket" type:"string" required:"true"`

	// Object key of the object.
	Key *string `location:"uri" locationName:"Key" type:"string" required:"true"`

	// The path of the file to be downloaded.
	DownloadFile *string `type:"string" locationName:"DownloadFile" required:"true"`

	// The size of each part.
	PartSize *int64 `type:"integer" locationName:"PartSize"`

	// The number of tasks to download the file.
	TaskNum *int64 `type:"integer" locationName:"TaskNum"`

	// Whether to enable checkpoint.
	EnableCheckpoint *bool `type:"boolean" locationName:"EnableCheckpoint"`

	// The directory to store the checkpoint file.
	CheckpointDir *string `type:"string" locationName:"CheckpointDir"`

	// The checkpoint file path.
	CheckpointFile *string `type:"string" locationName:"CheckpointFile"`

	// 下载的范围，range[0]为开始位置，range[1]为结束位置
	// range[0] 小于 0 时表示从文件头开始下载
	// range[1] 小于 0 时表示下载到文件末尾
	// range[0] 和 range[1] 都小于 0 时表示下载整个文件
	// range[0] 和 range[1] 都大于等于 0 时表示下载指定范围的文件
	// range[0] 大于 range[1] 且 range[1] 非负时表示下载整个文件
	// 例如：
	// range=[0, 99] 表示下载文件的前100个字节
	// range=[100, 199] 表示下载文件的第101个字节至第200个字节
	// range=[100, -1] 表示下载文件的第101个字节至文件末尾
	// range=[-1, 100] 表示下载文件的后100个字节
	// Downloads the specified range bytes of an object.
	Range []int64 `locationName:"Range" type:"list"`

	// Sets the Content-Type header of the response.
	ResponseContentType *string `location:"querystring" locationName:"response-content-type" type:"string"`

	// Sets the Content-Language header of the response.
	ResponseContentLanguage *string `location:"querystring" locationName:"response-content-language" type:"string"`

	// Sets the Expires header of the response.
	ResponseExpires *time.Time `location:"querystring" locationName:"response-expires" type:"timestamp" timestampFormat:"iso8601"`

	// Sets the Cache-Control header of the response.
	ResponseCacheControl *string `location:"querystring" locationName:"response-cache-control" type:"string"`

	// Sets the Content-Disposition header of the response
	ResponseContentDisposition *string `location:"querystring" locationName:"response-content-disposition" type:"string"`

	// Sets the Content-Encoding header of the response.
	ResponseContentEncoding *string `location:"querystring" locationName:"response-content-encoding" type:"string"`

	// Return the object only if it has been modified since the specified time,
	// otherwise return a 304 (not modified).
	IfModifiedSince *time.Time `location:"header" locationName:"If-Modified-Since" type:"timestamp" timestampFormat:"rfc822"`

	// Return the object only if it has not been modified since the specified time,
	// otherwise return a 412 (precondition failed).
	IfUnmodifiedSince *time.Time `location:"header" locationName:"If-Unmodified-Since" type:"timestamp" timestampFormat:"rfc822"`

	// Return the object only if its entity tag (ETag) is the same as the one specified,
	// otherwise return a 412 (precondition failed).
	IfMatch *string `location:"header" locationName:"If-Match" type:"string"`

	// Return the object only if its entity tag (ETag) is different from the one
	// specified, otherwise return a 304 (not modified).
	IfNoneMatch *string `location:"header" locationName:"If-None-Match" type:"string"`

	// Specify the encoding type of the client.
	// If you want to compress and transmit the returned content using gzip,
	// you need to add a request header: Accept-Encoding:gzip。
	// KS3 will determine whether to return gzip compressed data based on the
	// Content-Type and Object size (not less than 1 KB) of the object.
	// Value: gzip、br、deflate
	AcceptEncoding *string `location:"header" locationName:"Accept-Encoding" type:"string"`

	// Specifies the algorithm to use to when encrypting the object, eg: AES256.
	SSECustomerAlgorithm *string `location:"header" locationName:"x-amz-server-side-encryption-customer-algorithm" type:"string"`

	// Specifies the customer-provided encryption key for KS3 to use in encrypting data.
	SSECustomerKey *string `location:"header" locationName:"x-amz-server-side-encryption-customer-key" type:"string"`

	// Specifies the 128-bit MD5 digest of the encryption key according to RFC 1321.
	SSECustomerKeyMD5 *string `location:"header" locationName:"x-amz-server-side-encryption-customer-key-MD5" type:"string"`

	// Bandwidth limit for single-part download, in bits. For example, 10 * 1024 * 1024 * 8 means 10MB/s.
	TrafficLimit *int64 `location:"header" locationName:"x-kss-traffic-limit" type:"integer"`

	// Progress callback function
	ProgressFn aws.ProgressFunc `location:"function"`
}

type DownloadFileOutput struct {
	Bucket *string

	Key *string

	ETag *string

	ChecksumCRC64ECMA *string

	ObjectMeta map[string]*string
}

func (c *S3) DownloadFile(request *DownloadFileInput) (*DownloadFileOutput, error) {
	return c.DownloadFileWithContext(context.Background(), request)
}

func (c *S3) DownloadFileWithContext(ctx context.Context, request *DownloadFileInput) (*DownloadFileOutput, error) {
	return newDownloader(c, ctx, request).downloadFile()
}

type Downloader struct {
	client *S3

	context context.Context

	downloadFileRequest *DownloadFileInput

	downloadCheckpoint *DownloadCheckpoint

	CompletedSize int64

	downloadFileSize int64

	downloadFileMeta map[string]*string

	mu sync.Mutex

	error error
}

func newDownloader(s3 *S3, ctx context.Context, request *DownloadFileInput) *Downloader {
	return &Downloader{
		client:              s3,
		context:             ctx,
		downloadFileRequest: request,
	}
}

func (d *Downloader) downloadFile() (*DownloadFileOutput, error) {
	err := d.validate()
	if err != nil {
		return nil, err
	}

	d.downloadFileMeta, err = d.headObject()
	if err != nil {
		return nil, err
	}

	dcp, err := newDownloadCheckpoint(d)
	if err != nil {
		return nil, err
	}
	d.downloadCheckpoint = dcp

	if aws.ToBoolean(d.downloadFileRequest.EnableCheckpoint) {
		cpFilePath := aws.ToString(d.downloadFileRequest.CheckpointFile)
		if cpFilePath == "" {
			cpFilePath, err = generateDownloadCpFilePath(d.downloadFileRequest)
			if err != nil {
				return nil, err
			}
		}
		dcp.CpFilePath = cpFilePath

		err = dcp.load()
		if err != nil {
			return nil, err
		}

		if !FileExists(dcp.DownloadFilePath + TempFileSuffix) {
			dcp.PartETagList = make([]*CompletedPart, 0)
			dcp.remove()
		}
	}

	err = d.createDownloadDir(dcp.DownloadFilePath + TempFileSuffix)
	if err != nil {
		return nil, err
	}

	objectRange := d.getObjectRange()
	d.downloadFileSize = objectRange[1] - objectRange[0] + 1
	partSize := aws.ToLong(d.downloadFileRequest.PartSize)
	totalPartNum := (d.downloadFileSize-1)/partSize + 1
	tasks := make(chan DownloadPartTask, totalPartNum)

	var i int64
	for i = 0; i < totalPartNum; i++ {
		partNum := i + 1
		start := objectRange[0] + i*partSize
		end := Min(start+partSize-1, objectRange[1])
		actualPartSize := end - start + 1
		if d.getPartETag(partNum) != nil {
			d.publishProgress(actualPartSize)
		} else {
			downloadPartTask := DownloadPartTask{
				partNumber:     partNum,
				start:          start,
				end:            end,
				actualPartSize: actualPartSize,
			}
			tasks <- downloadPartTask
		}
	}
	close(tasks)

	var wg sync.WaitGroup
	for i = 0; i < aws.ToLong(d.downloadFileRequest.TaskNum); i++ {
		wg.Add(1)
		go d.runTask(tasks, &wg)
	}
	wg.Wait()

	if d.error != nil {
		return nil, d.error
	}

	if d.downloadFileRequest.Range == nil && d.client.Config.CrcCheckEnabled {
		clientCrc64 := d.getCrc64Ecma(dcp.PartETagList)
		serverCrc64, _ := strconv.ParseUint(aws.ToString(d.downloadFileMeta[HTTPHeaderAmzChecksumCrc64ecma]), 10, 64)
		d.client.Config.LogDebug("check file crc64, client crc64:%d, server crc64:%d", clientCrc64, serverCrc64)
		if serverCrc64 != 0 && clientCrc64 != serverCrc64 {
			return nil, errors.New(fmt.Sprintf("crc64 check failed, client crc64:%d, server crc64:%d", clientCrc64, serverCrc64))
		}
	}

	err = d.complete()
	if err != nil {
		return nil, err
	}

	return d.getDownloadFileOutput(), nil
}

func (d *Downloader) validate() error {
	request := d.downloadFileRequest
	if request == nil {
		return errors.New("download file request is required")
	}

	if aws.ToString(request.Bucket) == "" {
		return errors.New("bucket is required")
	}

	if aws.ToString(request.Key) == "" {
		return errors.New("key is required")
	}

	err := d.normalizeDownloadPath()
	if err != nil {
		return err
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

func (d *Downloader) getDownloadFileOutput() *DownloadFileOutput {
	return &DownloadFileOutput{
		Bucket:            d.downloadFileRequest.Bucket,
		Key:               d.downloadFileRequest.Key,
		ETag:              d.downloadFileMeta[HTTPHeaderEtag],
		ChecksumCRC64ECMA: d.downloadFileMeta[HTTPHeaderAmzChecksumCrc64ecma],
		ObjectMeta:        d.downloadFileMeta,
	}
}

func (d *Downloader) getActualPartSize(fileSize int64, partSize int64, partNum int64) int64 {
	offset := (partNum - 1) * partSize
	actualPartSize := partSize
	if offset+partSize >= fileSize {
		actualPartSize = fileSize - offset
	}
	return actualPartSize
}

func (d *Downloader) getPartETag(partNumber int64) *CompletedPart {
	for _, partETag := range d.downloadCheckpoint.PartETagList {
		if *partETag.PartNumber == partNumber {
			return partETag
		}
	}
	return nil
}

type DownloadPartTask struct {
	partNumber int64

	actualPartSize int64

	start int64

	end int64
}

func (d *Downloader) runTask(tasks <-chan DownloadPartTask, wg *sync.WaitGroup) {
	defer wg.Done()
	for task := range tasks {
		if d.error != nil {
			return
		}

		partETag, err := d.downloadPart(task)
		if err != nil {
			d.setError(err)
			return
		}

		d.updatePart(partETag)
	}
}

func (d *Downloader) downloadPart(task DownloadPartTask) (CompletedPart, error) {
	request := d.downloadFileRequest
	dcp := d.downloadCheckpoint
	tempFilePath := dcp.DownloadFilePath + TempFileSuffix
	var completedPart CompletedPart
	resp, err := d.client.GetObjectWithContext(d.context, &GetObjectInput{
		Bucket:                     aws.String(dcp.BucketName),
		Key:                        aws.String(dcp.ObjectKey),
		Range:                      aws.String(fmt.Sprintf("bytes=%d-%d", task.start, task.end)),
		ResponseContentType:        request.ResponseContentType,
		ResponseContentLanguage:    request.ResponseContentLanguage,
		ResponseExpires:            request.ResponseExpires,
		ResponseCacheControl:       request.ResponseCacheControl,
		ResponseContentDisposition: request.ResponseContentDisposition,
		ResponseContentEncoding:    request.ResponseContentEncoding,
		IfModifiedSince:            request.IfModifiedSince,
		IfUnmodifiedSince:          request.IfUnmodifiedSince,
		IfMatch:                    request.IfMatch,
		IfNoneMatch:                request.IfNoneMatch,
		AcceptEncoding:             request.AcceptEncoding,
		SSECustomerAlgorithm:       request.SSECustomerAlgorithm,
		SSECustomerKey:             request.SSECustomerKey,
		SSECustomerKeyMD5:          request.SSECustomerKeyMD5,
		TrafficLimit:               request.TrafficLimit,
	})

	if err != nil {
		return completedPart, err
	}
	defer resp.Body.Close()

	var crc64 hash.Hash64
	crc64 = crc.NewCRC(crc.CrcTable(), 0)
	resp.Body = aws.TeeReader(resp.Body, crc64, task.actualPartSize, nil)

	fd, err := os.OpenFile(tempFilePath, os.O_WRONLY|os.O_CREATE, FilePermMode)
	if err != nil {
		return completedPart, err
	}
	defer fd.Close()

	_, err = fd.Seek((task.partNumber-1)*dcp.PartSize, io.SeekStart)
	if err != nil {
		return completedPart, err
	}

	_, err = io.Copy(fd, resp.Body)
	if err != nil {
		return completedPart, err
	}

	completedPart.PartNumber = aws.Long(task.partNumber)
	completedPart.ChecksumCRC64ECMA = aws.String(strconv.FormatUint(crc64.Sum64(), 10))
	d.publishProgress(task.actualPartSize)

	return completedPart, nil
}

func (d *Downloader) updatePart(partETag CompletedPart) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.downloadCheckpoint.PartETagList = append(d.downloadCheckpoint.PartETagList, &partETag)
	d.downloadCheckpoint.dump()
}

func (d *Downloader) setError(err error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.error == nil {
		d.error = err
	}
}

func (d *Downloader) publishProgress(actualPartSize int64) {
	if d.downloadFileRequest.ProgressFn != nil {
		atomic.AddInt64(&d.CompletedSize, actualPartSize)
		d.downloadFileRequest.ProgressFn(actualPartSize, d.CompletedSize, d.downloadFileSize)
	}
}

func (d *Downloader) getCrc64Ecma(parts []*CompletedPart) uint64 {
	if parts == nil || len(parts) == 0 {
		return 0
	}

	sort.Sort(CompletedParts(d.downloadCheckpoint.PartETagList))

	crcTemp, _ := strconv.ParseUint(*parts[0].ChecksumCRC64ECMA, 10, 64)
	for i := 1; i < len(parts); i++ {
		crc2, _ := strconv.ParseUint(*parts[i].ChecksumCRC64ECMA, 10, 64)
		partSize := d.getActualPartSize(d.downloadFileSize, aws.ToLong(d.downloadFileRequest.PartSize), *parts[i].PartNumber)
		crcTemp = crc.CRC64Combine(crcTemp, crc2, (uint64)(partSize))
	}

	return crcTemp
}

func (d *Downloader) complete() error {
	fileName := aws.ToString(d.downloadFileRequest.DownloadFile)
	tempFileName := fileName + TempFileSuffix
	err := os.Rename(tempFileName, fileName)
	if err != nil {
		return err
	}
	d.downloadCheckpoint.remove()
	return nil
}

func (d *Downloader) headObject() (map[string]*string, error) {
	request := d.downloadFileRequest
	resp, err := d.client.HeadObjectWithContext(d.context, &HeadObjectInput{
		Bucket:               request.Bucket,
		Key:                  request.Key,
		IfModifiedSince:      request.IfModifiedSince,
		IfUnmodifiedSince:    request.IfUnmodifiedSince,
		IfMatch:              request.IfMatch,
		IfNoneMatch:          request.IfNoneMatch,
		SSECustomerAlgorithm: request.SSECustomerAlgorithm,
		SSECustomerKey:       request.SSECustomerKey,
		SSECustomerKeyMD5:    request.SSECustomerKeyMD5,
	})
	if err != nil {
		return nil, err
	}
	return resp.Metadata, err
}

func (d *Downloader) createDownloadDir(filePath string) error {
	dir := filepath.Dir(filePath)
	if !DirExists(dir) {
		err := os.MkdirAll(dir, DirPermMode)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *Downloader) getObjectRange() []int64 {
	objectRange := d.downloadFileRequest.Range
	objectSize := d.downloadCheckpoint.ObjectSize
	if objectRange == nil {
		return []int64{0, objectSize - 1}
	}

	if !d.isValidRange(objectRange, objectSize) {
		d.client.Config.LogWarn("Invalid range value: %v, ignore it and request for entire object", objectRange)
		return []int64{0, objectSize - 1}
	}
	objectStart := objectRange[0]
	objectEnd := objectRange[1]

	if objectStart < 0 {
		return []int64{objectSize - objectEnd, objectSize - 1}
	}

	if objectEnd < 0 {
		return []int64{objectStart, objectSize - 1}
	}

	return []int64{objectStart, Min(objectEnd, objectSize-1)}
}

func (d *Downloader) isValidRange(objectRange []int64, objectSize int64) bool {
	if len(objectRange) != 2 {
		return false
	}

	objectStart := objectRange[0]
	objectEnd := objectRange[1]

	if objectStart < 0 && objectEnd < 0 || objectEnd >= 0 && objectStart > objectEnd {
		return false
	}

	return objectStart < objectSize
}

func (d *Downloader) normalizeDownloadPath() error {
	downloadPath := aws.ToString(d.downloadFileRequest.DownloadFile)
	if downloadPath == "" {
		downloadPath = aws.ToString(d.downloadFileRequest.Key)
	}
	// 规范化路径
	normalizedPath := filepath.Clean(downloadPath)
	// 获取绝对路径
	absPath, err := filepath.Abs(normalizedPath)
	if err != nil {
		return err
	}
	d.downloadFileRequest.DownloadFile = aws.String(absPath)

	return nil
}

// DownloadDirInput 目录下载输入参数。
type DownloadDirInput struct {
	// 存储桶名称，必填。
	Bucket *string `location:"uri" locationName:"Bucket" type:"string" required:"true"`

	// 对象名称前缀，列举和下载以此开头的对象。默认为""。
	Prefix *string `type:"string"`

	// 本地下载目录路径，默认为当前目录"."。
	DownloadDir *string `type:"string" required:"true"`

	// 分块大小，默认5MB。
	PartSize *int64 `type:"integer"`

	// 单文件分块下载并发数，默认3。
	TaskNum *int64 `type:"integer"`

	// 目录级下载并发数，即同时下载的文件数，默认3。
	Jobs *int64 `type:"integer"`

	// 目录下载跳过策略，默认Never不跳过。可选值：IfExists/IfSizeEquals/IfNewer/IfNewerAndSizeEquals/IfCrc64Equals。
	SkipRule *string `type:"string"`

	// 是否不记录下载成功文件详情，默认记录。大目录下载时设为true可节省内存。
	IgnoreSuccessFiles *bool `type:"boolean"`

	// 是否启用断点续传，默认不启用。
	EnableCheckpoint *bool `type:"boolean"`

	// 断点续传记录文件的存放目录。
	CheckpointDir *string `type:"string"`

	// 指定解密对象时使用的算法。
	SSECustomerAlgorithm *string `location:"header" locationName:"x-amz-server-side-encryption-customer-algorithm" type:"string"`

	// 指定客户提供的加密密钥。
	SSECustomerKey *string `location:"header" locationName:"x-amz-server-side-encryption-customer-key" type:"string"`

	// 指定加密密钥的128位MD5摘要。
	SSECustomerKeyMD5 *string `location:"header" locationName:"x-amz-server-side-encryption-customer-key-MD5" type:"string"`

	// 目录下载进度回调，每次文件完成时调用。
	ProgressFn DirProgressFunc `location:"function"`
}

// DownloadDir 下载KS3目录到本地。
func (c *S3) DownloadDir(request *DownloadDirInput) (*DirResult, error) {
	return c.DownloadDirWithContext(context.Background(), request)
}

// DownloadDirWithContext 下载KS3目录到本地，支持上下文取消。
func (c *S3) DownloadDirWithContext(ctx context.Context, request *DownloadDirInput) (*DirResult, error) {
	return newDirDownloader(c, ctx, request).downloadDir()
}

// DirDownloader 目录下载实现。
type DirDownloader struct {
	client      *S3              // KS3客户端。
	context     context.Context  // 上下文，用于取消。
	request     *DownloadDirInput // 下载输入参数。
	producerErr error            // 生产者错误。
	done        chan struct{}    // 生产者完成信号。

	localDir string // 解析后的本地目录绝对路径。

	*TransferManager
}

func newDirDownloader(s3 *S3, ctx context.Context, request *DownloadDirInput) *DirDownloader {
	return &DirDownloader{
		client:          s3,
		context:         ctx,
		request:         request,
		done:            make(chan struct{}),
		TransferManager: newTransferManager(request.ProgressFn),
	}
}

func (d *DirDownloader) downloadDir() (*DirResult, error) {
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

func (d *DirDownloader) produceObjects(fileCh chan<- dirFileInfo) {
	defer close(d.done)
	d.producerErr = d.listObjects(fileCh)
	close(fileCh)
}

func (d *DirDownloader) validate() error {
	request := d.request
	if request == nil {
		return errors.New("download dir request is required")
	}

	if aws.ToString(request.Bucket) == "" {
		return errors.New("bucket is required")
	}

	if request.Prefix == nil {
		request.Prefix = aws.String("")
	}

	downloadDir, err := toAbs(aws.ToString(request.DownloadDir))
	if err != nil {
		return err
	}

	if !DirExists(downloadDir) {
		if mkErr := os.MkdirAll(downloadDir, DirPermMode); mkErr != nil {
			return mkErr
		}
	}
	d.localDir = downloadDir

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

	if aws.ToLong(request.Jobs) <= 0 {
		request.Jobs = aws.Long(DefaultJobs)
	}

	if request.SkipRule == nil {
		request.SkipRule = aws.String(SkipNever)
	}

	return nil
}

func (d *DirDownloader) listObjects(fileCh chan<- dirFileInfo) error {
	prefix := aws.ToString(d.request.Prefix)
	paginator := d.client.NewListObjectsPaginator(&ListObjectsInput{
		Bucket: d.request.Bucket,
		Prefix: aws.String(prefix),
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
			key := aws.ToString(obj.Key)
			// 跳过目录标记对象
			if strings.HasSuffix(key, "/") {
				continue
			}
			// 计算本地文件路径
			relPath := key
			if prefix != "" && strings.HasPrefix(key, prefix) {
				relPath = key[len(prefix):]
			}
			relPath = strings.TrimPrefix(relPath, "/")
			localPath := filepath.Join(d.localDir, relPath)

			// 创建父目录
			parentDir := filepath.Dir(localPath)
			if !DirExists(parentDir) {
				if mkErr := os.MkdirAll(parentDir, DirPermMode); mkErr != nil {
					return mkErr
				}
			}

			select {
			case fileCh <- dirFileInfo{
				filePath:     localPath,
				objectKey:    key,
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

func (d *DirDownloader) runWorker(fileCh <-chan dirFileInfo, wg *sync.WaitGroup) {
	defer wg.Done()
	for fi := range fileCh {
		select {
		case <-d.context.Done():
			return
		default:
		}
		if err := d.downloadSingleFile(fi); err != nil {
			d.TransferManager.addFailure(fi.filePath, fi.objectKey, fi.objectSize, err)
		}
	}
}

func (d *DirDownloader) downloadSingleFile(fi dirFileInfo) error {
	fileSize := fi.objectSize

	if skip := d.shouldSkipDownload(fi); skip {
		d.TransferManager.addSkip(fi.filePath, fi.objectKey, fileSize)
		return nil
	}

	_, err := d.client.DownloadFileWithContext(d.context, &DownloadFileInput{
		Bucket:               d.request.Bucket,
		Key:                  aws.String(fi.objectKey),
		DownloadFile:         aws.String(fi.filePath),
		PartSize:             d.request.PartSize,
		TaskNum:              d.request.TaskNum,
		EnableCheckpoint:     d.request.EnableCheckpoint,
		CheckpointDir:        d.request.CheckpointDir,
		SSECustomerAlgorithm: d.request.SSECustomerAlgorithm,
		SSECustomerKey:       d.request.SSECustomerKey,
		SSECustomerKeyMD5:    d.request.SSECustomerKeyMD5,
	})
	if err != nil {
		return err
	}
	d.TransferManager.addSuccess(fi.filePath, fi.objectKey, fileSize, aws.ToBoolean(d.request.IgnoreSuccessFiles))
	return nil
}

func (d *DirDownloader) shouldSkipDownload(fi dirFileInfo) bool {
	rule := aws.ToString(d.request.SkipRule)
	if rule == "" || rule == SkipNever {
		return false
	}

	localInfo, err := os.Stat(fi.filePath)
	if err != nil {
		return false
	}
	localSize := localInfo.Size()
	localModTime := localInfo.ModTime()

	switch rule {
	case SkipIfExists:
		return true
	case SkipIfSizeEquals:
		return fi.objectSize == localSize
	case SkipIfNewer:
		return !localModTime.Before(fi.lastModified)
	case SkipIfNewerAndSizeEquals:
		return !localModTime.Before(fi.lastModified) && fi.objectSize == localSize
	case SkipIfCrc64Equals:
		resp, headErr := d.client.HeadObjectWithContext(d.context, &HeadObjectInput{
			Bucket: d.request.Bucket,
			Key:    aws.String(fi.objectKey),
		})
		if headErr != nil || resp.Metadata == nil {
			return false
		}
		serverCrc := aws.ToString(resp.Metadata[HTTPHeaderAmzChecksumCrc64ecma])
		if serverCrc == "" {
			return false
		}
		localCrc := computeLocalCrc64(fi.filePath)
		return localCrc != "" && localCrc == serverCrc
	}
	return false
}


