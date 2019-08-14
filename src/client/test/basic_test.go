package test

import (
	"proj_bcbm/src/client/common"
	"proj_bcbm/src/server/msg"
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
