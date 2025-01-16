package retry

import "time"

// LinearRetry 线性增长等待时间重试策略
type LinearRetry struct {
	RetryTimes   int // 重试次数
	InitialDelay int // 初始重试间隔时间，单位 ms
}

func NewLinearRetry(retryTimes, initialDelay int) LinearRetry {
	return LinearRetry{
		RetryTimes:   retryTimes,
		InitialDelay: initialDelay,
	}
}

func (r LinearRetry) ShouldRetry(err error) bool {
	return ShouldRetry(err)
}

func (r LinearRetry) MaxRetryTimes() int {
	return r.RetryTimes
}

func (r LinearRetry) RetryDelay(attempts int) time.Duration {
	return time.Duration(r.InitialDelay * attempts)
}
