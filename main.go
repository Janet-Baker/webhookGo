package main

import (
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
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

var messageHandlers = map[string]func(*gin.Context){
	"BililiveRecorder": webhookHandler.BililiveRecorderWebhookHandler,
	"Blrec":            webhookHandler.BlrecWebhookHandler,
	"DDTV3":            webhookHandler.DDTV3WebhookHandler,
	"DDTV5":            webhookHandler.DDTV5WebhookHandler,
}

func main() {
	config := loadConfig()
	r := gin.Default()
	for _, receiver := range config.receivers {
		if function, ok := messageHandlers[receiver.Type]; ok {
			log.Info(receiver.Type + "已启用，监听 http://" + config.listenAddress + receiver.Path)
			r.POST(receiver.Path, function)
		}
	}

	r.NoRoute(func(c *gin.Context) {
		log.Warnln("Unknown access to", c.Request.Method, `"`+c.Request.URL.Path+`"`,
			"\nfrom", c.RemoteIP(), "User-Agent:", c.GetHeader("User-Agent"))
		c.Status(403)
	})

	err := r.Run(config.listenAddress)
	if err != nil {
		log.Fatal("监听端口异常，", err)
	}
}
