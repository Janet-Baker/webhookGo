package bilibiliInfo

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"github.com/valyala/fastjson"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// 请求相关
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

// 数据缓存相关
var (
	// avatarDict = map[uid]avatar 根据uid存储头像
	avatarDict = make(map[uint64]string)

	// roomidUidDict = map[roomid]uid 根据roomid存储uid
	roomidUidDict = make(map[uint64]uint64)
)

// reqHeaderSetter 设置请求头
func reqHeaderSetter(header *http.Header) {
	// Header 是从 Edge 抄来的
	header.Set("accept", "application/json, text/plain, */*")
	// 暂不启用压缩，因为没加gzip模块
	// header.Add("Accept-Encoding", "gzip, deflate, br")
	header.Set("Accept-Encoding", "")
	header.Set("accept-language", "zh-CN,zh;q=0.9,en-US;q=0.8,en;q=0.7")
	header.Set("Cache-Control", "no-cache")
	header.Set("dnt", "1")
	header.Set("origin", "https://live.bilibili.com")
	header.Set("referer", "https://live.bilibili.com/")
	header.Set("Pragma", "no-cache")
	header.Set("sec-ch-ua", "\"Not/A)Brand\";v=\"99\", \"Microsoft Edge\";v=\"115\", \"Chromium\";v=\"115\"")
	header.Set("sec-ch-ua-mobile", "?0")
	header.Set("sec-ch-ua-platform", "\"Windows\"")
	header.Set("Sec-Fetch-Dest", "empty")
	header.Set("Sec-Fetch-Mode", "cors")
	header.Set("Sec-Fetch-Site", "same-site")
	header.Set("Sec-Gpc", "1")
	header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Safari/537.36 Edg/117.0.2045.55")
	return
}

// GetUidByRoomid 通过roomid获取uid
func GetUidByRoomid(roomId uint64, webhookId string) (uint64, error) {
	// 检查缓存的字典中是否已经存在
	uid, ok := roomidUidDict[roomId]
	if ok {
		// 存在，直接返回
		return uid, nil
	}
	// 不存在，重新获取
	// 构造请求
	var urlBuilder strings.Builder
	urlBuilder.WriteString("https://api.live.bilibili.com/room/v1/Room/room_init?id=")
	urlBuilder.WriteString(strconv.FormatUint(roomId, 10))
	req, errRequest := http.NewRequest("GET", urlBuilder.String(), nil)
	if errRequest != nil {
		log.Error(webhookId, "请求用户uid 构造请求失败", errRequest.Error())
		return 0, errRequest
	}
	reqHeaderSetter(&req.Header)
	// 发起请求
	resp, err := BiliBiliClient.Do(req)
	if err != nil {
		log.Error(webhookId, "请求用户uid 请求失败", err.Error())
		return 0, err
	}
	defer func(Body io.ReadCloser) {
		errCloser := Body.Close()
		if errCloser != nil {
			log.Error(webhookId, "请求用户uid 关闭消息发送响应失败", errCloser.Error())
		}
	}(resp.Body)
	// 读取请求
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error(webhookId, "请求用户uid 读取响应消息失败", err.Error())
		return 0, err
	}
	log.Trace(webhookId, "请求用户uid 响应", content)

	var p fastjson.Parser
	getter, errOfJsonParser := p.ParseBytes(content)
	if errOfJsonParser != nil {
		return 0, errOfJsonParser
	}

	// 检查错误码
	code := getter.GetInt("code")
	if 0 != code {
		log.Error(webhookId, "请求用户uid 失败", getter.GetStringBytes("message"))
		return 0, errors.New(string(getter.GetStringBytes("msg")))
	}
	// 读取 uid
	uid = getter.GetUint64("data", "uid")
	roomidUidDict[roomId] = uid
	return uid, nil

}

