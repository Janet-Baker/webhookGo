package main

import (
	log "github.com/sirupsen/logrus"
	"net/http"
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
	/*log.SetLevel(log.DebugLevel)
	log.Warnf("已开启Debug模式.")*/
}

func main() {
	log.Infof("启动，监听：http://127.0.0.1:14000/ddtv")
	log.Infof("启动，监听：http://127.0.0.1:14000/bililiverecoder")
	log.Infof("启动，监听：http://127.0.0.1:14000/blrec")
	//log.Infof("启动，监听：http://127.0.0.1:14000/")
	http.HandleFunc("/ddtv", webhookHandler.DDTVWebhookHandler)
	http.HandleFunc("/bililiverecoder", webhookHandler.BililiveRecoderWebhookHandler)
	http.HandleFunc("/blrec", webhookHandler.BlrecWebhookHandler)
	//http.HandleFunc("/", handler)
	// 监听127.0.0.1:14000
	err := http.ListenAndServe("127.0.0.1:14000", nil)
	if err != nil {
		log.Fatalf("监听端口异常，%v", err)
	}
}
