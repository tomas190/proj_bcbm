package test

import (
	"fmt"
	"github.com/gorilla/websocket"
	"math/rand"
	"proj_bcbm/src/client/common"
	"proj_bcbm/src/server/msg"
	"testing"
	"time"
)

// 进入房间之后监听广播并打印

func TestRoom(t *testing.T) {
	conn, _, err := websocket.DefaultDialer.Dial("ws://"+"0.0.0.0"+":"+"10086", nil)
	if err != nil {
		fmt.Println("[TestRoom]连接错误", err)
	}

	// loginMsg := &msg.Login{UserID:955509280, Password:"123456"}
	loginMsg := &msg.LoginTest{UserID: 955509287}
	loginBS := common.ByteMsg(loginMsg)
	err = conn.WriteMessage(websocket.TextMessage, loginBS)
	if err != nil {
		fmt.Println("[TestRoom]写消息错误1", err)
	}

	joinMsg := &msg.JoinRoom{RoomID: 1}
	joinBS := common.ByteMsg(joinMsg)
	err = conn.WriteMessage(websocket.TextMessage, joinBS)
	if err != nil {
		fmt.Println("[TestRoom]写消息错误2", err)
	}

	rand.Seed(time.Now().Unix())
	for i := 0; i < 100; i++ {
		time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
		betMsg := &msg.Bet{Area: uint32(rand.Intn(8) + 1), Chip: uint32(rand.Intn(5) + 1)}
		betBS := common.ByteMsg(betMsg)
		err := conn.WriteMessage(websocket.TextMessage, betBS)
		if err != nil {
			fmt.Println("[TestRoom]写消息错误3", err)
		}
	}

	//respChan := make(chan interface{}, 1)
	//
	//go func() {
	//	for {
	//		_, message, err := conn.ReadMessage()
	//		if err != nil {
	//			fmt.Println("[TestRoom]读数据错误", err)
	//		}
	//
	//		id := binary.BigEndian.Uint16(message[:2])
	//		resp := common.TransIDToMsg(id)
	//		err = proto.Unmarshal(message[2:], resp)
	//		if err != nil {
	//			fmt.Println("[TestRoom]解析数据错误", err)
	//		}
	//		respChan <- resp
	//	}
	//}()
	//a := <-respChan
	//fmt.Printf("recv: %+v content: %v\n", reflect.TypeOf(a), a)
}
