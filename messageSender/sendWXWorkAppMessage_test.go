package messageSender

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"testing"
)

func TestMarshalWXWorkAppMessageStruct(t *testing.T) {
	var messageStruct = WXWorkAppMessageStruct{
		Touser:                 "app.ToUser",
		Msgtype:                "markdown",
		Agentid:                "app.AgentID",
		Markdown:               Markdown{"# " + "message.GetTitle()" + "\n\n" + "message.GetContent()"},
		EnableDuplicateCheck:   1,
		DuplicateCheckInterval: 3600,
	}
	bodyBuffer, err := json.Marshal(messageStruct)
	if err != nil {
		log.Error("发送企业微信应用消息 消息体编码失败", err)
		t.Fail()
		return
	}
	t.Log(string(bodyBuffer))
}
