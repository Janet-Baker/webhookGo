package messageSender

import (
	"bytes"
	"errors"
	jsoniter "github.com/json-iterator/go"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
	"webhookTemplate/secrets"
)

var updateAccessTokenLock sync.Mutex

func updateAccessToken() error {
	// https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=ID&corpsecret=SECRET
	log.Info("更新企业微信应用的access_token")
	// 构造请求地址
	var urlBuilder strings.Builder
	urlBuilder.Grow(363)
	urlBuilder.WriteString("https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=")
	urlBuilder.WriteString(secrets.WeworkCorpId)
	urlBuilder.WriteString("&corpsecret=")
	urlBuilder.WriteString(secrets.AppSecret)
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
	errcode := jsoniter.Get(content, "errcode").ToString()
	if "0" != errcode {
		log.Errorf("更新企业微信应用的access_token失败：%s", jsoniter.Get(content, "errmsg").ToString())
		return errors.New(jsoniter.Get(content, "errmsg").ToString())
	}
	secrets.WeworkAccessToken = jsoniter.Get(content, "access_token").ToString()
	secrets.WeworkAccessTokenExpireAt = time.Now().Unix() + jsoniter.Get(content, "expires_in").ToInt64()
	log.Tracef("企业微信AccessToken：%s", secrets.WeworkAccessToken)
	log.Debugf("有效期至：%v", secrets.WeworkAccessTokenExpireAt)
	return nil
}

func SendWeWorkMessage(message Message) {
	// 检查token是否过期
	if time.Now().Unix() > secrets.WeworkAccessTokenExpireAt {
		// 更新之前需要加锁，防止有线程正在更新
		updateAccessTokenLock.Lock()
		defer updateAccessTokenLock.Unlock()
		// 再次判断过期时间，防止被其他线程更新过了
		if time.Now().Unix() > secrets.WeworkAccessTokenExpireAt {
			err := updateAccessToken()
			if err != nil {
				return
			}
		}
	}
	// 制作要发送的 Markdown 消息
	var bodyBuffer bytes.Buffer
	bodyBuffer.WriteString("{\"touser\":\"@all\",\"msgtype\":\"markdown\",\"agentid\":\"")
	bodyBuffer.WriteString(secrets.AgentID)
	bodyBuffer.WriteString("\",\"markdown\":{\"content\":\"# ")
	bodyBuffer.WriteString(message.Title)
	bodyBuffer.WriteString("\n")
	bodyBuffer.WriteString(message.Content)
	bodyBuffer.WriteString("\"},\"enable_duplicate_check\":1,\"duplicate_check_interval\":3600}")

	// target: 发送目标，企业微信API https://qyapi.weixin.qq.com/cgi-bin/message/send?access_token=
	var targetBuilder strings.Builder
	targetBuilder.Grow(318)
	targetBuilder.WriteString("https://qyapi.weixin.qq.com/cgi-bin/message/send?access_token=")
	targetBuilder.WriteString(secrets.WeworkAccessToken)

	// 发送请求
	log.Tracef("%s 发送企业微信应用消息：请求地址：%s", message.ID, targetBuilder.String())
	resp, err := http.Post(targetBuilder.String(), "application/json", &bodyBuffer)
	defer func(Body io.ReadCloser) {
		errCloser := Body.Close()
		if errCloser != nil {
			log.Errorf("%s 发送企业微信应用消息:关闭连接失败：%s", message.ID, errCloser.Error())
		}
	}(resp.Body)
	if err != nil {
		log.Errorf("%s 发送企业微信应用消息：请求失败：%s", message.ID, err.Error())
		return
	}
	// 读取响应消息
	content, errReader := io.ReadAll(resp.Body)
	if errReader != nil {
		log.Errorf("%s 发送企业微信应用消息：读取响应内容失败：%s", message.ID, errReader.Error())
		return
	}
	errcode := jsoniter.Get(content, "errcode").ToString()
	if errcode != "0" {
		log.Errorf("%s 发送企业微信应用消息：服务器返回错误：%s", message.ID, content)
		return
	}
	if log.IsLevelEnabled(log.TraceLevel) {
		log.Tracef("%s 发送企业微信应用消息成功：响应消息：%s", message.ID, content)
	} else {
		log.Debugf("%s 发送企业微信应用消息成功：消息id：%s", message.ID, jsoniter.Get(content, "msgid").ToString())
	}
	return
}
