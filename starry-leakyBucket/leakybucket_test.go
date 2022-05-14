package leakybucket

import (
	"context"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	const (
		max = 10
		tps = 500.0
	)
	lb := New(&Config{
		MaxTokens: max,
		TokensPer: tps,
	})
	defer lb.Close()

	requirements := []int64{4, 9, 3, 5, 4}
	expected := []int{2, 0, 4, 3, 1}
	ret := make(chan int, len(requirements))

	for i, n := range requirements {
		go func(i int, n int64) {
			time.Sleep(time.Duration(i) * time.Millisecond)
			err := lb.Wait(n)
			if err != nil {
				t.Fatalf("Fail to wait for tokens. [err:%v] | [i:%v] | [n:%v]", err, i, n)
			}
			ret <- i
		}(i, n)
	}

	for _, i := range expected {
		if r := <-ret; i != r {
			t.Fatalf("Unexpected wait sequence. [expected:%v] | [actual:%v]", i, r)
		}
	}
}

func TestFill(t *testing.T) {
	const (
		max  = 4
		init = 1
	)
	lb := New(&Config{
		MaxTokens:  max,
		InitTokens: init,
		TokensPer:  0,
	})
	defer lb.Close()

	if err := lb.Get(1); err != nil {
		t.Fatalf("Fail to get token, err:%v", err)
	}

	if err := lb.Get(-1); err != nil {
		t.Fatalf("Fail to get token, err:%v", err)
	}

	if err := lb.Get(1); err != ErrNotEnough {
		t.Fatalf("Get token should fail, err:%v", err)
	}

	if err := lb.Set(4); err != nil {
		t.Fatalf("Fail to set token, err:%v", err)
	}

	if err := lb.Set(1); err != ErrOverFlow {
		t.Fatalf("Set token should fail, err:%v", err)
	}

	if err := lb.Set(-1); err != nil {
		t.Fatalf("Fail to set token, err:%v", err)
	}

	if n := lb.Fill(4); n != 0 {
		t.Fatalf("Bucket is full and n should be 0. [n:%v]", n)
	}

	if err := lb.Get(5); err != ErrNotEnough {
		t.Fatalf("Get token should fail, err:%v", err)
	}

	if err := lb.Get(3); err != nil {
		t.Fatalf("Fail to get token, err:%v", err)
	}

	if n := lb.Fill(1); n != 1 {
		t.Fatalf("Bucket is full and n should be 3. [n:%v]", n)
	}

	if n := lb.Fill(3); n != 2 {
		t.Fatalf("Bucket is full and n should be 3. [n:%v]", n)
	}

	if n := lb.Fill(-5); n != 0 {
		t.Fatalf("Bucket is full and n should be 3. [n:%v]", n)
	}
}

func TestResize(t *testing.T) {
	const init = 1
	lb := New(&Config{
		MaxTokens:  init,
		InitTokens: init,
	})
	defer lb.Close()

	if err := lb.Get(2); err != ErrNotEnough {
		t.Fatalf("Get token should fail. [err:%v]", err)
	}

	lb.ResizeMax(2)
	lb.Set(1)

	if err := lb.Get(2); err != nil {
		t.Fatalf("Get token fail. [err:%v]", err)
	}

	if err := lb.Get(1); err != ErrNotEnough {
		t.Fatalf("Get token should fail. [err:%v]", err)
	}

	lb.Fill(3)

	if err := lb.Get(2); err != nil {
		t.Fatalf("Get token fail. [err:%v]", err)
	}

	lb.Fill(3)
	lb.ResizeMax(1)

	if err := lb.Get(2); err != ErrNotEnough {
		t.Fatalf("Get token should fail. [err:%v]", err)
	}

	lb.Fill(3)
	if err := lb.Get(2); err != ErrNotEnough {
		t.Fatalf("Get token should fail. [err:%v]", err)
	}
}

func TestWaitLeaky(t *testing.T) {
	const (
		init = 4
		tps  = -500.0
	)
	lb := New(&Config{
		InitTokens: init,
		TokensPer:  tps,
	})
	defer lb.Close()

	unit := time.Duration(float64(time.Second) / -tps)
	time.Sleep(2 * unit)

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	if err := lb.WaitContext(ctx, 3); err != context.DeadlineExceeded {
		t.Fatalf("Wait token should fail with timeout. [err:%v]", err)
	}
	if err := lb.Wait(1); err != nil {
		t.Fatalf("Fail to wait token. [err:%v]", err)
	}
}

func TestClose(t *testing.T) {
	const (
		max  = 4
		init = 1
	)
	lb := New(&Config{
		MaxTokens:  max,
		InitTokens: init,
	})
	go func() {
		time.Sleep(time.Millisecond)
		lb.Close()
	}()

	if err := lb.Wait(5); err != ErrClosed {
		t.Fatalf("Wait token should fail with timeout. [err:%v]", err)
	}
	if err := lb.Close(); err != nil {
		t.Fatalf("Close a closed leaky bucket should be ok. [err:%v]", err)
	}
}

func TestWait(t *testing.T) {
	const (
		max  = 4
		init = 1
		tps  = 500.0
	)
	lb := New(&Config{
		MaxTokens:  max,
		InitTokens: init,
		TokensPer:  tps,
	})
	defer lb.Close()

	if err := lb.Wait(1); err != nil {
		t.Fatalf("Fail to wait token. [err:%v]", err)
	}
	if err := lb.Wait(-1); err != nil {
		t.Fatalf("Fail to wait token. [err:%v]", err)
	}

	start := time.Now()
	unit := time.Duration(float64(time.Second) / tps)

	if err := lb.Wait(1); err != nil {
		t.Fatalf("Fail to wait token. [err:%v]", err)
	} else if diff := time.Now().Sub(start); diff < unit || diff >= 2*unit {
		t.Fatalf("Token should be available in [unit, unit+1). [diff:%v] [unit:%v]", diff, unit)
	}

	ctx, cancel1 := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel1()

	if err := lb.WaitContext(ctx, 1); err != context.DeadlineExceeded {
		t.Fatalf("Wait token should fail with timeout. [err:%v]", err)
	}

	if err := lb.Wait(1); err != nil {
		t.Fatalf("Fail to wait token. [err:%v]", err)
	}

	ctx, cancel2 := context.WithTimeout(context.Background(), 15*time.Millisecond)
	defer cancel2()

	if err := lb.WaitContext(ctx, -1); err != nil {
		t.Fatalf("Fail to wait token. [err:%v]", err)
	}

	if err := lb.WaitContext(ctx, 1); err != nil {
		t.Fatalf("Fail to wait token. [err:%v]", err)
	}

	if err := lb.WaitContext(ctx, 5); err != context.DeadlineExceeded {
		t.Fatalf("Wait token should fail with timeout. [err:%v]", err)
	}

}
