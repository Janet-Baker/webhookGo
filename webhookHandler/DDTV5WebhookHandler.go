package webhookHandler

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"
	"webhookGo/bilibiliInfo"
	"webhookGo/messageSender"
)

// Core/LogModule/Opcode.cs
const (
	/* Config */
	// ReadingConfigurationFile 读取配置文件
	ReadingConfigurationFile = 10101
	// UpdateToConfigurationFile 更新到配置文件
	UpdateToConfigurationFile = 10102
	// ReadingRoomFiles 读取房间文件
	ReadingRoomFiles = 10103
	// UpdateToRoomFile 更新到房间文件
	UpdateToRoomFile = 10104
	// ModifyConfiguration 修改配置
	ModifyConfiguration = 10105
	// UpdateDetect 检测到更新
	UpdateDetect = 10106
	/* Room */
	// SuccessfullyAddedRoom 新增房间配置成功
	SuccessfullyAddedRoom = 20101
	// FailedToAddRoomConfiguration 新增房间配置失败
	FailedToAddRoomConfiguration = 20102
	// ModifyRoomRecordingConfiguration 修改房间录制配置
	ModifyRoomRecordingConfiguration = 20103
	// ModifyRoomBulletScreenConfiguration 修改房间弹幕配置
	ModifyRoomBulletScreenConfiguration = 20104
	// ModifyRoomPromptConfiguration 修改房间提示配置
	ModifyRoomPromptConfiguration = 20105
	// ManuallyTriggeringRecordingTasks 手动触发录制任务
	ManuallyTriggeringRecordingTasks = 20106
	// SuccessfullyDeletedRoom 删除房间成功
	SuccessfullyDeletedRoom = 20107
	// FailedToDeleteRoom 删除房间失败
	FailedToDeleteRoom = 20108
	// CancelRecordingSuccessful 取消录制成功
	CancelRecordingSuccessful = 20109
	// CancelRecordingFail 取消录制失败
	CancelRecordingFail = 20110
	// SuccessfullyTriggeredQuickCut 触发快剪成功
	SuccessfullyTriggeredQuickCut = 20111
	// TriggerQuickCutFail 触发快剪失败
	TriggerQuickCutFail = 20112
	// SuccessfullyAddedRecordingTask 新增录制任务成功
	SuccessfullyAddedRecordingTask = 20113
	// FailedToAddRecordingTask 新增录制任务失败
	FailedToAddRecordingTask = 20114
	/* Account */
	// UserConsentAgreement 用户同意协议
	UserConsentAgreement = 30101
	// UserDoesNotAgreeToAgreement 用户未同意协议
	UserDoesNotAgreeToAgreement = 30102
	// TriggerLoginAgain 触发重新登陆
	TriggerLoginAgain = 30103
	// LoginSuccessful 登陆成功
	LoginSuccessful = 30104
	// UpdateLoginStateCache 更新登录态缓存
	UpdateLoginStateCache = 30105
	// InvalidLoginStatus 登录态失效
	InvalidLoginStatus = 30106
	// ScanCodeConfirmation 扫码登陆确认
	ScanCodeConfirmation = 30107
	// QrCodeWaitingForScann 二维码等待扫码
	QrCodeWaitingForScann = 30108
	// ScannedCodeWaitingForConfirmation 已扫码等待确认
	ScannedCodeWaitingForConfirmation = 30109
	// QrCodeExpir 二维码已过期
	QrCodeExpir = 30110
	/* Download */
	// SaveBulletScreenFile 保存弹幕相关文件
	SaveBulletScreenFile = 40101
	// StartLiveEvent 触发开播事件
	StartLiveEvent = 40102
	// StartBroadcastingReminder 开播提醒
	StartBroadcastingReminder = 40103
	// StartRecording 开始录制
	StartRecording = 40104
	// RecordingEnd 录制结束
	RecordingEnd = 40105
	// StopLiveEvent 停止直播事件
	StopLiveEvent = 40106
	// Reconnect 录制触发重新连接
	Reconnect = 40107
	// HlsTaskStart HLS任务成功开始
	HlsTaskStart = 40108
	// FlvTaskStart FLV任务成功开始
	FlvTaskStart = 40109
)

