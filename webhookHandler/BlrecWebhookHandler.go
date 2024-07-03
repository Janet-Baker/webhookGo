package webhookHandler

import (
	"bytes"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
	"webhookGo/bilibiliInfo"
	"webhookGo/messageSender"
)

type BlrecMessageStruct struct {
	Id   string `json:"id" binding:"required"`
	Date string `json:"date" binding:"required"`
	Type string `json:"type" binding:"required"`
	Data struct {
		UserInfo struct {
			Name   string `json:"name"`
			Gender string `json:"gender"`
			Face   string `json:"face"`
			Uid    int64  `json:"uid"`
			Level  int    `json:"level"`
			Sign   string `json:"sign"`
		} `json:"user_info,omitempty"` // "LiveBeganEvent", "LiveEndedEvent"
		RoomInfo struct {
			Uid            int64  `json:"uid"`
			RoomId         int64  `json:"room_id"`
			ShortRoomId    int    `json:"short_room_id"`
			AreaId         int    `json:"area_id"`
			AreaName       string `json:"area_name"`
			ParentAreaId   int    `json:"parent_area_id"`
			ParentAreaName string `json:"parent_area_name"`
			LiveStatus     int    `json:"live_status"`
			LiveStartTime  int64  `json:"live_start_time"`
			Online         int    `json:"online"`
			Title          string `json:"title"`
			Cover          string `json:"cover"`
			Tags           string `json:"tags"`
			Description    string `json:"description"`
			// 下播时更新
			isLocked bool
			lockTill int64
		} `json:"room_info,omitempty"` // "LiveBeganEvent", "LiveEndedEvent", "RoomChangeEvent", "RecordingStartedEvent", "RecordingFinishedEvent", "RecordingCancelledEvent"
		RoomId    int64    `json:"room_id,omitempty"`   // "VideoFileCreatedEvent", "VideoFileCompletedEvent", "DanmakuFileCreatedEvent", "DanmakuFileCompletedEvent", "RawDanmakuFileCreatedEvent", "RawDanmakuFileCompletedEvent", "CoverImageDownloadedEvent", "VideoPostprocessingCompletedEvent", "PostprocessingCompletedEvent"
		Path      string   `json:"path,omitempty"`      // "VideoFileCreatedEvent", "VideoFileCompletedEvent", "DanmakuFileCreatedEvent", "DanmakuFileCompletedEvent", "RawDanmakuFileCreatedEvent", "RawDanmakuFileCompletedEvent", "CoverImageDownloadedEvent", "VideoPostprocessingCompletedEvent", "SpaceNoEnoughEvent"
		Files     []string `json:"files,omitempty"`     // "PostprocessingCompletedEvent"
		Threshold int64    `json:"threshold,omitempty"` // "SpaceNoEnoughEvent"
		Usage     struct {
			Total int64 `json:"total"`
			Used  int64 `json:"used"`
			Free  int64 `json:"free"`
		} `json:"usage,omitempty"` // "SpaceNoEnoughEvent"
		Name   string `json:"name,omitempty"`   // "Error"
		Detail string `json:"detail,omitempty"` // "Error"
	} `json:"data"`
}

func (message *BlrecMessageStruct) GetTitle() string {
	var sb strings.Builder
	switch message.Type {
	case "LiveEndedEvent":
		sb.WriteString(message.Data.UserInfo.Name)
		if message.Data.RoomInfo.isLocked {
			sb.WriteString(" 喜提直播间封禁")
		} else {
			sb.WriteString(" 直播结束")
		}
	case "SpaceNoEnoughEvent":
		return "blrec 硬盘空间不足"
	case "Error":
		return "blrec 出现异常 " + message.Data.Name
	default:
		sb.WriteString(message.Data.UserInfo.Name)
		sb.WriteString(" ")
		sb.WriteString(blrecEventNameMap[message.Type])
	}
	return sb.String()
}

