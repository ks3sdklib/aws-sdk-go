package s3util

import (
	"crypto/md5"
	"encoding/base64"
	"io"
	"os"
)

// GetBase64MD5Str 计算Base64格式字符串的MD5值
func GetBase64MD5Str(str string) string {
	// 创建一个MD5哈希对象
	hash := md5.New()

	// 将字符串转换为字节数组并计算MD5哈希值
	hash.Write([]byte(str))
	md5Hash := hash.Sum(nil)

	// 将MD5哈希值转换为Base64格式
	base64Str := base64.StdEncoding.EncodeToString(md5Hash)

	return base64Str
}

// GetBase64Str 计算Base64格式字符串
func GetBase64Str(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}

// GetBase64FileMD5Str 计算Base64格式文件的MD5值
func GetBase64FileMD5Str(filePath string) (string, error) {
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

	// 将MD5哈希值转换为Base64格式
	base64Str := base64.StdEncoding.EncodeToString(md5Hash)

	return base64Str, err
}
