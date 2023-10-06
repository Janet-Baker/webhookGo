package bilibiliInfo

import (
	"errors"
	jsoniter "github.com/json-iterator/go"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"strings"
	"time"
)

var tr = &http.Transport{
	ForceAttemptHTTP2:     true,
	MaxIdleConns:          8,
	IdleConnTimeout:       10 * time.Second,
	TLSHandshakeTimeout:   10 * time.Second,
	ExpectContinueTimeout: 1 * time.Second,
	Proxy:                 http.ProxyFromEnvironment,
	// 不需要 Keep alive
	DisableKeepAlives: true,
	// 因为没有处理压缩，所以禁用掉
	DisableCompression: true,
}

var BiliBiliClient = &http.Client{
	Transport: tr,
	Timeout:   10 * time.Second,
}

// AvatarDict = map[uid]avatar
var AvatarDict = make(map[string]string)

// RoomidUidDict = map[roomid]uid
var RoomidUidDict = make(map[string]string)

/*func init() {
	// 我要干什么来着？
}*/

// GetAvatarByUid 通过uid获取头像
func GetAvatarByUid(uid string, webhookId string) (string, error) {
	// 检查缓存的字典中是否已经存在
	avatar, ok := AvatarDict[uid]
	if ok {
		// 存在，直接返回
		return avatar, nil
	} else {
		// 不存在，重新获取
		// 构造请求 "https://api.live.bilibili.com/live_user/v1/Master/info?uid="+uid
		var urlBuilder strings.Builder
		urlBuilder.WriteString("https://api.live.bilibili.com/live_user/v1/Master/info?uid=")
		urlBuilder.WriteString(uid)
		req, errRequest := http.NewRequest("GET", urlBuilder.String(), nil)
		if errRequest != nil {
			log.Errorf("%s 请求用户头像 构造请求失败：%s", webhookId, errRequest.Error())
			return "", errRequest
		}
		// Header 是从 Edge 抄来的
		req.Header.Add("accept", "application/json, text/plain, */*")
		// 暂不启用压缩，因为没加gzip模块
		// req.Header.Add("Accept-Encoding", "gzip, deflate, br")
		req.Header.Add("Accept-Encoding", "")
		req.Header.Add("accept-language", "zh-CN,zh;q=0.9,en-US;q=0.8,en;q=0.7")
		req.Header.Add("Cache-Control", "no-cache")
		req.Header.Add("dnt", "1")
		req.Header.Add("origin", "https://live.bilibili.com")
		req.Header.Add("referer", "https://live.bilibili.com/")
		req.Header.Add("Pragma", "no-cache")
		req.Header.Add("sec-ch-ua", "\"Not/A)Brand\";v=\"99\", \"Microsoft Edge\";v=\"115\", \"Chromium\";v=\"115\"")
		req.Header.Add("sec-ch-ua-mobile", "?0")
		req.Header.Add("sec-ch-ua-platform", "\"Windows\"")
		req.Header.Add("Sec-Fetch-Dest", "empty")
		req.Header.Add("Sec-Fetch-Mode", "cors")
		req.Header.Add("Sec-Fetch-Site", "same-site")
		req.Header.Add("Sec-Gpc", "1")
		req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Safari/537.36 Edg/117.0.2045.55")
		// 发起请求
		resp, err := BiliBiliClient.Do(req)
		if err != nil {
			log.Errorf("%s 请求用户头像 请求失败：%s", webhookId, err.Error())
			return "", err
		}
		defer func(Body io.ReadCloser) {
			errCloser := Body.Close()
			if errCloser != nil {
				log.Errorf("%s 请求用户头像 关闭消息发送响应失败：%s", webhookId, errCloser.Error())
			}
		}(resp.Body)
		// 读取请求
		content, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Errorf("%s 请求用户头像 读取响应消息失败：%s", webhookId, err.Error())
			return "", err
		}
		log.Tracef("%s 请求用户头像 响应：%s", webhookId, content)
		// 检查错误码
		code := jsoniter.Get(content, "code").ToString()
		if "0" != code {
			log.Errorf("%s 请求用户头像 失败：%s", webhookId, jsoniter.Get(content, "message").ToString())
			return "", errors.New(jsoniter.Get(content, "message").ToString())
		}
		// 读取头像
		avatar := jsoniter.Get(content, "data", "info", "face").ToString()
		// 加入缓存
		AvatarDict[uid] = avatar
		return avatar, nil
	}
}

