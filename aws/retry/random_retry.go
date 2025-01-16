package retry

import (
	"math/rand"
	"time"
)

// RandomRetry 随机等待时间重试策略
type RandomRetry struct {
	RetryTimes   int // 重试次数
	InitialDelay int // 初始重试间隔时间，单位 ms
}

func NewRandomRetry(retryTimes, initialDelay int) RandomRetry {
	return RandomRetry{
		RetryTimes:   retryTimes,
		InitialDelay: initialDelay,
	}
}

func (r RandomRetry) ShouldRetry(err error) bool {
	return ShouldRetry(err)
}

func (r RandomRetry) MaxRetryTimes() int {
	return r.RetryTimes
}

func (r RandomRetry) RetryDelay(attempts int) time.Duration {
	return time.Duration(rand.Intn(r.InitialDelay))
}
