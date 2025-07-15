package heart

import (
	"time"
)

type HeartbeatSample struct {
	Timestamp time.Time
	Delay     time.Duration
	Success   bool
}

type HeartbeatRing struct {
	samples []HeartbeatSample
	size    int
	index   int
	full    bool
}

func NewHeartbeatRing(size int) *HeartbeatRing {
	return &HeartbeatRing{
		samples: make([]HeartbeatSample, size),
		size:    size,
	}
}

func (r *HeartbeatRing) Add(success bool, delay time.Duration) {
	r.samples[r.index] = HeartbeatSample{
		Timestamp: time.Now(),
		Delay:     delay,
		Success:   success,
	}
	r.index = (r.index + 1) % r.size
	if r.index == 0 {
		r.full = true
	}
}

func (r *HeartbeatRing) Status(isRun bool) string {
	if !isRun {
		return "❌ 服务未启动"
	}
	if !r.full {
		return "⏳ 连接评估中"
	}
	var success int
	for _, sample := range r.samples {
		if sample.Success {
			success++
		}
	}
	count := len(r.samples)
	failures := count - success
	if failures == 0 {
		return "✅ 连接稳定"
	} else if failures < count/3 {
		return "⚠️ 轻微波动"
	} else if failures < count/2 {
		return "❗ 抖动严重"
	} else if failures < count {
		return "❌ 频繁掉线"
	} else {
		return "❌ 服务异常"
	}
}
