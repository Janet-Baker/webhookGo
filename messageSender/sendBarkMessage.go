package messageSender

import (
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"net/url"
	"strings"
	"webhookTemplate/secrets"
)

func SendBarkMessage(message Message) {
	// resp, err := http.Get("https://api.day.app/" + secrets.BarkSecrets + "/" + url.QueryEscape(message.Title) + "/" + url.QueryEscape(message.Content))
	length := len(secrets.Secrets.Barks)
	for i := 0; i < length; i++ {
		go sendBarkMessage(secrets.Secrets.Barks[i], message)
	}
}

func sendBarkMessage(barkServer secrets.BarkServer, message Message) {
	if barkServer.BarkSecrets == "" {
		log.Warn(barkServer.ServerUrl, "的BarkSecrets为空，跳过Bark消息发送")
		return
	}
	var urlBuilder strings.Builder
	urlBuilder.WriteString(barkServer.ServerUrl)
	urlBuilder.WriteString(barkServer.BarkSecrets)
	urlBuilder.WriteString("/")
	urlBuilder.WriteString(url.QueryEscape(message.Title))
	urlBuilder.WriteString("/")
	urlBuilder.WriteString(url.QueryEscape(message.Content))
	if message.IconURL != "" {
		urlBuilder.WriteString("?icon=")
		urlBuilder.WriteString(url.QueryEscape(message.IconURL))
	}
	resp, err := http.Get(urlBuilder.String())
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
