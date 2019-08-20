package common

import (
	"encoding/binary"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
	"reflect"
)

const (
	host = "0.0.0.0"
	tcpPort = "10085"
	wsPort = "10086"
)

func ByteMsg(msg interface{}) []byte {
	payload, err := proto.Marshal(msg.(proto.Message))
	if err != nil {
		fmt.Println("Marshal error ", err)
	}

	// 创建一个新的字节数组，也可以在payload操作
	m := make([]byte, len(payload))
	if len(payload) > 0 {
		binary.BigEndian.PutUint16(m, uint16(len(payload)))
	}

	// 封入 id 字段
	// -------------------------
	// | id | protobuf message |
	// -------------------------
	id := transMsgToID(fmt.Sprintf("%v", reflect.TypeOf(msg)))
	tagId := make([]byte, 2)
	binary.BigEndian.PutUint16(tagId, id)
	m = append(tagId, m...)
	// 封入 payload
	copy(m[2:], payload)

	// 打印 - 用于调试
	//fmt.Println("*************", id, reflect.TypeOf(msg), len(payload))
	//for i, b := range m {
	//	fmt.Println(i, "-", b, string(b))
	//}

	return m
}

func WSWriteRead(bs []byte)  {
	conn, _, err := websocket.DefaultDialer.Dial("ws://"+host+":"+wsPort, nil)
	if err != nil {
		fmt.Println("[WSWriteRead]连接错误", err)
	}
	err = conn.WriteMessage(websocket.TextMessage, bs)
	if err != nil {
		fmt.Println("[WSWriteRead]写消息错误", err)
	}

	respChan := make(chan interface{}, 1)

	go func() {
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				fmt.Println("[WSWriteRead]读数据错误", err)
			}

			id := binary.BigEndian.Uint16(message[:2])
			resp := TransIDToMsg(id)
			err = proto.Unmarshal(message[2:], resp)
			if err != nil {
				fmt.Println("[WSWriteRead]解析数据错误", err)
			}
			respChan <- resp
		}
	}()
	a := <-respChan
	fmt.Printf("recv: %+v content: %v\n", reflect.TypeOf(a), a)
}
