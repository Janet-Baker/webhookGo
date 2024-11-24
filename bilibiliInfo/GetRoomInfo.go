package bilibiliInfo

import (
	"encoding/json"
	"errors"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"
)

var ContactBilibili = true

// 请求相关
var tr = &http.Transport{
	// ForceAttemptHTTP2:     true,
	MaxIdleConns:          8,
	IdleConnTimeout:       10 * time.Second,
	TLSHandshakeTimeout:   10 * time.Second,
	ExpectContinueTimeout: 1 * time.Second,
	Proxy:                 http.ProxyFromEnvironment,
	// 不需要 Keep alive
	// DisableKeepAlives: true,
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
func getAndUnmarshal(url string, v any) error {
	req, errRequest := http.NewRequest("GET", url, nil)
	if errRequest != nil {
		log.Error("请求失败", errRequest.Error())
		return errRequest
	}
	reqHeaderSetter(&req.Header)
	// 发起请求
	resp, err := BiliBiliClient.Do(req)
	if err != nil {
		log.Error("请求失败", err.Error())
		return err
	}
	defer func(Body io.ReadCloser) {
		errCloser := Body.Close()
		if errCloser != nil {
			log.Error("关闭消息发送响应失败", errCloser.Error())
		}
	}(resp.Body)
	// 读取请求
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error("读取响应消息失败", err.Error())
		return err
	}
	if log.IsLevelEnabled(log.TraceLevel) {
		log.Trace("响应", string(content))
	}

	err = json.Unmarshal(content, v)
	if err != nil {
		log.Error("解析失败:", string(content), err.Error())
		return err
	}
	return nil
}

// RoomInit 房间初始化信息
// IsRoomLocked, GetLiveStatus, GetUidByRoomid
type RoomInit struct {
	lastUpdate int64
	rwMutex    *sync.RWMutex
	Code       int    `json:"code"`
	Msg        string `json:"msg"`
	Message    string `json:"message"`
	Data       struct {
		RoomId      int64 `json:"room_id"`
		ShortId     int   `json:"short_id"`
		Uid         int64 `json:"uid"`
		NeedP2P     int   `json:"need_p2p"`
		IsHidden    bool  `json:"is_hidden"`
		IsLocked    bool  `json:"is_locked"`
		IsPortrait  bool  `json:"is_portrait"`
		LiveStatus  int   `json:"live_status"`
		HiddenTill  int   `json:"hidden_till"`
		LockTill    int64 `json:"lock_till"`
		Encrypted   bool  `json:"encrypted"`
		PwdVerified bool  `json:"pwd_verified"`
		LiveTime    int64 `json:"live_time"`
		RoomShield  int   `json:"room_shield"`
		IsSp        int   `json:"is_sp"`
		SpecialType int   `json:"special_type"`
	} `json:"data"`
}

// MasterInfo 取头像
// GetAvatarByUid
type MasterInfo struct {
	lastUpdate int64
	rwMutex    *sync.RWMutex
	Code       int    `json:"code"`
	Msg        string `json:"msg"`
	Message    string `json:"message"`
	Data       struct {
		Info struct {
			Uid            int64  `json:"uid"`
			Uname          string `json:"uname"`
			Face           string `json:"face"`
			OfficialVerify struct {
				Type int    `json:"type"`
				Desc string `json:"desc"`
			} `json:"official_verify"`
			Gender int `json:"gender"`
		} `json:"info"`
		Exp struct {
			MasterLevel struct {
				Level   int   `json:"level"`
				Color   int   `json:"color"`
				Current []int `json:"current"`
				Next    []int `json:"next"`
			} `json:"master_level"`
		} `json:"exp"`
		FollowerNum  int    `json:"follower_num"`
		RoomId       int64  `json:"room_id"`
		MedalName    string `json:"medal_name"`
		GloryCount   int    `json:"glory_count"`
		Pendant      string `json:"pendant"`
		LinkGroupNum int    `json:"link_group_num"`
		RoomNews     struct {
			Content   string `json:"content"`
			Ctime     string `json:"ctime"`
			CtimeText string `json:"ctime_text"`
		} `json:"room_news"`
	} `json:"data"`
}

