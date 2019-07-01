package internal

import (
	"fmt"
	"github.com/name5566/leaf/gate"
	pb_msg "server/msg/Protocal"
	"time"
)

func init() {
	skeleton.RegisterChanRPC("NewAgent", rpcNewAgent)
	skeleton.RegisterChanRPC("CloseAgent", rpcCloseAgent)

	skeleton.RegisterChanRPC("Ping", rpcPing)
	skeleton.RegisterChanRPC("UserLogin", rpcUserLogin)

}

func rpcNewAgent(args []interface{}) {
	a := args[0].(gate.Agent)

	p := CreatePlayer()
	p.connAgent = a

	//将玩家本身作为userData附加到agent上，避免后面收到信息再查找玩家
	p.connAgent.SetUserData(p)

	//建立新连接后随即开始心跳
	//p.StartBreathe()
}

func rpcCloseAgent(args []interface{}) {
	a := args[0].(gate.Agent)
	_ = a

}

// 心跳检测
func rpcPing(args []interface{}) {
	a := args[1].(gate.Agent)

	p, ok := a.UserData().(*Player)
	if ok {

		//每次客户端发送心跳，则将客户端心跳置为0
		p.onClientBreathe()

		pingTime := time.Now().UnixNano() / 1e6

		pong := &pb_msg.PongS2C{
			ServerTime: pingTime,
		}
		// 给发送者回应一个 Hello 消息
		a.WriteMsg(pong)
	}
}

func rpcUserLogin(args []interface{}) {
	m := args[0].(*pb_msg.LoginC2S)
	a := args[1].(gate.Agent)
	p, ok := a.UserData().(*Player)
	if ok {
		p.ID = m.GetLoginInfo().GetId()
		PlayerRegister(p.ID, p)
	}

	//查看数据库用户ID是否存在，存在直接数据库返回数据,不存在插入数据在返回
	data, err := FindUserInfoData(m)
	if err != nil {
		fmt.Println("not FindUserInfoData:", err)
		return
	}

	rsp := &pb_msg.LoginResultS2C{}
	rsp.PlayerInfo = new(pb_msg.PlayerInfo)
	rsp.PlayerInfo.Id = data.Id
	rsp.PlayerInfo.Name = data.Name
	rsp.PlayerInfo.HeadImg = data.HeadImg
	rsp.PlayerInfo.Balance = data.Balance

	fmt.Println("rpcUserLogin data ~ :", rsp)
	a.WriteMsg(rsp)

	//判断用户是否断线登录，通过判断用户房间是否为空，不为空，则返回房间信息
	if p.room != nil {
		gr := &GameRoom{}
		gr.PlayerJoin(p)
		p.connAgent.WriteMsg(gr)
	}
}
