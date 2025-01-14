package messageSender

import (
	"bytes"
	"errors"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"net/url"
)

const meowServerUrl = "http://api.chuckfang.com/"

type MeowServer struct {
	Username string `yaml:"username"`
}

func (m MeowServer) RegisterServer() {
	RegisterMessageServer(m)
}

func (m MeowServer) SendMessage(message Message) {
	err := m.sendMessage(message)
	if err != nil {
		log.Error("发送MeoW消息失败：", err)
	}
}

func (m MeowServer) sendMessage(message Message) error {
	if message == nil {
		return errors.New("发送MeoW消息失败：消息为空")
	}
	target, err := url.JoinPath(meowServerUrl, m.Username, message.GetTitle(), message.GetContent())
	if err != nil {
		log.Error("发送MeoW消息失败：", err)
		return err
	}
	resp, err := http.Get(target)
	if err != nil {
		log.Error("发送MeoW消息失败：", err)
		return err
	}
	defer func(Body io.Closer) {
		_ = Body.Close()
	}(resp.Body)

	if log.IsLevelEnabled(log.DebugLevel) {
		buf := bytesBufferPool.Get().(*bytes.Buffer)
		buf.Reset()                    // Reset the buffer for reuse
		defer bytesBufferPool.Put(buf) // Return the buffer to the pool
		_, _ = buf.ReadFrom(resp.Body)
		log.Debug("发送MeoW消息响应：", buf.String())
	}

	return nil
}
