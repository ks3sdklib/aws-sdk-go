package s3

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/ks3sdklib/aws-sdk-go/aws"
	"golang.org/x/crypto/pbkdf2"
	"io"
	"regexp"
	"strings"
	"time"
)

type GenerateShareUrlInput struct {
	// 分享的桶名称
	Bucket *string `location:"uri" locationName:"bucket" type:"string" required:"true"`

	// 分享的前缀，如果不指定，则分享整个桶下的对象
	Prefix *string `location:"querystring" locationName:"prefix" type:"string"`

	// 分享链接的过期时间，单位为秒，默认值为900秒（15分钟）
	Expires *int64 `locationName:"expires" type:"integer"`

	// 提取码，格式为6位字母和数字组合，如果设置了提取码，则生成带提取码的分享链接
	AccessCode *string `locationName:"accessCode" type:"string"`

	// 分享策略
	Policy *string `locationName:"policy" type:"string"`

	// 扩展的查询参数
	ExtendQueryParams map[string]*string `location:"extendQueryParams" type:"map"`
}

func (c *S3) GenerateShareUrl(input *GenerateShareUrlInput) (string, error) {
	op := &aws.Operation{
		HTTPMethod: "",
		HTTPPath:   "/{Bucket}",
	}

	if input == nil {
		input = &GenerateShareUrlInput{}
	}

	if IsEmpty(input.Bucket) {
		return "", errors.New("bucket is required")
	}

	if input.Expires == nil {
		input.Expires = aws.Long(15 * 60) // 默认15分钟
	}

	if input.ExtendQueryParams == nil {
		input.ExtendQueryParams = make(map[string]*string)
	}

	req := c.newRequest(op, input, nil)
	req.SignType = "share"

	var policy string
	if IsEmpty(input.Policy) {
		if IsEmpty(input.Prefix) {
			policy = fmt.Sprintf(`{"conditions":[{"bucket":"%s"}]}`, *input.Bucket)
		} else {
			policy = fmt.Sprintf(`{"conditions":[{"bucket":"%s"},["starts-with","$key","%s"]]}`, *input.Bucket, *input.Prefix)
			input.ExtendQueryParams["prefix"] = input.Prefix
		}
		input.Policy = aws.String(policy)
	}
	input.ExtendQueryParams["X-Amz-Policy"] = aws.String(GetBase64Str(*input.Policy))

	if IsV4Signature(c.Config.SignerVersion) {
		req.ExpireTime = *input.Expires
	} else {
		req.ExpireTime = *input.Expires + time.Now().Unix()
	}
	req.Sign()
	url := req.HTTPRequest.URL.String()

	accessCode := aws.ToString(input.AccessCode)
	if accessCode != "" {
		token, err := EncryptUrlToToken(url, accessCode)
		if err != nil {
			return "", err
		}
		url = fmt.Sprintf("%s?token=%s", ShareUrl, token)
	}

	return url, nil
}

func BuildPolicy(bucketName string, prefixes []string, keys []string) (string, error) {
	if bucketName == "" {
		return "", errors.New("bucketName is required")
	}

	if len(prefixes) == 0 && len(keys) == 0 {
		return "", errors.New("prefixes or keys must be provided")
	}

	var conditions []string
	conditions = append(conditions, fmt.Sprintf(`{"bucket":"%s"}`, bucketName))
	for _, prefix := range prefixes {
		conditions = append(conditions, fmt.Sprintf(`["starts-with","$key","%s"]`, prefix))
	}
	for _, key := range keys {
		conditions = append(conditions, fmt.Sprintf(`["eq","$key","%s"]`, key))
	}
	policy := fmt.Sprintf(`{"conditions":[%s]}`, strings.Join(conditions, ","))
	return policy, nil
}

// EncryptUrlToToken 将分享链接使用提取码进行加密
func EncryptUrlToToken(url string, accessCode string) (string, error) {
	if accessCode == "" {
		return "", errors.New("accessCode is required")
	}

	matched, err := regexp.MatchString("^[a-zA-Z0-9]{6}$", accessCode)
	if err != nil {
		return "", err
	}

	if !matched {
		return "", errors.New("accessCode must be 6 characters long and contain letters and numbers only")
	}

	salt := make([]byte, 16)
	_, err = io.ReadFull(rand.Reader, salt)
	if err != nil {
		return "", err
	}

	key := pbkdf2.Key([]byte(accessCode), salt, 10000, 32, sha256.New)
	ciphertext, err := aesEncrypt(base64.StdEncoding.EncodeToString([]byte(url)), key, []byte(Iv))
	if err != nil {
		return "", err
	}

	token := fmt.Sprintf("%s_%s_%s", ciphertext, base64.StdEncoding.EncodeToString(salt), base64.StdEncoding.EncodeToString([]byte(Iv)))
	return token, err
}

// DecryptTokenToUrl 将token解密为分享链接
func DecryptTokenToUrl(token string, accessCode string) (string, error) {
	tokenSlice := bytes.Split([]byte(token), []byte("_"))
	if len(tokenSlice) != 3 {
		return "", errors.New("invalid token")
	}

	ciphertext, err := base64.StdEncoding.DecodeString(string(tokenSlice[0]))
	if err != nil {
		return "", err
	}

	salt, err := base64.StdEncoding.DecodeString(string(tokenSlice[1]))
	if err != nil {
		return "", err
	}

	iv, err := base64.StdEncoding.DecodeString(string(tokenSlice[2]))
	if err != nil {
		return "", err
	}

	key := pbkdf2.Key([]byte(accessCode), salt, 10000, 32, sha256.New)

	base64Url, err := aesDecrypt(string(ciphertext), key, iv)
	if err != nil {
		return "", err
	}

	url, err := base64.StdEncoding.DecodeString(string(base64Url))
	if err != nil {
		return "", err
	}

	return string(url), nil
}

// AesEncrypt 使用AES加密数据
func aesEncrypt(str string, key, iv []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	plaintext := pkcs7Padding([]byte(str), block.BlockSize())
	blockMode := cipher.NewCBCEncrypter(block, iv)
	ciphertext := make([]byte, len(plaintext))
	blockMode.CryptBlocks(ciphertext, plaintext)

	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// AesDecrypt 使用AES解密数据
func aesDecrypt(ciphertext string, key, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	if len(ciphertext)%block.BlockSize() != 0 {
		return nil, errors.New("ciphertext is not a multiple of the block size")
	}

	blockMode := cipher.NewCBCDecrypter(block, iv)
	plaintext := make([]byte, len(ciphertext))
	blockMode.CryptBlocks(plaintext, []byte(ciphertext))
	plaintext, err = pkcs7UnPadding(plaintext)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// pkcs7Padding 对数据进行PKCS7填充
func pkcs7Padding(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	paddingText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, paddingText...)
}

// pkcs7UnPadding 对数据进行PKCS7去填充
func pkcs7UnPadding(data []byte) ([]byte, error) {
	length := len(data)
	if length == 0 {
		return nil, errors.New("pkcs7: Data is empty")
	}
	padding := int(data[length-1])
	if padding > length {
		return nil, errors.New("pkcs7: Invalid padding")
	}
	return data[:length-padding], nil
}
