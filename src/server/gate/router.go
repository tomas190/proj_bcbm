package gate

import (
	"server/game"
	"server/msg"
)

func init() {
	msg.Processor.SetRouter(&msg.Ping{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.JoinRoom{}, game.ChanRPC)
	msg.Processor.SetRouter(&msg.LeaveRoom{}, game.ChanRPC)
}
