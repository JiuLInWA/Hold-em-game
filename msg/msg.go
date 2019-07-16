package msg

import (
	"github.com/name5566/leaf/log"
	"github.com/name5566/leaf/network/protobuf"
	pb_msg "server/msg/Protocal"
)

// 使用默认的 Json 消息处理器 (默认还提供了 Protobuf 消息处理器)
var Processor = protobuf.NewProcessor()

func init() {
	log.Debug("msg init~")
	// 注册 UserLogin 协议
	//Processor.Register(&pb_msg.Test{})
	Processor.Register(&pb_msg.PingC2S{})             //--0
	Processor.Register(&pb_msg.PongS2C{})             //--1
	Processor.Register(&pb_msg.SvrMsgS2C{})           //--2
	Processor.Register(&pb_msg.LoginC2S{})            //--3
	Processor.Register(&pb_msg.LoginResultS2C{})      //--4
	Processor.Register(&pb_msg.QuickStartC2S{})       //--5
	Processor.Register(&pb_msg.CreateRoomC2S{})       //--6
	Processor.Register(&pb_msg.JoinRoomC2S{})         //--7
	Processor.Register(&pb_msg.EnterRoomS2C{})        //--8
	Processor.Register(&pb_msg.ExitRoomC2S{})         //--9
	Processor.Register(&pb_msg.ExitRoomS2C{})         //--10
	Processor.Register(&pb_msg.OtherPlayerJoinS2C{})  //--11
	Processor.Register(&pb_msg.OtherPlayerLeaveS2C{}) //--12
	Processor.Register(&pb_msg.SitDownC2S{})          //--13
	Processor.Register(&pb_msg.SitDownS2C{})          //--14
	Processor.Register(&pb_msg.StandUpC2S{})          //--15
	Processor.Register(&pb_msg.StandUpS2C{})          //--16
	Processor.Register(&pb_msg.GameStepChangeS2C{})   //--17

}
