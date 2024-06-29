package webhookHandler

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"testing"
)

func TestUnmarshalObject(t *testing.T) {
	var content = []byte(`{
	"cmd": "SaveBulletScreenFile",
	"code": 40101,
	"data": {
		"Title": {
			"Value": "100%鲜橙汁！"
		},
		"description": {
			"Value": ""
		},
		"live_time": {
			"Value": 0
		},
		"live_status": {
			"Value": 0
		},
		"live_status_end_event": false,
		"short_id": {
			"Value": 0
		},
		"area_name": {
			"Value": "单机联机"
		},
		"area_v2_name": {
			"Value": "其他单机"
		},
		"area_v2_parent_name": {
			"Value": "单机游戏"
		},
		"face": {
			"Value": "https://i2.hdslb.com/bfs/face/711b9992d2fb5bbc4ea72d4905826bdc633bc51f.jpg"
		},
		"cover_from_user": {
			"Value": "https://i0.hdslb.com/bfs/live/new_room_cover/adca62f1ee80e67d19a0d56ba43d01074403c19f.jpg"
		},
		"keyframe": {
			"Value": "https://i0.hdslb.com/bfs/live-key-frame/keyframe06161937000004983935pf8onv.jpg"
		},
		"lock_till": {
			"Value": "0000-00-00 00:00:00"
		},
		"is_locked": {
			"Value": false
		},
		"CurrentMode": 0,
		"DownInfo": {
			"IsDownload": true,
			"IsCut": false,
			"taskType": 1,
			"DownloadSize": 0,
			"RealTimeDownloadSpe": 71278.38469563503,
			"Status": 2,
			"StartTime": "2024-06-26T21:28:30.151071+08:00",
			"EndTime": "1970-01-01T00:00:00Z",
			"LiveChatListener": {
				"RoomId": 4983935,
				"Title": "100鲜橙汁！",
				"Name": "小胖子等六名用户",
				"File": "./Rec/4983935_小胖子等六名用户/2024_06_26/2024_06_26_21_28_22_100%鲜橙汁！",
				"State": true,
				"Register": ["DetectRoom_LiveStart"],
				"DanmuMessage": {
					"FileName": null,
					"TimeStopwatch": null,
					"Danmu": [],
					"SuperChat": [],
					"Gift": [],
					"GuardBuy": []
				},
				"TimeStopwatch": {
					"IsRunning": true,
					"Elapsed": "00:00:00.0001413",
					"ElapsedMilliseconds": 0,
					"ElapsedTicks": 1446
				},
				"SaveCount": 2
			},
			"DownloadFileList": {
				"TranscodingCount": 0,
				"VideoFile": [],
				"DanmuFile": [
					"C:\\Users\\10632\\Documents\\GitHub\\DDTV\\Desktop\\bin\\Debug\\net8.0-windows7.0\\Rec\\4983935_小胖子等六名用户\\2024_06_26\\2024_06_26_21_28_22_100%鲜橙汁！_1.xml"
				],
				"SCFile": [],
				"GiftFile": [],
				"GuardFile": [],
				"CurrentOperationVideoFile": "/rec_file/4983935_小胖子等六名用户/2024_06_26/2024_06_26_21_28_22_100%鲜橙汁！_original.mp4"
			}
		},
		"Name": "小胖子等六名用户",
		"Description": "",
		"RoomId": 4983935,
		"UID": 8511743,
		"IsAutoRec": true,
		"IsRemind": false,
		"IsRecDanmu": true,
		"Like": false,
		"Shell": "",
		"AppointmentRecord": false
	},
	"message": "保存弹幕相关文件"
}`)
	var message DDTV5MessageStruct
	err := json.Unmarshal(content, &message)
	if err != nil {
		log.Error("解析 DDTV5 webhook 请求失败：", err)
		t.Fail()
		return
	}
	t.Logf("%+v", message)
}

func TestReceivingRequest(t *testing.T) {
	receivingRequest := func(content []byte) {
		println()
		t.Log(string(content))
		println()
		return
	}

	ddtv5WebhookHandler = receivingRequest
	r := gin.Default()
	r.POST("/ddtv5", DDTV5WebhookHandler)
	r.Run("127.0.0.1:14000")
}
