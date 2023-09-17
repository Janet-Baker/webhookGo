package webhookHandler

import (
	jsoniter "github.com/json-iterator/go"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"strings"
	"time"
	"webhookTemplate/messageSender"
)

// BlrecTaskRunner 根据响应体内容，执行任务
func blrecTaskRunner(content []byte) {
	log.Info("收到 blrec webhook 请求")
	if log.IsLevelEnabled(log.DebugLevel) {
		log.Debug(string(content))
	}

	// 判断是否是重复的webhook请求
	webhookId := jsoniter.Get(content, "id").ToString()
	webhookMessageIdListLock.Lock()
	if webhookMessageIdList.IsContain(webhookId) {
		webhookMessageIdListLock.Unlock()
		var logBuilder strings.Builder
		logBuilder.WriteString("重复的webhook请求：")
		logBuilder.WriteString(webhookId)
		log.Warn(logBuilder.String())
		return
	} else {
		webhookMessageIdList.EnQueue(webhookId)
		webhookMessageIdListLock.Unlock()
	}

	// 判断事件类型
	hookType := jsoniter.Get(content, "type").ToString()
	switch hookType {
	// LiveBeganEvent 主播开播
	case "LiveBeganEvent":
		// 构造日志
		var logBuilder strings.Builder
		logBuilder.WriteString("blrec 主播开播：")
		logBuilder.WriteString(jsoniter.Get(content, "data", "user_info", "name").ToString())
		log.Info(logBuilder.String())

		// 构造消息
		var msgTitleBuilder strings.Builder
		msgTitleBuilder.WriteString(jsoniter.Get(content, "data", "user_info", "name").ToString())
		msgTitleBuilder.WriteString(" 开播了")
		var msgContentBuilder strings.Builder
		msgContentBuilder.WriteString("- 主播：[")
		msgContentBuilder.WriteString(jsoniter.Get(content, "data", "user_info", "name").ToString())
		msgContentBuilder.WriteString("](https://live.bilibili.com/")
		msgContentBuilder.WriteString(jsoniter.Get(content, "data", "room_info", "room_id").ToString())
		msgContentBuilder.WriteString(")\n- 标题：")
		msgContentBuilder.WriteString(jsoniter.Get(content, "data", "room_info", "title").ToString())
		msgContentBuilder.WriteString("\n- 分区：")
		msgContentBuilder.WriteString(jsoniter.Get(content, "data", "room_info", "parent_area_name").ToString())
		msgContentBuilder.WriteString(" - ")
		msgContentBuilder.WriteString(jsoniter.Get(content, "data", "room_info", "area_name").ToString())
		msgContentBuilder.WriteString("\n- 开播时间：")
		msgContentBuilder.WriteString(time.Unix(jsoniter.Get(content, "data", "room_info", "live_start_time").ToInt64(), 0).Local().Format("2006-01-02 15:04:05"))

		var msg = messageSender.Message{
			Title:   msgTitleBuilder.String(),
			Content: msgContentBuilder.String(),
		}
		msg.Send()
		break

	// LiveEndedEvent 主播下播
	case "LiveEndedEvent":
		var logBuilder strings.Builder
		logBuilder.WriteString("blrec 主播下播：")
		logBuilder.WriteString(jsoniter.Get(content, "data", "user_info", "name").ToString())
		log.Info(logBuilder.String())

		// 构造消息
		var msgTitleBuilder strings.Builder
		msgTitleBuilder.WriteString(jsoniter.Get(content, "data", "user_info", "name").ToString())
		msgTitleBuilder.WriteString(" 下播了")
		var msgContentBuilder strings.Builder
		msgContentBuilder.WriteString("- 主播：[")
		msgContentBuilder.WriteString(jsoniter.Get(content, "data", "user_info", "name").ToString())
		msgContentBuilder.WriteString("](https://live.bilibili.com/")
		msgContentBuilder.WriteString(jsoniter.Get(content, "data", "room_info", "room_id").ToString())
		msgContentBuilder.WriteString(")\n- 标题：")
		msgContentBuilder.WriteString(jsoniter.Get(content, "data", "room_info", "title").ToString())
		msgContentBuilder.WriteString("\n- 分区：")
		msgContentBuilder.WriteString(jsoniter.Get(content, "data", "room_info", "parent_area_name").ToString())
		msgContentBuilder.WriteString(" - ")
		msgContentBuilder.WriteString(jsoniter.Get(content, "data", "room_info", "area_name").ToString())

		var msg = messageSender.Message{
			Title:   msgTitleBuilder.String(),
			Content: msgContentBuilder.String(),
		}
		msg.Send()
		break

	// RoomChangeEvent 直播间信息改变
	case "RoomChangeEvent":
		if log.GetLevel() == log.DebugLevel {
			var logBuilder strings.Builder
			logBuilder.WriteString("blrec 直播间信息改变：")
			logBuilder.WriteString(jsoniter.Get(content, "data", "user_info", "room_id").ToString())
			log.Debug(logBuilder.String())
		}
		break

	// RecordingStartedEvent 录制开始
	case "RecordingStartedEvent":
		var logBuilder strings.Builder
		logBuilder.WriteString("blrec 录制开始：room_id ")
		logBuilder.WriteString(jsoniter.Get(content, "data", "room_info", "room_id").ToString())
		log.Info(logBuilder.String())
		break

	// RecordingFinishedEvent 录制结束
	case "RecordingFinishedEvent":
		var logBuilder strings.Builder
		logBuilder.WriteString("blrec 录制结束：room_id ")
		logBuilder.WriteString(jsoniter.Get(content, "data", "room_info", "room_id").ToString())
		log.Info(logBuilder.String())
		break

	// RecordingCancelledEvent 录制取消
	case "RecordingCancelledEvent":
		var logBuilder strings.Builder
		logBuilder.WriteString("blrec 录制取消：room_id ")
		logBuilder.WriteString(jsoniter.Get(content, "data", "room_info", "room_id").ToString())
		log.Info(logBuilder.String())
		break

	// VideoFileCreatedEvent 视频文件创建
	case "VideoFileCreatedEvent":
		if log.IsLevelEnabled(log.DebugLevel) {
			var logBuilder strings.Builder
			logBuilder.WriteString("blrec 视频文件创建：")
			logBuilder.WriteString(jsoniter.Get(content, "data", "path").ToString())
			log.Debug(logBuilder.String())
		}
		break

	// VideoFileCompletedEvent 视频文件完成
	case "VideoFileCompletedEvent":
		var logBuilder strings.Builder
		logBuilder.WriteString("blrec 视频文件完成：")
		logBuilder.WriteString(jsoniter.Get(content, "data", "path").ToString())
		log.Info(logBuilder.String())
		// 构造消息
		/*var msgTitleBuilder strings.Builder
		msgTitleBuilder.WriteString(jsoniter.Get(content, "data", "room_id").ToString())
		msgTitleBuilder.WriteString(" 视频文件完成")
		var msgContentBuilder strings.Builder
		msgContentBuilder.WriteString("文件路径：")
		msgContentBuilder.WriteString(jsoniter.Get(content, "data", "path").ToString())
		var msg = messageSender.Message{
			Title:   msgTitleBuilder.String(),
			Content: msgContentBuilder.String(),
		}
		msg.Send()*/
		break

	// DanmakuFileCreatedEvent 弹幕文件创建
	case "DanmakuFileCreatedEvent":
		if log.IsLevelEnabled(log.DebugLevel) {
			var logBuilder strings.Builder
			logBuilder.WriteString("blrec 弹幕文件创建：")
			logBuilder.WriteString(jsoniter.Get(content, "data", "path").ToString())
			log.Debug(logBuilder.String())
		}
		break

	// DanmakuFileCompletedEvent 弹幕文件完成
	case "DanmakuFileCompletedEvent":
		var logBuilder strings.Builder
		logBuilder.WriteString("blrec 弹幕文件完成：")
		logBuilder.WriteString(jsoniter.Get(content, "data", "path").ToString())
		log.Info(logBuilder.String())
		break

	// RawDanmakuFileCreatedEvent 原始弹幕文件创建
	case "RawDanmakuFileCreatedEvent":
		if log.IsLevelEnabled(log.DebugLevel) {
			var logBuilder strings.Builder
			logBuilder.WriteString("blrec 原始弹幕文件创建：")
			logBuilder.WriteString(jsoniter.Get(content, "data", "path").ToString())
			log.Debug(logBuilder.String())
		}
		break

	// RawDanmakuFileCompletedEvent 原始弹幕文件完成
	case "RawDanmakuFileCompletedEvent":
		if log.IsLevelEnabled(log.DebugLevel) {
			var logBuilder strings.Builder
			logBuilder.WriteString("blrec 原始弹幕文件完成：")
			logBuilder.WriteString(jsoniter.Get(content, "data", "path").ToString())
			log.Debug(logBuilder.String())
		}
		break

	// VideoPostprocessingCompletedEvent 视频后处理完成
	case "VideoPostprocessingCompletedEvent":
		if log.IsLevelEnabled(log.DebugLevel) {
			var logBuilder strings.Builder
			logBuilder.WriteString("blrec 视频后处理完成：")
			logBuilder.WriteString(jsoniter.Get(content, "data", "path").ToString())
			log.Debug(logBuilder.String())
		}
		break

	// SpaceNoEnoughEvent 硬盘空间不足
	case "SpaceNoEnoughEvent":
		var logBuilder strings.Builder
		logBuilder.WriteString("blrec 硬盘空间不足：文件路径：")
		logBuilder.WriteString(jsoniter.Get(content, "data", "path").ToString())
		logBuilder.WriteString("；可用空间：")
		logBuilder.WriteString(jsoniter.Get(content, "data", "usage", "free").ToString())
		log.Warn(logBuilder.String())
		break

	// Error 程序出现异常
	case "Error":
		var logBuilder strings.Builder
		logBuilder.WriteString("blrec 程序出现异常：")
		logBuilder.WriteString(jsoniter.Get(content, "data").ToString())
		log.Warn(logBuilder.String())
		break

	default:
		var logBuilder strings.Builder
		logBuilder.WriteString("未知的 blrec webhook 请求类型：")
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
