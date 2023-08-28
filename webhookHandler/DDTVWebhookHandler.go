package webhookHandler

import (
	"fmt"
	jsoniter "github.com/json-iterator/go"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"sync"
	"webhookTemplate/messageSender"
)

func DDTVWebhookHandler(w http.ResponseWriter, request *http.Request) {
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
	var ddtvwg sync.WaitGroup
	ddtvwg.Add(1)
	go func() {
		// 读取请求内容
		content, err := ioutil.ReadAll(request.Body)
		ddtvwg.Done()
		if err != nil {
			log.Errorf("读取 DDTV webhook 请求失败：%s", err.Error())
			return
		}
		log.Infof("收到 DDTV webhook 请求")
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
		hookType := jsoniter.Get(content, "type").ToInt()
		switch hookType {
		//	0 StartLive 主播开播
		case 0:
			log.Debugf("DDTV 主播开播：%s", jsoniter.Get(content, "room_Info", "uname").ToString())
			/*				var msg = messageSender.Message{
								Title: fmt.Sprintf("%s 开播了", jsoniter.Get(content, "room_Info", "uname").ToString()),
								Content: fmt.Sprintf("- 主播：%s\n\n- 标题：%s\n\n- 分区：%s - %s\n\n- 开播时间：%s",
									jsoniter.Get(content, "room_Info", "uname").ToString(),
									jsoniter.Get(content, "room_Info", "title").ToString(),
									jsoniter.Get(content, "room_Info", "area_v2_parent_name").ToString(),
									jsoniter.Get(content, "room_Info", "area_v2_name").ToString(),
									jsoniter.Get(content, "hook_time").ToString()),
							}
							msg.Send()*/
			break

		//	1 StopLive 主播下播
		case 1:
			// 主播正常下播
			log.Debugf("DDTV 主播下播：%s", jsoniter.Get(content, "room_Info", "uname").ToString())
			break

		//	2 StartRec 开始录制
		case 2:
			log.Debugf("DDTV 开始录制：%s", jsoniter.Get(content, "room_Info", "uname").ToString())
			break

		//	3 RecComplete 录制结束
		case 3:
			log.Debugf("DDTV 录制结束：%s", jsoniter.Get(content, "room_Info", "uname").ToString())
			break

		//	4 CancelRec 录制被取消
		case 4:
			log.Debugf("DDTV 录制被取消：%s", jsoniter.Get(content, "room_Info", "uname").ToString())
			break

		//	5 TranscodingComplete 完成转码
		case 5:
			log.Debugf("DDTV 完成转码：%s", jsoniter.Get(content, "room_Info", "uname").ToString())
			/*var msg = messageSender.Message{
				Title: fmt.Sprintf("%s 转码完成", jsoniter.Get(content, "room_Info", "uname").ToString()),
				Content: fmt.Sprintf("主播：%s\n标题：%s\n转码完成时间：%s",
					jsoniter.Get(content, "room_Info", "uname").ToString(),
					jsoniter.Get(content, "room_Info", "title").ToString(),
					jsoniter.Get(content, "hook_time").ToString()),
			}
			msg.Send()*/
			break

		//	6 SaveDanmuComplete 保存弹幕文件完成
		case 6:
			log.Debugf("DDTV 保存弹幕文件完成：%s", jsoniter.Get(content, "room_Info", "uname").ToString())
			break

		//	7 SaveSCComplete 保存SC文件完成
		case 7:
			log.Debugf("DDTV 保存SC文件完成：%s", jsoniter.Get(content, "room_Info", "uname").ToString())
			break

		//	8 SaveGiftComplete 保存礼物文件完成
		case 8:
			log.Debugf("DDTV 保存礼物文件完成：%s", jsoniter.Get(content, "room_Info", "uname").ToString())
			break

		//	9 SaveGuardComplete 保存大航海文件完成
		case 9:
			log.Debugf("DDTV 保存大航海文件完成：%s", jsoniter.Get(content, "room_Info", "uname").ToString())
			break

		//	10 RunShellComplete 执行Shell命令完成
		case 10:
			log.Debugf("DDTV 执行Shell命令完成：%s", jsoniter.Get(content, "room_Info", "uname").ToString())
			break

		//	11 DownloadEndMissionSuccess 下载任务成功结束
		case 11:
			log.Debugf("DDTV 下载任务成功结束：%s", jsoniter.Get(content, "room_Info", "uname").ToString())
			if jsoniter.Get(content, "room_Info", "is_locked").ToBool() {
				// 主播被封号了
				var msg = messageSender.Message{
					Title: fmt.Sprintf("%s 喜提直播间封禁！", jsoniter.Get(content, "room_Info", "uname").ToString()),
					Content: fmt.Sprintf("- 主播：%s\n\n- 标题：%s\n\n- 分区：%s - %s\n\n- 封禁时间：%s\n\n- 封禁到：%s",
						jsoniter.Get(content, "room_Info", "uname").ToString(),
						jsoniter.Get(content, "room_Info", "title").ToString(),
						jsoniter.Get(content, "room_Info", "area_v2_parent_name").ToString(),
						jsoniter.Get(content, "room_Info", "area_v2_name").ToString(),
						jsoniter.Get(content, "hook_time").ToString(),
						jsoniter.Get(content, "room_Info", "lock_till").ToString()),
				}
				msg.Send()
			}
			break

		//	12 SpaceIsInsufficientWarn 剩余空间不足
		case 12:
			log.Debugf("DDTV 剩余空间不足：%s", content)
			break

		//	13 LoginFailure 登陆失效
		case 13:
			log.Debugf("DDTV 登陆失效")
			break

		//	14 LoginWillExpireSoon 登陆即将失效
		case 14:
			log.Debugf("DDTV 登陆即将失效")
			break

		//	15 UpdateAvailable 有可用新版本
		case 15:
			log.Debugf("DDTV 有可用新版本：%s", jsoniter.Get(content, "version").ToString())
			break

		//	16 ShellExecutionComplete 执行Shell命令结束
		case 16:
			log.Debugf("DDTV 执行Shell命令结束：%+v", content)
			break

		//	别的不关心，所以没写
		default:
			log.Warnf("DDTV 未知的webhook请求：%+v", content)
		}
	}()
	ddtvwg.Wait()
}
