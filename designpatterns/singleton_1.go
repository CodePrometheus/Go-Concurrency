package designpatterns

import "sync"

type singleton struct{}

var instance *singleton

var once sync.Once

func GetInstanceOnce() *singleton {
	once.Do(func() {
		instance = &singleton{}
	})
	return instance
}

var mu sync.Mutex

func GetInstanceDoubleLock() *singleton {
	if instance == nil {
		mu.Lock()
		defer mu.Unlock()
		if instance == nil {
			instance = &singleton{}
		}
	}
	return instance
}

// GetInstanceLazy 懒汉加锁
func GetInstanceLazy() *singleton {
	mu.Lock()
	defer mu.Unlock()
	if instance == nil {
		instance = &singleton{}
	}
	return instance
}

// GetInstanceHungry 饿汉加锁
var ins *singleton = &singleton{}

// GetInstanceHungry 如果singleton创建初始化比较复杂耗时时，加载时间会延长
func GetInstanceHungry() *singleton {
	return ins
}
