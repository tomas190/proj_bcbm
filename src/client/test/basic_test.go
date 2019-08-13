package test

import (
	"client/common"
	"server/msg"
	"testing"
)

func TestPing(t *testing.T)  {
	m := msg.Ping{}
	bs := common.ByteMsg(&m)
	common.WSWriteRead(bs)
}

func TestLogin(t *testing.T) {
	m := msg.LoginTest{UserID:12345}
	bs := common.ByteMsg(&m)
	common.WSWriteRead(bs)
}
