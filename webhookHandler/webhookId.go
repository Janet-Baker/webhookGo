package webhookHandler

import (
	"sync"
	"time"
)

type Cache struct {
	sync.Map
	ticker *time.Ticker
}

const expireTime = int64(time.Hour) * 24

// NewAutoCleanupCache 创建一个Cache,自动清理过期缓存
func NewAutoCleanupCache() *Cache {
	var c = &Cache{}
	go c.autoCleanup()
	return c
}

// CleanUp 遍历缓存，删除过期的缓存
func (c *Cache) CleanUp() {
	c.Range(func(key, value interface{}) bool {
		if time.Now().Unix()-value.(int64) > expireTime {
			c.Delete(key.(string))
		}
		return true
	})
}

// autoCleanup 定时清理过期缓存
// autoCleanup 内维持了一个*time.Ticker，每隔expireTime时间清理一次过期缓存
func (c *Cache) autoCleanup() {
	c.ticker = time.NewTicker(time.Duration(expireTime))
	for range c.ticker.C {
		c.CleanUp()
	}
}

var idCache = NewAutoCleanupCache()

// registerId 注册一个ID，如果ID已经存在则返回true
func registerId(id string) (alreadyExist bool) {
	_, alreadyExist = idCache.LoadOrStore(id, time.Now().Unix())
	return
}
