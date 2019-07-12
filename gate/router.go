package gate

import (
	"server/game"
	"server/msg"
	pb_msg "server/msg/Protocal"
)

func init() {
	// 指定消息 Hello 路由到 game 模块
	//msg.Processor.SetRouter(&pb_msg.Test{},game.ChanRPC)
	msg.Processor.SetRouter(&pb_msg.PingC2S{}, game.ChanRPC)
	msg.Processor.SetRouter(&pb_msg.LoginC2S{}, game.ChanRPC)
	msg.Processor.SetRouter(&pb_msg.QuickStartC2S{}, game.ChanRPC)
	msg.Processor.SetRouter(&pb_msg.CreateRoomC2S{}, game.ChanRPC)
	msg.Processor.SetRouter(&pb_msg.JoinRoomC2S{}, game.ChanRPC)
	msg.Processor.SetRouter(&pb_msg.ExitRoomC2S{}, game.ChanRPC)
}
