package test

import (
	"client/common"
	"encoding/binary"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
	"server/msg"
	"testing"
)

const Host = "127.0.0.1"
const TCPPort = "8888"
const WSPort = "10086"

func TestLogin(t *testing.T) {
	conn, _, err := websocket.DefaultDialer.Dial("ws://"+Host+":"+WSPort, nil)
	if err != nil {
		fmt.Println("[NewUserClient]连接错误", err)
	}

	m := msg.Ping{}
	bs := common.WSMsg(&m)
	err = conn.WriteMessage(websocket.TextMessage, bs)
	if err != nil {
		fmt.Println("[TestLogin]写消息错误", err)
	}


	go func() {
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				fmt.Println("[TestLogin]读数据错误", err)
			}

			id := binary.BigEndian.Uint16(message[:2])
			resp := common.TransIDToMsg(id)
			err = proto.Unmarshal(message[2:], resp.(proto.Message))
			if err != nil {
				fmt.Println("[TestLogin]解析数据错误", err)
			}
		}
	}()
}
