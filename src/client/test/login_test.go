package test

import (
	"client/common"
	"server/msg"
	"testing"
)

func TestLogin(t *testing.T) {
	m := msg.LoginTest{UserID:12345}
	bs := common.ByteMsg(&m)
	common.WSWriteRead(bs)
}
