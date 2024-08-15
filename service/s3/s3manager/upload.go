package s3manager

import (
	"bytes"
	"fmt"
	"github.com/ks3sdklib/aws-sdk-go/aws"
	"github.com/ks3sdklib/aws-sdk-go/aws/awserr"
	"github.com/ks3sdklib/aws-sdk-go/aws/awsutil"
	"github.com/ks3sdklib/aws-sdk-go/internal/apierr"
	"github.com/ks3sdklib/aws-sdk-go/internal/crc"
	"github.com/ks3sdklib/aws-sdk-go/service/s3"
	"io"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

var MaxUploadParts = 10000 // max number of parts in a multipart upload

var MaxUploadPartSize int64 = 1024 * 1024 * 1024 * 5 // Max part size, 5GB

var MinUploadPartSize int64 = 1024 * 100 // Min part size, 100KB

var DefaultUploadPartSize int64 = 1024 * 1024 * 5 // Default part size, 5MB

var DefaultUploadConcurrency = 5 // Default number of goroutines

// DefaultUploadOptions The default set of options used when opts is nil in Upload().
var DefaultUploadOptions = &UploadOptions{
	PartSize:          DefaultUploadPartSize,
	Parallel:          DefaultUploadConcurrency,
	Jobs:              3,
	LeavePartsOnError: false,
	S3:                nil,
}

type MultiUploadFailure interface {
	awserr.Error

	// UploadID Returns the upload id for the S3 multipart upload that failed.
	UploadID() string
}

// So that the Error interface type can be included as an anonymous field
// in the multiUploadError struct and not conflict with the error.Error() method.
type awsError awserr.Error

// A multiUploadError wraps the upload ID of a failed s3 multipart upload.
// Composed of BaseError for code, message, and original error
//
// Should be used for an error that occurred failing a S3 multipart upload,
// and an upload ID is available. If an uploadID is not available a more relevant
type multiUploadError struct {
	awsError

	// ID for multipart upload which failed.
	uploadID string
}

// Error returns the string representation of the error.
//
// # See apierr.BaseError ErrorWithExtra for output format
//
// Satisfies the error interface.
func (m multiUploadError) Error() string {
	extra := fmt.Sprintf("upload id: %s", m.uploadID)
	return awserr.SprintError(m.Code(), m.Message(), extra, m.OrigErr())
}

// String returns the string representation of the error.
// Alias for Error to satisfy the stringer interface.
func (m multiUploadError) String() string {
	return m.Error()
}

// UploadID returns the id of the S3 upload which failed.
func (m multiUploadError) UploadID() string {
	return m.uploadID
}

// UploadInput contains all input for upload requests to Amazon S3.
type UploadInput struct {
	// The canned ACL to apply to the object.
	ACL *string `location:"header" locationName:"x-amz-acl" type:"string"`

	Size int64

	Bucket *string `location:"uri" locationName:"Bucket" type:"string" required:"true"`

	// Specifies caching behavior along the request/reply chain.
	CacheControl *string `location:"header" locationName:"Cache-Control" type:"string"`

	// Specifies presentational information for the object.
	ContentDisposition *string `location:"header" locationName:"Content-Disposition" type:"string"`

	// Specifies what content encodings have been applied to the object and thus
	// what decoding mechanisms must be applied to obtain the media-type referenced
	// by the Content-Type header field.
	ContentEncoding *string `location:"header" locationName:"Content-Encoding" type:"string"`

	// The language the content is in.
	ContentLanguage *string `location:"header" locationName:"Content-Language" type:"string"`

	// A standard MIME type describing the format of the object data.
	ContentType *string `location:"header" locationName:"Content-Type" type:"string"`

	// The date and time at which the object is no longer cacheable.
	Expires *time.Time `location:"header" locationName:"Expires" type:"timestamp" timestampFormat:"rfc822"`

	// Gives the grantee READ, READ_ACP, and WRITE_ACP permissions on the object.
	GrantFullControl *string `location:"header" locationName:"x-amz-grant-full-control" type:"string"`

	// Allows grantee to read the object data and its metadata.
	GrantRead *string `location:"header" locationName:"x-amz-grant-read" type:"string"`

	// Allows grantee to read the object ACL.
	GrantReadACP *string `location:"header" locationName:"x-amz-grant-read-acp" type:"string"`

	// Allows grantee to write the ACL for the applicable object.
	GrantWriteACP *string `location:"header" locationName:"x-amz-grant-write-acp" type:"string"`

	Key *string `location:"uri" locationName:"Key" type:"string" required:"true"`

	// A map of metadata to store with the object in S3.
	Metadata map[string]*string `location:"headers" locationName:"x-amz-meta-" type:"map"`

	// Confirms that the requester knows that she or he will be charged for the
	// request. Bucket owners need not specify this parameter in their requests.
	// Documentation on downloading objects from requester pays buckets can be found
	// at http://docs.aws.amazon.com/AmazonS3/latest/dev/ObjectsinRequesterPaysBuckets.html
	RequestPayer *string `location:"header" locationName:"x-amz-request-payer" type:"string"`

	// Specifies the algorithm to use to when encrypting the object (e.g., AES256,
	// aws:kms).
	SSECustomerAlgorithm *string `location:"header" locationName:"x-amz-server-side-encryption-customer-algorithm" type:"string"`

	// Specifies the customer-provided encryption key for Amazon S3 to use in encrypting
	// data. This value is used to store the object, and then it is discarded; Amazon
	// does not store the encryption key. The key must be appropriate for use with
	// the algorithm specified in the x-amz-server-side​-encryption​-customer-algorithm
	// header.
	SSECustomerKey *string `location:"header" locationName:"x-amz-server-side-encryption-customer-key" type:"string"`

	// Specifies the 128-bit MD5 digest of the encryption key according to RFC 1321.
	// Amazon S3 uses this header for a message integrity check to ensure the encryption
	// key was transmitted without error.
	SSECustomerKeyMD5 *string `location:"header" locationName:"x-amz-server-side-encryption-customer-key-MD5" type:"string"`

	// Specifies the AWS KMS key ID to use for object encryption. All GET and PUT
	// requests for an object protected by AWS KMS will fail if not made via SSL
	// or using SigV4. Documentation on configuring any of the officially supported
	// AWS SDKs and CLI can be found at http://docs.aws.amazon.com/AmazonS3/latest/dev/UsingAWSSDK.html#specify-signature-version
	SSEKMSKeyID *string `location:"header" locationName:"x-amz-server-side-encryption-aws-kms-key-id" type:"string"`

	// The Server-side encryption algorithm used when storing this object in S3
	// (e.g., AES256, aws:kms).
	ServerSideEncryption *string `location:"header" locationName:"x-amz-server-side-encryption" type:"string"`

	// The type of storage to use for the object. Defaults to 'STANDARD'.
	StorageClass *string `location:"header" locationName:"x-amz-storage-class" type:"string"`

	// If the bucket is configured as a website, redirects requests for this object
	// to another object in the same bucket or to an external URL. Amazon S3 stores
	// the value of this header in the object metadata.
	WebsiteRedirectLocation *string `location:"header" locationName:"x-amz-website-redirect-location" type:"string"`

	// The readable body payload to send to S3.
	Body io.Reader
}

// UploadOutput represents a response from the Upload() call.
type UploadOutput struct {
	// The URL where the object was uploaded to.
	Location string

	// The ID for a multipart upload to S3. In the case of an error the error
	// can be cast to the MultiUploadFailure interface to extract the upload ID.
	UploadID string

	ETag string
}

// UploadOptions keeps tracks of extra options to pass to an Upload() call.
type UploadOptions struct {
	// The buffer size (in bytes) to use when buffering data into chunks and
	// sending them as parts to KS3. The minimum allowed part size is 5MB, and
	// if this value is set to zero, the DefaultPartSize value will be used.
	PartSize int64

	//Number of concurrent tasks for internal operation of a single file
	Parallel int
	//Number of concurrent tasks in multi-file operation
	Jobs int
	// Setting this value to true will cause the SDK to avoid calling
	// AbortMultipartUpload on a failure, leaving all successfully uploaded
	// parts on S3 for manual recovery.
	//
	// Note that storing parts of an incomplete multipart upload counts towards
	// space usage on S3 and will add additional costs if not cleaned up.
	LeavePartsOnError bool

	//Set whether to upload hidden files
	UploadHidden bool
	//Set whether to upload existing files
	SkipAlreadyFile bool
	// The client to use when uploading to S3. Leave this as nil to use the
	// default S3 client.
	S3 *s3.S3
}

// NewUploader creates a new Uploader object to upload data to S3. Pass in
// an optional opts structure to customize the uploader behavior.
func NewUploader(opts *UploadOptions) *Uploader {
	if opts == nil {
		opts = DefaultUploadOptions
	} else {
		if opts.PartSize == 0 {
			opts.PartSize = DefaultUploadOptions.PartSize
		}
		if opts.Parallel == 0 {
			opts.Parallel = DefaultUploadOptions.Parallel
		}
		if opts.Jobs == 0 {
			opts.Jobs = DefaultUploadOptions.Jobs
		}
	}
	return &Uploader{opts: opts}
}

// The Uploader structure that calls Upload(). It is safe to call Upload()
// on this structure for multiple objects and across concurrent goroutines.
type Uploader struct {
	opts *UploadOptions
}

// Upload uploads an object to S3, intelligently buffering large files into
// smaller chunks and sending them in parallel across multiple goroutines. You
// can configure the buffer size and concurrency through the opts parameter.
//
// If opts is set to nil, DefaultUploadOptions will be used.
//
// It is safe to call this method for multiple objects and across concurrent
// goroutines.
func (u *Uploader) Upload(input *UploadInput) (*UploadOutput, error) {
	return u.UploadWithContext(aws.BackgroundContext(), input)
}

func (u *Uploader) UploadWithContext(ctx aws.Context, input *UploadInput) (*UploadOutput, error) {
	i := uploader{in: input, opts: *u.opts, ctx: ctx}
	return i.upload()
}

// internal structure to manage an upload to S3.
type uploader struct {
	in   *UploadInput
	opts UploadOptions
	ctx  aws.Context

	readerPos int64 // current reader position
	totalSize int64 // set to -1 if the size is not known
}

// internal logic for deciding whether to upload a single part or use a
// multipart upload.
func (u *uploader) upload() (*UploadOutput, error) {
	u.init()

	if u.in.Size != 0 {
		u.opts.PartSize = u.in.Size
	}

	if u.opts.PartSize < MinUploadPartSize {
		msg := fmt.Sprintf("part size must be at least %d bytes", MinUploadPartSize)
		return nil, awserr.New("ConfigError", msg, nil)
	}

	if u.opts.PartSize > MaxUploadPartSize {
		msg := fmt.Sprintf("part size must be at most %d bytes", MaxUploadPartSize)
		return nil, awserr.New("ConfigError", msg, nil)
	}

	// Do one read to determine if we have more than one part
	buf, trunkSize, err := u.nextReader()
	if err == io.EOF || err == io.ErrUnexpectedEOF { // single part
		return u.singlePart(buf)
	} else if err != nil {
		return nil, awserr.New("ReadRequestBody", "read upload data failed", err)
	}

	mu := multiuploader{uploader: u}
	return mu.upload(buf, trunkSize)
}

// init will initialize all default options.
func (u *uploader) init() {
	if u.opts.S3 == nil {
		u.opts.S3 = s3.New(nil)
	}
	if u.opts.Parallel == 0 {
		u.opts.Parallel = DefaultUploadConcurrency
	}
	if u.opts.PartSize == 0 {
		u.opts.PartSize = DefaultUploadPartSize
	}

	// Try to get the total size for some optimizations
	u.initSize()
}

// initSize tries to detect the total stream size, setting u.totalSize. If
// the size is not known, totalSize is set to -1.
func (u *uploader) initSize() {
	u.totalSize = -1

	switch r := u.in.Body.(type) {
	case io.Seeker:
		pos, _ := r.Seek(0, 1)
		defer r.Seek(pos, 0)

		n, err := r.Seek(0, 2)
		if err != nil {
			return
		}
		u.totalSize = n

		// try to adjust partSize if it is too small
		if u.totalSize/u.opts.PartSize >= int64(MaxUploadParts) {
			u.opts.PartSize = u.totalSize / int64(MaxUploadParts)
		}
	}
}

// nextReader returns a seekable reader representing the next packet of data.
// This operation increases the shared u.readerPos counter, but note that it
// does not need to be wrapped in a mutex because nextReader is only called
// from the main thread.
func (u *uploader) nextReader() (io.ReadSeeker, int64, error) {
	switch r := u.in.Body.(type) {
	case io.ReaderAt:
		var err error

		n := u.opts.PartSize
		if u.totalSize >= 0 {
			bytesLeft := u.totalSize - u.readerPos

			if bytesLeft == 0 {
				err = io.EOF
				n = bytesLeft
			} else if bytesLeft <= u.opts.PartSize {
				err = io.ErrUnexpectedEOF
				n = bytesLeft
			}
		}

		buf := io.NewSectionReader(r, u.readerPos, n)
		u.readerPos += n

		return buf, n, err

	default:
		packet := make([]byte, u.opts.PartSize)
		n, err := io.ReadFull(u.in.Body, packet)
		u.readerPos += int64(n)

		return bytes.NewReader(packet[0:n]), int64(n), err
	}
}

// singlePart contains upload logic for uploading a single chunk via
// a regular PutObject request. Multipart requests require at least two
// parts, or at least 5MB of data.
func (u *uploader) singlePart(buf io.ReadSeeker) (*UploadOutput, error) {
	params := &s3.PutObjectInput{}
	awsutil.Copy(params, u.in)
	params.Body = buf

	req, _ := u.opts.S3.PutObjectRequest(params)
	req.SetContext(u.ctx)
	if err := req.Send(); err != nil {
		return nil, err
	}

	url := req.HTTPRequest.URL.String()
	return &UploadOutput{Location: url}, nil
}

// internal structure to manage a specific multipart upload to S3.
type multiuploader struct {
	*uploader
	wg       sync.WaitGroup
	m        sync.Mutex
	err      error
	uploadID string
	parts    CompleteUploadParts
}

// keeps track of a single chunk of data being sent to S3.
type chunk struct {
	buf       io.ReadSeeker
	num       int64
	trunkSize int64
}

// CompleteUploadParts completedParts is a wrapper to make parts sortable by their part number,
// since KS3 required this list to be sent in sorted order.
type CompleteUploadParts []*CompleteUploadPart

func (a CompleteUploadParts) Len() int      { return len(a) }
func (a CompleteUploadParts) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a CompleteUploadParts) Less(i, j int) bool {
	return *a[i].Part.PartNumber < *a[j].Part.PartNumber
}

