package main

import (
	"fmt"
	"github.com/json-iterator/go"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"webhookTemplate/messageSender"
)

func webhookHandler(w http.ResponseWriter, request *http.Request) {

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		} else {
			w.WriteHeader(http.StatusOK)
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
				Content: fmt.Sprintf("主播：%s \n标题：%s \n分区：%s - %s \n封禁时间：%s\n封禁到：%s",
					jsoniter.Get(content, "room_Info", "uname").ToString(),
					jsoniter.Get(content, "room_Info", "title").ToString(),
					jsoniter.Get(content, "room_info", "area_v2_parent_name").ToString(),
					jsoniter.Get(content, "room_info", "area_v2_name").ToString(),
					jsoniter.Get(content, "hook_time").ToString(),
					jsoniter.Get(content, "room_Info", "lock_till").ToString()),
			}
			msg.Send()
		} else {
			// 主播正常下播
			var msg = messageSender.Message{
				Title: fmt.Sprintf("%s 下播了", jsoniter.Get(content, "room_Info", "uname").ToString()),
				Content: fmt.Sprintf("主播：%s\n标题：%s\n分区：%s - %s\n下播时间：%s",
					jsoniter.Get(content, "room_Info", "uname").ToString(),
					jsoniter.Get(content, "room_Info", "title").ToString(),
					jsoniter.Get(content, "room_info", "area_v2_parent_name").ToString(),
					jsoniter.Get(content, "room_info", "area_v2_name").ToString(),
					jsoniter.Get(content, "hook_time").ToString()),
			}
			msg.Send()
		}
		break
	case "StartLive":
		// 主播开播
		var msg = messageSender.Message{
			Title: fmt.Sprintf("%s 开播了", jsoniter.Get(content, "room_Info", "uname").ToString()),
			Content: fmt.Sprintf("主播：%s \n标题：%s \n分区：%s - %s\n开播时间：%s",
				jsoniter.Get(content, "room_Info", "uname").ToString(),
				jsoniter.Get(content, "room_Info", "title").ToString(),
				jsoniter.Get(content, "room_info", "area_v2_parent_name").ToString(),
				jsoniter.Get(content, "room_info", "area_v2_name").ToString(),
				jsoniter.Get(content, "hook_time").ToString()),
		}
		msg.Send()
		break
	default:
		// 未知类型
	}

}

func main() {
	log.Infof("Go!")
	// 当有请求访问ws时，执行此回调方法
	handler := http.HandlerFunc(webhookHandler)
	http.HandleFunc("/webhook", handler)
	// 监听127.0.0.1:14000
	err := http.ListenAndServe("127.0.0.1:14000", nil)
	if err != nil {
		log.Fatalf("监听端口异常，%v", err)
	} else {
		log.Infof("监听端口成功，端口：14000")
	}
}