// GetInfo
// GetUidByRoomid, GetAreaV2ParentName, GetAreaV2Name, GetLiveStatusString
type GetInfo struct {
	lastUpdate int64
	rwMutex    *sync.RWMutex
	Code       int    `json:"code"`
	Msg        string `json:"msg"`
	Message    string `json:"message"`
	Data       struct {
		Uid              int64    `json:"uid"`
		RoomId           int64    `json:"room_id"`
		ShortId          int      `json:"short_id"`
		Attention        int      `json:"attention"`
		Online           int      `json:"online"`
		IsPortrait       bool     `json:"is_portrait"`
		Description      string   `json:"description"`
		LiveStatus       int      `json:"live_status"`
		AreaId           int      `json:"area_id"`
		ParentAreaId     int      `json:"parent_area_id"`
		ParentAreaName   string   `json:"parent_area_name"`
		OldAreaId        int      `json:"old_area_id"`
		Background       string   `json:"background"`
		Title            string   `json:"title"`
		UserCover        string   `json:"user_cover"`
		Keyframe         string   `json:"keyframe"`
		IsStrictRoom     bool     `json:"is_strict_room"`
		LiveTime         string   `json:"live_time"`
		Tags             string   `json:"tags"`
		IsAnchor         int      `json:"is_anchor"`
		RoomSilentType   string   `json:"room_silent_type"`
		RoomSilentLevel  int      `json:"room_silent_level"`
		RoomSilentSecond int64    `json:"room_silent_second"`
		AreaName         string   `json:"area_name"`
		Pendants         string   `json:"pendants"`
		AreaPendants     string   `json:"area_pendants"`
		HotWords         []string `json:"hot_words"`
		HotWordsStatus   int      `json:"hot_words_status"`
		Verify           string   `json:"verify"`
		NewPendants      struct {
			Frame struct {
				Name       string `json:"name"`
				Value      string `json:"value"`
				Position   int    `json:"position"`
				Desc       string `json:"desc"`
				Area       int    `json:"area"`
				AreaOld    int    `json:"area_old"`
				BgColor    string `json:"bg_color"`
				BgPic      string `json:"bg_pic"`
				UseOldArea bool   `json:"use_old_area"`
			} `json:"frame"`
			Badge struct {
				Name     string `json:"name"`
				Position int    `json:"position"`
				Value    string `json:"value"`
				Desc     string `json:"desc"`
			} `json:"badge"`
			MobileFrame struct {
				Name       string `json:"name"`
				Value      string `json:"value"`
				Position   int    `json:"position"`
				Desc       string `json:"desc"`
				Area       int    `json:"area"`
				AreaOld    int    `json:"area_old"`
				BgColor    string `json:"bg_color"`
				BgPic      string `json:"bg_pic"`
				UseOldArea bool   `json:"use_old_area"`
			} `json:"mobile_frame"`
			MobileBadge struct {
				Name     string `json:"name"`
				Position int    `json:"position"`
				Value    string `json:"value"`
				Desc     string `json:"desc"`
			} `json:"mobile_badge"`
		} `json:"new_pendants"`
		UpSession            string `json:"up_session"`
		PkStatus             int    `json:"pk_status"`
		PkId                 int    `json:"pk_id"`
		BattleId             int    `json:"battle_id"`
		AllowChangeAreaTime  int    `json:"allow_change_area_time"`
		AllowUploadCoverTime int    `json:"allow_upload_cover_time"`
		StudioInfo           struct {
			Status     int           `json:"status"`
			MasterList []interface{} `json:"master_list"`
		} `json:"studio_info"`
	} `json:"data"`
}

