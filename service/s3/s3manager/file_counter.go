package s3manager

import "sync/atomic"

type FileCounter struct {
	TotalNum   int64
	SuccessNum int64
	FailNum    int64
}

func (fc *FileCounter) addTotalNum(num int64) {
	atomic.AddInt64(&fc.TotalNum, num)
}

func (fc *FileCounter) addSuccessNum(num int64) {
	atomic.AddInt64(&fc.SuccessNum, num)
}

func (fc *FileCounter) addFailNum(num int64) {
	atomic.AddInt64(&fc.FailNum, num)
}
