package messageSender

import (
	log "github.com/sirupsen/logrus"
)

type Message interface {
	GetTitle() string
	GetContent() string
	GetIconURL() string
}

type GeneralPushMessage struct {
	// 消息标题
	Title string
	// 消息内容
	Content string
	// IconURL，Bark等可以自定义头像的用
	IconURL string
}

func (message *GeneralPushMessage) GetTitle() string {
	return message.Title
}

func (message *GeneralPushMessage) GetContent() string {
	return message.Content
}

func (message *GeneralPushMessage) GetIconURL() string {
	return message.IconURL
}

func (message *GeneralPushMessage) SendToAllTargets() {
	if log.IsLevelEnabled(log.TraceLevel) {
		log.Tracef("发送消息：%+v", *message)
	}

	// 并发发送消息
	// 等 go 1.22 我就给这玩意改咯
	for i := 0; i < len(servers); i++ {
		go func(i int) {
			servers[i].SendMessage(message)
		}(i)
	}
	return
}

type MessageServer interface {
	SendMessage(message Message)
}

var servers []MessageServer

func GetAllServers() []MessageServer {
	return servers
}

func RegisterMessageServer(server MessageServer) {
	servers = append(servers, server)
}
