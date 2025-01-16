package retry

import "time"

// NoRetry 不重试
type NoRetry struct {
}

func NewNoRetry() NoRetry {
	return NoRetry{}
}

func (r NoRetry) ShouldRetry(err error) bool {
	return false
}

func (r NoRetry) MaxRetryTimes() int {
	return 0
}

func (r NoRetry) RetryDelay(attempts int) time.Duration {
	return time.Duration(0)
}
