package main

import (
	log "github.com/sirupsen/logrus"
	"net/http"
	"webhookTemplate/secrets"
	"webhookTemplate/terminal"
	"webhookTemplate/webhookHandler"
)

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
	secrets.WeworkAccessTokenExpiresIn = 0
}

func main() {
	log.Infof("启动，监听：127.0.0.1:14000/ddtv")
	log.Infof("启动，监听：127.0.0.1:14000/bililiverecoder")
	//log.Infof("启动，监听：127.0.0.1:14000/")
	// 当有请求访问时，执行此回调方法
	http.HandleFunc("/ddtv", webhookHandler.DDTVWebhookHandler)
	http.HandleFunc("/bililiverecoder", webhookHandler.BililiveRecoderWebhookHandler)
	//http.HandleFunc("/", handler)
	// 监听127.0.0.1:14000
	err := http.ListenAndServe("127.0.0.1:14000", nil)
	if err != nil {
		log.Fatalf("监听端口异常，%v", err)
	}
}