// upload will perform a multipart upload using the firstBuf buffer containing
// the first chunk of data.
func (u *multiuploader) upload(firstBuf io.ReadSeeker, firstTrunkSize int64) (*UploadOutput, error) {
	params := &s3.CreateMultipartUploadInput{}
	awsutil.Copy(params, u.in)

	// Create the multipart
	resp, err := u.opts.S3.CreateMultipartUploadWithContext(u.ctx, params)
	if err != nil {
		return nil, err
	}
	u.uploadID = *resp.UploadID

	// Create the workers
	ch := make(chan chunk, u.opts.Parallel)
	for i := 0; i < u.opts.Parallel; i++ {
		u.wg.Add(1)
		go u.readChunk(ch)
	}

	// Send part 1 to the workers
	var num int64 = 1
	ch <- chunk{buf: firstBuf, num: num, trunkSize: firstTrunkSize}

	// Read and queue the rest of the parts
	for u.geterr() == nil {
		// This upload exceeded maximum number of supported parts, error now.
		if num > int64(MaxUploadParts) {
			msg := fmt.Sprintf("exceeded total allowed parts (%d). "+
				"Adjust PartSize to fit in this limit", MaxUploadParts)
			u.seterr(awserr.New("TotalPartsExceeded", msg, nil))
			break
		}

		num++

		buf, trunkSize, err := u.nextReader()
		if err == io.EOF {
			break
		}

		ch <- chunk{buf: buf, num: num, trunkSize: trunkSize}

		if err != nil && err != io.ErrUnexpectedEOF {
			u.seterr(awserr.New(
				"ReadRequestBody",
				"read multipart upload data failed",
				err))
			break
		}
	}

	// Close the channel, wait for workers, and complete upload
	close(ch)
	u.wg.Wait()
	complete := u.complete()

	if err := u.geterr(); err != nil {
		return nil, &multiUploadError{
			awsError: awserr.New(
				"MultipartUpload",
				"upload multipart failed",
				err),
			uploadID: u.uploadID,
		}
	}
	return &UploadOutput{
		Location: *complete.Location,
		UploadID: u.uploadID,
	}, nil
}

