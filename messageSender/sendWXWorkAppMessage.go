package messageSender

import (
	"bytes"
	"encoding/json"
	"errors"
	log "github.com/sirupsen/logrus"
	"github.com/valyala/fastjson"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

type noCopy struct{}

type wxWorkAppToken struct {
	noCopy noCopy
	sync.RWMutex
	accessToken   string
	tokenExpireAt atomic.Int64
	//tokenExpireAt int64
}

func (token *wxWorkAppToken) isExpired() bool {
	token.RLock()
	defer token.RUnlock()
	return time.Now().Unix() > token.tokenExpireAt.Load()
}

type WXWorkAppTarget struct {
	CorpId    string `yaml:"corpId"`
	AppSecret string `yaml:"appSecret"`
	AgentID   string `yaml:"agentId"`
	ToUser    string `yaml:"to_user"`
	token     *wxWorkAppToken
}

func (app WXWorkAppTarget) RegisterServer() {
	app.token = new(wxWorkAppToken)
	RegisterMessageServer(app)
}

func updateAccessToken(app WXWorkAppTarget) error {
	// https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=ID&corpsecret=SECRET
	log.Info("更新企业微信应用的access_token")
	// 构造请求地址
	url := "https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=" + app.CorpId + "&corpsecret=" + app.AppSecret
	log.Trace("更新企业微信应用的access_token：请求地址：", url)
	// 发送请求
	resp, err := http.Get(url)
	if err != nil {
		log.Error("更新企业微信应用的access_token：请求发送失败：", err.Error())
		return err
	}
	defer func(Body io.ReadCloser) {
		errCloser := Body.Close()
		if errCloser != nil {
			log.Error("更新企业微信应用的access_token：关闭连接失败：", err.Error())
		}
	}(resp.Body)
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("更新企业微信应用的access_token：读取响应消息失败：", err.Error())
		return err
	}
	log.Trace("更新企业微信应用的access_token：响应消息：", content)

	var p fastjson.Parser
	getter, errOfJsonParser := p.ParseBytes(content)
	if errOfJsonParser != nil {
		return errOfJsonParser
	}
	errCode := getter.GetInt("errcode")
	if 0 != errCode {
		log.Error("更新企业微信应用的access_token失败：", getter.GetStringBytes("errmsg"))
		return errors.New(string(getter.GetStringBytes("errmsg")))
	}
	app.token.accessToken = string(getter.GetStringBytes("access_token"))
	app.token.tokenExpireAt.Store(time.Now().Unix() + getter.GetInt64("expires_in"))
	//atomic.StoreInt64(&app.token.tokenExpireAt, time.Now().Unix()+getter.GetInt64("expires_in"))
	//app.token.tokenExpireAt = time.Now().Unix() + getter.GetInt64("expires_in")
	log.Trace("企业微信AccessToken：", app.token.accessToken)
	log.Debug("企业微信AccessToken有效期至：", app.token.tokenExpireAt.Load())
	return nil
}

type WXWorkAppMessageStruct struct {
	Touser                 string   `json:"touser"`
	Msgtype                string   `json:"msgtype"`
	Agentid                string   `json:"agentid"`
	Markdown               Markdown `json:"markdown"`
	EnableDuplicateCheck   int      `json:"enable_duplicate_check"`
	DuplicateCheckInterval int      `json:"duplicate_check_interval"`
}

type Markdown struct {
	Content string `json:"content"`
}

func (app WXWorkAppTarget) SendMessage(message Message) {
	if message == nil {
		return
	}
	if app.CorpId == "" || app.AppSecret == "" || app.AgentID == "" {
		return
	}
	// 检查token是否过期
	if app.token.isExpired() {
		func() { // 更新之前需要加锁，防止有线程正在更新
			app.token.Lock()
			defer app.token.Unlock()
			// 再次判断过期时间，防止被其他线程更新过了
			if time.Now().Unix() > app.token.tokenExpireAt.Load() {
				err := updateAccessToken(app)
				if err != nil {
					return
				}
			}
		}()
	}
	// 制作要发送的 Markdown 消息
	//var bodyBuffer bytes.Buffer
	//bodyBuffer.WriteString(`{"touser":"`)
	//bodyBuffer.WriteString(app.ToUser)
	//bodyBuffer.WriteString(`","msgtype":"markdown","agentid":"`)
	//bodyBuffer.WriteString(app.AgentID)
	//bodyBuffer.WriteString(`","markdown":{"content":"# `)
	//bodyBuffer.WriteString(message.GetTitle())
	//bodyBuffer.WriteString("\n")
	//bodyBuffer.WriteString(message.GetContent())
	//bodyBuffer.WriteString(`"},"enable_duplicate_check":1,"duplicate_check_interval":3600}`)
	var messageStruct = WXWorkAppMessageStruct{
		Touser:                 app.ToUser,
		Msgtype:                "markdown",
		Agentid:                app.AgentID,
		Markdown:               Markdown{"# " + message.GetTitle() + "\n\n" + message.GetContent()},
		EnableDuplicateCheck:   1,
		DuplicateCheckInterval: 3600,
	}
	bodyBuffer, err := json.Marshal(messageStruct)
	if err != nil {
		log.Error("发送企业微信应用消息 消息体编码失败", err)
		return
	}
	// target: 发送目标，企业微信API https://qyapi.weixin.qq.com/cgi-bin/message/send?access_token=
	targetUrl := "https://qyapi.weixin.qq.com/cgi-bin/message/send?access_token=" + app.token.accessToken

	// 发送请求
	log.Trace("发送企业微信应用消息 请求地址", targetUrl)
	resp, err := http.Post(targetUrl, "application/json", bytes.NewBuffer(bodyBuffer))
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
	defer func(Body io.ReadCloser) {
		errCloser := Body.Close()
		if errCloser != nil {
			log.Error("发送企业微信应用消息 关闭连接失败", errCloser.Error())
		}
	}(resp.Body)
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
