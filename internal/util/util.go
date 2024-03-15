package util

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"io"
	"os"
	"regexp"
	"sort"
	"strings"
)

var reTrim = regexp.MustCompile(`\s{2,}`)

func Trim(s string) string {
	return strings.TrimSpace(reTrim.ReplaceAllString(s, " "))
}

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

// SortedKeys returns a sorted slice of keys of a map.
func SortedKeys(m map[string]interface{}) []string {
	i, sorted := 0, make([]string, len(m))
	for k := range m {
		sorted[i] = k
		i++
	}
	sort.Strings(sorted)
	return sorted
}