// 数据缓存相关
var (
	// avatarDict = map[uid]avatar 根据uid存储头像
	avatarDict = sync.Map{}
	// roomidUidDict = map[roomid]uid 根据roomid存储uid
	roomidUidDict   = sync.Map{}
	uidUsernameDict = sync.Map{}
	// roomidGetInfoDict = map[roomid]GetInfo 根据roomid存储房间信息
	roomidGetInfoDict = sync.Map{}
	// roomidRoomInitDict = map[roomid]RoomInit 根据roomid存储房间信息
	roomidRoomInitDict = sync.Map{}
	// uidMasterInfoDict = map[uid]MasterInfo 根据uid存储主播信息
	uidMasterInfoDict = sync.Map{}

	newGetInfoApiLock  = sync.Mutex{}
	newRoomInitApiLock = sync.Mutex{}
	newMasterInfoLock  = sync.Mutex{}
)

const expirePeriod = 3600

func (roomInit *RoomInit) isExpired() bool {
	roomInit.rwMutex.RLock()
	defer roomInit.rwMutex.RUnlock()
	return time.Now().Unix()-roomInit.lastUpdate > expirePeriod
}
func (masterInfo *MasterInfo) isExpired() bool {
	masterInfo.rwMutex.RLock()
	defer masterInfo.rwMutex.RUnlock()
	return time.Now().Unix()-masterInfo.lastUpdate > expirePeriod
}
func (getInfo *GetInfo) isExpired() bool {
	getInfo.rwMutex.RLock()
	defer getInfo.rwMutex.RUnlock()
	return time.Now().Unix()-getInfo.lastUpdate > expirePeriod
}

