package s3util

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"io"
	"os"
)

// GetStrMD5 计算字符串的MD5
func GetStrMD5(str string) string {
	// 创建一个MD5哈希对象
	hash := md5.New()

	// 将字符串转换为字节数组并计算MD5哈希值
	hash.Write([]byte(str))
	md5Hash := hash.Sum(nil)

	// 将MD5哈希值转换为Base64格式
	base64Str := base64.StdEncoding.EncodeToString(md5Hash)

	return base64Str
}

// GetStrBase64 将字符串转换为Base64格式
func GetStrBase64(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}

// GetFileMD5 计算文件的MD5
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