var ddtv5Settings = make(map[int]Event)
var ddtv5IdEventTitleMap = map[int]string{
	ReadingConfigurationFile:            "ReadingConfigurationFile",
	UpdateToConfigurationFile:           "UpdateToConfigurationFile",
	ReadingRoomFiles:                    "ReadingRoomFiles",
	UpdateToRoomFile:                    "UpdateToRoomFile",
	ModifyConfiguration:                 "ModifyConfiguration",
	UpdateDetect:                        "UpdateDetect",
	SuccessfullyAddedRoom:               "SuccessfullyAddedRoom",
	FailedToAddRoomConfiguration:        "FailedToAddRoomConfiguration",
	ModifyRoomRecordingConfiguration:    "ModifyRoomRecordingConfiguration",
	ModifyRoomBulletScreenConfiguration: "ModifyRoomBulletScreenConfiguration",
	ModifyRoomPromptConfiguration:       "ModifyRoomPromptConfiguration",
	ManuallyTriggeringRecordingTasks:    "ManuallyTriggeringRecordingTasks",
	SuccessfullyDeletedRoom:             "SuccessfullyDeletedRoom",
	FailedToDeleteRoom:                  "FailedToDeleteRoom",
	CancelRecordingSuccessful:           "CancelRecordingSuccessful",
	CancelRecordingFail:                 "CancelRecordingFail",
	SuccessfullyTriggeredQuickCut:       "SuccessfullyTriggeredQuickCut",
	TriggerQuickCutFail:                 "TriggerQuickCutFail",
	SuccessfullyAddedRecordingTask:      "SuccessfullyAddedRecordingTask",
	FailedToAddRecordingTask:            "FailedToAddRecordingTask",
	UserConsentAgreement:                "UserConsentAgreement",
	UserDoesNotAgreeToAgreement:         "UserDoesNotAgreeToAgreement",
	TriggerLoginAgain:                   "TriggerLoginAgain",
	LoginSuccessful:                     "LoginSuccessful",
	UpdateLoginStateCache:               "UpdateLoginStateCache",
	InvalidLoginStatus:                  "InvalidLoginStatus",
	ScanCodeConfirmation:                "ScanCodeConfirmation",
	QrCodeWaitingForScann:               "QrCodeWaitingForScann",
	ScannedCodeWaitingForConfirmation:   "ScannedCodeWaitingForConfirmation",
	QrCodeExpir:                         "QrCodeExpir",
	SaveBulletScreenFile:                "SaveBulletScreenFile",
	StartLiveEvent:                      "StartLiveEvent",
	StartBroadcastingReminder:           "StartBroadcastingReminder",
	StartRecording:                      "StartRecording",
	RecordingEnd:                        "RecordingEnd",
	StopLiveEvent:                       "StopLiveEvent",
	Reconnect:                           "Reconnect",
	HlsTaskStart:                        "HlsTaskStart",
	FlvTaskStart:                        "FlvTaskStart",
}
var ddtv5IdEventNameMap = map[int]string{
	ReadingConfigurationFile:            "读取配置文件",
	UpdateToConfigurationFile:           "更新到配置文件",
	ReadingRoomFiles:                    "读取房间文件",
	UpdateToRoomFile:                    "更新到房间文件",
	ModifyConfiguration:                 "修改配置",
	UpdateDetect:                        "检测到更新",
	SuccessfullyAddedRoom:               "新增房间配置成功",
	FailedToAddRoomConfiguration:        "新增房间配置失败",
	ModifyRoomRecordingConfiguration:    "修改房间录制配置",
	ModifyRoomBulletScreenConfiguration: "修改房间弹幕配置",
	ModifyRoomPromptConfiguration:       "修改房间提示配置",
	ManuallyTriggeringRecordingTasks:    "手动触发录制任务",
	SuccessfullyDeletedRoom:             "删除房间成功",
	FailedToDeleteRoom:                  "删除房间失败",
	CancelRecordingSuccessful:           "取消录制成功",
	CancelRecordingFail:                 "取消录制失败",
	SuccessfullyTriggeredQuickCut:       "触发快剪成功",
	TriggerQuickCutFail:                 "触发快剪失败",
	SuccessfullyAddedRecordingTask:      "新增录制任务成功",
	FailedToAddRecordingTask:            "新增录制任务失败",
	UserConsentAgreement:                "用户同意协议",
	UserDoesNotAgreeToAgreement:         "用户未同意协议",
	TriggerLoginAgain:                   "触发重新登陆",
	LoginSuccessful:                     "登陆成功",
	UpdateLoginStateCache:               "更新登录态缓存",
	InvalidLoginStatus:                  "登录态失效",
	ScanCodeConfirmation:                "扫码登陆确认",
	QrCodeWaitingForScann:               "二维码等待扫码",
	ScannedCodeWaitingForConfirmation:   "已扫码等待确认",
	QrCodeExpir:                         "二维码已过期",
	SaveBulletScreenFile:                "保存弹幕相关文件",
	StartLiveEvent:                      "开始直播",
	StartBroadcastingReminder:           "开播提醒",
	StartRecording:                      "开始录制",
	RecordingEnd:                        "录制结束",
	StopLiveEvent:                       "直播结束",
	Reconnect:                           "录制触发重新连接",
	HlsTaskStart:                        "HLS任务成功开始",
	FlvTaskStart:                        "FLV任务成功开始",
}