func forceGetInfo(roomId int64) (GetInfo, error) {
	if roomId <= 0 {
		return GetInfo{}, errors.New("roomId 不可为 0")
	}
	//	https://api.live.bilibili.com/room/v1/Room/get_info
	urlBuilder := "https://api.live.bilibili.com/room/v1/Room/get_info?room_id=" + strconv.FormatInt(roomId, 10)
	var getInfoResult GetInfo
	err := getAndUnmarshal(urlBuilder, &getInfoResult)
	if err != nil {
		return GetInfo{}, err
	}
	getInfoResult.lastUpdate = time.Now().Unix()
	getInfoResult.rwMutex = new(sync.RWMutex)
	// 检查错误码
	code := getInfoResult.Code
	if 0 != code {
		log.Error("请求房间信息 失败", getInfoResult.Message)
		return GetInfo{}, errors.New(getInfoResult.Message)
	}
	return getInfoResult, nil
}
func forceRoomInit(roomId int64) (RoomInit, error) {
	if roomId <= 0 {
		return RoomInit{}, errors.New("roomId 不可为 0")
	}
	// 构造请求
	urlBuilder := "https://api.live.bilibili.com/room/v1/Room/room_init?id=" + strconv.FormatInt(roomId, 10)
	var roomInitResult RoomInit
	err := getAndUnmarshal(urlBuilder, &roomInitResult)
	if err != nil {
		return RoomInit{}, err
	}
	roomInitResult.lastUpdate = time.Now().Unix()
	roomInitResult.rwMutex = new(sync.RWMutex)
	// 检查错误码
	code := roomInitResult.Code
	if 0 != code {
		log.Error("请求房间信息 失败", roomInitResult.Message)
		return RoomInit{}, errors.New(roomInitResult.Message)
	}
	roomidRoomInitDict.Store(roomId, roomInitResult)
	return roomInitResult, nil
}
func forceMasterInfo(uid int64) (MasterInfo, error) {
	if uid <= 0 {
		return MasterInfo{}, errors.New("uid 不可为 0")
	}
	// 构造请求
	urlBuilder := "https://api.live.bilibili.com/live_user/v1/Master/info?uid=" + strconv.FormatInt(uid, 10)
	var masterInfoResult MasterInfo
	err := getAndUnmarshal(urlBuilder, &masterInfoResult)
	if err != nil {
		return MasterInfo{}, err
	}
	masterInfoResult.lastUpdate = time.Now().Unix()
	masterInfoResult.rwMutex = new(sync.RWMutex)

	// 检查错误码
	code := masterInfoResult.Code
	if 0 != code {
		log.Error("请求主播信息 失败", masterInfoResult.Message)
		return MasterInfo{}, errors.New(masterInfoResult.Message)
	}
	uidMasterInfoDict.Store(uid, masterInfoResult)
	return masterInfoResult, nil
}
func getInfo(roomId int64) (*GetInfo, error) {
	if !ContactBilibili {
		return nil, http.ErrNotSupported
	}
	iGetInfo, ok := roomidGetInfoDict.Load(roomId)
	if ok {
		getInfoResult := iGetInfo.(GetInfo)
		// 存在，检查是否过期
		for getInfoResult.isExpired() {
			// 过期？拿写锁再看一次
			func() {
				getInfoResult.rwMutex.Lock()
				defer getInfoResult.rwMutex.Unlock()
				iGetInfo, _ = roomidGetInfoDict.Load(roomId)
				getInfoResult2 := iGetInfo.(GetInfo)
				if time.Now().Unix()-getInfoResult2.lastUpdate > expirePeriod {
					_, err := forceGetInfo(roomId)
					if err != nil {
						return
					}
				}
			}()
			iGetInfo, _ = roomidGetInfoDict.Load(roomId)
			getInfoResult = iGetInfo.(GetInfo)
		}
		return &getInfoResult, nil
	} else {
		newGetInfoApiLock.Lock()
		var needUnlock = true
		defer func() {
			if needUnlock {
				newGetInfoApiLock.Unlock()
			}
		}()
		iGetInfo, ok = roomidGetInfoDict.Load(roomId)
		if ok {
			needUnlock = false
			newGetInfoApiLock.Unlock()
			return getInfo(roomId)
		} else {
			// 初次获取
			getInfoResult, err := forceGetInfo(roomId)
			if err != nil {
				return nil, err
			}
			return &getInfoResult, nil
		}
	}
}
func roomInit(roomId int64) (*RoomInit, error) {
	if !ContactBilibili {
		return nil, http.ErrNotSupported
	}
	iRoomInit, ok := roomidRoomInitDict.Load(roomId)
	if ok {
		roomInitResult := iRoomInit.(RoomInit)
		// 存在，检查是否过期
		for roomInitResult.isExpired() && roomInitResult.lastUpdate != 0 {
			// 过期？拿写锁再看一次
			func() {
				roomInitResult.rwMutex.Lock()
				defer roomInitResult.rwMutex.Unlock()
				iRoomInit, _ = roomidRoomInitDict.Load(roomId)
				roomInitResult2 := iRoomInit.(RoomInit)
				if time.Now().Unix()-roomInitResult2.lastUpdate > expirePeriod {
					_, _ = forceRoomInit(roomId)
				}
			}()
			iRoomInit, _ = roomidRoomInitDict.Load(roomId)
			roomInitResult = iRoomInit.(RoomInit)
		}
		return &roomInitResult, nil
	} else {
		newRoomInitApiLock.Lock()
		var needUnlock = true
		defer func() {
			if needUnlock {
				newRoomInitApiLock.Unlock()
			}
		}()
		iRoomInit, ok = roomidRoomInitDict.Load(roomId)
		if ok {
			needUnlock = false
			newRoomInitApiLock.Unlock()
			return roomInit(roomId)
		} else {
			// 初次获取
			roomInitResult, err := forceRoomInit(roomId)
			if err != nil {
				return nil, err
			}
			return &roomInitResult, nil
		}
	}
}
func masterInfo(uid int64) (*MasterInfo, error) {
	if !ContactBilibili {
		return nil, http.ErrNotSupported
	}
	iMasterInfo, ok := uidMasterInfoDict.Load(uid)
	if ok {
		masterInfoResult := iMasterInfo.(MasterInfo)
		// 存在，检查是否过期
		for masterInfoResult.isExpired() {
			// 过期？拿写锁再看一次
			func() {
				masterInfoResult.rwMutex.Lock()
				defer masterInfoResult.rwMutex.Unlock()
				iMasterInfo, _ = uidMasterInfoDict.Load(uid)
				masterInfoResult2 := iMasterInfo.(MasterInfo)
				if time.Now().Unix()-masterInfoResult2.lastUpdate > expirePeriod {
					_, err := forceMasterInfo(uid)
					if err != nil {
						return
					}
				}
			}()
			iMasterInfo, _ = uidMasterInfoDict.Load(uid)
			masterInfoResult = iMasterInfo.(MasterInfo)
		}
		return &masterInfoResult, nil
	} else {
		newMasterInfoLock.Lock()
		var needUnlock = true
		defer func() {
			if needUnlock {
				newMasterInfoLock.Unlock()
			}
		}()
		iMasterInfo, ok = uidMasterInfoDict.Load(uid)
		if ok {
			needUnlock = false
			newMasterInfoLock.Unlock()
			return masterInfo(uid)
		} else {
			// 初次获取
			masterInfoResult, err := forceMasterInfo(uid)
			if err != nil {
				return nil, err
			}
			return &masterInfoResult, nil
		}
	}
}

