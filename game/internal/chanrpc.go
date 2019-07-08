package internal

import (
	"fmt"
	"github.com/google/uuid"
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

var ch chan int

var ConnPlayerMap = make(map[string]*Player, 0)

func rpcNewAgent(args []interface{}) {
	a := args[0].(gate.Agent)

	fmt.Println("rpcNewAgent ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~")
	p := CreatePlayer()
	p.connAgent = a

	//将玩家本身作为userData附加到agent上，避免后面收到信息再查找玩家
	p.connAgent.SetUserData(p)

	//这个全局，不用，此处面对连接
	uuidStr := uuid.New().String()
	p.ConnId = uuidStr
	ConnPlayerMap[uuidStr] = p
	StartBreathe(uuidStr)

	//ch = make(chan int)
	//
	//go func(chan int, string) {
	//	lastTime := time.Now().Unix()
	//	pp := ConnPlayerMap[p.ConnId]
	//	for {
	//		var (
	//			x  int
	//			ok bool
	//		)
	//		select {
	//		case x, ok = <-ch:
	//			if !ok {
	//				fmt.Println("通道关闭了")
	//				return
	//			}
	//			fmt.Println("pong ~ :", x)
	//			//每次心跳更新最新当前时间
	//			lastTime = time.Now().Unix()
	//			break
	//		default:
	//			curTime := time.Now().Unix()
	//			if curTime-lastTime > 10 {
	//				fmt.Println("超时时间为~ :", curTime-lastTime, "秒")
	//				close(ch)
	//				//断开连接,删除用户信息
	//				pp.PlayerExitRoom()
	//				pp.connAgent.Destroy()
	//				DeletePlayer(pp)
	//				return
	//			}
	//		}
	//	}
	//}(ch, p.ConnId)

}

func rpcCloseAgent(args []interface{}) {
	a := args[0].(gate.Agent)
	_ = a

}

// 心跳检测
func rpcPing(args []interface{}) {
	a := args[1].(gate.Agent)

	//defer func() {
	//	if recover() == nil {
	//		return
	//	}
	//}()
	//timerNow := time.Now().Unix()
	//ch <- int(timerNow)
	//
	//fmt.Println("ping ~ :", timerNow)

	p, ok := a.UserData().(*Player)
	fmt.Println("进来一个ping", p.ConnId)
	ConnPlayerMap[p.ConnId].uClientDelay = 0
	fmt.Println(fmt.Sprintf("我是Ping:::ConnId %v 次数 %v ", p.ConnId, p.uClientDelay))
	if ok {
		//p.onClientBreathe()
		//fmt.Println("PING uClientDelay :", p.uClientDelay)

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
	//TODO 占时测试用~ 这样遍历所有房间，速度会变慢
	gameHall.GetPlayerRoomInfo(p)

}
