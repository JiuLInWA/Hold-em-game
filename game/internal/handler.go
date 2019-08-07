package internal

import (
	"fmt"
	"github.com/name5566/leaf/gate"
	"github.com/name5566/leaf/log"
	"reflect"
	pb_msg "server/msg/Protocal"
	"time"
)

func init() {
	//向当前模块（game 模块）注册 Test 消息的消息处理函数 handleTest
	//handler(&pb_msg.Test{},handleTest)
	handler(&pb_msg.PingC2S{}, handlePing)
	handler(&pb_msg.LoginC2S{}, handleLogin)
	handler(&pb_msg.QuickStartC2S{}, handleQuickStart)
	handler(&pb_msg.CreateRoomC2S{}, handleCreatRoom)
	handler(&pb_msg.JoinRoomC2S{}, handleJoinRoom)

	handler(&pb_msg.ExitRoomC2S{}, handleExitRoom)
	handler(&pb_msg.SitDownC2S{}, handleSitDown)
	handler(&pb_msg.StandUpC2S{}, handleStandUp)
	handler(&pb_msg.PlayerActionC2S{}, handleAction)
}

// 异步处理
func handler(m interface{}, h interface{}) {
	skeleton.RegisterChanRPC(reflect.TypeOf(m), h)
}

//var ch chan *Player

func handlePing(args []interface{}) {
	m := args[0].(*pb_msg.PingC2S)
	a := args[1].(gate.Agent)

	log.Debug("Hello : %v", m)

	p, ok := a.UserData().(*Player)
	if ok {
		//ch = make(chan *Player)
		//ch <- p

		p.onClientBreathe()
		//fmt.Println("Ping~~~ id", p.ID, "------------", p.uClientDelay)

		pingTime := time.Now().UnixNano() / 1e6

		pong := &pb_msg.PongS2C{
			ServerTime: &pingTime,
		}
		// 给发送者回应一个 Hello 消息
		a.WriteMsg(pong)
	}
}

func handleLogin(args []interface{}) {
	m := args[0].(*pb_msg.LoginC2S)
	a := args[1].(gate.Agent)

	p, ok := a.UserData().(*Player)
	log.Debug("handleLogin 用户成功登陆~ :%v", p.ID)

	if ok {
		p.ID = m.GetLoginInfo().GetId()
		p.name = m.GetLoginInfo().GetId()
		PlayerRegister(p.ID, p)
	}
	//查看数据库用户ID是否存在，存在直接数据库返回数据,不存在插入数据在返回
	data, err := FindUserInfo(m)
	if err != nil {
		fmt.Println("not FindUserInfoData:", err)
		return
	}

	rsp := &pb_msg.LoginResultS2C{}
	rsp.PlayerInfo = new(pb_msg.PlayerInfo)
	rsp.PlayerInfo.Id = &data.Id
	rsp.PlayerInfo.Name = &data.Name
	rsp.PlayerInfo.Face = &data.Face
	rsp.PlayerInfo.Balance = &data.Balance

	fmt.Println("UserLogin data ~ :", rsp)
	a.WriteMsg(rsp)

	player := p.GetUserRoomInfo()
	if player != nil {
		//将新的用户链接赋给旧的用户链接，再将旧的用户数据在塞到链接上
		p = player
		p.IsOnLine = true
		p.connAgent = a
		p.connAgent.SetUserData(p)
		enter := p.RspEnterRoom()
		p.connAgent.WriteMsg(enter)
	}
}

func handleQuickStart(args []interface{}) {
	m := args[0].(*pb_msg.QuickStartC2S)
	a := args[1].(gate.Agent)

	p, ok := a.UserData().(*Player)
	log.Debug("handleQuickStart 快速匹配房间~ :%v", p.ID)

	if ok {
		r := new(RoomInfo)
		r.CfgId = *m.RoomInfo.CfgId
		r.MaxPlayer = *m.RoomInfo.MaxPlayer
		r.ActionTimeS = *m.RoomInfo.ActionTimeS

		gameHall.PlayerQuickStart(p, r)
	}
}