// GetUidByRoomid 通过roomid获取uid
func GetUidByRoomid(roomId int64) (int64, error) {
	if roomId == 0 {
		return 0, errors.New("roomid 不应该为 0")
	}
	// 检查缓存的字典中是否已经存在
	var uid int64
	iuid, ok := roomidUidDict.Load(roomId)
	if ok {
		uid = iuid.(int64)
		// 存在，直接返回
		return uid, nil
	}
	iGetInfo, ok := roomidGetInfoDict.Load(roomId)
	if ok {
		getInfoResult := iGetInfo.(GetInfo)
		uid = getInfoResult.Data.Uid
		roomidUidDict.Store(roomId, uid)
		return uid, nil
	}
	iRoomInit, ok := roomidRoomInitDict.Load(roomId)
	if ok {
		roomInitResult := iRoomInit.(RoomInit)
		uid = roomInitResult.Data.Uid
		roomidUidDict.Store(roomId, uid)
		return uid, nil
	}

	// 不存在，重新获取
	if !ContactBilibili {
		return 0, nil
	}
	// 构造请求

	roomInitResult, err := roomInit(roomId)
	if err != nil {
		return 0, err
	}

	// 读取 uid
	uid = roomInitResult.Data.Uid
	//roomidUidDict[roomId] = uid
	roomidUidDict.Store(roomId, uid)
	return uid, nil
}

// GetAvatarByUid 通过uid获取头像
func GetAvatarByUid(uid int64) (string, error) {
	if uid == 0 {
		return "", errors.New("uid 不应该为 0")
	}
	// 检查缓存的字典中是否已经存在
	var avatar string
	iavatar, ok := avatarDict.Load(uid)
	if ok {
		avatar = iavatar.(string)
		// 存在，直接返回
		return avatar, nil
	}
	// 不存在，重新获取
	// 构造请求 "https://api.live.bilibili.com/live_user/v1/Master/info?uid="+uid
	masterInfoResult, err := masterInfo(uid)
	if err != nil {
		return "", err
	}

	// 读取头像
	avatar = masterInfoResult.Data.Info.Face

	// 加入缓存
	//avatarDict[uid] = avatar
	avatarDict.Store(uid, avatar)
	return avatar, nil
}

// GetAvatarByRoomID 通过roomid获取头像
func GetAvatarByRoomID(roomId int64) (string, error) {
	if roomId == 0 {
		return "", errors.New("roomid 不应该为 0")
	}
	uid, err := GetUidByRoomid(roomId)
	if err != nil {
		return "", err
	}
	avatar, err := GetAvatarByUid(uid)
	if err != nil {
		return "", err
	}
	return avatar, nil
}