// 独立设置一个变量，方便测试插桩
var ddtv5WebhookHandler func(content []byte) = ddtv5TaskRunner

// DDTV5Data DDTV5 数据 可能为null所以需要单独拿出来。
type DDTV5Data struct {
	Title struct {
		Value string `json:"Value"`
	} `json:"Title"`
	Description struct {
		Value string `json:"Value"`
	} `json:"description"`
	LiveTime struct {
		Value int64 `json:"Value"`
	} `json:"live_time"`
	LiveStatus struct {
		Value int `json:"Value"`
	} `json:"live_status"`
	LiveStatusEndEvent bool `json:"live_status_end_event"`
	ShortId            struct {
		Value int `json:"Value"`
	} `json:"short_id"`
	AreaV2Name struct {
		Value string `json:"Value"`
	} `json:"area_v2_name"`
	AreaV2ParentName struct {
		Value string `json:"Value"`
	} `json:"area_v2_parent_name"`
	Face struct {
		Value string `json:"Value"`
	} `json:"face"`
	CoverFromUser struct {
		Value string `json:"Value"`
	} `json:"cover_from_user"`
	Keyframe struct {
		Value string `json:"Value"`
	} `json:"keyframe"`
	LockTill struct {
		Value string `json:"Value"`
	} `json:"lock_till"`
	IsLocked struct {
		Value bool `json:"Value"`
	} `json:"is_locked"`
	CurrentMode int `json:"CurrentMode"`
	DownInfo    struct {
		IsDownload          bool      `json:"IsDownload"`
		IsCut               bool      `json:"IsCut"`
		TaskType            int       `json:"taskType"`
		DownloadSize        int       `json:"DownloadSize"`
		RealTimeDownloadSpe float64   `json:"RealTimeDownloadSpe"`
		Status              int       `json:"Status"`
		StartTime           time.Time `json:"StartTime"`
		EndTime             time.Time `json:"EndTime"`
		LiveChatListener    struct {
			RoomId       int      `json:"RoomId"`
			Title        string   `json:"Title"`
			Name         string   `json:"Name"`
			File         string   `json:"File"`
			State        bool     `json:"State"`
			Register     []string `json:"Register"`
			DanmuMessage struct {
				FileName      interface{} `json:"FileName"`
				TimeStopwatch interface{} `json:"TimeStopwatch"`
				Danmu         []struct {
					Time      float64 `json:"time"`
					Type      int     `json:"type"`
					Size      int     `json:"size"`
					Color     int     `json:"color"`
					Timestamp int64   `json:"timestamp"`
					Pool      int     `json:"pool"`
					Uid       int64   `json:"uid"`
					Message   string  `json:"Message"`
					Nickname  string  `json:"Nickname"`
					LV        int     `json:"LV"`
				} `json:"Danmu"`
				SuperChat []interface{} `json:"SuperChat"`
				Gift      []interface{} `json:"Gift"`
				GuardBuy  []interface{} `json:"GuardBuy"`
			} `json:"DanmuMessage"`
			TimeStopwatch struct {
				IsRunning           bool   `json:"IsRunning"`
				Elapsed             string `json:"Elapsed"`
				ElapsedMilliseconds int64  `json:"ElapsedMilliseconds"`
				ElapsedTicks        int64  `json:"ElapsedTicks"`
			} `json:"TimeStopwatch"`
			SaveCount int `json:"SaveCount"`
		} `json:"LiveChatListener"`
		DownloadFileList struct {
			TranscodingCount          int      `json:"TranscodingCount"`
			VideoFile                 []string `json:"VideoFile"`
			DanmuFile                 []string `json:"DanmuFile"`
			SCFile                    []string `json:"SCFile"`
			GiftFile                  []string `json:"GiftFile"`
			GuardFile                 []string `json:"GuardFile"`
			CurrentOperationVideoFile string   `json:"CurrentOperationVideoFile"`
		} `json:"DownloadFileList"`
	} `json:"DownInfo"`
	Name              string `json:"Name"`
	Description1      string `json:"Description"`
	RoomId            int64  `json:"RoomId"`
	UID               int64  `json:"UID"`
	IsAutoRec         bool   `json:"IsAutoRec"`
	IsRemind          bool   `json:"IsRemind"`
	IsRecDanmu        bool   `json:"IsRecDanmu"`
	Like              bool   `json:"Like"`
	Shell             string `json:"Shell"`
	AppointmentRecord bool   `json:"AppointmentRecord"`
}

