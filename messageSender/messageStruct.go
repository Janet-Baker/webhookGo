package messageSender

import (
	"bytes"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
	"webhookTemplate/secrets"
)

type Message struct {
	// 消息标题
	Title string
	// 消息内容
	Content string
}

func SendBarkMessage(message Message) error {
	log.Debugf("发送 Bark 消息：%+v", message)
	resp, err := http.Get("https://api.day.app/" + secrets.BarkSecrets + "/" + url.QueryEscape(message.Title) + "/" + url.QueryEscape(message.Content))
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Errorf("关闭消息发送响应失败：%s", err.Error())
		}
	}(resp.Body)
	if err != nil {
		log.Errorf("发送消息失败：%s", err.Error())
		return err
	} else {
		log.Debugf("发送Bark消息成功：%+v", message)
	}
	return nil
}

func UpdateAccessToken() error {
	// https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=ID&corpsecret=SECRET
	log.Debugf("更新企业微信应用的access_token")
	resp, err := http.Get("https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=" + secrets.WeworkCorpId + "&corpsecret=" + secrets.AppSecret)
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
	errcode := jsoniter.Get(content, "errcode").ToInt()
	if errcode == 0 {
		secrets.WeworkAccessToken = jsoniter.Get(content, "access_token").ToString()
		secrets.WeworkAccessTokenExpiresIn = time.Now().Unix() + jsoniter.Get(content, "expires_in").ToInt64()
	}
	log.Debugf("企业微信AccessToken：%s", secrets.WeworkAccessToken)
	log.Debugf("有效期至：%v", secrets.WeworkAccessTokenExpiresIn)
	return nil
}

func SendWeWorkMessage(message Message) error {
	// https://qyapi.weixin.qq.com/cgi-bin/message/send?access_token=
	log.Infof("发送企业微信应用消息：%s", message)
	// 检查token是否过期
	if time.Now().Unix() > secrets.WeworkAccessTokenExpiresIn {
		err := UpdateAccessToken()
		if err != nil {
			return err
		}
	}
	var body = fmt.Sprintf("{\"touser\":\"@all\","+
		"\"msgtype\":\"markdown\","+
		"\"agentid\":\"%s\","+
		"\"markdown\":{\"content\":\"# %s\n\n%s\"},"+
		"\"enable_duplicate_check\":1,"+
		"\"duplicate_check_interval\":600}", secrets.AgentID, message.Title, message.Content)
	target := fmt.Sprintf("https://qyapi.weixin.qq.com/cgi-bin/message/send?access_token=%s&debug=1", secrets.WeworkAccessToken)
	resp, err := http.Post(target, "application/json", bytes.NewReader([]byte(body)))
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Errorf("关闭消息发送响应失败：%s", err.Error())
		}
	}(resp.Body)
	content, err := ioutil.ReadAll(resp.Body)
	errcode := jsoniter.Get(content, "errcode").ToInt()
	if errcode != 0 {
		log.Errorf("发送企业微信应用消息失败：%s", jsoniter.Get(content, "errmsg").ToString())
		return err
	}
	if err != nil {
		return err
	}
	return nil
}

func (m *Message) Send() error {
	log.Infof("发送消息：%+v", *m)
	err1 := SendBarkMessage(*m)
	err2 := SendWeWorkMessage(*m)
	if !(err1 == nil || err2 == nil) {
		return err1
	}
	return nil
}
