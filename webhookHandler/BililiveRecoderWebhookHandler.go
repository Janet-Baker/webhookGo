package webhookHandler

import (
	log "github.com/sirupsen/logrus"
	"github.com/valyala/fastjson"
	"io"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"
	"webhookGo/bilibiliInfo"
	"webhookGo/messageSender"
)

// bililiveRecoderTaskRunner 根据响应体内容，执行任务
func bililiveRecoderTaskRunner(content []byte) {
	log.Trace(string(content))
	var p fastjson.Parser
	getter, errOfJsonParser := p.ParseBytes(content)
	if errOfJsonParser != nil {
		return
	}
	webhookId := string(getter.GetStringBytes("EventId"))
	log.Debug(webhookId, "收到 BililiveRecoder webhook 请求")

	// 判断是否是重复的webhook请求
	if registerId(webhookId) {
		return
	}

	// 判断事件类型
	eventType := string(getter.GetStringBytes("EventType"))
	eventSettings, _ := bililiveRecoderSettings[eventType]
	switch eventType {
	//录制开始 SessionStarted
	case "SessionStarted":
		fallthrough
	//录制结束 SessionEnded
	case "SessionEnded":
		fallthrough
	//直播开始 StreamStarted
	case "StreamStarted":
		if eventSettings.Care {
			var logBuilder strings.Builder
			logBuilder.WriteString(webhookId)
			logBuilder.WriteString(" B站录播姬 ")
			logBuilder.WriteString(bililiveRecoderEventName[eventType])
			logBuilder.WriteString("：")
			logBuilder.Write(getter.GetStringBytes("EventData", "Name"))
			log.Info(logBuilder.String())
		}
		if eventSettings.Notify {
			var msgTitleBuilder strings.Builder
			msgTitleBuilder.Write(getter.GetStringBytes("EventData", "Name"))
			msgTitleBuilder.WriteString(" ")
			msgTitleBuilder.WriteString(bililiveRecoderEventName[eventType])
			var msgContentBuilder strings.Builder
			msgContentBuilder.WriteString("- 主播：[")
			msgContentBuilder.Write(getter.GetStringBytes("EventData", "Name"))
			msgContentBuilder.WriteString("](https://live.bilibili.com/")
			msgContentBuilder.WriteString(strconv.FormatUint(getter.GetUint64("EventData", "RoomId"), 10))
			msgContentBuilder.WriteString(")\n- 标题：")
			msgContentBuilder.Write(getter.GetStringBytes("EventData", "Title"))
			msgContentBuilder.WriteString("\n- 分区：")
			msgContentBuilder.Write(getter.GetStringBytes("EventData", "AreaNameParent"))
			msgContentBuilder.WriteString(" - ")
			msgContentBuilder.Write(getter.GetStringBytes("EventData", "AreaNameChild"))

			var msg = messageSender.Message{
				Title:   msgTitleBuilder.String(),
				Content: msgContentBuilder.String(),
				IconURL: bilibiliInfo.GetAvatarByRoomID(getter.GetUint64("EventData", "RoomId")),
			}
			msg.Send()
		}
		break

	//直播结束 StreamEnded
	case "StreamEnded":
		if eventSettings.Care {
			var logBuilder strings.Builder
			logBuilder.WriteString(webhookId)
			logBuilder.WriteString(" B站录播姬 直播结束 ")
			logBuilder.Write(getter.GetStringBytes("EventData", "Name"))
			log.Info(logBuilder.String())
		}
		if eventSettings.Notify {
			var msgTitleBuilder strings.Builder
			msgTitleBuilder.Write(getter.GetStringBytes("EventData", "Name"))
			isRoomLocked, lockTill := bilibiliInfo.IsRoomLocked(getter.GetUint64("EventData", "RoomId"))
			if isRoomLocked {
				msgTitleBuilder.WriteString(" 直播间被封禁")
			} else {
				msgTitleBuilder.WriteString(" 直播结束")
			}
			var msgContentBuilder strings.Builder
			msgContentBuilder.WriteString("- 主播：[")
			msgContentBuilder.Write(getter.GetStringBytes("EventData", "Name"))
			msgContentBuilder.WriteString("](https://live.bilibili.com/")
			msgContentBuilder.WriteString(strconv.FormatUint(getter.GetUint64("EventData", "RoomId"), 10))
			msgContentBuilder.WriteString(")\n- 标题：")
			msgContentBuilder.Write(getter.GetStringBytes("EventData", "Title"))
			msgContentBuilder.WriteString("\n- 分区：")
			msgContentBuilder.Write(getter.GetStringBytes("EventData", "AreaNameParent"))
			msgContentBuilder.WriteString(" - ")
			msgContentBuilder.Write(getter.GetStringBytes("EventData", "AreaNameChild"))
			if isRoomLocked {
				if lockTill > 0 {
					msgContentBuilder.WriteString("\n- 封禁到：")
					msgContentBuilder.WriteString(time.Unix(lockTill, 0).Local().Format("2006-01-02 15:04:05"))
				}
			}

			var msg = messageSender.Message{
				Title:   msgTitleBuilder.String(),
				Content: msgContentBuilder.String(),
				IconURL: bilibiliInfo.GetAvatarByRoomID(getter.GetUint64("EventData", "RoomId")),
			}
			msg.Send()
		}
		break

	//文件打开 FileOpening
	case "FileOpening":
		if eventSettings.Care {
			var logBuilder strings.Builder
			logBuilder.WriteString(webhookId)
			logBuilder.WriteString(" B站录播姬 ")
			logBuilder.WriteString(bililiveRecoderEventName[eventType])
			logBuilder.WriteString(" ")
			logBuilder.Write(getter.GetStringBytes("EventData", "RelativePath"))
			log.Info(logBuilder.String())
		}
		if eventSettings.Notify {
			var msgContentBuilder strings.Builder
			msgContentBuilder.WriteString("- 主播：[")
			msgContentBuilder.Write(getter.GetStringBytes("EventData", "Name"))
			msgContentBuilder.WriteString("](https://live.bilibili.com/")
			msgContentBuilder.WriteString(strconv.FormatUint(getter.GetUint64("EventData", "RoomId"), 10))
			msgContentBuilder.WriteString(")\n- 标题：")
			msgContentBuilder.Write(getter.GetStringBytes("EventData", "Title"))
			msgContentBuilder.WriteString("\n- 分区：")
			msgContentBuilder.Write(getter.GetStringBytes("EventData", "AreaNameParent"))
			msgContentBuilder.WriteString(" - ")
			msgContentBuilder.Write(getter.GetStringBytes("EventData", "AreaNameChild"))
			msgContentBuilder.WriteString("\n- 文件：")
			msgContentBuilder.Write(getter.GetStringBytes("EventData", "RelativePath"))
			msgContentBuilder.WriteString("\n- 文件打开时间：")
			msgContentBuilder.Write(getter.GetStringBytes("EventData", "FileOpenTime"))

			var msg = messageSender.Message{
				Title:   "B站录播姬 文件打开",
				Content: msgContentBuilder.String(),
				IconURL: bilibiliInfo.GetAvatarByRoomID(getter.GetUint64("EventData", "RoomId")),
			}
			msg.Send()
		}
	//文件关闭 FileClosed
	case "FileClosed":
		if eventSettings.Care {
			var logBuilder strings.Builder
			logBuilder.WriteString(webhookId)
			logBuilder.WriteString(" B站录播姬 ")
			logBuilder.WriteString(bililiveRecoderEventName[eventType])
			logBuilder.WriteString(" ")
			logBuilder.Write(getter.GetStringBytes("EventData", "RelativePath"))
			log.Info(logBuilder.String())
		}
		if eventSettings.Notify {
			var msgTitleBuilder strings.Builder
			msgTitleBuilder.WriteString(bililiveRecoderEventName[eventType])
			msgTitleBuilder.WriteString(" ")
			msgTitleBuilder.Write(getter.GetStringBytes("EventData", "RelativePath"))
			var msgContentBuilder strings.Builder
			msgContentBuilder.WriteString("- 主播：[")
			msgContentBuilder.Write(getter.GetStringBytes("EventData", "Name"))
			msgContentBuilder.WriteString("](https://live.bilibili.com/")
			msgContentBuilder.WriteString(strconv.FormatUint(getter.GetUint64("EventData", "RoomId"), 10))
			msgContentBuilder.WriteString(")\n- 标题：")
			msgContentBuilder.Write(getter.GetStringBytes("EventData", "Title"))
			msgContentBuilder.WriteString("\n- 分区：")
			msgContentBuilder.Write(getter.GetStringBytes("EventData", "AreaNameParent"))
			msgContentBuilder.WriteString(" - ")
			msgContentBuilder.Write(getter.GetStringBytes("EventData", "AreaNameChild"))
			msgContentBuilder.WriteString("\n- 文件：")
			msgContentBuilder.Write(getter.GetStringBytes("EventData", "RelativePath"))
			msgContentBuilder.WriteString("\n- 时长：")
			msgContentBuilder.WriteString(strconv.FormatUint(getter.GetUint64("EventData", "Duration"), 10))
			msgContentBuilder.WriteString("秒")
			msgContentBuilder.WriteString("\n- 文件大小：")
			msgContentBuilder.WriteString(formatStorageSpace(getter.GetInt64("EventData", "Size")))
			msgContentBuilder.WriteString("\n- 文件关闭时间：")
			msgContentBuilder.Write(getter.GetStringBytes("EventData", "FileCloseTime"))

			var msg = messageSender.Message{
				Title:   msgTitleBuilder.String(),
				Content: msgContentBuilder.String(),
				IconURL: bilibiliInfo.GetAvatarByRoomID(getter.GetUint64("EventData", "RoomId")),
			}
			msg.Send()
		}
		break

	default:
		var logBuilder strings.Builder
		logBuilder.WriteString(webhookId)
		logBuilder.WriteString(" BililiveRecoder 未知的webhook请求类型：")
		logBuilder.WriteString(eventType)
		log.Warn(logBuilder.String())
	}
	if eventSettings.HaveCommand {
		log.Info(webhookId, "执行命令：", eventSettings.ExecCommand)
		cmd := exec.Command(eventSettings.ExecCommand)
		err := cmd.Run()
		if err != nil {
			log.Error(webhookId, "执行命令失败：", err.Error())
		}
	}
}

// BililiveRecoderWebhookHandler 处理 BililiveRecoder 的 webhook 请求
func BililiveRecoderWebhookHandler(w http.ResponseWriter, request *http.Request) {
	// defer request.Body.Close()
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(request.Body)
	if request.Method != "POST" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// return 200 at first
	w.WriteHeader(http.StatusOK)

	// 读取请求内容
	content, err := io.ReadAll(request.Body)
	if err != nil {
		log.Error("读取 BililiveRecoder webhook 请求失败：", err.Error())
		return
	}
	go bililiveRecoderTaskRunner(content)
}

var bililiveRecoderEventName = map[string]string{
	"SessionStarted": "录制开始",
	"FileOpening":    "文件打开",
	"FileClosed":     "文件关闭",
	"SessionEnded":   "录制结束",
	"StreamStarted":  "直播开始",
	"StreamEnded":    "直播结束",
}

var bililiveRecoderSettings map[string]Event

func UpdateBililiveRecoderSettings(events map[string]Event) {
	bililiveRecoderSettings = events
}
