package test

import (
	"proj_bcbm/src/client/common"
	"proj_bcbm/src/server/msg"
	"testing"
)

func TestPing(t *testing.T) {
	m := msg.Ping{}
	bs := common.ByteMsg(&m)
	common.WSWriteRead(bs)
}

func TestLogin(t *testing.T) {
	m := msg.LoginTest{UserID: 12345}
	bs := common.ByteMsg(&m)
	common.WSWriteRead(bs)
}

func TestTestLogin(t *testing.T) {
	m := msg.LoginTest{UserID: 908789}
	bs := common.ByteMsg(&m)
	common.WSWriteRead(bs)
}

func TestLogout(t *testing.T) {
	m := msg.Logout{}
	bs := common.ByteMsg(&m)
	common.WSWriteRead(bs)
}

func TestJoinRoom(t *testing.T) {
	m := msg.JoinRoom{RoomID: 1}
	bs := common.ByteMsg(&m)
	common.WSWriteRead(bs)
}

func TestBet(t *testing.T) {
	m := msg.Bet{Area: 1, ChipSize: 10}
	bs := common.ByteMsg(&m)
	common.WSWriteRead(bs)
}

func TestGrabBanker(t *testing.T) {
	m := msg.GrabBanker{}
	bs := common.ByteMsg(&m)
	common.WSWriteRead(bs)
}

func TestAutoBet(t *testing.T) {
	m := msg.AutoBet{}
	bs := common.ByteMsg(&m)
	common.WSWriteRead(bs)
}

func TestLeaveRoom(t *testing.T) {
	m := msg.LeaveRoom{}
	bs := common.ByteMsg(&m)
	common.WSWriteRead(bs)
}
