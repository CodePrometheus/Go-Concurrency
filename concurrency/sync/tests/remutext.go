package test

import (
	"sync"
)

// MutexMap 对外暴露的map
type MutexMap struct {
	items map[string]interface{} // 为了和上面的ConcurrentMap做比较,都采用string->interface的方式
	mu    *sync.RWMutex          // 读写锁
}

// NewMutexMap 新建一个map
func NewMutexMap() *MutexMap {
	return &MutexMap{
		items: map[string]interface{}{},
		mu:    new(sync.RWMutex),
	}
}

// Set 设置key,value
func (m MutexMap) Set(key string, value interface{}) {
	m.mu.Lock()          // 加锁(全锁定)
	m.items[key] = value // 赋值
	m.mu.Unlock()        // 解锁
}

// Get 获取key对应的value
func (m MutexMap) Get(key string) (value interface{}, ok bool) {
	m.mu.RLock()             // 加锁(读锁定)
	value, ok = m.items[key] // 取值
	m.mu.RUnlock()           // 解锁
	return value, ok
}

// Count 统计key个数
func (m MutexMap) Count() int {
	m.mu.RLock() // 加锁(读锁定)
	count := len(m.items)
	m.mu.RUnlock() // 解锁
	return count
}

// Keys 所有的key
func (m MutexMap) Keys() []string {
	m.mu.RLock() // 加锁(读锁定)
	keys := make([]string, len(m.items))
	for k := range m.items {
		keys = append(keys, k)
	}
	m.mu.RUnlock() // 解锁

	return keys
}
