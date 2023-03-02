package limiter

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestLimiter(t *testing.T) {
	cnt := 100
	wait := 10 * time.Millisecond
	limiter := New(float64(cnt))
	defer limiter.Close()

	var success, failed int32
	var wg sync.WaitGroup
	wg.Add(2 * cnt)

	for i := 0; i < 2*cnt; i++ {
		go func(i int) {
			defer wg.Done()
			err := limiter.Run(func() {
				time.Sleep(wait)
			})
			if err == nil {
				atomic.AddInt32(&success, 1)
			} else {
				atomic.AddInt32(&failed, 1)
			}
		}(i)
	}

	wg.Wait()

	if expected := int32(cnt); success != expected {
		t.Fatalf("Fail to set qps to expected cnt. [expected:%v] | [actual:%v]", expected, success)
	}

	if expected := int32(cnt); failed != expected {
		t.Fatalf("Fail to set qps to expected cnt. [expected:%v] | [actual:%v]", expected, failed)
	}
}

func TestLimiterQPS(t *testing.T) {
	const max = 100
	const multiplier = 50
	sleep := 10 * time.Millisecond

	for cnt := 0; cnt < 50; cnt++ {
		func() {
			limiter := New(max)
			defer limiter.Close()

			var wg sync.WaitGroup
			wg.Add(max * multiplier)

			var executed int64

			for i := 0; i < max; i++ {
				for j := 0; j < multiplier; j++ {
					go func() {
						limiter.Run(func() {
							time.Sleep(sleep)
							atomic.AddInt64(&executed, 1)
						})
						wg.Done()
					}()
				}
			}

			wg.Wait()

			if top := max * (time.Second * sleep) / time.Second; executed < max || executed > int64(top) {
				t.Fatalf("Limiter does not limit qps as expected. range:%v - %v, actual:%v", max, top, executed)
			}
		}()
	}
}
