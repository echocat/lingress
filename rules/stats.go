package rules

import (
	"encoding/json"
	"sync/atomic"
	"time"
)

type Statistics struct {
	numberOfUsages uint64
	totalDuration  time.Duration
}

func (instance *Statistics) MarkUsed(duration time.Duration) {
	atomic.AddUint64(&instance.numberOfUsages, 1)
	atomic.AddInt64((*int64)(&instance.totalDuration), int64(duration))
}

func (instance *Statistics) NumberOfUsages() uint64 {
	return atomic.LoadUint64(&instance.numberOfUsages)
}

func (instance *Statistics) TotalDuration() time.Duration {
	return time.Duration(atomic.LoadInt64((*int64)(&instance.totalDuration)))
}

func (instance *Statistics) MarshalJSON() ([]byte, error) {
	v := struct {
		NumberOfUsages uint64        `json:"numberOfUsages"`
		TotalDuration  time.Duration `json:"totalDuration"`
	}{
		NumberOfUsages: instance.NumberOfUsages(),
		TotalDuration:  instance.TotalDuration() / time.Microsecond,
	}
	return json.Marshal(v)
}