// readChunk runs in worker goroutines to pull chunks off of the ch channel
// and send() them as UploadPart requests.
func (u *multiuploader) readChunk(ch chan chunk) {
	defer u.wg.Done()
	for {
		data, ok := <-ch

		if !ok {
			break
		}

		if u.geterr() == nil {
			if err := u.send(data); err != nil {
				u.seterr(err)
			}
		}
	}
}

type CompleteUploadPart struct {
	Part      *s3.CompletedPart
	TrunkSize int64
}

// send performs an UploadPart request and keeps track of the completed
// part information.
func (u *multiuploader) send(c chunk) error {
	resp, err := u.opts.S3.UploadPartWithContext(u.ctx, &s3.UploadPartInput{
		Bucket:     u.in.Bucket,
		Key:        u.in.Key,
		Body:       c.buf,
		UploadID:   &u.uploadID,
		PartNumber: &c.num,
	})

	if err != nil {
		return err
	}

	partNumber := c.num
	completed := &s3.CompletedPart{ETag: resp.ETag, ChecksumCRC64ECMA: resp.ChecksumCRC64ECMA, PartNumber: &partNumber}

	completeUploadPart := &CompleteUploadPart{Part: completed, TrunkSize: c.trunkSize}

	u.m.Lock()
	u.parts = append(u.parts, completeUploadPart)
	u.m.Unlock()

	return nil
}

