package webhookHandler

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/valyala/fastjson"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"
	"webhookGo/bilibiliInfo"
	"webhookGo/messageSender"
)

// ddtv3TaskRunner 根据响应体内容，执行任务
func ddtv3TaskRunner(path string, content []byte) {
	if log.IsLevelEnabled(log.TraceLevel) {
		log.Trace(string(content))
	}
	var p fastjson.Parser
	getter, errOfJsonParser := p.ParseBytes(content)
	if errOfJsonParser != nil {
		return
	}
	webhookId := string(getter.GetStringBytes("id"))
	if len(webhookId) < 36 {
		log.Warnln("DDTV webhook 请求的 id 读取失败", webhookId)
		return
	}
	log.Debug(webhookId + " 收到 DDTV webhook 请求")

	// 判断是否是重复的webhook请求
	if registerId(webhookId) {
		return
	}

	// 判断事件类型
	eventType := getter.GetInt("type")
	eventSettings, _ := ddtv3Settings[path][eventType]
	switch eventType {
	//	0 StartLive 主播开播
	case 0:
		if eventSettings.Care {
			var logBuilder strings.Builder
			logBuilder.WriteString("DDTV3 主播开播：")
			logBuilder.Write(getter.GetStringBytes("user_info", "name"))
			log.Info(logBuilder.String())
		}
		if eventSettings.Notify {
			// 构建消息
			// 构造消息标题
			var msgTitleBuilder strings.Builder
			msgTitleBuilder.Write(getter.GetStringBytes("user_info", "name"))
			msgTitleBuilder.WriteString(" 开播了")
			// 构造消息内容
			var msgContentBuilder strings.Builder
			msgContentBuilder.WriteString("- 主播：[")
			msgContentBuilder.Write(getter.GetStringBytes("room_Info", "uname"))
			msgContentBuilder.WriteString("](https://live.bilibili.com/")
			msgContentBuilder.WriteString(strconv.FormatUint(getter.GetUint64("room_Info", "room_id"), 10))
			msgContentBuilder.WriteString(")\n- 标题：")
			msgContentBuilder.Write(getter.GetStringBytes("room_Info", "title"))
			msgContentBuilder.WriteString("\n- 分区：")
			msgContentBuilder.Write(getter.GetStringBytes("room_Info", "area_v2_parent_name"))
			msgContentBuilder.WriteString(" - ")
			msgContentBuilder.Write(getter.GetStringBytes("room_Info", "area_v2_name"))
			msgContentBuilder.WriteString("\n- 开播时间：")
			msgContentBuilder.WriteString(time.Unix(getter.GetInt64("room_Info", "live_time"), 0).Local().Format("2006-01-02 15:04:05"))
			// 发送消息
			var msg = messageSender.GeneralPushMessage{
				Title:   msgTitleBuilder.String(),
				Content: msgContentBuilder.String(),
				IconURL: string(getter.GetStringBytes("user_info", "face")),
			}
			msg.SendToAllTargets()
		}
		break

	//	1 StopLive 主播下播
	case 1:
		if eventSettings.Care { // 输出日志
			var logBuilder strings.Builder
			logBuilder.WriteString("DDTV 主播下播：")
			logBuilder.Write(getter.GetStringBytes("room_Info", "uname"))
			log.Info(logBuilder.String())
		}
		if eventSettings.Notify {
			// 构造消息标题
			var msgTitleBuilder strings.Builder
			msgTitleBuilder.Write(getter.GetStringBytes("room_Info", "uname"))
			// 封禁检测
			isLocked, lockTill := bilibiliInfo.IsRoomLocked(getter.GetInt64("room_Info", "room_id"))
			if isLocked {
				// 主播被封号了
				msgTitleBuilder.WriteString(" 喜提直播间封禁！")
			} else {
				// 主播正常下播
				msgTitleBuilder.WriteString(" 下播了")
			}
			// 构造消息内容
			var msgContentBuilder strings.Builder
			msgContentBuilder.WriteString("- 主播：[")
			msgContentBuilder.Write(getter.GetStringBytes("room_Info", "uname"))
			msgContentBuilder.WriteString("](https://live.bilibili.com/")
			msgContentBuilder.WriteString(strconv.FormatUint(getter.GetUint64("room_Info", "room_id"), 10))
			msgContentBuilder.WriteString(")\n- 标题：")
			msgContentBuilder.Write(getter.GetStringBytes("room_Info", "title"))
			msgContentBuilder.WriteString("\n- 分区：")
			msgContentBuilder.Write(getter.GetStringBytes("room_Info", "area_v2_parent_name"))
			msgContentBuilder.WriteString(" - ")
			msgContentBuilder.Write(getter.GetStringBytes("room_Info", "area_v2_name"))
			if isLocked {
				msgContentBuilder.WriteString("\n- 封禁到：")
				msgContentBuilder.WriteString(time.Unix(lockTill, 0).Local().Format("2006-01-02 15:04:05"))
			}
			// 发送消息
			var msg = messageSender.GeneralPushMessage{
				Title:   msgTitleBuilder.String(),
				Content: msgContentBuilder.String(),
				IconURL: string(getter.GetStringBytes("user_info", "face")),
			}
			msg.SendToAllTargets()
		}
		break

	//	2 StartRec 开始录制
	//	3 RecComplete 录制结束
	//	4 CancelRec 录制被取消
	//	5 TranscodingComplete 完成转码
	//	6 SaveDanmuComplete 保存弹幕文件完成
	//	7 SaveSCComplete 保存SC文件完成
	//	8 SaveGiftComplete 保存礼物文件完成
	//	9 SaveGuardComplete 保存大航海文件完成
	//	17 WarnedByAdmin 被管理员警告
	//	18 LiveCutOff 直播被管理员切断
	case 2, 3, 4, 5, 6, 7, 8, 9, 17, 18:
		if eventSettings.Care {
			var logBuilder strings.Builder
			logBuilder.WriteString("DDTV ")
			logBuilder.WriteString(ddtv3IdEventNameMap[eventType])
			logBuilder.WriteString("：")
			logBuilder.Write(getter.GetStringBytes("room_Info", "uname"))
			log.Info(logBuilder.String())
		}
		if eventSettings.Notify {
			// 构造消息标题
			var msgTitleBuilder strings.Builder
			msgTitleBuilder.Write(getter.GetStringBytes("room_Info", "uname"))
			msgTitleBuilder.WriteString(" ")
			msgTitleBuilder.WriteString(ddtv3IdEventNameMap[eventType])
			// 构造消息内容
			var msgContentBuilder strings.Builder
			msgContentBuilder.WriteString("- 主播：[")
			msgContentBuilder.Write(getter.GetStringBytes("room_Info", "uname"))
			msgContentBuilder.WriteString("](https://live.bilibili.com/")
			msgContentBuilder.WriteString(strconv.FormatUint(getter.GetUint64("room_Info", "room_id"), 10))
			msgContentBuilder.WriteString(")\n- 标题：")
			msgContentBuilder.Write(getter.GetStringBytes("room_Info", "title"))
			msgContentBuilder.WriteString("\n- 分区：")
			msgContentBuilder.Write(getter.GetStringBytes("room_Info", "area_v2_parent_name"))
			msgContentBuilder.WriteString(" - ")
			msgContentBuilder.Write(getter.GetStringBytes("room_Info", "area_v2_name"))
			var msg = messageSender.GeneralPushMessage{
				Title:   msgTitleBuilder.String(),
				Content: msgContentBuilder.String(),
				IconURL: string(getter.GetStringBytes("user_info", "face")),
			}
			msg.SendToAllTargets()
		}
		break

	//	10 RunShellComplete 执行Shell命令完成
	//	16 ShellExecutionComplete 执行Shell命令结束
	case 10, 16:
		if eventSettings.Care {
			var logBuilder strings.Builder
			logBuilder.WriteString("DDTV ")
			logBuilder.WriteString(ddtv3IdEventNameMap[eventType])
			logBuilder.WriteString("：")
			logBuilder.Write(getter.GetStringBytes("room_Info", "Shell"))
			log.Info(logBuilder.String())
		}
		if eventSettings.Notify {
			// 构造消息标题
			var msgTitleBuilder strings.Builder
			msgTitleBuilder.Write(getter.GetStringBytes("room_Info", "uname"))
			msgTitleBuilder.WriteString(" ")
			msgTitleBuilder.WriteString(ddtv3IdEventNameMap[eventType])
			// 构造消息内容
			var msgContentBuilder strings.Builder
			if string(getter.GetStringBytes("room_Info", "uname")) != "" {
				msgContentBuilder.WriteString("- 主播：[")
				msgContentBuilder.Write(getter.GetStringBytes("room_Info", "uname"))
				msgContentBuilder.WriteString("](https://live.bilibili.com/")
				msgContentBuilder.WriteString(strconv.FormatUint(getter.GetUint64("room_Info", "room_id"), 10))
				msgContentBuilder.WriteString(")\n- 标题：")
				msgContentBuilder.Write(getter.GetStringBytes("room_Info", "title"))
				msgContentBuilder.WriteString("\n- 分区：")
				msgContentBuilder.Write(getter.GetStringBytes("room_Info", "area_v2_parent_name"))
				msgContentBuilder.WriteString(" - ")
				msgContentBuilder.Write(getter.GetStringBytes("room_Info", "area_v2_name"))
				msgContentBuilder.WriteString("\n")
			}
			msgContentBuilder.WriteString("- 命令：")
			msgContentBuilder.Write(getter.GetStringBytes("room_Info", "Shell"))
			// 发送消息
			var msg = messageSender.GeneralPushMessage{
				Title:   msgTitleBuilder.String(),
				Content: msgContentBuilder.String(),
				IconURL: string(getter.GetStringBytes("user_info", "face")),
			}
			msg.SendToAllTargets()
		}
		break

	//	11 DownloadEndMissionSuccess 下载任务成功结束
	case 11:
		if eventSettings.Care {
			var logBuilder strings.Builder
			logBuilder.WriteString("DDTV 下载任务成功结束：")
			logBuilder.Write(getter.GetStringBytes("room_Info", "uname"))
			log.Info(logBuilder.String())
		}
		if eventSettings.Notify {
			// 构造消息
			// 构造消息内容
			var msgContentBuilder strings.Builder
			msgContentBuilder.WriteString("- 主播：[")
			msgContentBuilder.Write(getter.GetStringBytes("room_Info", "uname"))
			msgContentBuilder.WriteString("](https://live.bilibili.com/")
			msgContentBuilder.WriteString(strconv.FormatUint(getter.GetUint64("room_Info", "room_id"), 10))
			msgContentBuilder.WriteString(")\n- 标题：")
			msgContentBuilder.Write(getter.GetStringBytes("room_Info", "title"))
			msgContentBuilder.WriteString("\n- 分区：")
			msgContentBuilder.Write(getter.GetStringBytes("room_Info", "area_v2_parent_name"))
			msgContentBuilder.WriteString(" - ")
			msgContentBuilder.Write(getter.GetStringBytes("room_Info", "area_v2_name"))
			// 构造消息标题
			var msgTitleBuilder strings.Builder
			msgTitleBuilder.Write(getter.GetStringBytes("room_Info", "uname"))
			// 判断是否是封禁
			isRoomLocked, lockTill := bilibiliInfo.IsRoomLocked(getter.GetInt64("room_Info", "room_id"))
			if isRoomLocked {
				// 主播被封号了
				msgTitleBuilder.WriteString(" 喜提直播间封禁！")
				msgContentBuilder.WriteString("\n- 封禁到：")
				msgContentBuilder.WriteString(time.Unix(lockTill, 0).Local().Format("2006-01-02 15:04:05"))
			} else {
				msgTitleBuilder.WriteString(" 录制完成")
			}
			var msg = messageSender.GeneralPushMessage{
				Title:   msgTitleBuilder.String(),
				Content: msgContentBuilder.String(),
				IconURL: string(getter.GetStringBytes("user_info", "face")),
			}
			msg.SendToAllTargets()
		}
		break

	//	12 SpaceIsInsufficientWarn 剩余空间不足
	//	14 LoginWillExpireSoon 登陆即将失效
	case 12, 14:
		if eventSettings.Care {
			var logBuilder strings.Builder
			logBuilder.WriteString("DDTV ")
			logBuilder.WriteString(ddtv3IdEventNameMap[eventType])
			logBuilder.WriteString("：")
			logBuilder.Write(content)
			log.Warn(logBuilder.String())
		}
		if eventSettings.Notify {
			var msg = messageSender.GeneralPushMessage{
				Title:   "DDTV " + ddtv3IdEventNameMap[eventType],
				Content: string(content),
			}
			msg.SendToAllTargets()
		}
		break

	//	13 LoginFailure 登陆失效
	case 13:
		if eventSettings.Care {
			log.Error("DDTV 登录失效")
		}
		if eventSettings.Notify {
			var msg = messageSender.GeneralPushMessage{
				Title:   "DDTV 登录失效",
				Content: string(content),
			}
			msg.SendToAllTargets()
		}
		break

	//	15 UpdateAvailable 有可用新版本
	case 15:
		if eventSettings.Care {
			var logBuilder strings.Builder
			logBuilder.WriteString(" DDTV 有可用新版本：")
			logBuilder.Write(getter.GetStringBytes("version"))
			log.Info(logBuilder.String())
		}
		if eventSettings.Notify {
			var msg = messageSender.GeneralPushMessage{
				Title:   "DDTV 有可用新版本",
				Content: string(content),
			}
			msg.SendToAllTargets()
		}
		break

	//	19 RoomLocked 直播间被封禁
	case 19:
		if eventSettings.Care {
			var logBuilder strings.Builder
			logBuilder.WriteString(" DDTV 直播间被封禁：")
			logBuilder.Write(getter.GetStringBytes("room_Info", "uname"))
			log.Info(logBuilder.String())
		}
		if eventSettings.Notify {
			// 构造消息标题
			var msgTitleBuilder strings.Builder
			msgTitleBuilder.Write(getter.GetStringBytes("room_Info", "uname"))
			msgTitleBuilder.WriteString(" 喜提直播间封禁！")
			// 构造消息内容
			var msgContentBuilder strings.Builder
			msgContentBuilder.WriteString("- 主播：[")
			msgContentBuilder.Write(getter.GetStringBytes("room_Info", "uname"))
			msgContentBuilder.WriteString("](https://live.bilibili.com/")
			msgContentBuilder.WriteString(strconv.FormatUint(getter.GetUint64("room_Info", "room_id"), 10))
			msgContentBuilder.WriteString(")\n- 标题：")
			msgContentBuilder.Write(getter.GetStringBytes("room_Info", "title"))
			msgContentBuilder.WriteString("\n- 分区：")
			msgContentBuilder.Write(getter.GetStringBytes("room_Info", "area_v2_parent_name"))
			msgContentBuilder.WriteString(" - ")
			msgContentBuilder.Write(getter.GetStringBytes("room_Info", "area_v2_name"))
			msgContentBuilder.WriteString("\n- 封禁到：")
			msgContentBuilder.Write(getter.GetStringBytes("room_Info", "lock_till"))
			var msg = messageSender.GeneralPushMessage{
				Title:   msgTitleBuilder.String(),
				Content: msgContentBuilder.String(),
				IconURL: string(getter.GetStringBytes("user_info", "face")),
			}
			msg.SendToAllTargets()
		}
		break

	//	别的不关心，所以没写
	default:
		var logBuilder strings.Builder
		logBuilder.WriteString(webhookId)
		logBuilder.WriteString(" DDTV 未知的webhook请求类型：")
		logBuilder.Write(getter.GetStringBytes("type"))
		log.Warn(logBuilder.String())
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

// DDTV3WebhookHandler 处理 DDTV 的 webhook 请求
func DDTV3WebhookHandler(c *gin.Context) {
	// 读取请求内容
	content, err := c.GetRawData()
	if err != nil {
		log.Errorf("读取 DDTV3 webhook 请求失败：%s", err.Error())
		c.Status(http.StatusBadRequest)
		return
	}
	// return 200 at first
	c.Status(http.StatusOK)
	go ddtv3TaskRunner(c.FullPath(), content)
}

var ddtv3Settings = make(map[string]map[int]Event)
var ddtv3IdEventTitleMap = map[int]string{
	0:  "StartLive",
	1:  "StopLive",
	2:  "StartRec",
	3:  "RecComplete",
	4:  "CancelRec",
	5:  "TranscodingComplete",
	6:  "SaveDanmuComplete",
	7:  "SaveSCComplete",
	8:  "SaveGiftComplete",
	9:  "SaveGuardComplete",
	10: "RunShellComplete",
	11: "DownloadEndMissionSuccess",
	12: "SpaceIsInsufficientWarn",
	13: "LoginFailure",
	14: "LoginWillExpireSoon",
	15: "UpdateAvailable",
	16: "ShellExecutionComplete",
	17: "WarnedByAdmin",
	18: "LiveCutOff",
	19: "RoomLocked",
}
var ddtv3IdEventNameMap = map[int]string{
	0:  "主播开播",
	1:  "主播下播",
	2:  "开始录制",
	3:  "录制结束",
	4:  "录制被取消",
	5:  "完成转码",
	6:  "保存弹幕文件完成",
	7:  "保存SC文件完成",
	8:  "保存礼物文件完成",
	9:  "保存大航海文件完成",
	10: "执行Shell命令完成",
	11: "下载任务成功结束",
	12: "剩余空间不足",
	13: "登录失效",
	14: "登录即将失效",
	15: "有可用新版本",
	16: "执行Shell命令结束",
	17: "被管理员警告",
	18: "直播被管理员切断",
	19: "直播间被封禁",
}

func UpdateDDTV3Settings(path string, events map[string]Event) {
	if ddtv3Settings[path] == nil {
		ddtv3Settings[path] = make(map[int]Event)
	}
	for id, name := range ddtv3IdEventTitleMap {
		event, ok := events[name]
		if ok {
			t := ddtv3Settings[path]
			t[id] = event
			ddtv3Settings[path] = t
		}
	}
}
