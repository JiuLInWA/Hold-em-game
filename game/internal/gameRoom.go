package internal

import (
	"fmt"
	"math/rand"
	pb_msg "server/msg/Protocal"
	"time"
)

//RoomStat 表示房间状态
type RoomStat uint8

const (
	emRoomStateNone  RoomStat = 0 //房间初始状态
	emRoomStateRun   RoomStat = 1 //房间正在运行
	emRoomStateClose RoomStat = 2 //房间已关闭并摧毁
)

type RoomInfo struct {
	RoomId      string                  //房间ID
	CfgId       string                  //房间配置信息
	MaxPlayer   int32                   //房间最大玩家数
	ActionTimeS pb_msg.Enum_ActionTimeS //玩家行动时间
	Pwd         string                  //房间密码
}

type GameRoom struct {
	roomInfo     *RoomInfo
	PlayerList   []*Player //玩家列表
	PlayerStdUp  []*Player //站起的玩家
	PlayerSitDw  []*Player //自动坐下的玩家
	curPlayerNum int32     //房间当前玩家数

	isStepEnd      bool                 //是否本轮结束(将玩家筹码飞到注池)
	gameStep       pb_msg.Enum_GameStep //当前游戏阶段状态
	minRaise       float64              //加注最小值（本轮水位）
	activePos      int32                //当前正在行动的玩家座位号
	nextStepTs     int64                //下一个阶段的时间戳
	pot            float64              //赌注池当前总共有多少钱
	publicCardKeys []int32              //桌面公牌

	//房间状态
	Status RoomStat

	Pots []uint32 //奖池筹码数
	SB   uint32   //小盲注
	BB   uint32   //大盲注
}

//Init 房间初始化
func (gr *GameRoom) Init(p *Player, r *RoomInfo) {

	ri := new(RoomInfo)
	gr.roomInfo = ri

	roomId := fmt.Sprintf("%06v", rand.New(rand.NewSource(time.Now().UnixNano())).Int31n(1000000))
	gr.roomInfo.RoomId = roomId
	gr.roomInfo.CfgId = r.CfgId
	gr.roomInfo.MaxPlayer = r.MaxPlayer
	gr.roomInfo.ActionTimeS = r.ActionTimeS
	gr.roomInfo.Pwd = r.Pwd

	gr.PlayerList = make([]*Player, r.MaxPlayer)
	for i := 0; i < len(gr.PlayerList); i++ {
		gr.PlayerList[i] = nil
	}

	gr.curPlayerNum = 0

	gr.gameStep = pb_msg.Enum_GameStep_STEP_WAITING
	gr.pot = 0

	gr.Status = emRoomStateNone

	cd := CfgDataHandle(r.CfgId)
	gr.SB = uint32(cd.Bb / 2)
	gr.BB = uint32(cd.Bb)

}

//Broadcast 广播消息
func (gr *GameRoom) Broadcast(msg interface{}) {
	for _, p := range gr.PlayerList {
		if p != nil {
			p.SendMsg(msg)
		}
	}
}

//CanJoin 房间是否还能加入~返回坐位号
func (gr *GameRoom) IsCanJoin() bool {
	fmt.Println("check CanJoin~~~!!")
	fmt.Println("当前房间人数~ : ", gr.curPlayerNum)
	fmt.Println("房间限定人数~ : ", gr.roomInfo.MaxPlayer)
	return gr.curPlayerNum < gr.roomInfo.MaxPlayer
}

//FindAbleChair 寻找一个空位置
func (gr *GameRoom) FindAbleChair() uint8 {
	for chair, p := range gr.PlayerList {
		if p == nil {
			return uint8(chair)
		}
	}
	panic("The GameRoom make a logic error,don't find able chair, Should check canJoin first please")
}

//IsRoomActionTimes 房间是否同玩家速度
func (gr *GameRoom) IsRoomActionTimes(actionTimes int32) bool {
	fmt.Println("111~ :", gr.roomInfo.ActionTimeS)

	return int32(gr.roomInfo.ActionTimeS) == actionTimes
}

//IsPlayerMaxNum 房间限定人数是否与用户限定人数一致
func (gr *GameRoom) IsPlayerMaxNum(maxPlayerNum int32) bool {
	return gr.roomInfo.MaxPlayer == maxPlayerNum
}

//Running 房间运行
func (gr *GameRoom) Running() {
	fmt.Println("this room is Running! ~")
	gr.Status = emRoomStateRun

}

