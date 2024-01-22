package messageSender

import (
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"net/url"
	"sync"
)

// BarkServer Bark消息推送(iOS)
type BarkServer struct {
	ServerUrl   string `yaml:"url"`
	BarkSecrets string `yaml:"secrets"`
}

var barkServers []BarkServer

func RegisterBarkServer(barkServer BarkServer) {
	barkServers = append(barkServers, barkServer)
}

func SendBarkMessage(message Message) {
	// resp, err := http.Get("https://api.day.app/" + secrets.BarkSecrets + "/" + url.QueryEscape(message.Title) + "/" + url.QueryEscape(message.Content))
	length := len(barkServers)
	if length > 0 {
		wg := sync.WaitGroup{}
		for i := 0; i < length; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				sendBarkMessage(barkServers[i], message)
			}(i)
		}
	}
}

func sendBarkMessage(barkServer BarkServer, message Message) {
	if barkServer.BarkSecrets == "" {
		return
	}
	sendUrl := barkServer.ServerUrl + barkServer.BarkSecrets + "/" + url.QueryEscape(message.Title) + "/" + url.QueryEscape(message.Content)
	if message.IconURL != "" {
		sendUrl = sendUrl + "?icon=" + url.QueryEscape(message.IconURL)
	}
	resp, err := http.Get(sendUrl)
	defer func(Body io.ReadCloser) {
		errCloser := Body.Close()
		if errCloser != nil {
			log.Errorf("%s 发送Bark消息：关闭消息发送响应失败：%s", message.ID, errCloser.Error())
		}
	}(resp.Body)
	if err != nil {
		log.Errorf("%s 发送Bark消息失败：%s", message.ID, err.Error())
		return
	}
	if log.IsLevelEnabled(log.TraceLevel) {
		log.Tracef("%s 发送Bark消息响应：%+v", message.ID, resp)
	} else {
		log.Debugf("%s 发送Bark消息成功", message.ID)
	}
	return
}
