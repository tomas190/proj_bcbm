package common

import (
	"encoding/binary"
	"fmt"
	"github.com/golang/protobuf/proto"
	"reflect"
)

func WSMsg(msg interface{}) []byte {
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
	id := TransMsgToID(fmt.Sprintf("%v", reflect.TypeOf(msg)))
	tagId := make([]byte, 2)
	binary.BigEndian.PutUint16(tagId, id)
	m = append(tagId, m...)
	// 封入 payload
	copy(m[2:], payload)

	// 打印
	fmt.Println("*************", id, reflect.TypeOf(msg), len(payload))
	for i, b := range m {
		fmt.Println(i, "-", b, string(b))
	}

	return m
}
