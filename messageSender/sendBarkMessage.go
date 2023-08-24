package messageSender

import (
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"net/url"
	"webhookTemplate/secrets"
)

func SendBarkMessage(message Message) error {
	log.Debugf("发送 Bark 消息：%+v", message)
	resp, err := http.Get("https://api.day.app/" + secrets.BarkSecrets + "/" + url.QueryEscape(message.Title) + "/" + url.QueryEscape(message.Content))
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Errorf("关闭消息发送响应失败：%s", err.Error())
		}
	}(resp.Body)
	if err != nil {
		log.Errorf("发送消息失败：%s", err.Error())
		return err
	} else {
		log.Debugf("发送Bark消息成功：%+v", message)
	}
	return nil
}
