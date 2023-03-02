package stabilizer

import (
	"Starry-LeakyBucket"
	"context"
	"reflect"
)

// Stabilizer 结合 context 超时机制提供削峰填谷的能力
type Stabilizer struct {
	bucket *leakybucket.LeakyBucket
}

// New 确保同时并发的方法数不超过 max 个
func New(max int) *Stabilizer {
	return &Stabilizer{
		bucket: leakybucket.New(&leakybucket.Config{
			InitTokens: int64(max),
		}),
	}
}

// Run 只在同时并发不超过预设的 max 下会执行 f
// 若达到 max，则一直等待，直达其他 g 执行完毕留出空档
func (s *Stabilizer) Run(f func()) error {
	err := s.bucket.Wait(1)
	if err != nil {
		return err
	}
	defer s.bucket.Set(1)
	f()
	return nil
}

// RunContext 若达到 max，则一直等待，直达其他 g 执行完毕留出空档或者 ctx 超时
func (s *Stabilizer) RunContext(ctx context.Context, f func()) error {
	err := s.bucket.WaitContext(ctx, 1)
	if err != nil {
		if err == leakybucket.ErrDeadlineExceeded {
			s.bucket.Set(1)
		}
		return err
	}
	defer s.bucket.Set(1)
	f()
	return nil
}

// BindFunc 将任意函数与 s 绑定
func (s *Stabilizer) BindFunc(f interface{}) interface{} {
	v := reflect.ValueOf(f)
	if v.Kind() != reflect.Func {
		return nil
	}
	fv := reflect.MakeFunc(v.Type(), func(args []reflect.Value) (res []reflect.Value) {
		err := s.bucket.Wait(1)
		if err == nil {
			defer s.bucket.Set(1)
		}
		return v.Call(args)
	})
	return fv.Interface()
}

// Close 关闭，所有函数调用都会失败
func (s *Stabilizer) Close() error {
	return s.bucket.Close()
}
