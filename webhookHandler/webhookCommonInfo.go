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
	sync.Map
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
	ticker := time.NewTicker(time.Duration(expireTime))
	for range ticker.C {
		c.CleanUp()
	}
}

var idCache = NewAutoCleanupCache()

// registerId 注册一个ID，如果ID已经存在则返回true
func registerId(id string) (alreadyExist bool) {
	_, alreadyExist = idCache.LoadOrStore(id, time.Now().Unix())
	return
}

// Event 每个事件的设置项
type Event struct {
	Care        bool   `yaml:"care"`         // 是否在控制台打印事件
	Notify      bool   `yaml:"notify"`       // 是否向客户端推送消息
	HaveCommand bool   `yaml:"have_command"` // 是否需要执行命令
	ExecCommand string `yaml:"exec_command"` // 执行的命令
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
