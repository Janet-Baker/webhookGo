package messageSender

import (
	"bytes"
	"errors"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"net/url"
	"sync"
)

type barkMessageStruct struct {
	DeviceKey string `json:"device_key"`
	Title     string `json:"title,omitempty"` // 推送标题
	Body      string `json:"body"`            // 推送内容
	Icon      string `json:"icon,omitempty"`  // 自定义推送图标
	//Category  string `json:"category,omitempty"`  // 消息分类(?)
	//Group     string `json:"group,omitempty"`     // 推送消息分组
	//Sound     string `json:"sound,omitempty"`     // 推送铃声
	//Badge     int    `json:"badge,omitempty"`     // 设置角标
	//Url       string `json:"url,omitempty"`       // 点击通知跳转至URL
	//IsArchive int    `json:"isArchive,omitempty"` // 为1时自动保存通知消息
	//Copy      string `json:"copy,omitempty"`      // 长按通知可选择复制指定内容
	//AutoCopy  int    `json:"autoCopy,omitempty"`  // 为1时自动复制copy内容
}

// BarkServer Bark消息推送(iOS)
type BarkServer struct {
	ServerUrl string `yaml:"url"`
	DeviceKey string `yaml:"device_key"`
}

func (barkServer BarkServer) RegisterServer() {
	RegisterMessageServer(barkServer)
}

func (barkServer BarkServer) SendMessage(message Message) {
	_ = barkServer.sendMessage(message)
}

var barkMessagePool = sync.Pool{
	New: func() any {
		b := new(barkMessageStruct)
		return b
	},
}

func (barkServer BarkServer) sendMessage(message Message) error {
	if message == nil {
		return errors.New("发送Bark消息失败：消息为空")
	}
	if barkServer.DeviceKey == "" {
		return errors.New("无效的Bark密钥")
	}

	// Get a buffer from the pool
	buf := bytesBufferPool.Get().(*bytes.Buffer)
	buf.Reset()                    // Reset the buffer for reuse
	defer bytesBufferPool.Put(buf) // Return the buffer to the pool

	var barkMessage = barkMessagePool.Get().(*barkMessageStruct)
	defer barkMessagePool.Put(barkMessage)
	barkMessage.DeviceKey = barkServer.DeviceKey
	barkMessage.Title = message.GetTitle()
	barkMessage.Body = message.GetContent()
	barkMessage.Icon = message.GetIconURL()
	//var barkMessage = barkMessageStruct{
	//	DeviceKey: barkServer.DeviceKey,
	//	Title:     message.GetTitle(),
	//	Body:      message.GetContent(),
	//	Icon:      message.GetIconURL(),
	//}
	// Marshal the message into the buffer
	if err := encodeJson(barkMessage, buf); err != nil {
		log.Error("Encoding message failed", err)
		return err
	}

	//sendUrl := barkServer.ServerUrl + "/push"
	sendUrl, err := url.JoinPath(barkServer.ServerUrl, "/push")
	if err != nil {
		log.Error("发送Bark消息失败：构造目标链接时出错：", err)
		return err
	}

	resp, err := http.Post(sendUrl, "application/json", buf)
	if err != nil {
		log.Error("发送Bark消息失败：", err)
		return err
	}
	defer func(Body io.ReadCloser) {
		errCloser := Body.Close()
		if errCloser != nil {
			log.Error("发送Bark消息：关闭消息发送响应失败：", errCloser.Error())
		}
	}(resp.Body)
	if log.IsLevelEnabled(log.DebugLevel) {
		buf := bytesBufferPool.Get().(*bytes.Buffer)
		buf.Reset()                    // Reset the buffer for reuse
		defer bytesBufferPool.Put(buf) // Return the buffer to the pool
		_, _ = buf.ReadFrom(resp.Body)
		log.Debug("发送Bark消息响应：%+v", buf.String())
	}
	return nil
}
