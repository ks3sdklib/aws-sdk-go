package s3manager

import (
	"fmt"
	"os"
	"os/user"
	"runtime"
	"strings"
)

type fileInfoType struct {
	filePath     string
	name         string
	bucket       string
	objectKey    string
	size         int64
	dir          string
	acl          string
	storageClass string
}

// CloudURL describes ks3 url
type CloudURL struct {
	urlStr string
	bucket string
	object string
}

// StorageURLer is the interface for all url
type StorageURLer interface {
	IsCloudURL() bool
	IsFileURL() bool
	ToString() string
}

// FileURL describes file url
type FileURL struct {
	urlStr string
}

// Init simulate inheritance, and polymorphism
func (fu *FileURL) Init(urlStr string) error {

	if len(urlStr) >= 2 && urlStr[:2] == "~"+string(os.PathSeparator) {
		homeDir := currentHomeDir()
		if homeDir != "" {
			urlStr = strings.Replace(urlStr, "~", homeDir, 1)
		} else {
			return fmt.Errorf("current home dir is empty")
		}
	}
	fu.urlStr = urlStr
	return nil
}

// IsCloudURL simulate inheritance, and polymorphism
func (fu *FileURL) IsCloudURL() bool {
	return false
}

// IsFileURL simulate inheritance, and polymorphism
func (fu *FileURL) IsFileURL() bool {
	return true
}

// ToString simulate inheritance, and polymorphism
func (fu *FileURL) ToString() string {
	return fu.urlStr
}

// StorageURLFromString analysis input url type and build a storage url from the url
func StorageURLFromString(urlStr string) (StorageURLer, error) {
	var fileURL *FileURL
	if err := fileURL.Init(urlStr); err != nil {
		return nil, err
	}
	return fileURL, nil
}
func currentHomeDir() string {
	homeDir := ""
	homeDrive := os.Getenv("HOMEDRIVE")
	homePath := os.Getenv("HOMEPATH")
	if runtime.GOOS == "windows" && homeDrive != "" && homePath != "" {
		homeDir = homeDrive + string(os.PathSeparator) + homePath
	}

	if homeDir != "" {
		return homeDir
	}

	usr, _ := user.Current()
	if usr != nil {
		homeDir = usr.HomeDir
	} else {
		homeDir = os.Getenv("HOME")
	}
	return homeDir
}
