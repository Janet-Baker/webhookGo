package bilibiliInfo

import (
	"fmt"
	"testing"
)

func TestGetUidByRoomid(t *testing.T) {
	ContactBilibili = true
	var roomid int64 = 4983935
	var rightUid int64 = 8511743
	uid, err := GetUidByRoomid(roomid)
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}
	if uid != rightUid {
		t.Errorf("uid: %d, rightUid: %d", uid, rightUid)
		t.Fail()
		return
	}
}

func TestGetAvatarByUid(t *testing.T) {
	ContactBilibili = true
	var uid int64 = 8511743
	avatar, err := GetAvatarByUid(uid)
	if err != nil {
		t.Error(err)
		t.Fail()
		return
	}
	if avatar == "" {
		t.Error("avatar is empty")
		t.Fail()
		return
	}
	t.Log(avatar)
}

func TestGetAreaName1(t *testing.T) {
	ContactBilibili = false
	var roomid int64 = 4983935
	areaName := GetAreaV2Name(roomid)
	t.Log(areaName)
	if areaName == "获取失败" {
		t.Fail()
	}
}

func TestGetAreaName2(t *testing.T) {
	ContactBilibili = true
	var roomid int64 = 4983935
	areaName := GetAreaV2Name(roomid)
	t.Log(areaName)
	if areaName == "未知" || areaName == "获取失败" {
		t.Fail()
	}
}

func Test_forceGetInfo(t *testing.T) {
	infoGot, err := forceGetInfo(27209189)
	if err != nil {
		fmt.Println(err)
		t.Fail()
	}
	fmt.Println(infoGot)
}
