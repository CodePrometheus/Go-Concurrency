package test

import (
	"math/rand"
	"strconv"
	"sync"
	"testing"
)

// go test --bench=. -benchmem -count=1
// 基准测试下，分段锁ConcurrentMap是多协程安全的，且效率最高；sync.map效率次之，传统的map+读写锁效率最低
// ConcurrentMap中锁的个数越多，效率越高，因为争夺同一把锁的概率降低了

// 10万次的赋值,10万次的读取
var times int = 100000

// 测试ConcurrentMap
func BenchmarkTestConcurrentMap(b *testing.B) {
	for k := 0; k < b.N; k++ {
		b.StopTimer()
		// 产生10000个不重复的键值对(string -> int)
		testKV := map[string]int{}
		for i := 0; i < 10000; i++ {
			testKV[strconv.Itoa(i)] = i
		}

		// 新建一个ConcurrentMap
		pMap := NewConcurrentMap()

		// set到map中
		for k, v := range testKV {
			pMap.Set(k, v)
		}

		// 开始计时
		b.StartTimer()

		wg := sync.WaitGroup{}
		wg.Add(2)

		// 赋值
		go func() {
			// 对随机key,赋值times次
			for i := 0; i < times; i++ {
				index := rand.Intn(times)
				pMap.Set(strconv.Itoa(index), index+1)
			}
			wg.Done()
		}()

		// 读取
		go func() {
			// 对随机key,读取times次
			for i := 0; i < times; i++ {
				index := rand.Intn(times)
				pMap.Get(strconv.Itoa(index))
			}
			wg.Done()
		}()

		// 等待两个协程处理完毕
		wg.Wait()
	}
}

// 测试map加锁
func BenchmarkTestMap(b *testing.B) {
	for k := 0; k < b.N; k++ {
		b.StopTimer()
		// 产生10000个不重复的键值对(string -> int)
		testKV := map[string]int{}
		for i := 0; i < 10000; i++ {
			testKV[strconv.Itoa(i)] = i
		}

		// 新建一个MutexMap
		pMap := NewMutexMap()

		// set到map中
		for k, v := range testKV {
			pMap.Set(k, v)
		}

		// 开始计时
		b.StartTimer()

		wg := sync.WaitGroup{}
		wg.Add(2)
		// 赋值
		go func() {
			// 对随机key,赋值100万次
			for i := 0; i < times; i++ {
				index := rand.Intn(times)
				pMap.Set(strconv.Itoa(index), index+1)
			}
			wg.Done()
		}()

		// 读取
		go func() {
			// 对随机key,读取100万次
			for i := 0; i < times; i++ {
				index := rand.Intn(times)
				pMap.Get(strconv.Itoa(index))
			}
			wg.Done()
		}()

		// 等待两个协程处理完毕
		wg.Wait()
	}
}

// 测试sync.map
func BenchmarkTestSyncMap(b *testing.B) {
	for k := 0; k < b.N; k++ {
		b.StopTimer()
		// 产生10000个不重复的键值对(string -> int)
		testKV := map[string]int{}
		for i := 0; i < 10000; i++ {
			testKV[strconv.Itoa(i)] = i
		}

		// 新建一个sync.Map
		pMap := &sync.Map{}

		// set到map中
		for k, v := range testKV {
			pMap.Store(k, v)
		}

		// 开始计时
		b.StartTimer()

		wg := sync.WaitGroup{}
		wg.Add(2)

		// 赋值
		go func() {
			// 对随机key,赋值10万次
			for i := 0; i < times; i++ {
				index := rand.Intn(times)
				pMap.Store(strconv.Itoa(index), index+1)
			}
			wg.Done()
		}()

		// 读取
		go func() {
			// 对随机key,读取10万次
			for i := 0; i < times; i++ {
				index := rand.Intn(times)
				pMap.Load(strconv.Itoa(index))
			}
			wg.Done()
		}()

		// 等待两个协程处理完毕
		wg.Wait()
	}
}
