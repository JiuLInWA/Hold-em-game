package internal

import (
	"fmt"
	"github.com/name5566/leaf/gate"
	"github.com/name5566/leaf/log"
	"reflect"
	pb_msg "server/msg/Protocal"
)

func init() {
	//向当前模块（game 模块）注册 Test 消息的消息处理函数 handleTest
	//handler(&pb_msg.Test{},handleTest)
	handler(&pb_msg.QuickStartC2S{}, handleQuickStart)
	handler(&pb_msg.CreateRoomC2S{}, handleCreatRoom)
	handler(&pb_msg.JoinRoomC2S{}, handleJoinRoom)
	handler(&pb_msg.ExitRoomC2S{}, handleExitRoom)
}

// 异步处理
func handler(m interface{}, h interface{}) {
	skeleton.RegisterChanRPC(reflect.TypeOf(m), h)
}

func handleQuickStart(args []interface{}) {
	m := args[0].(*pb_msg.QuickStartC2S)
	a := args[1].(gate.Agent)

	p, ok := a.UserData().(*Player)
	log.Debug("handleQuickStart 快速匹配房间 :%v", p.ID)

	if ok {
		r := new(RoomInfo)
		r.RoomId = m.RoomInfo.RoomId
		r.CfgId = m.RoomInfo.CfgId
		r.MaxPlayer = m.RoomInfo.MaxPlayer
		r.ActionTimeS = m.RoomInfo.ActionTimeS
		r.Pwd = m.RoomInfo.Pwd

		//TODO 自己调试设定的余额
		//p.balance = 4000
		gameHall.PlayerQuickStart(p, r)
	}
}

func handleCreatRoom(args []interface{}) {
	m := args[0].(*pb_msg.CreateRoomC2S)
	a := args[1].(gate.Agent)

	p, ok := a.UserData().(*Player)
	log.Debug("handleCreatRoom 用户创建房间 ~ :%v", p.ID)

	if ok {
		r := new(RoomInfo)
		r.RoomId = m.RoomInfo.RoomId
		r.CfgId = m.RoomInfo.CfgId
		r.MaxPlayer = m.RoomInfo.MaxPlayer
		r.ActionTimeS = m.RoomInfo.ActionTimeS
		r.Pwd = m.RoomInfo.Pwd

		gameHall.PlayerCreatRoom(p, r)
	}
}

func handleJoinRoom(args []interface{}) {
	m := args[0].(*pb_msg.JoinRoomC2S)
	a := args[1].(gate.Agent)

	p, ok := a.UserData().(*Player)
	log.Debug("handleJoinRoom 用户加入房间 ~ :%v", p.ID)

	if ok {
		msg := gameHall.PlayerJoinRoom(p, m.RoomId, m.Pwd)
		if msg == 0 {
			log.Debug("PlayerJoinRoom error")
			return
		}
	}
}

func handleExitRoom(args []interface{}) {
	a := args[1].(gate.Agent)

	log.Debug("handleExitRoom 用户退出房间 ~")

	p, ok := a.UserData().(*Player)
	if ok {
		fmt.Println(p.ID, "exit room! ~")
		p.PlayerExitRoom()
	}
}
