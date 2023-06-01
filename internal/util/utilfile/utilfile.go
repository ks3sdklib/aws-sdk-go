package utilfile

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"os"
)

func GetFileMD5(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	md5Hash := hash.Sum(nil)
	return hex.EncodeToString(md5Hash), nil
}
