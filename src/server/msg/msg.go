package msg

import (
	"fmt"
	"github.com/name5566/leaf/network/protobuf"
	"reflect"
	"strings"
)

var Processor = protobuf.NewProcessor()

func init() {
	// 基本消息
	Processor.Register(&Error{})
	Processor.Register(&Ping{})
	Processor.Register(&Pong{})

	Processor.Register(&Login{})
	Processor.Register(&LoginR{})
	Processor.Register(&Logout{})
	Processor.Register(&LogoutR{})
	Processor.Register(&RoomChangeHB{})

	Processor.Register(&JoinRoom{})
	Processor.Register(&JoinRoomR{})
	Processor.Register(&LeaveRoom{})
	Processor.Register(&LeaveRoomR{})

	Processor.Register(&GrabBanker{})
	Processor.Register(&AutoBet{})
	Processor.Register(&AutoBetR{})

	// 下注
	Processor.Register(&Bet{})

	// 特定情况触发的广播消息
	Processor.Register(&BetInfoB{}) // 有人投注广播一次
	Processor.Register(&BankersB{}) // 有人上庄或下庄广播一次
	Processor.Register(&Players{})  // 玩家列表请求
	Processor.Register(&PlayersR{}) // 玩家列表响应
	Processor.Register(&RoomStatusB{})

	// print ID 打印出想要的任意格式
	Processor.Range(printMsgIDPB)
	// Processor.Range(printMsgID)
	// Processor.Range(printMsg)
}

func printMsgIDPB(id uint16, t reflect.Type) {
	tStr := fmt.Sprintf("%v", t)
	tStr = strings.Replace(tStr, "*", "", 1)
	tStr = strings.Replace(tStr, ".", "", 1)
	tStr = strings.Title(tStr)
	fmt.Printf("\t%-20v = %d;\n", tStr, id)
}

func printMsgID(id uint16, t reflect.Type) {
	fmt.Printf("\t\"%v\" : %d,\n", t, id)
}

func printMsg(id uint16, t reflect.Type) {
	tStr := fmt.Sprintf("%v", t)
	tStr = strings.Replace(tStr, "*", "", 1)
	fmt.Printf("case %d: resp = %v{}\n", id, tStr)
}