// GetAvatarByUid 通过uid获取头像
func GetAvatarByUid(uid uint64, webhookId string) (string, error) {
	// 检查缓存的字典中是否已经存在
	avatar, ok := avatarDict[uid]
	if ok {
		// 存在，直接返回
		return avatar, nil
	}
	// 不存在，重新获取
	// 构造请求 "https://api.live.bilibili.com/live_user/v1/Master/info?uid="+uid
	var urlBuilder strings.Builder
	urlBuilder.WriteString("https://api.live.bilibili.com/live_user/v1/Master/info?uid=")
	urlBuilder.WriteString(strconv.FormatUint(uid, 10))
	req, errRequest := http.NewRequest("GET", urlBuilder.String(), nil)
	if errRequest != nil {
		log.Error(webhookId, "请求用户头像 构造请求失败：", errRequest.Error())
		return "", errRequest
	}
	reqHeaderSetter(&req.Header)
	// 发起请求
	resp, err := BiliBiliClient.Do(req)
	if err != nil {
		log.Error(webhookId, "请求用户头像 请求失败：", err.Error())
		return "", err
	}
	defer func(Body io.ReadCloser) {
		errCloser := Body.Close()
		if errCloser != nil {
			log.Error(webhookId, "请求用户头像 关闭消息发送响应失败：", errCloser.Error())
		}
	}(resp.Body)
	// 读取请求
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error(webhookId, "请求用户头像 读取响应消息失败：", err.Error())
		return "", err
	}
	log.Trace(webhookId, "请求用户头像 响应：", content)

	var p fastjson.Parser
	getter, errOfJsonParser := p.ParseBytes(content)
	if errOfJsonParser != nil {
		return "", errOfJsonParser
	}

	// 检查错误码
	code := getter.GetInt("code")

	if 0 != code {
		log.Error(webhookId, "请求用户头像 失败：", getter.GetStringBytes("message"))
		return "", errors.New(string(getter.GetStringBytes("msg")))
	}
	// 读取头像
	avatar = string(getter.GetStringBytes("data", "info", "face"))

	// 加入缓存
	avatarDict[uid] = avatar
	return avatar, nil

}

// GetAvatarByRoomID 通过roomid获取头像
func GetAvatarByRoomID(roomid uint64, webhookId string) string {
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

// IsRoomLocked 检查主播房间是否被封禁 返回状态和时间戳
func IsRoomLocked(roomId uint64, webhookId string) (bool, int64) {
	// 每次下播时更新
	time.Sleep(1 * time.Second)
	// 构造请求
	var urlBuilder strings.Builder
	// 不要用 https://api.live.bilibili.com/room/v1/Room/getBannedInfo?roomid= 因为获取不到数据了
	urlBuilder.WriteString("https://api.live.bilibili.com/room/v1/Room/room_init?id=")
	urlBuilder.WriteString(strconv.FormatUint(roomId, 10))
	req, errRequest := http.NewRequest("GET", urlBuilder.String(), nil)
	if errRequest != nil {
		log.Error(webhookId, "请求直播间封禁状态 构造请求失败", errRequest.Error())
		return false, 0
	}
	reqHeaderSetter(&req.Header)
	// 发起请求
	resp, err := BiliBiliClient.Do(req)
	if err != nil {
		log.Error(webhookId, "请求直播间封禁状态 请求失败", err.Error())
		return false, 0
	}
	defer func(Body io.ReadCloser) {
		errCloser := Body.Close()
		if errCloser != nil {
			log.Error(webhookId, "请求直播间封禁状态 关闭消息发送响应失败", errCloser.Error())
		}
	}(resp.Body)
	// 读取请求
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error(webhookId, "请求直播间封禁状态 读取响应消息失败", err.Error())
		return false, 0
	}
	log.Trace(webhookId, "请求直播间封禁状态 响应", content)

	var p fastjson.Parser
	getter, errOfJsonParser := p.ParseBytes(content)
	if errOfJsonParser != nil {
		return false, 0
	}

	// 检查错误码
	code := getter.GetInt("code")
	//code := roomInfo.Code
	if 0 != code {
		log.Error(webhookId, "请求直播间封禁状态 失败", getter.GetStringBytes("message"))
		return false, 0
	}

	// 读取 uid
	uid := getter.GetUint64("data", "uid")
	roomidUidDict[roomId] = uid
	// 读取锁定状态
	isLocked := getter.GetBool("data", "is_locked")
	lockTill := getter.GetInt64("data", "lock_till")
	return isLocked, lockTill
}
