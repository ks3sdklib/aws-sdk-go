package s3

import (
	"fmt"
	"hash"
	"hash/crc64"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// TaskResult 单个文件传输结果。
type TaskResult struct {
	Path string // 本地文件路径。
	Key  string // 对象Key。
	Size int64  // 文件大小。
}

// FileError 单个文件传输失败记录。
type FileError struct {
	Path  string // 本地文件路径。
	Key   string // 对象Key。
	Error error  // 失败原因。
}

// DirResult 目录传输输出参数，上传下载共用。
type DirResult struct {
	TotalNum        int64        // 文件总数。
	TotalFileSize   int64        // 文件总大小。
	SuccessNum      int64        // 成功文件数。
	SuccessFileSize int64        // 成功文件总大小。
	SkipNum         int64        // 跳过文件数。
	SkipFileSize    int64        // 跳过文件总大小。
	FailNum         int64        // 失败文件数。
	FailFileSize    int64        // 失败文件总大小。
	SuccessFiles    []TaskResult // 成功文件详情。
	SkipFiles       []TaskResult // 跳过文件详情。
	ErrorFiles      []FileError  // 失败文件详情。
}

// DirError 目录传输部分失败的错误。
type DirError struct {
	FailNum int64       // 失败文件数。
	Errors  []FileError // 失败文件详情。
}

func (e *DirError) Error() string {
	return fmt.Sprintf("dir transfer failed: %d file(s) failed", e.FailNum)
}

// TransferManager 目录传输公共状态。
type TransferManager struct {
	result DirResult
	mu        sync.Mutex
	progressFn DirProgressFunc
}

// DirProgressFunc 目录传输进度回调函数类型。
type DirProgressFunc func(stat DirResult)

func newDirTransferCore() TransferManager {
	return TransferManager{
		result: DirResult{
			SuccessFiles: make([]TaskResult, 0),
			SkipFiles:    make([]TaskResult, 0),
			ErrorFiles:   make([]FileError, 0),
		},
	}
}

func newTransferManager(fn DirProgressFunc) *TransferManager {
	c := newDirTransferCore()
	c.progressFn = fn
	return &c
}

func (c *TransferManager) snapshot() DirResult {
	return DirResult{
		TotalNum:        atomic.LoadInt64(&c.result.TotalNum),
		TotalFileSize:   atomic.LoadInt64(&c.result.TotalFileSize),
		SuccessNum:      atomic.LoadInt64(&c.result.SuccessNum),
		SuccessFileSize: atomic.LoadInt64(&c.result.SuccessFileSize),
		SkipNum:         atomic.LoadInt64(&c.result.SkipNum),
		SkipFileSize:    atomic.LoadInt64(&c.result.SkipFileSize),
		FailNum:         atomic.LoadInt64(&c.result.FailNum),
		FailFileSize:    atomic.LoadInt64(&c.result.FailFileSize),
	}
}

func (c *TransferManager) addSuccess(path, key string, size int64, ignore bool) {
	atomic.AddInt64(&c.result.SuccessNum, 1)
	atomic.AddInt64(&c.result.SuccessFileSize, size)
	atomic.AddInt64(&c.result.TotalNum, 1)
	atomic.AddInt64(&c.result.TotalFileSize, size)
	if !ignore {
		c.mu.Lock()
		c.result.SuccessFiles = append(c.result.SuccessFiles, TaskResult{Path: path, Key: key, Size: size})
		c.mu.Unlock()
	}
	c.publishProgress()
}

func (c *TransferManager) addSkip(path, key string, size int64) {
	atomic.AddInt64(&c.result.SkipNum, 1)
	atomic.AddInt64(&c.result.SkipFileSize, size)
	atomic.AddInt64(&c.result.TotalNum, 1)
	atomic.AddInt64(&c.result.TotalFileSize, size)
	c.mu.Lock()
	c.result.SkipFiles = append(c.result.SkipFiles, TaskResult{Path: path, Key: key, Size: size})
	c.mu.Unlock()
	c.publishProgress()
}

func (c *TransferManager) addFailure(path, key string, size int64, err error) {
	atomic.AddInt64(&c.result.FailNum, 1)
	atomic.AddInt64(&c.result.FailFileSize, size)
	atomic.AddInt64(&c.result.TotalNum, 1)
	atomic.AddInt64(&c.result.TotalFileSize, size)
	c.mu.Lock()
	c.result.ErrorFiles = append(c.result.ErrorFiles, FileError{Path: path, Key: key, Error: err})
	c.mu.Unlock()
	c.publishProgress()
}

func (c *TransferManager) publishProgress() {
	if c.progressFn != nil {
		c.progressFn(c.snapshot())
	}
}

func (c *TransferManager) setResult() (*DirResult, error) {
	r := &c.result
	if r.FailNum > 0 {
		return r, &DirError{FailNum: r.FailNum, Errors: r.ErrorFiles}
	}
	return r, nil
}

// dirFileInfo 目录中的单个文件信息。
type dirFileInfo struct {
	filePath     string    // 本地文件路径。
	objectKey    string    // 对象Key。
	objectSize   int64     // 文件大小，下载时从ListObjects填充。
	lastModified time.Time  // 对象最后修改时间，下载时从ListObjects填充。
}

func computeLocalCrc64(filePath string) string {
	f, err := os.Open(filePath)
	if err != nil {
		return ""
	}
	defer f.Close()

	crc64Ins := crc64.New(crc64.MakeTable(crc64.ECMA))
	w, _ := crc64Ins.(hash.Hash)
	if _, err := io.Copy(w, f); err != nil {
		return ""
	}
	return strconv.FormatUint(crc64Ins.Sum64(), 10)
}

func toAbs(rootDir string) (string, error) {
	if rootDir == "~" || strings.HasPrefix(rootDir, "~/") {
		currentUser, err := user.Current()
		if err != nil {
			return rootDir, err
		}
		rootDir = strings.Replace(rootDir, "~", currentUser.HomeDir, 1)
	}
	rootDir, err := filepath.Abs(rootDir)
	if err != nil {
		return rootDir, err
	}
	if !strings.HasSuffix(rootDir, string(os.PathSeparator)) && len(rootDir) > 0 {
		rootDir = rootDir + string(os.PathSeparator)
	}
	return rootDir, nil
}

func makeObjectName(rootDir, prefix, filePath string) string {
	filePath = filepath.ToSlash(filePath)
	rootDir = filepath.ToSlash(rootDir)
	relativePath := strings.Replace(filePath, rootDir, "", 1)
	relativePath = strings.TrimPrefix(relativePath, "/")
	return prefix + relativePath
}

func getTimeValue(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}
