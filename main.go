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

func main() {
	config := loadConfig()
	r := gin.Default()
	if config.bililiveRecoder.enable {
		log.Info("B站录播姬已启用，监听 http://" + config.listenAddress + config.bililiveRecoder.path)
		r.POST(config.bililiveRecoder.path, webhookHandler.BililiveRecoderWebhookHandler)
	}
	if config.blrec.enable {
		log.Info("blrec已启用，监听 http://" + config.listenAddress + config.blrec.path)
		r.POST(config.blrec.path, webhookHandler.BlrecWebhookHandler)
	}
	if config.ddtv3.enable {
		log.Info("DDTV已启用，监听 http://" + config.listenAddress + config.ddtv3.path)
		r.POST(config.ddtv3.path, webhookHandler.DDTV3WebhookHandler)
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
