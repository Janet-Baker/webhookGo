package messageSender

import (
	log "github.com/sirupsen/logrus"
	"sync"
)

type Message struct {
	// 消息标题
	Title string
	// 消息内容
	Content string
}

func (m *Message) Send() {
	log.Debugf("发送消息：%+v", *m)

	// 并发发送消息
	var wg sync.WaitGroup
	wg.Add(2)
	// 发送 Bark 消息
	go func() {
		defer wg.Done() //程序退出的时候执行
		SendBarkMessage(*m)
	}()
	// 发送企业微信应用消息
	go func() {
		defer wg.Done() //程序退出的时候执行
		SendWeWorkMessage(*m)
	}()
	// 等待协程结束
	wg.Wait()

	return
}
