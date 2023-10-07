package webhookHandler

import (
	log "github.com/sirupsen/logrus"
	"github.com/valyala/fastjson"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
	"webhookTemplate/bilibiliInfo"
	"webhookTemplate/messageSender"
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
	{
		log.Info(webhookId, "收到 blrec webhook 请求")
	}

	// 判断是否是重复的webhook请求
	webhookMessageIdListLock.Lock()
	if webhookMessageIdList.IsContain(webhookId) {
		webhookMessageIdListLock.Unlock()
		var logBuilder strings.Builder
		logBuilder.WriteString(webhookId)
		logBuilder.WriteString(" 重复的webhook请求")
		log.Warn(logBuilder.String())
		return
	} else {
		webhookMessageIdList.EnQueue(webhookId)
		webhookMessageIdListLock.Unlock()
	}

	// 判断事件类型
	hookType := string(getter.GetStringBytes("type"))
	switch hookType {
	// LiveBeganEvent 主播开播
	case "LiveBeganEvent":
		// 构造日志
		var logBuilder strings.Builder
		logBuilder.WriteString(webhookId)
		logBuilder.WriteString(" blrec 主播开播：")
		logBuilder.Write(getter.GetStringBytes("data", "user_info", "name"))
		log.Info(logBuilder.String())

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
		break

	// LiveEndedEvent 主播下播
	case "LiveEndedEvent":
		var logBuilder strings.Builder
		logBuilder.WriteString(webhookId)
		logBuilder.WriteString(" blrec 主播下播：")
		logBuilder.Write(getter.GetStringBytes("data", "user_info", "name"))
		log.Info(logBuilder.String())

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
		break

	// RoomChangeEvent 直播间信息改变
	case "RoomChangeEvent":
		var logBuilder strings.Builder
		logBuilder.WriteString(webhookId)
		logBuilder.WriteString(" blrec 直播间信息改变：")
		logBuilder.WriteString(strconv.FormatUint(getter.GetUint64("data", "room_info", "room_id"), 10))
		log.Info(logBuilder.String())
		break

	// RecordingStartedEvent 录制开始
	case "RecordingStartedEvent":
		var logBuilder strings.Builder
		logBuilder.WriteString(webhookId)
		logBuilder.WriteString(" blrec 录制开始：room_id ")
		logBuilder.WriteString(time.Unix(getter.GetInt64("data", "room_info", "room_id"), 0).Local().Format("2006-01-02 15:04:05"))
		log.Info(logBuilder.String())
		break

	// RecordingFinishedEvent 录制结束
	case "RecordingFinishedEvent":
		var logBuilder strings.Builder
		logBuilder.WriteString(webhookId)
		logBuilder.WriteString(" blrec 录制结束：room_id ")
		logBuilder.WriteString(time.Unix(getter.GetInt64("data", "room_info", "room_id"), 0).Local().Format("2006-01-02 15:04:05"))
		log.Info(logBuilder.String())
		break

	// RecordingCancelledEvent 录制取消
	case "RecordingCancelledEvent":
		var logBuilder strings.Builder
		logBuilder.WriteString(webhookId)
		logBuilder.WriteString(" blrec 录制取消：room_id ")
		logBuilder.WriteString(time.Unix(getter.GetInt64("data", "room_info", "room_id"), 0).Local().Format("2006-01-02 15:04:05"))
		log.Info(logBuilder.String())
		break

	// VideoFileCreatedEvent 视频文件创建
	case "VideoFileCreatedEvent":
		var logBuilder strings.Builder
		logBuilder.WriteString(webhookId)
		logBuilder.WriteString(" blrec 视频文件创建：")
		logBuilder.Write(getter.GetStringBytes("data", "path"))
		log.Debug(logBuilder.String())

	// VideoFileCompletedEvent 视频文件完成
	case "VideoFileCompletedEvent":
		var logBuilder strings.Builder
		logBuilder.WriteString(webhookId)
		logBuilder.WriteString(" blrec 视频文件完成：")
		logBuilder.Write(getter.GetStringBytes("data", "path"))
		log.Info(logBuilder.String())
		break

	// DanmakuFileCreatedEvent 弹幕文件创建
	case "DanmakuFileCreatedEvent":
		var logBuilder strings.Builder
		logBuilder.WriteString(webhookId)
		logBuilder.WriteString(" blrec 弹幕文件创建：")
		logBuilder.Write(getter.GetStringBytes("data", "path"))
		log.Debug(logBuilder.String())
		break

	// DanmakuFileCompletedEvent 弹幕文件完成
	case "DanmakuFileCompletedEvent":
		var logBuilder strings.Builder
		logBuilder.WriteString(webhookId)
		logBuilder.WriteString(" blrec 弹幕文件完成：")
		logBuilder.Write(getter.GetStringBytes("data", "path"))
		log.Info(logBuilder.String())
		break

	// RawDanmakuFileCreatedEvent 原始弹幕文件创建
	case "RawDanmakuFileCreatedEvent":
		var logBuilder strings.Builder
		logBuilder.WriteString(webhookId)
		logBuilder.WriteString(" blrec 原始弹幕文件创建：")
		logBuilder.Write(getter.GetStringBytes("data", "path"))
		log.Debug(logBuilder.String())
		break

	// RawDanmakuFileCompletedEvent 原始弹幕文件完成
	case "RawDanmakuFileCompletedEvent":
		var logBuilder strings.Builder
		logBuilder.WriteString(webhookId)
		logBuilder.WriteString(" blrec 原始弹幕文件完成：")
		logBuilder.Write(getter.GetStringBytes("data", "path"))
		log.Debug(logBuilder.String())
		break

	// VideoPostprocessingCompletedEvent 视频后处理完成
	case "VideoPostprocessingCompletedEvent":
		var logBuilder strings.Builder
		logBuilder.WriteString(webhookId)
		logBuilder.WriteString(" blrec 视频后处理完成：")
		logBuilder.Write(getter.GetStringBytes("data", "path"))
		log.Debug(logBuilder.String())
		break

	// SpaceNoEnoughEvent 硬盘空间不足
	case "SpaceNoEnoughEvent":
		var logBuilder strings.Builder
		logBuilder.WriteString(webhookId)
		logBuilder.WriteString(" blrec 硬盘空间不足：文件路径：")
		logBuilder.Write(getter.GetStringBytes("data", "path"))
		logBuilder.WriteString("；可用空间：")
		logBuilder.WriteString(time.Unix(getter.GetInt64("data", "usage", "free"), 0).Local().Format("2006-01-02 15:04:05"))
		log.Warn(logBuilder.String())
		break

	// Error 程序出现异常
	case "Error":
		var logBuilder strings.Builder
		logBuilder.WriteString(webhookId)
		logBuilder.WriteString(" blrec 程序出现异常：")
		logBuilder.Write(getter.GetStringBytes("data"))
		log.Warn(logBuilder.String())
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
		break

	default:
		var logBuilder strings.Builder
		logBuilder.WriteString(webhookId)
		logBuilder.WriteString(" 未知的 blrec webhook 请求类型：")
		logBuilder.WriteString(hookType)
		log.Warn(logBuilder.String())
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
	// return 200 at first
	w.WriteHeader(http.StatusOK)

	// 读取请求内容
	content, err := io.ReadAll(request.Body)
	if err != nil {
		log.Errorf("读取 blrec webhook 请求失败：%s", err.Error())
		return
	}
	go blrecTaskRunner(content)
}