type DDTV5MessageStruct struct {
	Cmd     string     `json:"cmd"`
	Code    int        `json:"code"`
	Data    *DDTV5Data `json:"data"`
	Message string     `json:"message"`
}

func (message *DDTV5MessageStruct) GetTitle() string {
	if message.Data == nil && message.Code < 10000 {
		return "DDTV5 收到未知请求"
	}
	if message.Code > 10000 && message.Code < 40000 {
		return "DDTV5 " + ddtv5IdEventNameMap[message.Code]
	}
	if message.Code > 40000 {
		name := message.Data.Name
		if len(name) == 0 {
			name = message.Data.DownInfo.LiveChatListener.Name
		}
		if message.Code == RecordingEnd || message.Code == StopLiveEvent {
			if message.Data.IsLocked.Value {
				return name + "喜提直播间封禁"
			}
			// 封禁检测
			if message.Data.IsLocked.Value {
				return name + "喜提直播间封禁"
			} else {
				return name + " 下播了"
			}
		} else {
			return name + " " + ddtv5IdEventNameMap[message.Code]
		}
	}
	return "DDTV5 " + message.Message
}

func (message *DDTV5MessageStruct) GetContent() string {
	if message.Data == nil {
		return message.Message
	} else {
		var msgContentBuilder strings.Builder
		msgContentBuilder.WriteString("- 主播：[")
		msgContentBuilder.WriteString(message.Data.Name)
		msgContentBuilder.WriteString("](https://live.bilibili.com/")
		msgContentBuilder.WriteString(strconv.FormatInt(message.Data.RoomId, 10))
		msgContentBuilder.WriteString(")\n- 标题：")
		msgContentBuilder.WriteString(message.Data.Title.Value)
		msgContentBuilder.WriteString("\n- 分区：")
		if message.Data.AreaV2ParentName.Value == "" {
			message.Data.AreaV2ParentName.Value = bilibiliInfo.GetAreaParentName(message.Data.UID)
		}
		msgContentBuilder.WriteString(message.Data.AreaV2ParentName.Value)
		msgContentBuilder.WriteString(" - ")
		if message.Data.AreaV2Name.Value == "" {
			message.Data.AreaV2Name.Value = bilibiliInfo.GetAreaName(message.Data.UID)
		}
		msgContentBuilder.WriteString(message.Data.AreaV2Name.Value)
		if message.Code == StartLiveEvent {
			msgContentBuilder.WriteString("\n- 开播时间：")
			msgContentBuilder.WriteString(time.Unix(message.Data.LiveTime.Value, 0).Format("2006-01-02 15:04:05"))
		}
		if message.Code == RecordingEnd || message.Code == StopLiveEvent {
			if message.Data.IsLocked.Value {
				msgContentBuilder.WriteString("\n- 直播间封禁至：")
				msgContentBuilder.WriteString(message.Data.LockTill.Value)
			}
		}
		return msgContentBuilder.String()
	}
}

