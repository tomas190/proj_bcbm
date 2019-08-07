package gate

import (
	"server/game"
	"server/msg"
)

func init() {
	msg.Processor.SetRouter(&msg.Ping{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.LoginTest{}, game.ChanRPC)

	msg.Processor.SetRouter(&msg.Login{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.Logout{}, game.ChanRPC)

	msg.Processor.SetRouter(&msg.JoinRoom{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.LeaveRoom{}, game.ChanRPC)

	msg.Processor.SetRouter(&msg.GrabDealer{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.AutoBet{}, game.ChanRPC)
}
