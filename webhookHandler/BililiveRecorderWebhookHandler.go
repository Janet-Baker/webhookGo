package webhookHandler

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"strings"
	"time"
	"webhookGo/bilibiliInfo"
	"webhookGo/messageSender"
)

type BililiveRecorderMessageStruct struct {
	EventType      string    `json:"EventType" binding:"required"`
	EventTimestamp time.Time `json:"EventTimestamp" binding:"required"`
	EventId        string    `json:"EventId" binding:"required,min=32,max=36"`
	EventData      struct {
		RoomId           int64  `json:"RoomId"`
		ShortId          int    `json:"ShortId"`
		Name             string `json:"Name"`
		Title            string `json:"Title"`
		AreaNameParent   string `json:"AreaNameParent"`
		AreaNameChild    string `json:"AreaNameChild"`
		Recording        bool   `json:"Recording"`
		Streaming        bool   `json:"Streaming"`
		DanmakuConnected bool   `json:"DanmakuConnected"`
		// 以下内容只有特定事件才有
		SessionId     string  `json:"SessionId"`
		RelativePath  string  `json:"RelativePath"`
		FileOpenTime  string  `json:"FileOpenTime"`
		FileCloseTime string  `json:"FileCloseTime"`
		FileSize      int64   `json:"FileSize"`
		Duration      float64 `json:"Duration"`
		// 下播时更新
		isLocked bool
		lockTill int64
	} `json:"EventData"`
}

func (message *BililiveRecorderMessageStruct) GetTitle() string {
	if message.EventType == "StreamEnded" {
		if message.EventData.isLocked {
			return message.EventData.Name + " 喜提直播间封禁"
		}
		return message.EventData.Name + " 直播结束"
	}
	return message.EventData.Name + " " + bililiveRecorderEventName[message.EventType]
}

func (message *BililiveRecorderMessageStruct) GetContent() string {
	var contentBuilder strings.Builder
	contentBuilder.WriteString("- 主播：[")
	contentBuilder.WriteString(message.EventData.Name)
	contentBuilder.WriteString("](https://live.bilibili.com/")
	contentBuilder.WriteString(strconv.FormatInt(message.EventData.RoomId, 10))
	contentBuilder.WriteString(")\n- 标题：")
	contentBuilder.WriteString(message.EventData.Title)
	contentBuilder.WriteString("\n- 分区：")
	contentBuilder.WriteString(message.EventData.AreaNameParent)
	contentBuilder.WriteString(" - ")
	contentBuilder.WriteString(message.EventData.AreaNameChild)

	switch message.EventType {
	case "SessionStarted", "SessionEnded":
		contentBuilder.WriteString("\n- 录制任务：")
		contentBuilder.WriteString(message.EventData.SessionId)

	case "FileOpening":
		contentBuilder.WriteString("\n- 录制任务：")
		contentBuilder.WriteString(message.EventData.SessionId)
		contentBuilder.WriteString("\n- 文件：")
		contentBuilder.WriteString(message.EventData.RelativePath)
		contentBuilder.WriteString("\n- 文件打开时间：")
		contentBuilder.WriteString(message.EventData.FileOpenTime)
		fallthrough
	case "FileClosed":
		contentBuilder.WriteString("\n- 时长：")
		contentBuilder.WriteString(secondsToString(message.EventData.Duration))
		contentBuilder.WriteString("\n- 文件大小：")
		contentBuilder.WriteString(formatStorageSpace(message.EventData.FileSize))
		contentBuilder.WriteString("\n- 文件关闭时间：")
		contentBuilder.WriteString(message.EventData.FileCloseTime)
		break

	case "StreamStarted":
		liveStatus, liveTime := bilibiliInfo.GetLiveStatusString(message.EventData.RoomId)
		if liveStatus == 1 {
			contentBuilder.WriteString("\n- 开播时间：")
			contentBuilder.WriteString(liveTime)
		}

	case "StreamEnded":
		if message.EventData.isLocked {
			contentBuilder.WriteString("\n- 封禁到：")
			contentBuilder.WriteString(time.Unix(message.EventData.lockTill, 0).Local().Format("2006-01-02 15:04:05"))
		}

	}
	return contentBuilder.String()
}

func (message *BililiveRecorderMessageStruct) GetIconURL() string {
	return bilibiliInfo.GetAvatarByRoomID(message.EventData.RoomId)
}

func (message *BililiveRecorderMessageStruct) SendToAllTargets() {
	var msg = messageSender.GeneralPushMessage{
		Title:   message.GetTitle(),
		Content: message.GetContent(),
		IconURL: message.GetIconURL(),
	}
	go msg.SendToAllTargets()
}

// BililiveRecorderWebhookHandler 处理 BililiveRecoder 的 webhook 请求
func BililiveRecorderWebhookHandler(c *gin.Context) {
	var message BililiveRecorderMessageStruct
	// 读取请求内容
	err := c.BindJSON(&message)
	if err != nil {
		log.Error("读取 BililiveRecoder webhook 请求失败：", err.Error())
		c.Status(http.StatusBadRequest)
		return
	}
	// return 200 at first
	c.Status(http.StatusOK)
	// 处理请求
	if registerId(message.EventId) {
		return
	}

	var eventSettings Event = bililiveRecorderSettings[c.FullPath()][message.EventType]
	if eventSettings.Care {
		log.Info("BililiveRecorder ", message.EventType)
	}
	if eventSettings.Notify {
		if message.EventType == "StreamEnded" {
			message.EventData.isLocked, message.EventData.lockTill = bilibiliInfo.IsRoomLocked(message.EventData.RoomId)
		}
		message.SendToAllTargets()
	}
	if eventSettings.HaveCommand {
		go execCommand(eventSettings.ExecCommand)
	}
}

var bililiveRecorderEventName = map[string]string{
	"SessionStarted": "录制开始",
	"FileOpening":    "文件打开",
	"FileClosed":     "文件关闭",
	"SessionEnded":   "录制结束",
	"StreamStarted":  "直播开始",
	"StreamEnded":    "直播结束",
}

var bililiveRecorderSettings = make(map[string]map[string]Event)

func UpdateBililiveRecorderSettings(path string, events map[string]Event) {
	bililiveRecorderSettings[path] = events
}
