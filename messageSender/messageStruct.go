package messageSender

import (
	log "github.com/sirupsen/logrus"
)

type Message interface {
	GetTitle() string
	GetContent() string
	GetIconURL() string
	SendToAllTargets()
}

type OldMessageToRefactor struct {
	// 消息标题
	Title string
	// 消息内容
	Content string
	// IconURL，Bark等可以自定义头像的用
	IconURL string
}

func (message *OldMessageToRefactor) GetTitle() string {
	return message.Title
}

func (message *OldMessageToRefactor) GetContent() string {
	return message.Content
}

func (message *OldMessageToRefactor) GetIconURL() string {
	return message.IconURL
}

func (message *OldMessageToRefactor) SendToAllTargets() {
	if log.IsLevelEnabled(log.TraceLevel) {
		log.Tracef("发送消息：%+v", *message)
	}

	// 并发发送消息
	for i := 0; i < len(servers); i++ {
		go func(i int) {
			servers[i].SendMessage(message)
		}(i)
	}
	return
}

type MessageServer interface {
	SendMessage(message *OldMessageToRefactor)
}

var servers []MessageServer

func GetAllServers() []MessageServer {
	return servers
}

func RegisterMessageServer(server MessageServer) {
	servers = append(servers, server)
}
