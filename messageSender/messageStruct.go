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

func (m *Message) Send() error {
	log.Infof("发送消息：%+v", *m)

	// 并发发送消息
	var wg sync.WaitGroup
	wg.Add(2)
	// 发送 Bark 消息
	var err1 error
	go func() {
		defer wg.Done() //程序退出的时候执行
		err1 = SendBarkMessage(*m)
	}()
	// 发送企业微信应用消息
	var err2 error
	go func() {
		defer wg.Done() //程序退出的时候执行
		err2 = SendWeWorkMessage(*m)
	}()
	// 等待协程结束
	wg.Wait()

	if !(err1 == nil || err2 == nil) {
		return err1
	}
	return nil
}
