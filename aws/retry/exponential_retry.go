package retry

import (
	"math"
	"time"
)

// ExponentialRetry 指数级增长等待时间重试策略
type ExponentialRetry struct {
	RetryTimes   int // 重试次数
	InitialDelay int // 初始重试间隔时间，单位 ms
}

func NewExponentialRetry(retryTimes, initialDelay int) ExponentialRetry {
	return ExponentialRetry{
		RetryTimes:   retryTimes,
		InitialDelay: initialDelay,
	}
}

func (r ExponentialRetry) ShouldRetry(err error) bool {
	return ShouldRetry(err)
}

func (r ExponentialRetry) MaxRetryTimes() int {
	return r.RetryTimes
}

func (r ExponentialRetry) RetryDelay(attempts int) time.Duration {
	return time.Duration(r.InitialDelay * int(math.Pow(2, float64(attempts-1))))
}
