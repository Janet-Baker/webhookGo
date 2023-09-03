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
	var urlBuilder strings.Builder
	urlBuilder.WriteString("https://api.day.app/")
	urlBuilder.WriteString(secrets.BarkSecrets)
	urlBuilder.WriteString("/")
	urlBuilder.WriteString(url.QueryEscape(message.Title))
	urlBuilder.WriteString("/")
	urlBuilder.WriteString(url.QueryEscape(message.Content))
	resp, err := http.Get(urlBuilder.String())
	defer func(Body io.ReadCloser) {
		errCloser := Body.Close()
		if errCloser != nil {
			log.Errorf("发送Bark消息：关闭消息发送响应失败：%s", errCloser.Error())
		}
	}(resp.Body)
	if err != nil {
		log.Errorf("发送Bark消息失败：%s", err.Error())
		return
	}
	if log.IsLevelEnabled(log.DebugLevel) {
		log.Debugf("发送Bark消息成功：%+v", message)
	}
	return
}
