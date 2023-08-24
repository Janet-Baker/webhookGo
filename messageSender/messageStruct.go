package messageSender

import (
	log "github.com/sirupsen/logrus"
)

type Message struct {
	// 消息标题
	Title string
	// 消息内容
	Content string
}

func (m *Message) Send() error {
	log.Infof("发送消息：%+v", *m)
	err1 := SendBarkMessage(*m)
	err2 := SendWeWorkMessage(*m)
	if !(err1 == nil || err2 == nil) {
		return err1
	}
	return nil
}
