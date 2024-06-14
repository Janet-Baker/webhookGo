package webhookHandler

import (
	"fmt"
	"sync"
	"time"
)

type Cache struct {
	m     map[string]int64
	mutex sync.Mutex
}

func NewAutoCleanupCache() *Cache {
	c := &Cache{
		m: make(map[string]int64),
	}
	go c.autoCleanup()
	return c
}

func (c *Cache) Add(id string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.m[id] = time.Now().Unix()
}

func (c *Cache) Get(id string) (int64, bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	val, ok := c.m[id]
	return val, ok
}

func (c *Cache) Exist(id string) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	_, ok := c.m[id]
	return ok
}

func (c *Cache) Delete(id string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	delete(c.m, id)
}

func (c *Cache) CleanUp() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	for id, t := range c.m {
		if time.Now().Unix()-t >= 3600 {
			delete(c.m, id)
		}
	}
}

func (c *Cache) autoCleanup() {
	ticker := time.NewTicker(time.Hour)
	for range ticker.C {
		c.CleanUp()
	}
}

var idCache = NewAutoCleanupCache()

func registerId(id string) (exist bool) {
	exist = idCache.Exist(id)
	if exist {
		return
	}
	idCache.Add(id)
	return
}

type Event struct {
	Care        bool   `yaml:"care"`
	Notify      bool   `yaml:"notify"`
	HaveCommand bool   `yaml:"have_command"`
	ExecCommand string `yaml:"exec_command"`
}

// 定义一个函数，接受一个整数参数，表示字节数
func formatStorageSpace(bytes int64) string {
	// 定义一个字符串数组，表示不同的单位
	units := []string{"B", "KB", "MB", "GB", "TB"}
	// 定义一个变量，表示当前的单位索引
	index := 0
	// 定义一个浮点数变量，表示当前的字节数
	value := float64(bytes)
	// 循环，直到字节数小于1024或者单位索引达到最大值
	for value >= 1024 && index < len(units)-1 {
		// 字节数除以1024，单位索引加一
		value /= 1024
		index++
	}
	// 返回格式化后的字符串，保留两位小数
	return fmt.Sprintf("%.2f%s", value, units[index])
}
