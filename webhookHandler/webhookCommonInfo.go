package webhookHandler

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Cache struct {
	m     map[string]int64
	mutex sync.RWMutex
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
	c.mutex.RLock()
	defer c.mutex.RUnlock()
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

// 定义一个字符串数组，表示不同的单位
var units = []string{"B", "KB", "MB", "GB", "TB"}

// 定义一个函数，接受一个整数参数，表示字节数
func formatStorageSpace(bytes int64) string {
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

// secondsToString 将用秒表示的时间转换为字符串（?天）(?时)(?分)(?秒)(剩余小数)
// 例如 1.234567 => "1秒234"
func secondsToString(seconds float64) string {
	var timeBuilder strings.Builder
	s := int(seconds)
	ms := int((seconds - float64(s)) * 1000)
	if s >= 86400 {
		timeBuilder.WriteString(strconv.Itoa(s / 86400))
		timeBuilder.WriteString("天")
		s = s - (int(s/86400))*86400
	}
	if s >= 3600 {
		timeBuilder.WriteString(strconv.Itoa(s / 3600))
		timeBuilder.WriteString("时")
		s = s - (int(s/3600))*3600
	}
	if s >= 60 {
		timeBuilder.WriteString(strconv.Itoa(s / 60))
		timeBuilder.WriteString("分")
		s = s - (int(s/60))*60
	}
	timeBuilder.WriteString(strconv.Itoa(s))
	timeBuilder.WriteString("秒")

	if ms > 1 {
		timeBuilder.WriteString(strconv.Itoa(ms))
	}
	return timeBuilder.String()
}

func execCommand(command string) {
	log.Info("执行命令：", command)
	cmd := exec.Command(command)
	err := cmd.Run()
	if err != nil {
		log.Error("执行命令失败：", err.Error())
	}
}
