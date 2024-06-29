package messageSender

import (
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"net/url"
)

// BarkServer Bark消息推送(iOS)
type BarkServer struct {
	ServerUrl   string `yaml:"url"`
	BarkSecrets string `yaml:"secrets"`
}

func (barkServer *BarkServer) RegisterBarkServer() {
	RegisterMessageServer(barkServer)
}

func (barkServer *BarkServer) SendMessage(message Message) {
	if message == nil {
		return
	}
	if barkServer.BarkSecrets == "" {
		return
	}
	sendUrl := barkServer.ServerUrl + barkServer.BarkSecrets + "/" + url.QueryEscape(message.GetTitle()) + "/" + url.QueryEscape(message.GetContent())
	if message.GetIconURL() != "" {
		sendUrl = sendUrl + "?icon=" + url.QueryEscape(message.GetIconURL())
	}
	resp, err := http.Get(sendUrl)
	if err != nil {
		log.Error("发送Bark消息失败：", err)
		return
	}
	defer func(Body io.ReadCloser) {
		errCloser := Body.Close()
		if errCloser != nil {
			log.Error("发送Bark消息：关闭消息发送响应失败：", errCloser.Error())
		}
	}(resp.Body)
	if log.IsLevelEnabled(log.TraceLevel) {
		log.Tracef("发送Bark消息响应：%+v", resp)
	} else {
		log.Debug("发送Bark消息成功")
	}
	return
}
