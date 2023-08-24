package messageSender

import (
	"bytes"
	"errors"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"time"
	"webhookTemplate/secrets"
)

func updateAccessToken() error {
	// https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=ID&corpsecret=SECRET
	log.Debugf("更新企业微信应用的access_token")
	resp, err := http.Get("https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=" + secrets.WeworkCorpId + "&corpsecret=" + secrets.AppSecret + "&debug=1")
	if err != nil {
		log.Errorf("更新企业微信应用的access_token失败：%s", err.Error())
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Errorf("关闭消息发送响应失败：%s", err.Error())
		}
	}(resp.Body)
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("更新企业微信应用的access_token失败：读取响应消息失败：%s", err.Error())
		return err
	}
	errcode := jsoniter.Get(content, "errcode").ToString()
	if "0" != errcode {
		log.Errorf("更新企业微信应用的access_token失败：%s", jsoniter.Get(content, "errmsg").ToString())
		return errors.New(jsoniter.Get(content, "errmsg").ToString())
	}
	secrets.WeworkAccessToken = jsoniter.Get(content, "access_token").ToString()
	secrets.WeworkAccessTokenExpiresIn = time.Now().Unix() + jsoniter.Get(content, "expires_in").ToInt64()
	log.Debugf("企业微信AccessToken：%s", secrets.WeworkAccessToken)
	log.Debugf("有效期至：%v", secrets.WeworkAccessTokenExpiresIn)
	return nil
}

func SendWeWorkMessage(message Message) error {
	log.Debugf("发送企业微信应用消息：%s", message)
	// 检查token是否过期
	if time.Now().Unix() > secrets.WeworkAccessTokenExpiresIn {
		err := updateAccessToken()
		if err != nil {
			return err
		}
	}
	// 制作要发送的 Markdown 消息
	var body = fmt.Sprintf("{\"touser\":\"@all\","+
		"\"msgtype\":\"markdown\","+
		"\"agentid\":\"%s\","+
		"\"markdown\":{\"content\":\"# %s\n\n%s\"},"+
		"\"enable_duplicate_check\":1,"+
		"\"duplicate_check_interval\":600}", secrets.AgentID, message.Title, message.Content)
	// target: 发送目标，企业微信API https://qyapi.weixin.qq.com/cgi-bin/message/send?access_token=
	target := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/message/send?access_token=%s&debug=1", secrets.WeworkAccessToken)
	resp, err := http.Post(target, "application/json", bytes.NewReader([]byte(body)))
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Errorf("关闭消息发送响应失败：%s", err.Error())
		}
	}(resp.Body)
	// 读取响应消息
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("发送企业微信应用消息失败：%s", err.Error())
		return err
	}
	errcode := jsoniter.Get(content, "errcode").ToString()
	if errcode != "0" {
		log.Errorf("发送企业微信应用消息失败：%s", jsoniter.Get(content, "errmsg").ToString())
		return err
	}
	if err != nil {
		return err
	}
	return nil
}