// GetUsernameByRoomId 通过uid获取用户名
func GetUsernameByUid(uid int64) (string, error) {
	if uid == 0 {
		return "", errors.New("uid 不应该为 0")
	}
	// 检查缓存的字典中是否已经存在
	var username string
	iusername, ok := uidUsernameDict.Load(uid)
	if ok {
		username = iusername.(string)
		// 存在，直接返回
		return username, nil
	}
	// 不存在，重新获取
	masterInfoResult, err := masterInfo(uid)
	if err != nil {
		return "", err
	}
	username = masterInfoResult.Data.Info.Uname
	//uidUsernameDict[uid] = username
	uidUsernameDict.Store(uid, username)
	return username, nil
}

// GetUsernameByRoomId 通过roomid获取用户名
func GetUsernameByRoomId(roomId int64) (string, error) {
	if roomId == 0 {
		return "", errors.New("roomid 不应该为 0")
	}
	uid, err := GetUidByRoomid(roomId)
	if err != nil {
		return "", err
	}
	username, err := GetUsernameByUid(uid)
	if err != nil {
		return "", err
	}
	return username, nil
}

// IsRoomLocked 检查主播房间是否被封禁 返回状态和时间戳
func IsRoomLocked(roomId int64) (isLocked bool, lockTill int64) {
	if roomId == 0 {
		return false, 0
	}
	// 构造请求
	roomInitResult, err := forceRoomInit(roomId)
	if err != nil {
		log.Error("请求房间信息 失败", err.Error())
		return false, 0
	}
	// 读取 uid
	uid := roomInitResult.Data.Uid
	//roomidUidDict[roomId] = uid
	roomidUidDict.Store(roomId, uid)
	// 读取锁定状态
	isLocked = roomInitResult.Data.IsLocked
	lockTill = roomInitResult.Data.LockTill
	return isLocked, lockTill
}

func GetAreaV2ParentName(roomId int64) string {
	if roomId == 0 {
		return "获取失败:roomId=0"
	}
	if !ContactBilibili {
		return "未知"
	}
	getInfoResult, err := getInfo(roomId)
	if err != nil {
		return "获取失败"
	}
	return getInfoResult.Data.ParentAreaName
}
func GetAreaV2Name(roomId int64) string {
	if roomId == 0 {
		return "获取失败:roomId=0"
	}
	if !ContactBilibili {
		return "未知"
	}
	getInfoResult, err := getInfo(roomId)
	if err != nil {
		return "获取失败"
	}
	return getInfoResult.Data.AreaName
}
func GetLiveStatusString(roomId int64) (liveStatus int, liveTime string) {
	if roomId == 0 {
		return 0, "获取失败:roomId=0"
	}
	if !ContactBilibili {
		return 0, "未知"
	}
	getInfoResult, err := getInfo(roomId)
	if err != nil {
		return 0, "获取失败" + err.Error()
	}
	return getInfoResult.Data.LiveStatus, getInfoResult.Data.LiveTime
}
func GetLiveStatus(roomId int64) (liveStatus int, liveTime int64) {
	if roomId == 0 {
		return 0, 0
	}
	iRoomInit, ok := roomidRoomInitDict.Load(roomId)
	if ok {
		roomInitResult := iRoomInit.(RoomInit)
		if !roomInitResult.isExpired() && roomInitResult.Data.LiveStatus == 1 {
			return roomInitResult.Data.LiveStatus, roomInitResult.Data.LiveTime
		}
	}

	// 不存在，重新获取
	// 构造请求
	roomInitResult, err := roomInit(roomId)
	if err != nil {
		return 0, 0
	}
	return roomInitResult.Data.LiveStatus, roomInitResult.Data.LiveTime
}
