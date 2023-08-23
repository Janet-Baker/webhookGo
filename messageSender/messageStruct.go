package messageSender

import (
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"net/url"
)

type Message struct {
	// 消息标题
	Title string
	// 消息内容
	Content string
}

func SendBarkMessage(message Message) error {
	log.Debugf("发送消息：%s", message)
	resp, err := http.Get("https://api.day.app/VLtHaCk3iNumHrXBPMmdhc/" + url.QueryEscape(message.Title) + "/" + url.QueryEscape(message.Content))
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
		log.Infof("发送消息成功：%s", message)
	}
	return nil
}
func (m *Message) Send() error {
	err := SendBarkMessage(*m)
	if err != nil {
		return err
	}
	return nil
}