// geterr is a thread-safe getter for the error object
func (u *multiuploader) geterr() error {
	u.m.Lock()
	defer u.m.Unlock()

	return u.err
}

// seterr is a thread-safe setter for the error object
func (u *multiuploader) seterr(e error) {
	u.m.Lock()
	defer u.m.Unlock()

	u.err = e
}

// fail will abort the multipart unless LeavePartsOnError is set to true.
func (u *multiuploader) fail() {
	if u.opts.LeavePartsOnError {
		return
	}

	u.opts.S3.AbortMultipartUploadWithContext(u.ctx, &s3.AbortMultipartUploadInput{
		Bucket:   u.in.Bucket,
		Key:      u.in.Key,
		UploadID: &u.uploadID,
	})
}

func (u *multiuploader) allParts() []*s3.CompletedPart {
	var ps []*s3.CompletedPart
	for _, part := range u.parts {
		ps = append(ps, part.Part)
	}
	return ps
}

func (u *multiuploader) combineCRCInUploadParts(parts []*CompleteUploadPart) uint64 {
	if parts == nil || len(parts) == 0 {
		return 0
	}

	crcTemp, _ := strconv.ParseUint(*parts[0].Part.ChecksumCRC64ECMA, 10, 64)
	for i := 1; i < len(parts); i++ {
		crc2, _ := strconv.ParseUint(*parts[i].Part.ChecksumCRC64ECMA, 10, 64)
		crcTemp = crc.CRC64Combine(crcTemp, crc2, (uint64)(parts[i].TrunkSize))
	}

	return crcTemp
}

