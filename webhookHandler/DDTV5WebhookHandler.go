package webhookHandler

import (
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

var ddtv5Settings = make(map[string]map[int]Event)
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

// TimeStopwatch
// DDTV5 的开发者把C#的Stopwatch也序列化传出来了……
type TimeStopwatch struct {
	Elapsed             string `json:"Elapsed"`
	ElapsedMilliseconds int64  `json:"ElapsedMilliseconds"`
	ElapsedTicks        int64  `json:"ElapsedTicks"`
	IsRunning           bool   `json:"IsRunning"`
}

// LiveChatListener 在开播时(StartLiveEvent)这玩意可以为null的
type LiveChatListener struct {
	RoomId       int      `json:"RoomId"`   // 直播间房间号(长号)
	Title        string   `json:"Title"`    // 标题
	Name         string   `json:"Name"`     // 昵称
	File         string   `json:"File"`     // 输出文件路径构造
	State        bool     `json:"State"`    // 正在监听
	Register     []string `json:"Register"` // "DetectRoom_LiveStart" | "DanmaOnlyWindow" | "VlcPlayWindow"
	DanmuMessage struct {
		FileName      string         `json:"FileName"` // @NullAble 已保存的弹幕文件位置
		TimeStopwatch *TimeStopwatch `json:"TimeStopwatch"`
		Danmu         []struct {
			Time      float64 `json:"time"`      // 弹幕在视频里的时间
			Type      int     `json:"type"`      // 弹幕类型
			Size      int     `json:"size"`      // 弹幕大小
			Color     int     `json:"color"`     // 弹幕颜色
			Timestamp int64   `json:"timestamp"` // 时间戳
			Pool      int     `json:"pool"`      // 弹幕池
			Uid       int64   `json:"uid"`       // 发送者UID
			Message   string  `json:"Message"`   // 弹幕信息
			Nickname  string  `json:"Nickname"`  // 发送人昵称
			LV        int     `json:"LV"`        // 发送人舰队等级
		} `json:"Danmu"` // 弹幕信息
		SuperChat []struct {
			Time         float64 `json:"Time"`         // 送礼的时候在视频里的时间
			Timestamp    int64   `json:"Timestamp"`    // 时间戳
			UserId       int64   `json:"UserId"`       // 打赏人UID
			UserName     string  `json:"UserName"`     // 打赏人昵称
			Price        float64 `json:"Price"`        // SC金额
			Message      string  `json:"Message"`      // SC消息内容
			MessageTrans string  `json:"MessageTrans"` // SC消息内容_翻译后
			TimeLength   int     `json:"TimeLength"`   // SC消息的持续时间
		} `json:"SuperChat"` // 醒目留言
		Gift []struct {
			Time      float64 `json:"Time"`      // 送礼的时候在视频里的时间
			Timestamp int64   `json:"Timestamp"` // 时间戳
			UserId    int64   `json:"UserId"`    // 送礼人UID
			UserName  string  `json:"UserName"`  // 送礼人昵称
			Amount    int     `json:"Amount"`    // 礼物数量
			Price     float64 `json:"Price"`     // 花费 单位:金瓜子
			GiftName  string  `json:"GiftName"`  // 礼物名称
		} `json:"Gift"` // 礼物信息
		GuardBuy []struct {
			Time       float64 `json:"Time"`       // 送礼的时候在视频里的时间
			Timestamp  int64   `json:"Timestamp"`  // 时间戳
			UserId     int64   `json:"UserId"`     // 上舰人UID
			UserName   string  `json:"UserName"`   // 上舰人昵称
			Number     int     `json:"Number"`     // 开通了几个月
			GuradName  string  `json:"GuradName"`  // 开通的舰队名称
			GuardLevel int     `json:"GuardLevel"` // 舰队等级：1-总督 2-提督 3-舰长
			Price      float64 `json:"Price"`      // 花费 单位:金瓜子
		} `json:"GuardBuy"` // 舰队信息
	} `json:"DanmuMessage"`
	TimeStopwatch *TimeStopwatch `json:"TimeStopwatch"`
	SaveCount     int            `json:"SaveCount"`
}

// DDTV5Data DDTV5 数据 可能为null所以需要单独拿出来。
type DDTV5Data struct {
	Title struct {
		Value string `json:"Value"`
	} `json:"Title"` // 标题
	Description struct {
		Value string `json:"Value"`
	} `json:"description"` // 主播简介
	LiveTime struct {
		Value int64 `json:"Value"`
	} `json:"live_time"` // 开播时间(未开播时为-62170012800,live_status为1时有效)
	LiveStatus struct {
		Value int `json:"Value"`
	} `json:"live_status"` // 直播状态(0:未直播   1:正在直播   2:轮播中)
	// LiveStatusEndEvent live_status_end_event 触发下播事件缓存（表示该房间目前为开播事件触发状态，但是事件还未处理完成）
	LiveStatusEndEvent bool `json:"live_status_end_event"`
	ShortId            struct {
		Value int `json:"Value"`
	} `json:"short_id"` // 直播间房间号(直播间短房间号，常见于签约主播)
	AreaV2Name struct {
		Value string `json:"Value"`
	} `json:"area_v2_name"` // 直播间新版分区名
	AreaV2ParentName struct {
		Value string `json:"Value"`
	} `json:"area_v2_parent_name"` // 直播间新版父分区名
	Face struct {
		Value string `json:"Value"`
	} `json:"face"` // 主播头像url
	CoverFromUser struct {
		Value string `json:"Value"`
	} `json:"cover_from_user"` // 直播封面图
	Keyframe struct {
		Value string `json:"Value"`
	} `json:"keyframe"` // 直播关键帧图
	LockTill struct {
		Value string `json:"Value"`
	} `json:"lock_till"` // 直播间锁定时间戳
	IsLocked struct {
		Value bool `json:"Value"`
	} `json:"is_locked"` // 是否锁定
	CurrentMode int `json:"CurrentMode"` // 当前模式（1:FLV 2:HLS）
	DownInfo    struct {
		IsDownload          bool              `json:"IsDownload"`          // 当前是否在下载
		IsCut               bool              `json:"IsCut"`               // 是否触发瞎几把剪
		TaskType            int               `json:"taskType"`            // 任务类型
		DownloadSize        int64             `json:"DownloadSize"`        // 当前房间下载任务总大小
		RealTimeDownloadSpe float64           `json:"RealTimeDownloadSpe"` // 实时下载速度
		Status              int               `json:"Status"`              // 任务状态
		StartTime           time.Time         `json:"StartTime"`           // 任务开始时间
		EndTime             time.Time         `json:"EndTime"`             // 任务结束时间
		LiveChatListener    *LiveChatListener `json:"LiveChatListener"`
		DownloadFileList    struct {
			TranscodingCount          int      `json:"TranscodingCount"`
			VideoFile                 []string `json:"VideoFile"`
			DanmuFile                 []string `json:"DanmuFile"`
			SCFile                    []string `json:"SCFile"`
			GiftFile                  []string `json:"GiftFile"`
			GuardFile                 []string `json:"GuardFile"`
			CurrentOperationVideoFile string   `json:"CurrentOperationVideoFile"`
		} `json:"DownloadFileList"`
	} `json:"DownInfo"`
	Name              string `json:"Name"`              // 昵称
	Description1      string `json:"Description"`       // 描述
	RoomId            int64  `json:"RoomId"`            // 直播间房间号(长号)
	UID               int64  `json:"UID"`               // 主播mid
	IsAutoRec         bool   `json:"IsAutoRec"`         // 是否自动录制
	IsRemind          bool   `json:"IsRemind"`          // 是否开播提醒
	IsRecDanmu        bool   `json:"IsRecDanmu"`        // 是否录制弹幕
	Like              bool   `json:"Like"`              // 特殊标记
	Shell             string `json:"Shell"`             // 该房间录制完成后会执行的Shell命令
	AppointmentRecord bool   `json:"AppointmentRecord"` // 是否预约下一次录制
}

type DDTV5MessageStruct struct {
	Cmd     string     `json:"cmd" binding:"required"`
	Code    int        `json:"code" binding:"required"`
	Data    *DDTV5Data `json:"data"`
	Message string     `json:"message" binding:"required"`
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
		if len(message.Data.AreaV2ParentName.Value) == 0 {
			message.Data.AreaV2ParentName.Value = bilibiliInfo.GetAreaV2ParentName(message.Data.UID)
		}
		msgContentBuilder.WriteString(message.Data.AreaV2ParentName.Value)
		msgContentBuilder.WriteString(" - ")
		if len(message.Data.AreaV2Name.Value) == 0 {
			message.Data.AreaV2Name.Value = bilibiliInfo.GetAreaV2Name(message.Data.UID)
		}
		msgContentBuilder.WriteString(message.Data.AreaV2Name.Value)
		if message.Code == StartLiveEvent {
			if message.Data.LiveTime.Value > 0 {
				msgContentBuilder.WriteString("\n- 开播时间：")
				msgContentBuilder.WriteString(time.Unix(message.Data.LiveTime.Value, 0).Format("2006-01-02 15:04:05"))
			} else {

			}
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

func DDTV5WebhookHandler(c *gin.Context) {
	// 读取请求内容
	var message DDTV5MessageStruct
	err := c.BindJSON(&message)
	if err != nil {
		log.Error("读取 DDTV5 webhook 请求失败：", err)
		c.Status(http.StatusBadRequest)
		return
	}
	// return 200 at first
	c.Status(http.StatusOK)
	// DDTV5中，webhook请求是异步的，不再强烈要求立刻返回。
	var eventSettings Event = ddtv5Settings[c.FullPath()][message.Code]
	if eventSettings.Care {
		log.Info("DDTV5 ", message.Message)
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

func UpdateDDTV5Settings(path string, events map[string]Event) {
	if ddtv5Settings[path] == nil {
		ddtv5Settings[path] = make(map[int]Event)
	}
	for id, name := range ddtv5IdEventTitleMap {
		event, ok := events[name]
		if ok {
			t := ddtv5Settings[path]
			t[id] = event
			ddtv5Settings[path] = t
		}
	}
}
