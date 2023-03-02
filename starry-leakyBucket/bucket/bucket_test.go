package bucket

import (
	"context"
	"testing"
	"time"
)

func TestConfig(t *testing.T) {
	tb := New(&Config{
		InitTokens: 1,
		TokensPer:  -500,
	})
	time.Sleep(10 * time.Millisecond)
	if err := tb.Get(1); err != nil {
		t.Fatalf("Fail to get token. [err:%v]", err)
	}
	tb.Close()
}

func TestTB(t *testing.T) {
	tb := New(&Config{
		MaxTokens:  10,
		InitTokens: 5,
		TokensPer:  500,
	})
	defer tb.Close()

	if err := tb.Get(1); err != nil {
		t.Fatalf("Fail to get token. [err:%v]", err)
	}

	if err := tb.Wait(1); err != nil {
		t.Fatalf("Fail to wait token. [err:%v]", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	if err := tb.WaitContext(ctx, 1); err != nil {
		t.Fatalf("Fail to wait token. [err:%v]", err)
	}

	ctx, cancel2 := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel2()

	if err := tb.WaitContext(ctx, 4); err != context.DeadlineExceeded {
		t.Fatalf("Wait token should timeout. [err:%v]", err)
	}
}