func (u *multiuploader) checkMultipartUploadCrc64(clientCrc uint64, res *s3.CompleteMultipartUploadOutput) error {
	var err error
	serverCrc := uint64(0)
	if res.Metadata["X-Amz-Checksum-Crc64ecma"] != nil {
		serverCrc, _ = strconv.ParseUint(*res.Metadata["X-Amz-Checksum-Crc64ecma"], 10, 64)
	}

	u.opts.S3.Config.WriteLog(aws.LogOn, "client crc:%d, server crc:%d\n", clientCrc, serverCrc)

	if serverCrc != 0 && clientCrc != serverCrc {
		err = apierr.New("CRCCheckError", fmt.Sprintf("client crc and server crc do not match, request id:[%s]", *res.Metadata["X-Kss-Request-Id"]), nil)
		u.opts.S3.Config.WriteLog(aws.LogOn, "error:%s\n", err.Error())
	}

	return err
}

// complete successfully completes a multipart upload and returns the response.
func (u *multiuploader) complete() *s3.CompleteMultipartUploadOutput {
	if u.geterr() != nil {
		u.fail()
		return nil
	}

	// Parts must be sorted in PartNumber order.
	sort.Sort(u.parts)

	resp, err := u.opts.S3.CompleteMultipartUploadWithContext(u.ctx, &s3.CompleteMultipartUploadInput{
		Bucket:          u.in.Bucket,
		Key:             u.in.Key,
		UploadID:        &u.uploadID,
		MultipartUpload: &s3.CompletedMultipartUpload{Parts: u.allParts()},
	})
	if err != nil {
		u.seterr(err)
		u.fail()
	}

	if u.opts.S3.Config.CrcCheckEnabled {
		clientCrc := u.combineCRCInUploadParts(u.parts)
		err = u.checkMultipartUploadCrc64(clientCrc, resp)
		if err != nil {
			u.seterr(err)
			u.fail()
		}
	}

	return resp
}

