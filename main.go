package main

import (
	"fmt"
	"github.com/json-iterator/go"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"time"
	"webhookTemplate/CustomizedQueue"
	"webhookTemplate/messageSender"
	"webhookTemplate/secrets"
	"webhookTemplate/terminal"
)

var webhookMessageIdList = CustomizedQueue.NewQueue(15)

func webhookHandler(w http.ResponseWriter, request *http.Request) {
	// defer request.Body.Close()
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
	}(request.Body)
	// 读取请求内容
	content, err := ioutil.ReadAll(request.Body)
	if err != nil {
		return
	}
	log.Infof("收到webhook请求")
	log.Debugf(string(content))

	// 判断是否是重复的webhook请求
	webhookId := jsoniter.Get(content, "id").ToString()
	if webhookMessageIdList.IsContain(webhookId) {
		log.Warnf("重复的webhook请求：%s", webhookId)
		w.WriteHeader(http.StatusOK)
		return
	} else {
		webhookMessageIdList.EnQueue(webhookId)
	}

	// 判断事件类型
	hookType := jsoniter.Get(content, "type").ToInt()
	switch hookType {
	//	0 StartLive 主播开播
	case 0:
		var msg = messageSender.Message{
			Title: fmt.Sprintf("%s 开播了", jsoniter.Get(content, "room_Info", "uname").ToString()),
			Content: fmt.Sprintf("- 主播：%s\n\n- 标题：%s\n\n- 分区：%s - %s\n\n- 开播时间：%s",
				jsoniter.Get(content, "room_Info", "uname").ToString(),
				jsoniter.Get(content, "room_Info", "title").ToString(),
				jsoniter.Get(content, "room_Info", "area_v2_parent_name").ToString(),
				jsoniter.Get(content, "room_Info", "area_v2_name").ToString(),
				jsoniter.Get(content, "hook_time").ToString()),
		}
		err := msg.Send()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.WriteHeader(http.StatusOK)
		break

	//	1 StopLive 主播下播
	case 1:
		// 主播正常下播
		log.Infof("主播下播：%s", jsoniter.Get(content, "room_Info", "uname").ToString())
		w.WriteHeader(http.StatusOK)
		break

	//	2 StartRec 开始录制
	case 2:
		log.Infof("开始录制：%s", jsoniter.Get(content, "room_Info", "uname").ToString())
		w.WriteHeader(http.StatusOK)
		break

	//	3 RecComplete 录制结束
	case 3:
		log.Infof("录制结束：%s", jsoniter.Get(content, "room_Info", "uname").ToString())
		w.WriteHeader(http.StatusOK)
		break

	//	4 CancelRec 录制被取消
	case 4:
		log.Infof("录制被取消：%s", jsoniter.Get(content, "room_Info", "uname").ToString())
		w.WriteHeader(http.StatusOK)
		break

	//	5 TranscodingComplete 完成转码
	case 5:
		if jsoniter.Get(content, "room_Info", "is_locked").ToBool() {
			// 主播被封号了
			var msg = messageSender.Message{
				Title: fmt.Sprintf("%s 喜提直播间封禁！转码完成", jsoniter.Get(content, "room_Info", "uname").ToString()),
				Content: fmt.Sprintf("- 主播：%s\n\n- 标题：%s\n\n- 分区：%s - %s\n\n- 封禁时间：%s\n\n- 封禁到：%s",
					jsoniter.Get(content, "room_Info", "uname").ToString(),
					jsoniter.Get(content, "room_Info", "title").ToString(),
					jsoniter.Get(content, "room_Info", "area_v2_parent_name").ToString(),
					jsoniter.Get(content, "room_Info", "area_v2_name").ToString(),
					jsoniter.Get(content, "hook_time").ToString(),
					jsoniter.Get(content, "room_Info", "lock_till").ToString()),
			}
			err := msg.Send()
			if err != nil {
				return
			}
		} else {
			var msg = messageSender.Message{
				Title: fmt.Sprintf("%s 转码完成", jsoniter.Get(content, "room_Info", "uname").ToString()),
				Content: fmt.Sprintf("主播：%s\n标题：%s\n转码完成时间：%s",
					jsoniter.Get(content, "room_Info", "uname").ToString(),
					jsoniter.Get(content, "room_Info", "title").ToString(),
					jsoniter.Get(content, "hook_time").ToString()),
			}
			err := msg.Send()
			if err != nil {
				return
			}
		}
		w.WriteHeader(http.StatusOK)
		break

	//	6 SaveDanmuComplete 保存弹幕文件完成
	case 6:
		log.Infof("保存弹幕文件完成：%s", jsoniter.Get(content, "room_Info", "uname").ToString())
		w.WriteHeader(http.StatusOK)
		break

	//	7 SaveSCComplete 保存SC文件完成
	case 7:
		log.Infof("保存SC文件完成：%s", jsoniter.Get(content, "room_Info", "uname").ToString())
		w.WriteHeader(http.StatusOK)
		break

	//	8 SaveGiftComplete 保存礼物文件完成
	case 8:
		log.Infof("保存礼物文件完成：%s", jsoniter.Get(content, "room_Info", "uname").ToString())
		w.WriteHeader(http.StatusOK)
		break

	//	9 SaveGuardComplete 保存大航海文件完成
	case 9:
		log.Infof("保存大航海文件完成：%s", jsoniter.Get(content, "room_Info", "uname").ToString())
		w.WriteHeader(http.StatusOK)
		break

	//	10 RunShellComplete 执行Shell命令完成
	case 10:
		log.Infof("执行Shell命令完成：%s", jsoniter.Get(content, "room_Info", "uname").ToString())
		w.WriteHeader(http.StatusOK)
		break

	//	11 DownloadEndMissionSuccess 下载任务成功结束
	case 11:
		log.Infof("下载任务成功结束：%s", jsoniter.Get(content, "room_Info", "uname").ToString())
		w.WriteHeader(http.StatusOK)
		break

	//	12 SpaceIsInsufficientWarn 剩余空间不足
	case 12:
		log.Infof("剩余空间不足：%s", content)
		w.WriteHeader(http.StatusOK)
		break

	//	13 LoginFailure 登陆失效
	case 13:
		log.Errorf("登陆失效")
		w.WriteHeader(http.StatusOK)
		break

	//	14 LoginWillExpireSoon 登陆即将失效
	case 14:
		log.Warnf("登陆即将失效")
		w.WriteHeader(http.StatusOK)
		break

	//	15 UpdateAvailable 有可用新版本
	case 15:
		log.Infof("有可用新版本：%s", jsoniter.Get(content, "version").ToString())
		w.WriteHeader(http.StatusOK)
		break

	//	16 ShellExecutionComplete 执行Shell命令结束
	case 16:
		log.Infof("执行Shell命令结束：%+v", content)
		w.WriteHeader(http.StatusOK)
		break

	//	别的不关心，所以没写
	default:
		log.Warnf("未知的webhook请求：%+v", content)
		w.WriteHeader(http.StatusOK)
	}
}

func init() {
	// 防止因为选择导致的进程挂起
	_ = terminal.DisableQuickEdit()
	// 设置日志
	log.SetFormatter(&log.TextFormatter{
		ForceColors:   true,
		FullTimestamp: true,
	})
	// 设置Debug模式
	//log.SetLevel(log.DebugLevel)
	//log.Warnf("已开启Debug模式.")
	// 手动初始化包变量，使包变量有访问者，防止被GC清理
	secrets.WeworkAccessToken = "0"
	secrets.WeworkAccessTokenExpiresIn = time.Now().Unix()
}

func main() {
	log.Infof("启动，监听：127.0.0.1:14000/webhook")
	//log.Infof("启动，监听：127.0.0.1:14000/")
	// 当有请求访问时，执行此回调方法
	handler := http.HandlerFunc(webhookHandler)
	http.HandleFunc("/webhook", handler)
	//http.HandleFunc("/", handler)
	// 监听127.0.0.1:14000
	err := http.ListenAndServe("127.0.0.1:14000", nil)
	if err != nil {
		log.Fatalf("监听端口异常，%v", err)
	}
}
