package bucket

import (
	"Starry-LeakyBucket"
	"context"
)

type TokenBucket struct {
	bucket *leakybucket.LeakyBucket
}

func New(config *Config) *TokenBucket {
	if config.TokensPer < 0 {
		config.TokensPer = 0
	}
	return &TokenBucket{
		bucket: leakybucket.New(&leakybucket.Config{
			MaxTokens:  config.MaxTokens,
			InitTokens: config.InitTokens,
			TokensPer:  config.TokensPer,
		}),
	}
}

func (tb *TokenBucket) Get(n int64) error {
	return tb.bucket.Get(n)
}

func (tb *TokenBucket) Wait(n int64) error {
	return tb.bucket.Wait(n)
}

func (tb *TokenBucket) WaitContext(ctx context.Context, n int64) error {
	return tb.bucket.WaitContext(ctx, n)
}

func (tb *TokenBucket) Close() error {
	return tb.bucket.Close()
}
