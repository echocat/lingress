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

func (this *Statistics) MarkUsed(duration time.Duration) {
	atomic.AddUint64(&this.numberOfUsages, 1)
	atomic.AddInt64((*int64)(&this.totalDuration), int64(duration))
}

func (this *Statistics) NumberOfUsages() uint64 {
	return atomic.LoadUint64(&this.numberOfUsages)
}

func (this *Statistics) TotalDuration() time.Duration {
	return time.Duration(atomic.LoadInt64((*int64)(&this.totalDuration)))
}

func (this *Statistics) MarshalJSON() ([]byte, error) {
	v := struct {
		NumberOfUsages uint64        `json:"numberOfUsages"`
		TotalDuration  time.Duration `json:"totalDuration"`
	}{
		NumberOfUsages: this.NumberOfUsages(),
		TotalDuration:  this.TotalDuration() / time.Microsecond,
	}
	return json.Marshal(v)
}