func (message *BlrecMessageStruct) GetContent() string {
	if message.Type == "Error" {
		return message.Data.Detail
	}
	var sb strings.Builder
	switch message.Type {
	case "SpaceNoEnoughEvent":
		sb.WriteString("剩余空间：")
		sb.WriteString(formatStorageSpace(message.Data.Usage.Free))
		sb.WriteString("，阈值：")
		sb.WriteString(formatStorageSpace(message.Data.Threshold))
		return sb.String()
	case "LiveBeganEvent", "LiveEndedEvent", "RoomChangeEvent", "RecordingStartedEvent", "RecordingFinishedEvent", "RecordingCancelledEvent":
		sb.WriteString("- 主播：[")
		sb.WriteString(message.Data.UserInfo.Name)
		sb.WriteString("](https://live.bilibili.com/")
		sb.WriteString(strconv.FormatInt(message.Data.RoomInfo.RoomId, 10))
		sb.WriteString(")\n- 标题：")
		sb.WriteString(message.Data.RoomInfo.Title)
		sb.WriteString("\n- 分区：")
		sb.WriteString(message.Data.RoomInfo.ParentAreaName)
		sb.WriteString(" - ")
		sb.WriteString(message.Data.RoomInfo.AreaName)
		switch message.Type {
		case "LiveBeganEvent":
			sb.WriteString("\n- 开播时间：")
			sb.WriteString(time.Unix(message.Data.RoomInfo.LiveStartTime, 0).Format("2006-01-02 15:04:05"))
		case "LiveEndedEvent":
			if message.Data.RoomInfo.isLocked {
				sb.WriteString("\n- 封禁到：")
				sb.WriteString(time.Unix(message.Data.RoomInfo.lockTill, 0).Format("2006-01-02 15:04:05"))
			}
		}
	case "VideoFileCreatedEvent", "VideoFileCompletedEvent", "DanmakuFileCreatedEvent", "DanmakuFileCompletedEvent", "RawDanmakuFileCreatedEvent", "RawDanmakuFileCompletedEvent", "CoverImageDownloadedEvent", "VideoPostprocessingCompletedEvent", "PostprocessingCompletedEvent":
		sb.WriteString("- 主播：[")
		sb.WriteString(message.Data.UserInfo.Name)
		sb.WriteString("](https://live.bilibili.com/")
		sb.WriteString(strconv.FormatInt(message.Data.RoomInfo.RoomId, 10))
		sb.WriteString(")\n")
		sb.WriteString("- 时间：")
		sb.WriteString(message.Date)
		sb.WriteString("\n- 文件：\n")
		if message.Type == "PostprocessingCompletedEvent" {
			for i := 0; i < len(message.Data.Files); i++ {
				sb.WriteString(message.Data.Files[i])
				sb.WriteString("\n")
			}
		} else {
			sb.WriteString(message.Data.Path)
		}
	}
	return sb.String()
}

func (message *BlrecMessageStruct) GetIconURL() string {
	return message.Data.UserInfo.Face
}

func (message *BlrecMessageStruct) SendToAllTargets() {
	var msg = messageSender.GeneralPushMessage{
		Title:   message.GetTitle(),
		Content: message.GetContent(),
		IconURL: message.GetIconURL(),
	}
	go msg.SendToAllTargets()
}

// BlrecWebhookHandler 处理 blrec 的 webhook 请求
func BlrecWebhookHandler(c *gin.Context) {
	if log.IsLevelEnabled(log.DebugLevel) {
		b, e := c.GetRawData()
		log.Debug(string(b), e)
		c.Request.Body = io.NopCloser(bytes.NewReader(b))
	}
	// 读取请求内容
	var message BlrecMessageStruct
	err := c.BindJSON(&message)
	if err != nil {
		log.Error("读取 blrec webhook 请求失败：", err.Error())
		log.Trace(c.GetRawData())
		c.Status(http.StatusBadRequest)
		return
	}
	c.Status(http.StatusOK)
	if registerId(message.Id) {
		return
	}
	if _, ok := blrecEventNameMap[message.Type]; !ok {
		b, e := c.GetRawData()
		s := string(b)
		if len(s) > 128 {
			s = s[:128]
		}
		log.Warn("未知事件类型：", message.Type, '\n', s, e)
		return
	}
	var eventSettings Event = blrecSettings[c.FullPath()][message.Type]
	if eventSettings.Care {
		log.Info("blrec" + blrecEventNameMap[message.Type])
	}
	if eventSettings.Notify {
		switch message.Type {
		case "LiveEndedEvent":
			message.Data.RoomInfo.isLocked, message.Data.RoomInfo.lockTill = bilibiliInfo.IsRoomLocked(message.Data.RoomInfo.RoomId)
		case "VideoFileCreatedEvent", "VideoFileCompletedEvent", "DanmakuFileCreatedEvent", "DanmakuFileCompletedEvent", "RawDanmakuFileCreatedEvent", "RawDanmakuFileCompletedEvent", "CoverImageDownloadedEvent", "VideoPostprocessingCompletedEvent", "PostprocessingCompletedEvent":
			message.Data.RoomInfo.RoomId = message.Data.RoomId
			fallthrough
		case "RoomChangeEvent", "RecordingStartedEvent", "RecordingFinishedEvent", "RecordingCancelledEvent":
			// get icon
			message.Data.UserInfo.Face, _ = bilibiliInfo.GetAvatarByRoomID(message.Data.RoomInfo.RoomId)
			message.Data.UserInfo.Name, _ = bilibiliInfo.GetUsernameByRoomId(message.Data.RoomInfo.RoomId)
		}

		message.SendToAllTargets()
	}
	if eventSettings.HaveCommand {
		go execCommand(eventSettings.ExecCommand)
	}
}

var blrecSettings = make(map[string]map[string]Event)
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
	"CoverImageDownloadedEvent":         "封面图片下载完成",
	"VideoPostprocessingCompletedEvent": "视频后处理完成",
	"PostprocessingCompletedEvent":      "后处理完成",
	"SpaceNoEnoughEvent":                "硬盘空间不足",
	"Error":                             "程序出现异常",
}

func UpdateBlrecSettings(path string, events map[string]Event) {
	blrecSettings[path] = events
}