func handleCreatRoom(args []interface{}) {
	m := args[0].(*pb_msg.CreateRoomC2S)
	a := args[1].(gate.Agent)

	p, ok := a.UserData().(*Player)
	log.Debug("handleCreatRoom 用户创建房间~ :%v", p.ID)

	if ok {
		r := new(RoomInfo)
		r.CfgId = *m.RoomInfo.CfgId
		r.MaxPlayer = *m.RoomInfo.MaxPlayer
		r.ActionTimeS = *m.RoomInfo.ActionTimeS
		r.Pwd = *m.RoomInfo.Pwd

		gameHall.PlayerCreatRoom(p, r)
	}
}

func handleJoinRoom(args []interface{}) {
	m := args[0].(*pb_msg.JoinRoomC2S)
	a := args[1].(gate.Agent)

	p, ok := a.UserData().(*Player)
	log.Debug("handleJoinRoom 用户加入房间~ :%v", p.ID)

	if ok {
		gameHall.PlayerJoinRoom(p, *m.RoomId, *m.Pwd)
	}
}

func handleExitRoom(args []interface{}) {
	a := args[1].(gate.Agent)

	log.Debug("handleExitRoom 用户退出房间~ ")

	p, ok := a.UserData().(*Player)
	if ok {
		p.PlayerExitRoom()
	}
}

func handleSitDown(args []interface{}) {
	m := args[0].(*pb_msg.SitDownC2S)
	a := args[1].(gate.Agent)

	p, ok := a.UserData().(*Player)
	log.Debug("handleSitDown 玩家坐下座位~ ")

	if ok {
		p.SitDownTable(*m.Position)
	}
}

func handleStandUp(args []interface{}) {
	a := args[1].(gate.Agent)

	p, ok := a.UserData().(*Player)
	log.Debug("handleStandUp 玩家站起观战~ ")

	if ok {
		p.StandUpBattle()
	}
}

func handleAction(args []interface{}) {
	m := args[0].(*pb_msg.PlayerActionC2S)
	a := args[1].(gate.Agent)

	p, ok := a.UserData().(*Player)
	log.Debug("handleAction 玩家开始行动~ :%v", p.ID)
	log.Debug("handleAction 玩家开始行动~ :%v %v", m.BetAmount, m.Action)

	if ok {
		amount := float64(*m.BetAmount)
		//GetPlayerAction(m)
		p.room.preChips = amount
		//p.actions = make()
		p.actions <- *m.Action
	}
}


//	a := args[1].(gate.Agent)
//	p, ok := a.UserData().(*Player)
//	if ok {
//		switch t := args[0].(type) {
//		case *pb_msg.QuickStartC2S:
//			r := new(RoomInfo)
//			r.CfgId = *t.RoomInfo.CfgId
//			r.MaxPlayer = *t.RoomInfo.MaxPlayer
//			r.ActionTimeS = *t.RoomInfo.ActionTimeS
//			gameHall.PlayerQuickStart(p, r)
//		case *pb_msg.CreateRoomC2S:
//			r := new(RoomInfo)
//			r.CfgId = *t.RoomInfo.CfgId
//			r.MaxPlayer = *t.RoomInfo.MaxPlayer
//			r.ActionTimeS = *t.RoomInfo.ActionTimeS
//			r.Pwd = *t.RoomInfo.Pwd
//			gameHall.PlayerCreatRoom(p, r)
//		case *pb_msg.JoinRoomC2S:
//			gameHall.PlayerJoinRoom(p, *t.RoomId, *t.Pwd)
//		default:
//			log.Error("GameHall 事件无法识别", t)
//		}
//	}
//}
//
//func handleRoomEvent(args []interface{}) {
//	a := args[1].(gate.Agent)
//	p, ok := a.UserData().(*Player)
//	if ok {
//		switch t := args[0].(type) {
//		case *pb_msg.ExitRoomC2S:
//			p.PlayerExitRoom()
//		case *pb_msg.SitDownC2S:
//			p.SitDownTable(*t.Position)
//		case *pb_msg.StandUpC2S:
//			p.StandUpBattle()
//		case *pb_msg.PlayerActionC2S:
//			p.room.handleRoomEvent(a, p, args[0])
//		default:
//			log.Error("GameHall 事件无法识别", t)
//		}
//	}
//}
