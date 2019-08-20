package test

import (
	"encoding/binary"
	"fmt"
	"proj_bcbm/src/client/common"
	"proj_bcbm/src/server/msg"
	"reflect"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
)

// 进入大厅之后接收广播并打印

func TestHall(t *testing.T) {
	conn, _, err := websocket.DefaultDialer.Dial("ws://"+"0.0.0.0"+":"+"10086", nil)
	if err != nil {
		fmt.Println("[TestRoom]连接错误", err)
	}

	// loginMsg := &msg.Login{UserID:955509280, Password:"123456"}
	loginMsg := &msg.LoginTest{UserID: 955509280}
	loginBS := common.ByteMsg(loginMsg)
	err = conn.WriteMessage(websocket.TextMessage, loginBS)
	if err != nil {
		fmt.Println("[TestRoom]写消息错误", err)
	}

	respChan := make(chan interface{}, 1)

	go func() {
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				fmt.Println("[TestRoom]读数据错误", err)
			}

			id := binary.BigEndian.Uint16(message[:2])
			resp := common.TransIDToMsg(id)
			err = proto.Unmarshal(message[2:], resp)
			if err != nil {
				fmt.Println("[TestRoom]解析数据错误", err)
			}
			respChan <- resp
		}
	}()
	a := <-respChan
	fmt.Printf("recv: %+v, content: %v\n", reflect.TypeOf(a), a)
}
