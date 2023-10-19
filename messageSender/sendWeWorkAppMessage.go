package messageSender

import (
	"bytes"
	"errors"
	log "github.com/sirupsen/logrus"
	"github.com/valyala/fastjson"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
	"webhookGo/secrets"
)

func updateAccessToken(app *secrets.WeworkApp) error {
	// https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=ID&corpsecret=SECRET
	log.Info("更新企业微信应用的access_token")
	// 构造请求地址
	var urlBuilder strings.Builder
	urlBuilder.Grow(363)
	urlBuilder.WriteString("https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=")
	urlBuilder.WriteString(app.CorpId)
	urlBuilder.WriteString("&corpsecret=")
	urlBuilder.WriteString(app.AppSecret)
	log.Tracef("更新企业微信应用的access_token：请求地址：%s", urlBuilder.String())
	// 发送请求
	resp, err := http.Get(urlBuilder.String())
	if err != nil {
		log.Errorf("更新企业微信应用的access_token：请求发送失败：%s", err.Error())
		return err
	}
	defer func(Body io.ReadCloser) {
		errCloser := Body.Close()
		if errCloser != nil {
			log.Errorf("更新企业微信应用的access_token：关闭连接失败：%s", err.Error())
		}
	}(resp.Body)
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("更新企业微信应用的access_token：读取响应消息失败：%s", err.Error())
		return err
	}
	log.Tracef("更新企业微信应用的access_token：响应消息：%s", content)

	var p fastjson.Parser
	getter, errOfJsonParser := p.ParseBytes(content)
	if errOfJsonParser != nil {
		return errOfJsonParser
	}
	errcode := getter.GetInt("errcode")
	if 0 != errcode {
		log.Error("更新企业微信应用的access_token失败：", getter.GetStringBytes("errmsg"))
		return errors.New(string(getter.GetStringBytes("errmsg")))
	}
	app.WeworkAccessToken = string(getter.GetStringBytes("access_token"))
	app.WeworkAccessTokenExpireAt = time.Now().Unix() + getter.GetInt64("expires_in")
	log.Trace("企业微信AccessToken：", app.WeworkAccessToken)
	log.Debug("有效期至：", app.WeworkAccessTokenExpireAt)
	return nil
}

func SendWeWorkAppMessage(message Message) {
	length := len(secrets.Secrets.WeworkApps)
	if length > 0 {
		wg := sync.WaitGroup{}
		for i := 0; i < length; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				sendWeWorkAppMessage(&secrets.Secrets.WeworkApps[i], message)
			}(i)
		}
		wg.Wait()
	}
}

func sendWeWorkAppMessage(app *secrets.WeworkApp, message Message) {
	if app.CorpId == "" || app.AppSecret == "" || app.AgentID == "" {
		return
	}
	// 检查token是否过期
	if time.Now().Unix() > app.WeworkAccessTokenExpireAt {
		// 更新之前需要加锁，防止有线程正在更新
		app.Lock()
		defer app.Unlock()
		// 再次判断过期时间，防止被其他线程更新过了
		if time.Now().Unix() > app.WeworkAccessTokenExpireAt {
			err := updateAccessToken(app)
			if err != nil {
				return
			}
		}
	}
	// 制作要发送的 Markdown 消息
	var bodyBuffer bytes.Buffer
	bodyBuffer.WriteString("{\"touser\":\"@all\",\"msgtype\":\"markdown\",\"agentid\":\"")
	bodyBuffer.WriteString(app.AgentID)
	bodyBuffer.WriteString("\",\"markdown\":{\"content\":\"# ")
	bodyBuffer.WriteString(message.Title)
	bodyBuffer.WriteString("\n")
	bodyBuffer.WriteString(message.Content)
	bodyBuffer.WriteString("\"},\"enable_duplicate_check\":1,\"duplicate_check_interval\":3600}")

	// target: 发送目标，企业微信API https://qyapi.weixin.qq.com/cgi-bin/message/send?access_token=
	var targetBuilder strings.Builder
	targetBuilder.Grow(318)
	targetBuilder.WriteString("https://qyapi.weixin.qq.com/cgi-bin/message/send?access_token=")
	targetBuilder.WriteString(app.WeworkAccessToken)

	// 发送请求
	log.Trace(message.ID, "发送企业微信应用消息 请求地址", targetBuilder.String())
	resp, err := http.Post(targetBuilder.String(), "application/json", &bodyBuffer)
	defer func(Body io.ReadCloser) {
		errCloser := Body.Close()
		if errCloser != nil {
			log.Error(message.ID, "发送企业微信应用消息 关闭连接失败", errCloser.Error())
		}
	}(resp.Body)
	if err != nil {
		log.Error(message.ID, "发送企业微信应用消息 请求失败", err.Error())
		return
	}
	// 读取响应消息
	content, errReader := io.ReadAll(resp.Body)
	if errReader != nil {
		log.Error(message.ID, "发送企业微信应用消息 读取响应内容失败", errReader.Error())
		return
	}
	var p fastjson.Parser
	getter, errOfJsonParser := p.ParseBytes(content)
	if errOfJsonParser != nil {
		return
	}
	errcode := getter.GetInt("errcode")
	if errcode != 0 {
		log.Error(message.ID, "发送企业微信应用消息 服务器返回错误", content)
		return
	}
	if log.IsLevelEnabled(log.TraceLevel) {
		log.Trace(message.ID, "发送企业微信应用消息成功 响应消息", content)
	} else {
		log.Debug(message.ID, "发送企业微信应用消息成功 消息id", getter.GetStringBytes("msgid"))
	}
	return
}