//PlayerJoin 玩家加入房间
func (gr *GameRoom) PlayerJoin(p *Player) uint8 {
	fmt.Println("Player Join Room ~")
	gr.curPlayerNum++
	p.chair = gr.FindAbleChair()
	gr.PlayerList[p.chair] = p

	if gr.Status != emRoomStateRun {
		// RUN
		gr.Running()
	} else {
		// 如果已经在Running，说明还有其他玩家，玩家入场广播消息给其他玩家
		data := &pb_msg.LoginResultS2C{}
		data.PlayerInfo = new(pb_msg.PlayerInfo)
		data.PlayerInfo.Id = p.ID
		data.PlayerInfo.Name = p.name
		data.PlayerInfo.HeadImg = p.headImg
		data.PlayerInfo.Balance = p.balance
		gr.Broadcast(data)
	}

	//需要返回玩家自己本身消息，和同房间其他玩家基础信息
	er := &pb_msg.EnterRoomS2C{}
	er.RoomData = new(pb_msg.RoomData)
	er.RoomData.RoomInfo = new(pb_msg.RoomInfo)
	er.RoomData.RoomInfo.RoomId = gr.roomInfo.RoomId
	er.RoomData.RoomInfo.CfgId = gr.roomInfo.CfgId
	er.RoomData.RoomInfo.MaxPlayer = gr.roomInfo.MaxPlayer
	er.RoomData.RoomInfo.ActionTimeS = gr.roomInfo.ActionTimeS
	er.RoomData.RoomInfo.Pwd = gr.roomInfo.Pwd
	er.RoomData.IsStepEnd = gr.isStepEnd
	er.RoomData.GameStep = gr.gameStep
	er.RoomData.MinRaise = gr.minRaise
	er.RoomData.ActivePos = gr.activePos
	er.RoomData.NextStepTs = gr.nextStepTs
	er.RoomData.Pot = gr.pot
	er.RoomData.PublicCardKeys = gr.publicCardKeys
	for _, players := range gr.PlayerList {
		if players != nil {
			data := &pb_msg.PlayerData{}
			data.PlayerInfo = new(pb_msg.PlayerInfo)
			data.PlayerInfo.Id = p.ID
			data.PlayerInfo.Name = p.name
			data.PlayerInfo.HeadImg = p.headImg
			data.PlayerInfo.Balance = p.balance
			data.Position = int32(p.chair)
			data.IsRaised = p.IsRaised
			data.PlayerStatus = p.playerStatus
			data.DropedBets = p.dropedBets
			data.DropedBetsSum = p.dropedBetsSum
			data.CardKeys = p.cardKeys
			data.CardSuitData = new(pb_msg.CardSuitData)
			p.cardSuitData = new(CardSuitData)
			data.CardSuitData.HandCardKeys = p.cardSuitData.HandCardKeys
			data.CardSuitData.PublicCardKeys = p.cardSuitData.PublicCardKeys
			data.CardSuitData.SuitPattern = p.cardSuitData.SuitPattern
			data.IsWinner = p.isWinner
			data.Blind = p.blind
			data.IsButton = p.isButton
			data.IsAllIn = p.isAllIn
			data.IsSelf = p.isSelf
			data.ResultMoney = p.resultMoney
			er.RoomData.PlayerDatas = append(er.RoomData.PlayerDatas, data)
		}
	}

	p.connAgent.WriteMsg(er)
	fmt.Println("this room data ~ :", er)

	return p.chair
}

//ExitFromRoom 玩家从房间退出
func (gr *GameRoom) ExitFromRoom(p *Player) {
	gr.curPlayerNum--
	fmt.Println("ExitFromRoom curPlayerNum ~ :", gr.curPlayerNum)
	gr.PlayerList[p.chair] = nil
	if gr.curPlayerNum == 0 {
		fmt.Println("Room PlayerNum is 0,so delete this room! ~ ")
		gr.Status = emRoomStateClose

		gh := GameHall{}
		gh.DeleteRoom(p.room.roomInfo.RoomId)
	} else {
		//给其他玩家广播该用户已下线！
		gr.Broadcast(p.ID)
	}

	data := &pb_msg.ExitRoomS2C{}
	data.PlayerInfo = new(pb_msg.PlayerInfo)
	data.PlayerInfo.Id = p.ID
	data.PlayerInfo.Name = p.name
	data.PlayerInfo.HeadImg = p.headImg
	data.PlayerInfo.Balance = p.balance

	p.connAgent.WriteMsg(data)
}
