package main

import (
	"fmt"
	"github.com/json-iterator/go"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"webhookTemplate/messageSender"
	"webhookTemplate/terminal"
)

func webhookHandler(w http.ResponseWriter, request *http.Request) {
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
	}(request.Body)

	content, err := ioutil.ReadAll(request.Body)
	if err != nil {
		return
	}

	typeName := jsoniter.Get(content, "type_name").ToString()
	switch typeName {
	// 主播下播
	case "StopLive":
		if jsoniter.Get(content, "room_Info", "IsLocked").ToBool() {
			// 主播被封号了
			var msg = messageSender.Message{
				Title: fmt.Sprintf("%s 被封号啦！快去围观吧", jsoniter.Get(content, "room_Info", "uname").ToString()),
				Content: fmt.Sprintf("- 主播：%s \n\n- 标题：%s \n\n- 分区：%s - %s \n\n- 封禁时间：%s\n\n- 封禁到：%s",
					jsoniter.Get(content, "room_Info", "uname").ToString(),
					jsoniter.Get(content, "room_Info", "title").ToString(),
					jsoniter.Get(content, "room_info", "area_v2_parent_name").ToString(),
					jsoniter.Get(content, "room_info", "area_v2_name").ToString(),
					jsoniter.Get(content, "hook_time").ToString(),
					jsoniter.Get(content, "room_Info", "lock_till").ToString()),
			}
			err := msg.Send()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
		} else {
			// 主播正常下播
			var msg = messageSender.Message{
				Title: fmt.Sprintf("%s 下播了", jsoniter.Get(content, "room_Info", "uname").ToString()),
				Content: fmt.Sprintf("- 主播：%s\n\n- 标题：%s\n\n- 分区：%s - %s\n\n- 下播时间：%s",
					jsoniter.Get(content, "room_Info", "uname").ToString(),
					jsoniter.Get(content, "room_Info", "title").ToString(),
					jsoniter.Get(content, "room_info", "area_v2_parent_name").ToString(),
					jsoniter.Get(content, "room_info", "area_v2_name").ToString(),
					jsoniter.Get(content, "hook_time").ToString()),
			}
			err := msg.Send()
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
		}
		w.WriteHeader(http.StatusOK)
		break
	case "StartLive":
		// 主播开播
		var msg = messageSender.Message{
			Title: fmt.Sprintf("%s 开播了", jsoniter.Get(content, "room_Info", "uname").ToString()),
			Content: fmt.Sprintf("- 主播：%s\n\n- 标题：%s\n\n- 分区：%s - %s\n\n- 开播时间：%s",
				jsoniter.Get(content, "room_Info", "uname").ToString(),
				jsoniter.Get(content, "room_Info", "title").ToString(),
				jsoniter.Get(content, "room_info", "area_v2_parent_name").ToString(),
				jsoniter.Get(content, "room_info", "area_v2_name").ToString(),
				jsoniter.Get(content, "hook_time").ToString()),
		}
		err := msg.Send()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.WriteHeader(http.StatusOK)
		break
	case "StartRec":
		// 开始录制
		log.Infof("开始录制：%s", jsoniter.Get(content, "room_Info", "uname").ToString())
		w.WriteHeader(http.StatusOK)
		break
	case "RecComplete":
		// 录制结束
		log.Infof("录制结束：%s", jsoniter.Get(content, "room_Info", "uname").ToString())
		w.WriteHeader(http.StatusOK)
		break
	case "CancelRec":
		// 录制被取消
		log.Infof("录制被取消：%s", jsoniter.Get(content, "room_Info", "uname").ToString())
		w.WriteHeader(http.StatusOK)
		break
	case "TranscodingComplete":
		// 完成转码
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
		w.WriteHeader(http.StatusOK)
		break
	case "SaveDanmuComplete":
		// 保存弹幕文件完成
		log.Infof("保存弹幕文件完成：%s", jsoniter.Get(content, "room_Info", "uname").ToString())
		w.WriteHeader(http.StatusOK)
		break
	case "SaveSCComplete":
		//	保存SC文件完成
		log.Infof("保存SC文件完成：%s", jsoniter.Get(content, "room_Info", "uname").ToString())
		w.WriteHeader(http.StatusOK)
		break
	case "SaveGiftComplete":
		//	保存礼物文件完成
		log.Infof("保存礼物文件完成：%s", jsoniter.Get(content, "room_Info", "uname").ToString())
		w.WriteHeader(http.StatusOK)
		break
	case "SaveGuardComplete":
		//	保存大航海文件完成
		log.Infof("保存大航海文件完成：%s", jsoniter.Get(content, "room_Info", "uname").ToString())
		w.WriteHeader(http.StatusOK)
		break
	case "RunShellComplete":
		//	执行Shell命令完成
		log.Infof("执行Shell命令完成：%s", jsoniter.Get(content, "room_Info", "uname").ToString())
		w.WriteHeader(http.StatusOK)
		break
	case "DownloadEndMissionSuccess":
		//	下载任务成功结束
		log.Infof("下载任务成功结束：%s", jsoniter.Get(content, "room_Info", "uname").ToString())
		w.WriteHeader(http.StatusOK)
		break
	case "SpaceIsInsufficientWarn":
		//	剩余空间不足
		log.Infof("剩余空间不足：%s", content)
		w.WriteHeader(http.StatusOK)
		break
	case "LoginFailure":
		//	登陆失效
		log.Errorf("登陆失效")
		w.WriteHeader(http.StatusOK)
		break
	case "LoginWillExpireSoon":
		//	登陆即将失效
		log.Warnf("登陆即将失效")
		w.WriteHeader(http.StatusOK)
		break
	case "UpdateAvailable":
		//	有可用新版本
		log.Infof("有可用新版本：%s", jsoniter.Get(content, "version").ToString())
		w.WriteHeader(http.StatusOK)
		break
	default:
		// 别的不关心，所以没写
		w.WriteHeader(http.StatusOK)
	}
}

func main() {
	_ = terminal.DisableQuickEdit()
	log.SetFormatter(&log.TextFormatter{
		ForceColors:   true,
		FullTimestamp: true,
	})
	log.Infof("启动，监听：127.0.0.1:14000/webhook")
	//log.Infof("启动，监听：127.0.0.1:14000/")
	// 当有请求访问ws时，执行此回调方法
	handler := http.HandlerFunc(webhookHandler)
	http.HandleFunc("/webhook", handler)
	//http.HandleFunc("/", handler)
	// 监听127.0.0.1:14000
	err := http.ListenAndServe("127.0.0.1:14000", nil)
	if err != nil {
		log.Fatalf("监听端口异常，%v", err)
	}
}
