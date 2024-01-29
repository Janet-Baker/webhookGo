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

// BlrecTaskRunner 根据响应体内容，执行任务
func blrecTaskRunner(content []byte) {
	log.Trace(string(content))
	var p fastjson.Parser
	getter, errOfJsonParser := p.ParseBytes(content)
	if errOfJsonParser != nil {
		return
	}
	webhookId := string(getter.GetStringBytes("id"))
	log.Info(webhookId, "收到 blrec webhook 请求")

	// 判断是否是重复的webhook请求
	if registerId(webhookId) {
		return
	}

	// 判断事件类型
	hookType := string(getter.GetStringBytes("type"))
	eventSettings, _ := blrecSettings[hookType]
	switch hookType {
	// LiveBeganEvent 主播开播
	case "LiveBeganEvent":
		if eventSettings.Care {
			// 构造日志
			var logBuilder strings.Builder
			logBuilder.WriteString(webhookId)
			logBuilder.WriteString(" blrec 主播开播：")
			logBuilder.Write(getter.GetStringBytes("data", "user_info", "name"))
			log.Info(logBuilder.String())
		}
		if eventSettings.Notify {
			// 构造消息
			var msgTitleBuilder strings.Builder
			msgTitleBuilder.Write(getter.GetStringBytes("data", "user_info", "name"))
			msgTitleBuilder.WriteString(" 开播了")
			var msgContentBuilder strings.Builder
			msgContentBuilder.WriteString("- 主播：[")
			msgContentBuilder.Write(getter.GetStringBytes("data", "user_info", "name"))
			msgContentBuilder.WriteString("](https://live.bilibili.com/")
			msgContentBuilder.WriteString(strconv.FormatUint(getter.GetUint64("data", "room_info", "room_id"), 10))
			msgContentBuilder.WriteString(")\n- 标题：")
			msgContentBuilder.Write(getter.GetStringBytes("data", "room_info", "title"))
			msgContentBuilder.WriteString("\n- 分区：")
			msgContentBuilder.Write(getter.GetStringBytes("data", "room_info", "parent_area_name"))
			msgContentBuilder.WriteString(" - ")
			msgContentBuilder.Write(getter.GetStringBytes("data", "room_info", "area_name"))
			msgContentBuilder.WriteString("\n- 开播时间：")
			msgContentBuilder.WriteString(time.Unix(getter.GetInt64("data", "room_info", "live_start_time"), 0).Local().Format("2006-01-02 15:04:05"))

			var msg = messageSender.Message{
				Title:   msgTitleBuilder.String(),
				Content: msgContentBuilder.String(),
				ID:      webhookId,
				IconURL: string(getter.GetStringBytes("data", "user_info", "face")),
			}
			msg.Send()
		}
		break

	// LiveEndedEvent 主播下播
	case "LiveEndedEvent":
		if eventSettings.Care {
			// 构造日志
			var logBuilder strings.Builder
			logBuilder.WriteString(webhookId)
			logBuilder.WriteString(" blrec 主播下播：")
			logBuilder.Write(getter.GetStringBytes("data", "user_info", "name"))
			log.Info(logBuilder.String())
		}
		if eventSettings.Notify {
			// 构造消息
			var msgTitleBuilder strings.Builder
			msgTitleBuilder.Write(getter.GetStringBytes("data", "user_info", "name"))
			isRoomLocked, lockTill := bilibiliInfo.IsRoomLocked(getter.GetUint64("data", "room_info", "room_id"), webhookId)
			if isRoomLocked {
				msgTitleBuilder.WriteString(" 直播间被封禁")
			} else {
				msgTitleBuilder.WriteString(" 下播了")
			}
			var msgContentBuilder strings.Builder
			msgContentBuilder.WriteString("- 主播：[")
			msgContentBuilder.Write(getter.GetStringBytes("data", "user_info", "name"))
			msgContentBuilder.WriteString("](https://live.bilibili.com/")
			msgContentBuilder.WriteString(strconv.FormatUint(getter.GetUint64("data", "room_info", "room_id"), 10))
			msgContentBuilder.WriteString(")\n- 标题：")
			msgContentBuilder.Write(getter.GetStringBytes("data", "room_info", "title"))
			msgContentBuilder.WriteString("\n- 分区：")
			msgContentBuilder.Write(getter.GetStringBytes("data", "room_info", "parent_area_name"))
			msgContentBuilder.WriteString(" - ")
			msgContentBuilder.Write(getter.GetStringBytes("data", "room_info", "area_name"))
			if isRoomLocked {
				msgContentBuilder.WriteString("\n- 封禁到：")
				msgContentBuilder.WriteString(time.Unix(lockTill, 0).Local().Format("2006-01-02 15:04:05"))
			}

			var msg = messageSender.Message{
				Title:   msgTitleBuilder.String(),
				Content: msgContentBuilder.String(),
				ID:      webhookId,
				IconURL: string(getter.GetStringBytes("data", "user_info", "face")),
			}
			msg.Send()
		}
		break

	// RoomChangeEvent 直播间信息改变
	case "RoomChangeEvent":
		fallthrough

	// RecordingStartedEvent 录制开始
	case "RecordingStartedEvent":
		fallthrough

	// RecordingFinishedEvent 录制结束
	case "RecordingFinishedEvent":
		fallthrough

	// RecordingCancelledEvent 录制取消
	case "RecordingCancelledEvent":
		if eventSettings.Care {
			var logBuilder strings.Builder
			logBuilder.WriteString(webhookId)
			logBuilder.WriteString(" blrec ")
			logBuilder.WriteString(blrecEventNameMap[hookType])
			logBuilder.WriteString("：")
			logBuilder.Write(getter.GetStringBytes("data", "user_info", "name"))
			logBuilder.WriteString(time.Unix(getter.GetInt64("data", "room_info", "room_id"), 0).Local().Format("2006-01-02 15:04:05"))
			log.Info(logBuilder.String())
		}
		if eventSettings.Notify {
			var msgTitleBuilder strings.Builder
			msgTitleBuilder.Write(getter.GetStringBytes("data", "user_info", "name"))
			msgTitleBuilder.WriteString(blrecEventNameMap[hookType])
			var msgContentBuilder strings.Builder
			msgContentBuilder.WriteString("- 主播：[")
			msgContentBuilder.Write(getter.GetStringBytes("data", "user_info", "name"))
			msgContentBuilder.WriteString("](https://live.bilibili.com/")
			msgContentBuilder.WriteString(strconv.FormatUint(getter.GetUint64("data", "room_info", "room_id"), 10))
			msgContentBuilder.WriteString(")\n- 标题：")
			msgContentBuilder.Write(getter.GetStringBytes("data", "room_info", "title"))
			msgContentBuilder.WriteString("\n- 分区：")
			msgContentBuilder.Write(getter.GetStringBytes("data", "room_info", "parent_area_name"))
			msgContentBuilder.WriteString(" - ")
			msgContentBuilder.Write(getter.GetStringBytes("data", "room_info", "area_name"))

			var msg = messageSender.Message{
				Title:   msgTitleBuilder.String(),
				Content: msgContentBuilder.String(),
				ID:      webhookId,
				IconURL: string(getter.GetStringBytes("data", "user_info", "face")),
			}
			msg.Send()
		}
		break

	// VideoFileCreatedEvent 视频文件创建
	case "VideoFileCreatedEvent":
		fallthrough

	// VideoFileCompletedEvent 视频文件完成
	case "VideoFileCompletedEvent":
		fallthrough

	// DanmakuFileCreatedEvent 弹幕文件创建
	case "DanmakuFileCreatedEvent":
		fallthrough

	// DanmakuFileCompletedEvent 弹幕文件完成
	case "DanmakuFileCompletedEvent":
		fallthrough

	// RawDanmakuFileCreatedEvent 原始弹幕文件创建
	case "RawDanmakuFileCreatedEvent":
		fallthrough

	// RawDanmakuFileCompletedEvent 原始弹幕文件完成
	case "RawDanmakuFileCompletedEvent":
		fallthrough

	// VideoPostprocessingCompletedEvent 视频后处理完成
	case "VideoPostprocessingCompletedEvent":
		if eventSettings.Care {
			var logBuilder strings.Builder
			logBuilder.WriteString(webhookId)
			logBuilder.WriteString(" blrec ")
			logBuilder.WriteString(blrecEventNameMap[hookType])
			logBuilder.WriteString("：")
			logBuilder.Write(getter.GetStringBytes("data", "path"))
			log.Info(logBuilder.String())
		}
		if eventSettings.Notify {
			var msgTitleBuilder strings.Builder
			msgTitleBuilder.WriteString(blrecEventNameMap[hookType])
			var msgContentBuilder strings.Builder
			msgContentBuilder.WriteString("- 文件路径：")
			msgContentBuilder.Write(getter.GetStringBytes("data", "path"))
			msgContentBuilder.WriteString("\n- 文件大小：")
			msgContentBuilder.WriteString(strconv.FormatUint(getter.GetUint64("data", "size"), 10))
			msgContentBuilder.WriteString(" 字节\n- 文件时长：")
			msgContentBuilder.WriteString(strconv.FormatUint(getter.GetUint64("data", "duration"), 10))
			msgContentBuilder.WriteString(" 秒")

			var msg = messageSender.Message{
				Title:   msgTitleBuilder.String(),
				Content: msgContentBuilder.String(),
				ID:      webhookId,
				IconURL: bilibiliInfo.GetAvatarByRoomID(getter.GetUint64("data", "room_id"), webhookId),
			}
			msg.Send()
		}
		break

	// SpaceNoEnoughEvent 硬盘空间不足
	case "SpaceNoEnoughEvent":
		if eventSettings.Care {
			var logBuilder strings.Builder
			logBuilder.WriteString(webhookId)
			logBuilder.WriteString(" blrec 硬盘空间不足：文件路径：")
			logBuilder.Write(getter.GetStringBytes("data", "path"))
			logBuilder.WriteString("；可用空间：")
			logBuilder.WriteString(formatStorageSpace(getter.GetInt64("data", "usage", "free")))
			log.Warn(logBuilder.String())
		}
		if eventSettings.Notify {
			var msgTitleBuilder strings.Builder
			msgTitleBuilder.WriteString("硬盘空间不足")
			var msgContentBuilder strings.Builder
			msgContentBuilder.WriteString("- 文件路径：")
			msgContentBuilder.Write(getter.GetStringBytes("data", "path"))
			msgContentBuilder.WriteString("\n- 磁盘总空间：")
			msgContentBuilder.WriteString(formatStorageSpace(getter.GetInt64("data", "usage", "total")))
			msgContentBuilder.WriteString("\n- 已用空间：")
			msgContentBuilder.WriteString(formatStorageSpace(getter.GetInt64("data", "usage", "used")))
			msgContentBuilder.WriteString("\n- 设定临界空间：")
			msgContentBuilder.WriteString(formatStorageSpace(getter.GetInt64("data", "threshold")))
			msgContentBuilder.WriteString("\n- 可用空间：")
			msgContentBuilder.WriteString(formatStorageSpace(getter.GetInt64("data", "usage", "free")))

			var msg = messageSender.Message{
				Title:   msgTitleBuilder.String(),
				Content: msgContentBuilder.String(),
				ID:      webhookId,
			}
			msg.Send()
		}
		break

	// Error 程序出现异常
	case "Error":
		if eventSettings.Care {
			var logBuilder strings.Builder
			logBuilder.WriteString(webhookId)
			logBuilder.WriteString(" blrec 程序出现异常：")
			logBuilder.Write(getter.GetStringBytes("data"))
			log.Warn(logBuilder.String())
		}
		if eventSettings.Notify {
			var msgContentBuilder strings.Builder
			msgContentBuilder.WriteString("```json\n")
			msgContentBuilder.Write(getter.GetStringBytes("data"))
			msgContentBuilder.WriteString("\n```")
			var msg = messageSender.Message{
				Title:   "blrec 程序出现异常",
				Content: msgContentBuilder.String(),
				ID:      webhookId,
			}
			msg.Send()
		}
		break

	default:
		var logBuilder strings.Builder
		logBuilder.WriteString(webhookId)
		logBuilder.WriteString(" 未知的 blrec webhook 请求类型：")
		logBuilder.WriteString(hookType)
		log.Warn(logBuilder.String(), content)
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

// BlrecWebhookHandler 处理 blrec 的 webhook 请求
func BlrecWebhookHandler(w http.ResponseWriter, request *http.Request) {
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
		log.Error("读取 blrec webhook 请求失败：", err.Error())
		return
	}
	go blrecTaskRunner(content)
}

var blrecSettings map[string]Event
var blrecEventNameMap = map[string]string{
	"LiveBeganEvent":                    "主播开播",
	"LiveEndedEvent":                    "主播下播",
	"RoomChangeEvent":                   "直播间信息改变",
	"RecordingStartedEvent":             "录制开始",
	"RecordingFinishedEvent":            "录制结束",
	"RecordingCancelledEvent":           "录制取消",
	"VideoFileCreatedEvent":             "视频文件创建",
	"VideoFileCompletedEvent":           "视频文件完成",
	"DanmakuFileCreatedEvent":           "弹幕文件创建",
	"DanmakuFileCompletedEvent":         "弹幕文件完成",
	"RawDanmakuFileCreatedEvent":        "原始弹幕文件创建",
	"RawDanmakuFileCompletedEvent":      "原始弹幕文件完成",
	"VideoPostprocessingCompletedEvent": "视频后处理完成",
	"SpaceNoEnoughEvent":                "硬盘空间不足",
	"Error":                             "程序出现异常",
}

func UpdateBlrecSettings(events map[string]Event) {
	blrecSettings = events
}
