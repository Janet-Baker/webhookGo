package bilibiliInfo

import (
	"encoding/json"
	"errors"
	log "github.com/sirupsen/logrus"
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

// 数据缓存相关
var (
	// avatarDict = map[uid]avatar 根据uid存储头像
	avatarDict = make(map[uint64]string)

	// roomidUidDict = map[roomid]uid 根据roomid存储uid
	roomidUidDict = make(map[uint64]uint64)

	// roomInfoDict 缓存主播信息
	roomInfoDict = make(map[uint64]biliLiveRoomInitStruct)
)

// 读取json用的结构体
type (
	// biliLiveMasterInfoStruct 根据uid取得主播信息（头像）
	// https://api.live.bilibili.com/live_user/v1/Master/info?uid=
	biliLiveMasterInfoStruct struct {
		Code    int    `json:"code"`
		Msg     string `json:"msg"`
		Message string `json:"message"`
		Data    struct {
			Info struct {
				Uid uint64 `json:"uid"`
				//Uname          string `json:"uname"`
				Face string `json:"face"`
				//OfficialVerify struct {
				//	Type int    `json:"type"`
				//	Desc string `json:"desc"`
				//} `json:"official_verify"`
				//Gender int `json:"gender"`
			} `json:"info"`
			//Exp struct {
			//	MasterLevel struct {
			//		Level   int   `json:"level"`
			//		Color   int   `json:"color"`
			//		Current []int `json:"current"`
			//		Next    []int `json:"next"`
			//	} `json:"master_level"`
			//} `json:"exp"`
			//FollowerNum  int    `json:"follower_num"`
			RoomId uint64 `json:"room_id"`
			//MedalName    string `json:"medal_name"`
			//GloryCount   int    `json:"glory_count"`
			//Pendant      string `json:"pendant"`
			//LinkGroupNum int    `json:"link_group_num"`
			//RoomNews     struct {
			//	Content   string `json:"content"`
			//	Ctime     string `json:"ctime"`
			//	CtimeText string `json:"ctime_text"`
			//} `json:"room_news"`
		} `json:"data"`
	}

	// biliLiveRoomInitStruct 根据直播间号，取得主播UID和封禁状态
	// https://api.live.bilibili.com/room/v1/Room/room_init?id=
	biliLiveRoomInitStruct struct {
		Code    int    `json:"code"`
		Msg     string `json:"msg"`
		Message string `json:"message"`
		Data    struct {
			RoomId  uint64 `json:"room_id"`
			ShortId int    `json:"short_id"`
			Uid     uint64 `json:"uid"`
			//NeedP2P     int    `json:"need_p2p"`
			//IsHidden    bool   `json:"is_hidden"`
			IsLocked bool `json:"is_locked"`
			//IsPortrait  bool   `json:"is_portrait"`
			//LiveStatus  int    `json:"live_status"`
			//HiddenTill  int    `json:"hidden_till"`
			LockTill int64 `json:"lock_till"`
			//Encrypted   bool   `json:"encrypted"`
			//PwdVerified bool   `json:"pwd_verified"`
			//LiveTime    int    `json:"live_time"`
			//RoomShield  int    `json:"room_shield"`
			//IsSp        int    `json:"is_sp"`
			//SpecialType int    `json:"special_type"`
		} `json:"data"`
	}
)

/*func init() {
	// 我要干什么来着？
}*/

// GetUidByRoomid 通过roomid获取uid
func GetUidByRoomid(roomId uint64, webhookId string) (uint64, error) {
	// 检查缓存的字典中是否已经存在
	uid, ok := roomidUidDict[roomId]
	if ok {
		// 存在，直接返回
		return uid, nil
	} else {
		// 不存在，重新获取
		// 构造请求
		var urlBuilder strings.Builder
		urlBuilder.WriteString("https://api.live.bilibili.com/room/v1/Room/room_init?id=")
		urlBuilder.WriteString(strconv.FormatUint(roomId, 10))
		req, errRequest := http.NewRequest("GET", urlBuilder.String(), nil)
		if errRequest != nil {
			log.Errorf("%s 请求用户uid 构造请求失败：%s", webhookId, errRequest.Error())
			return 0, errRequest
		}
		reqHeaderSetter(&req.Header)
		// 发起请求
		resp, err := BiliBiliClient.Do(req)
		if err != nil {
			log.Errorf("%s 请求用户uid 请求失败：%s", webhookId, err.Error())
			return 0, err
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
			return 0, err
		}
		log.Tracef("%s 请求用户uid 响应：%s", webhookId, content)

		var roomInfo biliLiveRoomInitStruct
		errUnmarshal := json.Unmarshal(content, &roomInfo)
		if errUnmarshal != nil {
			log.Errorf("%s 请求用户uid 解析json失败：%s", webhookId, errUnmarshal.Error())
			return 0, errUnmarshal
		}

		// 检查错误码
		code := roomInfo.Code
		if 0 != code {
			log.Errorf("%s 请求用户uid 失败：%s", webhookId, roomInfo.Message)
			return 0, errors.New(roomInfo.Msg)
		}
		roomInfoDict[roomId] = roomInfo
		// 读取头像
		uid := roomInfo.Data.Uid
		roomidUidDict[roomId] = uid
		return uid, nil
	}
}

// GetAvatarByUid 通过uid获取头像
func GetAvatarByUid(uid uint64, webhookId string) (string, error) {
	// 检查缓存的字典中是否已经存在
	avatar, ok := avatarDict[uid]
	if ok {
		// 存在，直接返回
		return avatar, nil
	} else {
		// 不存在，重新获取
		// 构造请求 "https://api.live.bilibili.com/live_user/v1/Master/info?uid="+uid
		var urlBuilder strings.Builder
		urlBuilder.WriteString("https://api.live.bilibili.com/live_user/v1/Master/info?uid=")
		urlBuilder.WriteString(strconv.FormatUint(uid, 10))
		req, errRequest := http.NewRequest("GET", urlBuilder.String(), nil)
		if errRequest != nil {
			log.Errorf("%s 请求用户头像 构造请求失败：%s", webhookId, errRequest.Error())
			return "", errRequest
		}
		reqHeaderSetter(&req.Header)
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

		var masterInfo biliLiveMasterInfoStruct
		errUnmarshal := json.Unmarshal(content, &masterInfo)
		if errUnmarshal != nil {
			return "", errUnmarshal
		}

		// 检查错误码
		code := masterInfo.Code
		if 0 != code {
			log.Errorf("%s 请求用户头像 失败：%s", webhookId, masterInfo.Message)
			return "", errors.New(masterInfo.Msg)
		}
		// 读取头像
		avatar := masterInfo.Data.Info.Face
		// 加入缓存
		avatarDict[uid] = avatar
		return avatar, nil
	}
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
	urlBuilder.WriteString("https://api.live.bilibili.com/room/v1/Room/room_init?id=")
	urlBuilder.WriteString(strconv.FormatUint(roomId, 10))
	req, errRequest := http.NewRequest("GET", urlBuilder.String(), nil)
	if errRequest != nil {
		log.Errorf("%s 请求用户封禁状态 构造请求失败：%s", webhookId, errRequest.Error())
		return false, 0
	}
	reqHeaderSetter(&req.Header)
	// 发起请求
	resp, err := BiliBiliClient.Do(req)
	if err != nil {
		log.Errorf("%s 请求用户封禁状态 请求失败：%s", webhookId, err.Error())
		return false, 0
	}
	defer func(Body io.ReadCloser) {
		errCloser := Body.Close()
		if errCloser != nil {
			log.Errorf("%s 请求用户封禁状态 关闭消息发送响应失败：%s", webhookId, errCloser.Error())
		}
	}(resp.Body)
	// 读取请求
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("%s 请求用户封禁状态 读取响应消息失败：%s", webhookId, err.Error())
		return false, 0
	}
	log.Tracef("%s 请求用户封禁状态 响应：%s", webhookId, content)
	var roomInfo biliLiveRoomInitStruct
	errUnmarshal := json.Unmarshal(content, &roomInfo)
	if errUnmarshal != nil {
		return false, 0
	}
	// 检查错误码
	code := roomInfo.Code
	if 0 != code {
		log.Errorf("%s 请求用户封禁状态 失败：%s", webhookId, roomInfo.Message)
		return false, 0
	}
	roomInfoDict[roomId] = roomInfo
	return roomInfo.Data.IsLocked, roomInfo.Data.LockTill
}
