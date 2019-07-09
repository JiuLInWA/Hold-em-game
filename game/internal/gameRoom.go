package internal

import (
	"fmt"
	"github.com/name5566/leaf/log"
	"math/rand"
	pb_msg "server/msg/Protocal"
	"time"
)

//RoomStat 表示房间状态
type RoomStat uint8

const (
	emRoomStateNone RoomStat = 0 //房间初始状态
	emRoomStateRun  RoomStat = 1 //房间正在运行
	emRoomStateOver RoomStat = 2 //房间结束游戏
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
	activePos      uint32               //当前正在行动的玩家座位号
	nextStepTs     int64                //下一个阶段的时间戳
	pot            float64              //赌注池当前总共有多少钱
	publicCardKeys []int32              //桌面公牌

	//房间状态
	Status RoomStat

	Pots   []uint32 //奖池筹码数
	Button uint32   //庄家座位号
	SB     uint32   //小盲注
	BB     uint32   //大盲注
}

//Init 房间初始化
func (gr *GameRoom) Init(r *RoomInfo) {

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

	gr.Button = 0
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

//IsRoomPwd 房间密码是否一致
func (gr *GameRoom) IsRoomPwd(pwd string) bool {
	fmt.Println("房间 pwd为~ :", gr.roomInfo.Pwd)
	fmt.Println("用户 pwd为~ :", pwd)
	return gr.roomInfo.Pwd == pwd
}

//FindAbleChair 寻找一个空位置
func (gr *GameRoom) FindAbleChair() uint32 {
	for chair, p := range gr.PlayerList {
		if p == nil {
			fmt.Println("座位号下标为~ :", uint32(chair))
			return uint32(chair)
		}
	}
	panic("The GameRoom make a logic error,don't find able chair, Should check canJoin first please")
}

//IsRoomActionTimes 房间是否同玩家速度
func (gr *GameRoom) IsRoomActionTimes(actionTimes int32) bool {
	return int32(gr.roomInfo.ActionTimeS) == actionTimes
}

//IsPlayerMaxNum 房间限定人数是否与用户限定人数一致
func (gr *GameRoom) IsPlayerMaxNum(maxPlayerNum int32) bool {
	return gr.roomInfo.MaxPlayer == maxPlayerNum
}

//PlayerLen 记录当前房间人数
func (gr *GameRoom) PlayerLen() int32 {
	var num int32
	for _, v := range gr.PlayerList {
		if v != nil {
			num++
		}
	}
	return num
}

//RoomMaxPlayer 房间最大人数
func (gr *GameRoom) RoomMaxPlayer() int32 {
	return gr.roomInfo.MaxPlayer
}

//DragInRoomChips 玩家带入筹码
func (gr *GameRoom) DragInRoomChips(p *Player) float64 {
	dataCfg := CfgDataHandle(gr.roomInfo.CfgId)

	//1、如果玩家余额 大于房间最大设定金额 MaxTakeIn，则带入金额就设为 房间最大设定金额
	//2、如果玩家余额 小于房间最大设定金额 MaxTakeIn，则带入金额就设为 玩家的所有余额
	if p.balance > float64(dataCfg.MaxTakeIn) {
		p.balance = p.balance - float64(dataCfg.MaxTakeIn)
		return float64(dataCfg.MaxTakeIn)
	}

	Balance := p.balance
	p.balance = p.balance - p.balance

	return Balance
}

//KickPlayer 踢出筹码小与大盲的玩家
func (gr *GameRoom) KickPlayer() {
	for _, v := range gr.PlayerList {
		if v != nil {
			if v.chips < float64(gr.BB) {
				msg := pb_msg.SvrMsgS2C{}
				msg.Code = RECODE_NOTCHIPS
				msg.TipType = pb_msg.Enum_SvrTipType_WARN
				v.connAgent.WriteMsg(&msg)

				log.Debug("玩家带入筹码已不足~")
				v.PlayerExitRoom()
			}
		}
	}
}

//Banker 庄家
func (gr *GameRoom) Banker(pos uint32, f func(p *Player) bool) {

	//max := gr.RoomMaxPlayer()
	//start := int(pos+1) % int(max)
	//
	//for i := start; i < int(max); i = (i + 1) % int(max) {
	//	if gr.PlayerList[i] != nil && gr.PlayerList[pos] != gr.PlayerList[i] {
	//		return gr.PlayerList[i]
	//	}
	//}
	//return nil
	volume := gr.RoomMaxPlayer()
	end := int((volume + int32(pos) - 1) % volume)
	i := int(pos)
	for ; i != end; i = (i + 1) % int(volume) {
		if gr.PlayerList[i] != nil && !f(gr.PlayerList[i]) {
			return
		}
	}
	// end
	if gr.PlayerList[i] != nil {
		f(gr.PlayerList[i])
	}
}

//Blind 小盲注和大盲注
func (gr *GameRoom) Blind(pos uint32) *Player {

	max := gr.RoomMaxPlayer()
	start := int(pos+1) % int(max)
	for i := start; i < int(max); i = (i + 1) % int(max) {
		if gr.PlayerList[i] != nil && gr.PlayerList[pos] != gr.PlayerList[i] {
			return gr.PlayerList[i]
		}
	}
	return nil
}

//betting 小大盲下注
func (gr *GameRoom) betting(p *Player, blind float64) {
	//当前行动玩家
	gr.activePos = p.chair
	//玩家筹码变动
	p.chips = p.chips - blind
	//本轮玩家下注额
	p.dropedBets = blind
	//玩家本局总下注额
	p.dropedBetsSum = p.dropedBetsSum + blind
	//总筹码变动
	gr.pot = gr.pot + blind

	//TODO 广播发送玩家盲注金额
	msg := gr.RspEnterRoom(p)
	gr.Broadcast(msg)

}

//Running 房间运行
func (gr *GameRoom) Running() {

	//踢掉筹码小与大盲的玩家
	gr.KickPlayer()

	n := gr.PlayerLen()
	fmt.Println("Running 当前房间玩家人数为 ~ :", n)

	//当前房间人数存在2人及2人以上才开始游戏
	if n >= 2 {
		log.Debug("this room is Running! ~")

		gr.pot = 0
		gr.minRaise = 0
		gr.publicCardKeys = []int32{}
		gr.Pots = []uint32{}

		gr.Status = emRoomStateRun
		gr.gameStep = pb_msg.Enum_GameStep_STEP_WAITING

		//1、产生庄家
		var dealer *Player
		button := gr.Button - 1
		gr.Banker((button+1)%uint32(gr.RoomMaxPlayer()), func(p *Player) bool {
			gr.Button = p.chair
			dealer = p
			return false
		})
		dealer.isButton = true
		dealer.isSelf = true

		//获取庄家数据，进行广播，因为重新开始会有多名玩家
		enter := gr.RspEnterRoom(dealer)
		gr.Broadcast(enter)
		log.Debug("庄家的座位号为 : %v", dealer.chair)

		//2、洗牌

		//3、产生小盲
		sb := gr.Blind(dealer.chair)
		sb.blind = pb_msg.Enum_Blind_SMALL_BLIND
		log.Debug("小盲注座位号为 : %v", sb.chair)
		//4、小盲注下注
		gr.betting(sb, float64(gr.SB))

		//5、产生大盲
		if n >= 3 {
			bb := gr.Blind(sb.chair)
			sb.blind = pb_msg.Enum_Blind_BIG_BLIND
			log.Debug("大盲注座位号为 : %v", bb.chair)
			//6、大盲注下注
			gr.betting(bb, float64(gr.BB))
		}

		// Round 1：preFlop 看手牌,下盲注
		gr.gameStep = pb_msg.Enum_GameStep_STEP_PRE_FLOP

		// Round 2：Flop 翻牌圈,牌桌上发3张公牌
		gr.gameStep = pb_msg.Enum_GameStep_STEP_FLOP

		// Round 3：Turn 转牌圈,牌桌上发第4张公共牌
		gr.gameStep = pb_msg.Enum_GameStep_STEP_TURN

		// Round 4：River 河牌圈,牌桌上发第5张公共牌
		gr.gameStep = pb_msg.Enum_GameStep_STEP_RIVER

		// showdown 摊开底牌,开牌比大小
		gr.gameStep = pb_msg.Enum_GameStep_STEP_SHOW_DOWN

		//6、游戏结束，停留5秒，重新开始游戏
		gr.Status = emRoomStateOver

	} else {
		return
	}
}

//RspEnterRoom 返回客户端房间数据
func (gr *GameRoom) RspEnterRoom(p *Player) *pb_msg.EnterRoomS2C {

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
	er.RoomData.ActivePos = int32(gr.activePos)
	er.RoomData.NextStepTs = gr.nextStepTs
	er.RoomData.Pot = gr.pot
	er.RoomData.PublicCardKeys = gr.publicCardKeys
	for _, players := range gr.PlayerList {
		if gr.PlayerList[p.chair] == players {
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
	}
	return er
}

//PlayerJoin 玩家加入房间
func (gr *GameRoom) PlayerJoin(p *Player) uint8 {

	// 玩家带入筹码
	p.chips = gr.DragInRoomChips(p)
	log.Debug("玩家带入筹码 : %v", p.chips)
	log.Debug("玩家剩余金额 : %v", p.balance)

	fmt.Println("Player Join Room ~")
	gr.curPlayerNum++
	p.chair = gr.FindAbleChair()
	gr.PlayerList[p.chair] = p

	p.room = gr

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

	// 返回前端房间信息
	roomData := gr.RspEnterRoom(p)
	p.connAgent.WriteMsg(roomData)
	fmt.Println("this room data ~ :", roomData)

	return uint8(p.chair)
}

//ExitFromRoom 玩家从房间退出
func (gr *GameRoom) ExitFromRoom(p *Player) {
	gr.curPlayerNum--
	fmt.Println("ExitFromRoom curPlayerNum ~ :", gr.curPlayerNum)

	//玩家退出房间, 将剩余的筹码转换为玩家金额
	p.balance = p.chips
	p.chips = 0
	log.Debug("玩家剩余余额 : %v", p.balance)

	gr.PlayerList[p.chair] = nil
	if gr.curPlayerNum == 0 {
		fmt.Println("Room PlayerNum is 0，so delete this room! ~ ")

		gameHall.DeleteRoom(p.room.roomInfo.RoomId)
	} else {
		//给其他玩家广播该用户已下线！
		data := &pb_msg.LoginResultS2C{}
		data.PlayerInfo = new(pb_msg.PlayerInfo)
		data.PlayerInfo.Id = p.ID
		data.PlayerInfo.Name = p.name
		data.PlayerInfo.HeadImg = p.headImg
		data.PlayerInfo.Balance = p.balance

		gr.Broadcast(data)
	}

	data := &pb_msg.ExitRoomS2C{}
	data.PlayerInfo = new(pb_msg.PlayerInfo)
	data.PlayerInfo.Id = p.ID
	data.PlayerInfo.Name = p.name
	data.PlayerInfo.HeadImg = p.headImg
	data.PlayerInfo.Balance = p.balance

	p.connAgent.WriteMsg(data)
	fmt.Println("ExitRoom data :", data)
}
