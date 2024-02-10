package main

import (
	log "github.com/sirupsen/logrus"
	"net/http"
	"webhookGo/terminal"
	"webhookGo/webhookHandler"
)

func init() {
	// 防止因为选择导致的进程挂起
	_ = terminal.DisableQuickEdit()
	// 设置控制台显示
	log.SetFormatter(&log.TextFormatter{
		ForceColors:   true,
		FullTimestamp: true,
	})
}

func main() {
	config := loadConfig()
	if config.bililiveRecoder.enable {
		log.Info("B站录播姬已启用，监听 http://" + config.listenAddress + config.bililiveRecoder.path)
		http.HandleFunc(config.bililiveRecoder.path, webhookHandler.BililiveRecoderWebhookHandler)
	}
	if config.blrec.enable {
		log.Info("blrec已启用，监听 http://" + config.listenAddress + config.blrec.path)
		http.HandleFunc(config.blrec.path, webhookHandler.BlrecWebhookHandler)
	}
	if config.ddtv.enable {
		log.Info("DDTV已启用，监听 http://" + config.listenAddress + config.ddtv.path)
		http.HandleFunc(config.ddtv.path, webhookHandler.DDTVWebhookHandler)
	}
	err := http.ListenAndServe(config.listenAddress, nil)
	if err != nil {
		log.Fatalf("监听端口异常，%+v", err)
	}
}
