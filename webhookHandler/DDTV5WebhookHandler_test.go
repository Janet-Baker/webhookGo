package webhookHandler

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"testing"
	"webhookGo/bilibiliInfo"
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
					"C:\\Users\\1\\DDTV\\Desktop\\Rec\\4983935_小胖子等六名用户\\2024_06_26\\2024_06_26_21_28_22_100%鲜橙汁！_1.xml"
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

func TestUnmarshalObject2(t *testing.T) {
	var content = []byte(`{"cmd":"ModifyRoomPromptConfiguration","code":20105,"data":{"cover_from_user":{"Value":"https://i0.hdslb.com/bfs/live/new_room_cover/8a6d4b65d2fb90e4dc7528726a67d4111b598e20.jpg"},"CurrentMode":0,"description":{"Value":""},"DownInfo":{"LiveChatListener":{"RoomId":27209189,"Name":"宅豆糕糕ki","File":"C:/Users/10632/Videos/bilibili/27209189-宅豆糕糕ki/2024-12-01/20241201-202908","State":false,"Register":["DetectRoom_LiveStart"],"DanmuMessage":{"FileName":null,"TimeStopwatch":null,"Danmu":[],"SuperChat":[],"Gift":[],"GuardBuy":[]},"TimeStopwatch":{"IsRunning":true,"Elapsed":"00:22:24.0147536","ElapsedMilliseconds":1344014,"ElapsedTicks":13440147607},"SaveCount":1},"IsDownload":true,"IsCut":false,"taskType":2,"DownloadSize":1018953322,"RealTimeDownloadSpe":781090.8556342461,"Status":2,"StartTime":"2024-12-01T20:29:40.6734313+08:00","EndTime":"2024-12-01T00:25:59.8455716+08:00","DownloadFileList":{"TranscodingCount":0,"VideoFile":[],"DanmuFile":[],"SCFile":[],"GiftFile":[],"GuardFile":[],"CurrentOperationVideoFile":"C:/Users/10632/Videos/bilibili//27209189-宅豆糕糕ki/2024-12-01/20241201-202940_original.flv","SnapshotGenerationInProgress":false}},"keyframe":{"Value":"https://i0.hdslb.com/bfs/live-key-frame/keyframe12012045000027209189fn07ku.jpg"},"live_status":{"Value":1},"live_status_end_event":false,"live_time":{"Value":1733056138},"short_id":{"Value":0},"Title":{"Value":"ow怀旧服~最后两天了似乎！?"},"Name":"宅豆糕糕ki","Description":"","RoomId":27209189,"UID":1855623989,"IsAutoRec":true,"IsRemind":true,"IsRecDanmu":true,"Like":false,"Shell":"","AppointmentRecord":false},"message":"修改房间提示设置，房间UID:1855623989[True]"}`)
	var message DDTV5MessageStruct
	err := json.Unmarshal(content, &message)
	if err != nil {
		log.Error("解析 DDTV5 webhook 请求失败：", err)
		t.Fail()
		return
	}
	t.Logf("%+v", message)
	var event Event = Event{
		Care:        true,
		Notify:      true,
		HaveCommand: false,
		ExecCommand: "",
	}
	bilibiliInfo.ContactBilibili = true
	handleDDTV5Messages(event, message)
}
