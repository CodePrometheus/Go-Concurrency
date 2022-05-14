package limiter

import (
	"Starry-LeakyBucket"
	"math"
)

type Limiter struct {
	bucket *leakybucket.LeakyBucket
}

func New(qps float64) *Limiter {
	max := int64(math.Floor(qps))
	return &Limiter{
		bucket: leakybucket.New(&leakybucket.Config{
			MaxTokens:  max,
			InitTokens: max,
			TokensPer:  qps,
		}),
	}
}

// Run 只在同时并发不超过预设的 max 情况下会执行 f
// 如果同时并发达到了 max，Run 直接返回结果，f 不会被执行
func (l *Limiter) Run(f func()) error {
	err := l.bucket.Get(1)
	if err != nil {
		return nil
	}
	f()
	return nil
}

func (l *Limiter) Close() error {
	return l.bucket.Close()
}