type UploadDirInput struct {
	// The path to the folder to be uploaded.
	RootDir string
	// The name of the bucket.
	Bucket string
	// Prefix of the object.
	Prefix string
	// The ACL of the object.
	ACL string
	// The StorageClass of the object.
	StorageClass string
}

func (u *Uploader) UploadDir(input *UploadDirInput) error {
	return u.UploadDirWithContext(aws.BackgroundContext(), input)
}

func (u *Uploader) UploadDirWithContext(ctx aws.Context, input *UploadDirInput) error {
	if input.RootDir == "" {
		return apierr.New("InvalidParameter", "RootDir is required", nil)
	}
	if input.Bucket == "" {
		return apierr.New("InvalidParameter", "Bucket is required", nil)
	}
	if !strings.HasSuffix(input.Prefix, "/") && len(input.Prefix) > 0 {
		input.Prefix = input.Prefix + "/"
	}
	rootDir, err := u.toAbs(input.RootDir)
	if err != nil {
		return err
	}

	chFiles := make(chan fileInfoType)
	var consumerWgc sync.WaitGroup
	var fileCounter FileCounter
	for i := 0; i < u.opts.Jobs; i++ {
		consumerWgc.Add(1)
		go func() {
			defer consumerWgc.Done()
			for file := range chFiles {
				fileCounter.addTotalNum(1)
				u.upload(ctx, file, &fileCounter)
			}
		}()
	}
	filepath.Walk(rootDir, func(path string, file os.FileInfo, _ error) (err error) {
		if !file.IsDir() {
			if !awsutil.IsHidden(path) || u.opts.UploadHidden {
				chFiles <- fileInfoType{
					filePath:     path,
					name:         file.Name(),
					bucket:       input.Bucket,
					size:         file.Size(),
					dir:          rootDir,
					objectKey:    makeObjectName(rootDir, input.Prefix, path),
					acl:          input.ACL,
					storageClass: input.StorageClass,
				}
			}
		}
		return
	})
	close(chFiles)
	consumerWgc.Wait()
	fmt.Printf("Done. Total num: %d, success num: %d, fail num: %d \n", fileCounter.TotalNum, fileCounter.SuccessNum, fileCounter.FailNum)
	return nil
}

func (u *Uploader) toAbs(rootDir string) (string, error) {
	if rootDir == "~" || strings.HasPrefix(rootDir, "~/") {
		currentUser, err := user.Current()
		if err != nil {
			log.Fatalf(err.Error())
			return rootDir, err
		}

		homeDir := currentUser.HomeDir
		rootDir = strings.Replace(rootDir, "~", homeDir, 1)
	}
	rootDir, err := filepath.Abs(rootDir)
	if err != nil {
		return rootDir, err
	}
	if !strings.HasSuffix(rootDir, "/") && len(rootDir) > 0 {
		rootDir = rootDir + "/"
	}
	return rootDir, nil
}

func (u *Uploader) uploadFile(ctx aws.Context, fileIfo fileInfoType, call func(success bool)) {
	file, err := os.Open(fileIfo.filePath)
	if err != nil {
		log.Print(err)
	}
	defer file.Close()
	if u.opts.SkipAlreadyFile {
		resp, err := u.opts.S3.HeadObjectWithContext(ctx, &s3.HeadObjectInput{
			Bucket: aws.String(fileIfo.bucket),
			Key:    aws.String(fileIfo.objectKey),
		})
		if err == nil && len(*resp.ETag) > 0 {
			call(true)
			return
		}
	}
	_, err = u.UploadWithContext(ctx, &UploadInput{
		Body:         file,
		Bucket:       aws.String(fileIfo.bucket),
		Key:          aws.String(fileIfo.objectKey),
		ACL:          aws.String(fileIfo.acl),
		StorageClass: aws.String(fileIfo.storageClass),
	})
	call(err == nil)

}
func makeObjectName(RootDir, Prefix, filePath string) string {
	resDir := strings.Replace(filePath, RootDir, "", 1)
	objectName := Prefix + resDir
	return objectName
}

func (u *Uploader) upload(ctx aws.Context, file fileInfoType, fileCounter *FileCounter) {
	u.uploadFile(ctx, file, func(success bool) {
		if success {
			fileCounter.addSuccessNum(1)
			fmt.Println(fmt.Sprintf("%s successfully uploaded ", file.objectKey))
		} else {
			fileCounter.addFailNum(1)
			fmt.Println(fmt.Sprintf("%s upload failed", file.objectKey))
		}
	})
}
