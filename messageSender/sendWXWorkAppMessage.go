package messageSender

import (
	"bytes"
	"errors"
	log "github.com/sirupsen/logrus"
	"github.com/valyala/fastjson"
	"io"
	"net/http"
	"sync"
	"time"
)

type wxWorkAppToken struct {
	sync.Mutex
	accessToken   string
	tokenExpireAt int64
}

type WXWorkAppTarget struct {
	CorpId    string `yaml:"corpId"`
	AppSecret string `yaml:"appSecret"`
	AgentID   string `yaml:"agentId"`
	ToUser    string `yaml:"to_user"`
	token     *wxWorkAppToken
}

var wxWorkAppTargets []WXWorkAppTarget

func RegisterWXWorkApp(target WXWorkAppTarget) {
	target.token = new(wxWorkAppToken)
	wxWorkAppTargets = append(wxWorkAppTargets, target)
}

func updateAccessToken(app WXWorkAppTarget) error {
	// https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=ID&corpsecret=SECRET
	log.Info("更新企业微信应用的access_token")
	// 构造请求地址
	url := "https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=" + app.CorpId + "&corpsecret=" + app.AppSecret
	log.Tracef("更新企业微信应用的access_token：请求地址：%s", url)
	// 发送请求
	resp, err := http.Get(url)
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
	app.token.accessToken = string(getter.GetStringBytes("access_token"))
	app.token.tokenExpireAt = time.Now().Unix() + getter.GetInt64("expires_in")
	log.Trace("企业微信AccessToken：", app.token.accessToken)
	log.Debug("企业微信AccessToken有效期至：", app.token.tokenExpireAt)
	return nil
}

func SendWXWorkAppMessage(message Message) {
	length := len(wxWorkAppTargets)
	if length > 0 {
		wg := sync.WaitGroup{}
		for i := 0; i < length; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				sendWXWorkAppMessage(wxWorkAppTargets[i], message)
			}(i)
		}
		wg.Wait()
	}
}

func sendWXWorkAppMessage(app WXWorkAppTarget, message Message) {
	if app.CorpId == "" || app.AppSecret == "" || app.AgentID == "" {
		return
	}
	// 检查token是否过期
	if time.Now().Unix() > app.token.tokenExpireAt {
		func() { // 更新之前需要加锁，防止有线程正在更新
			app.token.Lock()
			defer app.token.Unlock()
			// 再次判断过期时间，防止被其他线程更新过了
			if time.Now().Unix() > app.token.tokenExpireAt {
				err := updateAccessToken(app)
				if err != nil {
					return
				}
			}
		}()
	}
	// 制作要发送的 Markdown 消息
	var bodyBuffer bytes.Buffer
	bodyBuffer.WriteString(`{"touser":"`)
	bodyBuffer.WriteString(app.ToUser)
	bodyBuffer.WriteString(`","msgtype":"markdown","agentid":"`)
	bodyBuffer.WriteString(app.AgentID)
	bodyBuffer.WriteString(`","markdown":{"content":"# `)
	bodyBuffer.WriteString(message.Title)
	bodyBuffer.WriteString("\n")
	bodyBuffer.WriteString(message.Content)
	bodyBuffer.WriteString(`"},"enable_duplicate_check":1,"duplicate_check_interval":3600}`)

	// target: 发送目标，企业微信API https://qyapi.weixin.qq.com/cgi-bin/message/send?access_token=
	targetUrl := "https://qyapi.weixin.qq.com/cgi-bin/message/send?access_token=" + app.token.accessToken

	// 发送请求
	log.Trace("发送企业微信应用消息 请求地址", targetUrl)
	resp, err := http.Post(targetUrl, "application/json", &bodyBuffer)
	defer func(Body io.ReadCloser) {
		errCloser := Body.Close()
		if errCloser != nil {
			log.Error("发送企业微信应用消息 关闭连接失败", errCloser.Error())
		}
	}(resp.Body)
	if err != nil {
		log.Error("发送企业微信应用消息 请求失败", err.Error())
		return
	}
	// 读取响应消息
	content, errReader := io.ReadAll(resp.Body)
	if errReader != nil {
		log.Error("发送企业微信应用消息 读取响应内容失败", errReader.Error())
		return
	}
	var p fastjson.Parser
	getter, errOfJsonParser := p.ParseBytes(content)
	if errOfJsonParser != nil {
		return
	}
	errcode := getter.GetInt("errcode")
	if errcode != 0 {
		log.Error("发送企业微信应用消息 服务器返回错误", content)
		return
	}
	if log.IsLevelEnabled(log.TraceLevel) {
		log.Trace("发送企业微信应用消息成功 响应消息", content)
	} else {
		log.Debug("发送企业微信应用消息成功 消息id", getter.GetStringBytes("msgid"))
	}
	return
}