// GetUidByRoomid 通过roomid获取uid
func GetUidByRoomid(roomId string, webhookId string) (string, error) {
	// 检查缓存的字典中是否已经存在
	uid, ok := RoomidUidDict[roomId]
	if ok {
		// 存在，直接返回
		return uid, nil
	} else {
		// 不存在，重新获取
		// 构造请求
		var urlBuilder strings.Builder
		urlBuilder.WriteString("https://api.live.bilibili.com/room/v1/Room/get_info?room_id=")
		urlBuilder.WriteString(roomId)
		req, errRequest := http.NewRequest("GET", urlBuilder.String(), nil)
		if errRequest != nil {
			log.Errorf("%s 请求用户uid 构造请求失败：%s", webhookId, errRequest.Error())
			return "", errRequest
		}
		req.Header.Add("accept", "application/json, text/plain, */*")
		// 暂不启用压缩，因为没加gzip模块
		// req.Header.Add("Accept-Encoding", "gzip, deflate, br")
		req.Header.Add("Accept-Encoding", "")
		req.Header.Add("accept-language", "zh-CN,zh;q=0.9,en-US;q=0.8,en;q=0.7")
		req.Header.Add("Cache-Control", "no-cache")
		req.Header.Add("dnt", "1")
		req.Header.Add("origin", "https://live.bilibili.com")
		req.Header.Add("referer", "https://live.bilibili.com/")
		req.Header.Add("Pragma", "no-cache")
		req.Header.Add("sec-ch-ua", "\"Not/A)Brand\";v=\"99\", \"Microsoft Edge\";v=\"115\", \"Chromium\";v=\"115\"")
		req.Header.Add("sec-ch-ua-mobile", "?0")
		req.Header.Add("sec-ch-ua-platform", "\"Windows\"")
		req.Header.Add("Sec-Fetch-Dest", "empty")
		req.Header.Add("Sec-Fetch-Mode", "cors")
		req.Header.Add("Sec-Fetch-Site", "same-site")
		req.Header.Add("Sec-Gpc", "1")
		req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Safari/537.36 Edg/117.0.2045.55")
		// 发起请求
		resp, err := BiliBiliClient.Do(req)
		if err != nil {
			log.Errorf("%s 请求用户uid 请求失败：%s", webhookId, err.Error())
			return "", err
		}
		defer func(Body io.ReadCloser) {
			errCloser := Body.Close()
			if errCloser != nil {
				log.Errorf("%s 请求用户uid 关闭消息发送响应失败：%s", webhookId, errCloser.Error())
			}
		}(resp.Body)
		// 读取请求
		content, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Errorf("%s 请求用户uid 读取响应消息失败：%s", webhookId, err.Error())
			return "", err
		}
		log.Tracef("%s 请求用户uid 响应：%s", webhookId, content)
		// 检查错误码
		code := jsoniter.Get(content, "code").ToString()
		if "0" != code {
			log.Errorf("%s 请求用户uid 失败：%s", webhookId, jsoniter.Get(content, "message").ToString())
			return "", errors.New(jsoniter.Get(content, "message").ToString())
		}
		// 读取头像
		uid := jsoniter.Get(content, "data", "uid").ToString()
		AvatarDict[roomId] = uid
		return uid, nil
	}
}

// GetAvatarByRoomID 通过roomid获取头像
func GetAvatarByRoomID(roomid string, webhookId string) string {
	uid, err := GetUidByRoomid(roomid, webhookId)
	if err != nil {
		return ""
	}
	avatar, err := GetAvatarByUid(uid, webhookId)
	if err != nil {
		return ""
	}
	return avatar
}
