package webhookHandler

import (
	jsoniter "github.com/json-iterator/go"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
)

func BililiveRecoderWebhookHandler(w http.ResponseWriter, request *http.Request) {
	// defer request.Body.Close()
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
	}(request.Body)
	// 读取请求内容
	content, err := ioutil.ReadAll(request.Body)
	if err != nil {
		return
	}
	log.Infof("收到webhook请求")
	log.Debugf(string(content))

	// 判断是否是重复的webhook请求
	webhookId := jsoniter.Get(content, "EventId").ToString()
	if webhookMessageIdList.IsContain(webhookId) {
		log.Warnf("重复的webhook请求：%s", webhookId)
		w.WriteHeader(http.StatusOK)
		return
	} else {
		webhookMessageIdList.EnQueue(webhookId)
	}

	// 判断事件类型
	eventType := jsoniter.Get(content, "EventType").ToString()
	switch eventType {
	//录制开始 SessionStarted
	case "SessionStarted":
		break
	//文件打开 FileOpening
	case "FileOpening":
		break
	//文件关闭 FileClosed
	case "FileClosed":
		break
	//录制结束 SessionEnded
	case "SessionEnded":
		break
	//直播开始 StreamStarted
	case "StreamStarted":
		break
	//直播结束 StreamEnded
	case "StreamEnded":
		break
	//	别的不关心，所以没写
	default:
		log.Warnf("未知的webhook请求：%+v", content)
		w.WriteHeader(http.StatusOK)
	}
}