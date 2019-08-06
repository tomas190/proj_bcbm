package msg

import (
	"fmt"
	"github.com/name5566/leaf/network/protobuf"
	"reflect"
	"strings"
)

var Processor = protobuf.NewProcessor()

func init() {
	// 消息
	Processor.Register(&Error{})
	Processor.Register(&Ping{})
	Processor.Register(&Pong{})

	// 登录登出 进房离房 上庄下庄
	// 上庄10把自动下庄-可配置
	// 庄家金额 < 50000 自动下庄-可配置
	Processor.Register(&Login{})
	Processor.Register(&LoginR{})
	Processor.Register(&Logout{})
	Processor.Register(&LogoutR{})
	Processor.Register(&JoinRoom{})
	Processor.Register(&JoinRoomR{})
	Processor.Register(&LeaveRoom{})
	Processor.Register(&LeaveRoomR{})
	Processor.Register(&GrabDealer{})
	Processor.Register(&AutoBet{})
	Processor.Register(&AutoBetR{})

	// 特定情况触发的广播消息
	Processor.Register(&BetInfoB{})
	Processor.Register(&DealersB{})
	Processor.Register(&PlayersB{})

	// print ID
	Processor.Range(printMsgID)
}

func printMsgID(id uint16, t reflect.Type)  {
	tStr := fmt.Sprintf("%v", t)
	tStr = strings.Replace(tStr, "*", "", 1)
	tStr = strings.Replace(tStr, ".", "", 1)
	tStr = strings.Title(tStr)
	fmt.Println("\t", tStr, "=", id, ";")
}
