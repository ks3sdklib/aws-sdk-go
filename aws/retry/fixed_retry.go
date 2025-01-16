package retry

import "time"

// FixedRetry 固定等待时间重试策略
type FixedRetry struct {
	RetryTimes   int // 重试次数
	InitialDelay int // 初始重试间隔时间，单位 ms
}

func NewFixedRetry(retryTimes, initialDelay int) FixedRetry {
	return FixedRetry{
		RetryTimes:   retryTimes,
		InitialDelay: initialDelay,
	}
}

func (r FixedRetry) ShouldRetry(err error) bool {
	return ShouldRetry(err)
}

func (r FixedRetry) MaxRetryTimes() int {
	return r.RetryTimes
}

func (r FixedRetry) RetryDelay(attempts int) time.Duration {
	return time.Duration(r.InitialDelay)
}