func (message *DDTV5MessageStruct) GetIconURL() string {
	if message.Data == nil {
		return ""
	}
	if len(message.Data.Face.Value) > 0 {
		return message.Data.Face.Value
	}
	avatar, err := bilibiliInfo.GetAvatarByUid(message.Data.UID)
	if err != nil {
		return ""
	} else {
		return avatar
	}
}

func (message *DDTV5MessageStruct) SendToAllTargets() {
	var newMessage = messageSender.GeneralPushMessage{
		Title:   message.GetTitle(),
		Content: message.GetContent(),
		IconURL: message.GetIconURL(),
	}
	newMessage.SendToAllTargets()
}

func ddtv5TaskRunner(content []byte) {
	var message DDTV5MessageStruct
	err := json.Unmarshal(content, &message)
	if err != nil {
		log.Error("解析 DDTV5 webhook 请求失败：", err)
		return
	}
	var eventSettings Event = ddtv5Settings[message.Code]
	if eventSettings.Care {
		log.Info("DDTV5", message.Message)
	}
	if eventSettings.Notify {
		if message.Code == RecordingEnd || message.Code == StopLiveEvent {
			// 封禁检测
			isLocked, locktill := bilibiliInfo.IsRoomLocked(message.Data.RoomId)
			message.Data.IsLocked.Value = isLocked
			message.Data.LockTill.Value = time.Unix(locktill, 0).Format("2006-01-02 15:04:05")
		}
		message.SendToAllTargets()
	}
	if eventSettings.HaveCommand {
		log.Info("执行命令：", eventSettings.ExecCommand)
		cmd := exec.Command(eventSettings.ExecCommand)
		err := cmd.Run()
		if err != nil {
			log.Error("执行命令失败：", err.Error())
		}
	}
}

func DDTV5WebhookHandler(c *gin.Context) {
	// 读取请求内容
	content, err := c.GetRawData()
	if err != nil {
		log.Errorf("读取 DDTV5 webhook 请求失败：%s", err.Error())
		c.Status(http.StatusBadRequest)
		return
	}
	// return 200 at first
	c.Status(http.StatusOK)
	go ddtv5WebhookHandler(content)
}

func UpdateDDTV5Settings(events map[string]Event) {
	for id, name := range ddtv5IdEventTitleMap {
		event, ok := events[name]
		if ok {
			ddtv5Settings[id] = event
		}
	}
}
