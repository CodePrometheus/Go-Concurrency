package stabilizer

import (
	leakybucket "Starry-LeakyBucket"
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestStabilizer(t *testing.T) {
	cnt := 100
	wait := 10 * time.Millisecond
	stab := New(cnt)
	defer stab.Close()

	var success, failed int32
	var wg sync.WaitGroup
	wg.Add(2 * cnt)

	for i := 0; i < 2*cnt; i++ {
		go func(i int) {
			defer wg.Done()
			err := stab.Run(func() {
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

	if expected := int32(2 * cnt); success != expected {
		t.Fatalf("Fail to set qps to expected count. [expected:%v] [actual:%v]", expected, success)
	}
	if expected := int32(0); failed != expected {
		t.Fatalf("Fail to set qps to expected count. [expected:%v] [actual:%v]", expected, failed)
	}
}

func TestOutTime(t *testing.T) {
	stab := New(3)
	defer stab.Close()

	var wg sync.WaitGroup
	wg.Add(10)

	for i := 0; i < 10; i++ {
		go func(i int) {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			f := func() {
				time.Sleep(time.Duration(i+3) * time.Second)
			}
			err := stab.RunContext(ctx, f)
			if err != nil {
				println("goroutine:", i, err.Error())
			} else {
				println("goroutine:", i, "OK")
			}
		}(i)
	}

	time.Sleep(5 * time.Second)

	for i := 0; i < 1000; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		err := stab.RunContext(ctx, func() {
			time.Sleep(10 * time.Millisecond)
		})
		if err != nil {
			println(i, err.Error())
			if i > 20 {
				t.Fatalf("Run shouldn't fail. [err:%v]", err)
			}
		} else {
			println(i, "ok")
		}
	}
	wg.Wait()
}

func TestContext(t *testing.T) {
	stab := New(1)
	defer stab.Close()

	ctx := context.Background()
	if err := stab.RunContext(ctx, func() {}); err != nil {
		t.Fatalf("Fail to call RunContext. [err:%v]", err)
	}

	go func() {
		stab.Run(func() {
			time.Sleep(10 * time.Millisecond)
		})
	}()

	time.Sleep(time.Millisecond)
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	if err := stab.RunContext(ctx, func() {}); err != context.DeadlineExceeded {
		t.Fatalf("RunContext should fail with timeout. [err:%v]", err)
	}
}

func TestClose(t *testing.T) {
	stab := New(1)
	stab.Close()

	if err := stab.Run(func() {}); err != leakybucket.ErrClosed {
		t.Fatalf("Run should fail with error. [err:%v]", err)
	}
	if err := stab.RunContext(context.Background(), func() {}); err != leakybucket.ErrClosed {
		t.Fatalf("Run should fail with error. [err:%v]", err)
	}
}

const (
	args1 = 1
	args2 = "2"
	args3 = 3.4
)

var errR2 = errors.New("test")

func testBindFunc(t *testing.T, arg1 int, arg2 string) (r1 float64, r2 error) {
	if t == nil {
		panic("Fail to get t")
	}
	if arg1 != args1 {
		t.Fatalf("Fail to get arg1. [expected:%v] [actual:%v]", args1, arg1)
	}

	if arg2 != args2 {
		t.Fatalf("Fail to get arg2. [expected:%v] [actual:%v]", args2, arg2)
	}
	return args3, errR2
}

func TestBindFunc(t *testing.T) {
	stab := New(1)
	defer stab.Close()

	if f := stab.BindFunc(123); f != nil {
		t.Fatalf("BindFunc should fail. [func:%v]", f)
	}

	f := stab.BindFunc(testBindFunc).(func(t *testing.T, arg1 int, arg2 string) (r1 float64, r2 error))

	go func() {
		stab.Run(func() {
			time.Sleep(10 * time.Millisecond)
		})
	}()

	time.Sleep(time.Millisecond)
	r1, err := f(t, args1, args2)
	if r1 != args3 {
		t.Fatalf("Fail to get r1. [expected:%v] [actual:%v]", args3, r1)
	}

	if err != errR2 {
		t.Fatalf("Fail to get err. [expected:%v] [actual:%v]", errR2, err)
	}
}

func TestQPS(t *testing.T) {
	const max = 100
	const multi = 30
	sleep := 2 * time.Millisecond

	start := time.Now()
	stab := New(max)

	var wg sync.WaitGroup
	wg.Add(max * multi)

	for i := 0; i < multi; i++ {
		for j := 0; j < max; j++ {
			go func() {
				stab.Run(func() {
					time.Sleep(sleep)
					wg.Done()
				})
			}()
		}
	}

	wg.Wait()

	d := time.Now().Sub(start)
	if expected := multi * sleep; d < expected {
		t.Fatalf("Finish too soon. [expected:%v] [actual:%v]", expected, d)
	}
}
