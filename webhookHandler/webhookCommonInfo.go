package webhookHandler

import (
	"fmt"
	"sync"
	"time"
)

var webhookMessageIds = make(map[string]int64)
var webhookMessageIdsLock sync.Mutex

func registerId(id string) (exist bool) {
	webhookMessageIdsLock.Lock()
	defer webhookMessageIdsLock.Unlock()
	// query
	_, exist = webhookMessageIds[id]
	if exist {
		return
	}
	webhookMessageIds[id] = time.Now().Unix()
	go func(id string) {
		time.Sleep(time.Hour)
		webhookMessageIdsLock.Lock()
		defer webhookMessageIdsLock.Unlock()
		delete(webhookMessageIds, id)
	}(id)
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
