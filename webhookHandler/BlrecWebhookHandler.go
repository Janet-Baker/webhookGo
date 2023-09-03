package webhookHandler

import (
	"fmt"
	jsoniter "github.com/json-iterator/go"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
	"webhookTemplate/messageSender"
)

// BlrecWebhookHandler 处理 blrec 的 webhook 请求
func BlrecWebhookHandler(w http.ResponseWriter, request *http.Request) {
	// defer request.Body.Close()
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(request.Body)
	// return 200 at first
	w.WriteHeader(http.StatusOK)

	// process other steps in another goroutine
	var ioReaderWaitGroup sync.WaitGroup
	ioReaderWaitGroup.Add(1)
	go func() {
		// 读取请求内容
		content, err := ioutil.ReadAll(request.Body)
		// 读取完毕，解除阻塞
		ioReaderWaitGroup.Done()
		if err != nil {
			log.Errorf("读取 blrec webhook 请求失败：%s", err.Error())
			return
		}
		log.Infof("收到 blrec webhook 请求")
		log.Debugf(string(content))

		// 判断是否是重复的webhook请求
		webhookId := jsoniter.Get(content, "id").ToString()
		log.Debug(webhookId)
		webhookMessageIdListLock.Lock()
		if webhookMessageIdList.IsContain(webhookId) {
			webhookMessageIdListLock.Unlock()
			log.Warnf("重复的webhook请求：%s", webhookId)
			return
		} else {
			webhookMessageIdList.EnQueue(webhookId)
			webhookMessageIdListLock.Unlock()
		}

		// 判断事件类型
		hookType := jsoniter.Get(content, "type").ToString()
		switch hookType {
		// LiveBeganEvent 主播开播
		case "LiveBeganEvent":
			log.Infof("blrec 主播开播：%s", jsoniter.Get(content, "data", "user_info", "name").ToString())
			var msg = messageSender.Message{
				Title: fmt.Sprintf("%s 开播了", jsoniter.Get(content, "data", "user_info", "name").ToString()),
				Content: fmt.Sprintf("- 主播：[%s](https://live.bilibili.com/%s)\n- 标题：%s\n- 分区：%s - %s\n- 开播时间：%s",
					jsoniter.Get(content, "data", "user_info", "name").ToString(),
					jsoniter.Get(content, "data", "room_info", "room_id").ToString(),
					jsoniter.Get(content, "data", "room_info", "title").ToString(),
					jsoniter.Get(content, "data", "room_info", "parent_area_name").ToString(),
					jsoniter.Get(content, "data", "room_info", "area_name").ToString(),
					time.Unix(jsoniter.Get(content, "data", "room_info", "live_start_time").ToInt64(), 0).Local().Format("2006-01-02 15:04:05")),
			}
			msg.Send()
			break

		// LiveEndedEvent 主播下播
		case "LiveEndedEvent":
			log.Infof("blrec 主播下播：%s", jsoniter.Get(content, "data", "user_info", "name").ToString())
			var msg = messageSender.Message{
				Title: fmt.Sprintf("%s 下播了", jsoniter.Get(content, "data", "user_info", "name").ToString()),
				Content: fmt.Sprintf("- 主播：[%s](https://live.bilibili.com/%s)\n- 标题：%s\n- 分区：%s - %s",
					jsoniter.Get(content, "data", "user_info", "name").ToString(),
					jsoniter.Get(content, "data", "room_info", "room_id").ToString(),
					jsoniter.Get(content, "data", "room_info", "title").ToString(),
					jsoniter.Get(content, "data", "room_info", "parent_area_name").ToString(),
					jsoniter.Get(content, "data", "room_info", "area_name").ToString()),
			}
			msg.Send()
			break

		// RoomChangeEvent 直播间信息改变
		case "RoomChangeEvent":
			log.Debugf("blrec 直播间信息改变：%s", jsoniter.Get(content, "data", "user_info", "room_id").ToString())
			break

		// RecordingStartedEvent 录制开始
		case "RecordingStartedEvent":
			log.Infof("blrec 录制开始：room_id %s", jsoniter.Get(content, "data", "room_info", "room_id").ToString())
			break

		// RecordingFinishedEvent 录制结束
		case "RecordingFinishedEvent":
			log.Infof("blrec 录制结束：room_id %s", jsoniter.Get(content, "data", "room_info", "room_id").ToString())
			break

		// RecordingCancelledEvent 录制取消
		case "RecordingCancelledEvent":
			log.Infof("blrec 录制取消：room_id %s", jsoniter.Get(content, "data", "room_info", "room_id").ToString())
			break

		// VideoFileCreatedEvent 视频文件创建
		case "VideoFileCreatedEvent":
			log.Debugf("blrec 视频文件创建：%s", jsoniter.Get(content, "data", "path").ToString())
			break

		// VideoFileCompletedEvent 视频文件完成
		case "VideoFileCompletedEvent":
			log.Infof("blrec 视频文件完成：%s", jsoniter.Get(content, "data", "path").ToString())
			/*var msg = messageSender.Message{
				Title: fmt.Sprintf("%s 视频文件完成", jsoniter.Get(content, "data", "room_id").ToString()),
				Content: fmt.Sprintf("文件路径：%s",
					jsoniter.Get(content, "data", "path").ToString()),
			}
			msg.Send()*/
			break

		// DanmakuFileCreatedEvent 弹幕文件创建
		case "DanmakuFileCreatedEvent":
			log.Debugf("blrec 弹幕文件创建：%s", jsoniter.Get(content, "data", "path").ToString())
			break

		// DanmakuFileCompletedEvent 弹幕文件完成
		case "DanmakuFileCompletedEvent":
			log.Infof("blrec 弹幕文件完成：%s", jsoniter.Get(content, "data", "path").ToString())
			break

		// RawDanmakuFileCreatedEvent 原始弹幕文件创建
		case "RawDanmakuFileCreatedEvent":
			log.Debugf("blrec 原始弹幕文件创建：%s", jsoniter.Get(content, "data", "path").ToString())
			break

		// RawDanmakuFileCompletedEvent 原始弹幕文件完成
		case "RawDanmakuFileCompletedEvent":
			log.Debugf("blrec 原始弹幕文件完成：%s", jsoniter.Get(content, "data", "path").ToString())
			break

		// VideoPostprocessingCompletedEvent 视频后处理完成
		case "VideoPostprocessingCompletedEvent":
			log.Debugf("blrec 视频后处理完成：%s", jsoniter.Get(content, "data", "path").ToString())
			break

		// SpaceNoEnoughEvent 硬盘空间不足
		case "SpaceNoEnoughEvent":
			log.Warnf("blrec 硬盘空间不足：文件路径：%s；可用空间：%v",
				jsoniter.Get(content, "data", "path").ToString(),
				jsoniter.Get(content, "data", "usage", "free").ToUint64())
			break

		// Error 程序出现异常
		case "Error":
			log.Errorf("blrec 程序出现异常：%+v", jsoniter.Get(content, "data"))
			break

		default:
			log.Warnf("未知的 blrec webhook 请求类型：%s", hookType)
		}
	}()
	// 等待响应体读取完毕
	ioReaderWaitGroup.Wait()
}
