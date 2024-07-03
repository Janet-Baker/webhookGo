package messageSender

import (
	"github.com/sirupsen/logrus"
	"testing"
)

func TestSendBarkMessageWithEmptySecret(t *testing.T) {
	logrus.SetLevel(logrus.TraceLevel)
	var barkServer BarkServer = BarkServer{
		ServerUrl:   "https://api.day.app",
		BarkSecrets: "", // 留空，用于测试密钥为空的情况
	}
	var message Message = &GeneralPushMessage{
		Title:   "Test",
		Content: "https://live.bilibili.com/4983935",
		IconURL: "https://i2.hdslb.com/bfs/face/711b9992d2fb5bbc4ea72d4905826bdc633bc51f.jpg",
	}
	if err := barkServer.sendMessage(message); err == nil {
		t.Fail() // 密钥为空时，应当报错。如果返回nil说明不符合预期
	}
}

func TestSendBarkMessage(t *testing.T) {
	logrus.SetLevel(logrus.TraceLevel)
	var barkServer BarkServer = BarkServer{
		ServerUrl:   "https://api.day.app",
		BarkSecrets: "", // 需要填写有效的密钥
	}
	var message Message = &GeneralPushMessage{
		Title:   "Test",
		Content: "https://live.bilibili.com/4983935",
		IconURL: "https://i2.hdslb.com/bfs/face/711b9992d2fb5bbc4ea72d4905826bdc633bc51f.jpg",
	}
	barkServer.SendMessage(message)
}
