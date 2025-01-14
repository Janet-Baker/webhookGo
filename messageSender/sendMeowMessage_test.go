package messageSender

import (
	"testing"
	"time"
)

func TestMeowServer_sendMessage(t *testing.T) {
	var target = MeowServer{Username: "xiaopangzi"}
	var msg = GeneralPushMessage{
		Title:   "测试标题" + time.Now().String(),
		Content: "测试内容" + time.Now().GoString(),
		IconURL: "",
	}
	err := target.sendMessage(&msg)
	if err != nil {
		t.Fail()
	}
}
