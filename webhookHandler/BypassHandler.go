package webhookHandler

import (
	"bytes"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"io"
	"webhookGo/messageSender"
)

func BypassHandler(c *gin.Context) {
	content, err := c.GetRawData()
	if log.IsLevelEnabled(log.DebugLevel) {
		log.Debug(string(content), err)
		c.Request.Body = io.NopCloser(bytes.NewReader(content))
	}
	if err != nil {
		log.Error(err)
		return
	}
	var msg = messageSender.GeneralPushMessage{
		Title:   "分流抢票通知",
		Content: string(content),
	}
	msg.SendToAllTargets()
}
