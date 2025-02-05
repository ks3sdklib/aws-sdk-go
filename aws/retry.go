package aws

import (
	"errors"
	"github.com/ks3sdklib/aws-sdk-go/aws/awserr"
	"github.com/ks3sdklib/aws-sdk-go/internal/apierr"
	"math"
	"math/rand"
	"time"
)

// ExponentialRetryRules 指数级增长等待时间重试策略
func ExponentialRetryRules(attempts int) time.Duration {
	return time.Duration(10 * int(math.Pow(2, float64(attempts-1))))
}

// LinearRetryRules 线性增长等待时间重试策略
func LinearRetryRules(attempts int) time.Duration {
	return time.Duration(10 * attempts)
}

// FixedRetryRules 固定等待时间重试策略
func FixedRetryRules(attempts int) time.Duration {
	return time.Duration(10)
}

// RandomRetryRules 随机等待时间重试策略
func RandomRetryRules(attempts int) time.Duration {
	return time.Duration(rand.Intn(10))
}

// NoRetryRules 不等待
func NoRetryRules(attempts int) time.Duration {
	return time.Duration(0)
}

// retryableCodes is a collection of service response codes which are retry-able
// without any further action.
var retryableCodes = map[string]struct{}{
	"RequestError":                           {},
	"ProvisionedThroughputExceededException": {},
	"Throttling":                             {},
}

// credsExpiredCodes is a collection of error codes which signify the credentials
// need to be refreshed. Expired tokens require refreshing of credentials, and
// resigning before the request can be retried.
var credsExpiredCodes = map[string]struct{}{
	"ExpiredToken":          {},
	"ExpiredTokenException": {},
	"RequestExpired":        {},
}

func isCodeRetryable(code string) bool {
	if _, ok := retryableCodes[code]; ok {
		return true
	}

	return isCodeExpiredCreds(code)
}

func isCodeExpiredCreds(code string) bool {
	_, ok := credsExpiredCodes[code]
	return ok
}

// 重试错误码
var retryErrorCodes = []int{
	408, // RequestTimeout
	429, // TooManyRequests
}

// ShouldRetry 判断是否需要重试
// 重试条件：
// 1.状态码为5xx
// 2.状态码在retryErrorCodes中
// 3.错误码在retryableCodes中
func ShouldRetry(err error) bool {
	var requestError *apierr.RequestError
	if errors.As(err, &requestError) {
		if requestError.StatusCode() >= 500 {
			return true
		}

		for _, code := range retryErrorCodes {
			if requestError.StatusCode() == code {
				return true
			}
		}
	}

	if err, ok := err.(awserr.Error); ok {
		return isCodeRetryable(err.Code())
	}

	return false
}
