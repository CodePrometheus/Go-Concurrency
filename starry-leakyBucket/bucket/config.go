package bucket

type Config struct {
	MaxTokens  int64   // Token 最大量
	InitTokens int64   // Token 初始量
	TokensPer  float64 // Token 恢复速度，正代表一个令牌桶，负代表一个 leaky 桶
}
