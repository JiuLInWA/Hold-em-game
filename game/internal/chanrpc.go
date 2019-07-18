package internal

import (
	"fmt"
	"github.com/name5566/leaf/gate"
	"github.com/name5566/leaf/log"
)

func init() {
	skeleton.RegisterChanRPC("NewAgent", rpcNewAgent)
	skeleton.RegisterChanRPC("CloseAgent", rpcCloseAgent)

}

func rpcNewAgent(args []interface{}) {
	a := args[0].(gate.Agent)

	fmt.Println("rpcNewAgent ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~")
	p := CreatePlayer()
	p.connAgent = a

	//将玩家本身作为userData附加到agent上，避免后面收到信息再查找玩家
	p.connAgent.SetUserData(p)

	//开始呼吸
	//p.StartBreathe()

}

func rpcCloseAgent(args []interface{}) {
	a := args[0].(gate.Agent)
	//断开连接，删除用户信息，用户链接设为空
	p, ok := a.UserData().(*Player)
	if ok {
		log.Debug("Player close websocket ~~~ : %v", p.ID)
		DeletePlayer(p)
	}
	a.SetUserData(nil)
}
