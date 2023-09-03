package webhookHandler

import (
	jsoniter "github.com/json-iterator/go"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"strings"
	"sync"
)

// BililiveRecoderWebhookHandler 处理 BililiveRecoder 的 webhook 请求
func BililiveRecoderWebhookHandler(w http.ResponseWriter, request *http.Request) {
	// defer request.Body.Close()
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(request.Body)
	// return 200 at first
	w.WriteHeader(http.StatusOK)

	// process other steps in another goroutine
	var ioReaderWaitGroup sync.WaitGroup
	ioReaderWaitGroup.Add(1)
	go func() {
		// 读取请求内容
		content, err := io.ReadAll(request.Body)
		ioReaderWaitGroup.Done()
		if err != nil {
			var logBuilder strings.Builder
			logBuilder.WriteString("读取 BililiveRecoder webhook 请求失败：")
			logBuilder.WriteString(err.Error())
			log.Error(logBuilder.String())
			return
		}
		log.Info("收到 BililiveRecoder webhook 请求")
		if log.IsLevelEnabled(log.DebugLevel) {
			log.Debug(string(content))
		}

		// 判断是否是重复的webhook请求
		webhookId := jsoniter.Get(content, "EventId").ToString()
		webhookMessageIdListLock.Lock()
		if webhookMessageIdList.IsContain(webhookId) {
			webhookMessageIdListLock.Unlock()
			var logBuilder strings.Builder
			logBuilder.WriteString("重复的 BililiveRecoder webhook 请求：")
			logBuilder.WriteString(webhookId)
			log.Warn(logBuilder.String())
			return
		} else {
			webhookMessageIdList.EnQueue(webhookId)
			webhookMessageIdListLock.Unlock()
		}

		// 判断事件类型
		eventType := jsoniter.Get(content, "EventType").ToString()
		switch eventType {
		//录制开始 SessionStarted
		case "SessionStarted":
			var logBuilder strings.Builder
			logBuilder.WriteString("B站录播姬 录制开始 ")
			logBuilder.WriteString(jsoniter.Get(content, "EventData", "Name").ToString())
			log.Info(logBuilder.String())
			break

		//文件打开 FileOpening
		case "FileOpening":
			if log.IsLevelEnabled(log.DebugLevel) {
				var logBuilder strings.Builder
				logBuilder.WriteString("B站录播姬 文件打开 ")
				logBuilder.WriteString(jsoniter.Get(content, "EventData", "RelativePath").ToString())
				log.Debug(logBuilder.String())
			}
			break

		//文件关闭 FileClosed
		case "FileClosed":
			if log.IsLevelEnabled(log.DebugLevel) {
				var logBuilder strings.Builder
				logBuilder.WriteString("B站录播姬 文件关闭 ")
				logBuilder.WriteString(jsoniter.Get(content, "EventData", "RelativePath").ToString())
				log.Debug(logBuilder.String())
			}
			break

		//录制结束 SessionEnded
		case "SessionEnded":
			var logBuilder strings.Builder
			logBuilder.WriteString("B站录播姬 录制结束 ")
			logBuilder.WriteString(jsoniter.Get(content, "EventData", "Name").ToString())
			log.Info(logBuilder.String())
			break

		//直播开始 StreamStarted
		case "StreamStarted":
			var logBuilder strings.Builder
			logBuilder.WriteString("B站录播姬 直播开始 ")
			logBuilder.WriteString(jsoniter.Get(content, "EventData", "Name").ToString())
			log.Info(logBuilder.String())

			/*var msgTitleBuilder strings.Builder
			msgTitleBuilder.WriteString(jsoniter.Get(content, "EventData", "Name").ToString())
			msgTitleBuilder.WriteString(" 开播了")
			var msgContentBuilder strings.Builder
			msgContentBuilder.WriteString("- 主播：")
			msgContentBuilder.WriteString(jsoniter.Get(content, "EventData", "Name").ToString())
			msgContentBuilder.WriteString("\n- 标题：")
			msgContentBuilder.WriteString(jsoniter.Get(content, "EventData", "Title").ToString())
			msgContentBuilder.WriteString("\n- 分区：")
			msgContentBuilder.WriteString(jsoniter.Get(content, "EventData", "AreaNameParent").ToString())
			msgContentBuilder.WriteString(" - ")
			msgContentBuilder.WriteString(jsoniter.Get(content, "EventData", "AreaNameChild").ToString())
			msgContentBuilder.WriteString("\n- 开播时间：")
			msgContentBuilder.WriteString(jsoniter.Get(content, "EventTimestamp").ToString())

			var msg = messageSender.Message{
				Title:  msgTitleBuilder.String(),
				Content: msgContentBuilder.String(),
			}
			msg.Send()*/
			break

		//直播结束 StreamEnded
		case "StreamEnded":
			var logBuilder strings.Builder
			logBuilder.WriteString("B站录播姬 直播结束 ")
			logBuilder.WriteString(jsoniter.Get(content, "EventData", "Name").ToString())
			log.Info(logBuilder.String())

			/*var msgTitleBuilder strings.Builder
			msgTitleBuilder.WriteString(jsoniter.Get(content, "EventData", "Name").ToString())
			msgTitleBuilder.WriteString(" 直播结束")
			var msgContentBuilder strings.Builder
			msgContentBuilder.WriteString("- 主播：")
			msgContentBuilder.WriteString(jsoniter.Get(content, "EventData", "Name").ToString())
			msgContentBuilder.WriteString("\n- 标题：")
			msgContentBuilder.WriteString(jsoniter.Get(content, "EventData", "Title").ToString())
			msgContentBuilder.WriteString("\n- 分区：")
			msgContentBuilder.WriteString(jsoniter.Get(content, "EventData", "AreaNameParent").ToString())
			msgContentBuilder.WriteString(" - ")
			msgContentBuilder.WriteString(jsoniter.Get(content, "EventData", "AreaNameChild").ToString())

			var msg = messageSender.Message{
				Title:  msgTitleBuilder.String(),
				Content: msgContentBuilder.String(),
			}
			msg.Send()*/
			break

		//	别的不关心，所以没写
		default:
			var logBuilder strings.Builder
			logBuilder.WriteString("BililiveRecoder 未知的webhook请求类型：")
			logBuilder.WriteString(eventType)
			log.Warn(logBuilder.String())
		}
	}()
	// 等待响应体读取完毕
	ioReaderWaitGroup.Wait()
}
