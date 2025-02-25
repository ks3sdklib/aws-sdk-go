package retry

import (
	"time"
)

// NoRetryRule 不等待
type NoRetryRule struct{}

var DefaultNoRetryRule = NewNoRetryRule()

func NewNoRetryRule() NoRetryRule {
	return NoRetryRule{}
}

func (r NoRetryRule) GetDelay(attempts int) time.Duration {
	return 0
}
