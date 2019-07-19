package internal

import (
	"fmt"
	"github.com/name5566/leaf/log"
	"math/rand"
	"server/game/algorithm"
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
	AllPlayer    []*Player //站起来的玩家
	PlayerList   []*Player //座位玩家列表
	curPlayerNum int32     //房间当前玩家数

	Cards          algorithm.Cards      //公共牌
	isStepEnd      bool                 //是否本轮结束(将玩家筹码飞到注池)
	gameStep       pb_msg.Enum_GameStep //当前游戏阶段状态
	minRaise       float64              //加注最小值（本轮水位）
	activePos      int32                //当前正在行动的玩家座位号
	nextStepTs     int64                //下一个阶段的时间戳
	pot            float64              //赌注池当前总共有多少钱
	publicCardKeys []int32              //桌面公牌

	//房间状态
	Status RoomStat

	Timeout time.Duration

	preChips float64 //上一个玩家的下注金额
	remain   int32   //记录每个阶段玩家的下注的数量
	allin    int32   //allin玩家的数量
	Chips    []int32 //所有玩家本局下的总筹码,对应player玩家
	Pots     []int32 //奖池筹码数,第一项为主池，其他项(若存在)为边池
	Button   uint32  //庄家座位号
	SB       uint32  //小盲注
	BB       uint32  //大盲注
}

//Init 房间初始化
func (gr *GameRoom) Init(r *RoomInfo) {

	gr.roomInfo = new(RoomInfo)
	roomId := fmt.Sprintf("%06v", rand.New(rand.NewSource(time.Now().UnixNano())).Int31n(1000000))
	gr.roomInfo.RoomId = roomId
	gr.roomInfo.CfgId = r.CfgId
	gr.roomInfo.MaxPlayer = r.MaxPlayer
	gr.roomInfo.ActionTimeS = r.ActionTimeS
	gr.roomInfo.Pwd = r.Pwd

	gr.AllPlayer = nil
	gr.PlayerList = make([]*Player, r.MaxPlayer)
	for i := 0; i < len(gr.PlayerList); i++ {
		gr.PlayerList[i] = nil
	}
	gr.curPlayerNum = 0

	gr.Cards = nil
	gr.gameStep = pb_msg.Enum_GameStep_STEP_WAITING
	gr.activePos = -1
	gr.pot = 0
	gr.publicCardKeys = []int32{}

	gr.Status = emRoomStateNone
	gr.remain = 0
	gr.allin = 0
	gr.Pots = []int32{}

	gr.Button = 0
	cd := CfgDataHandle(r.CfgId)
	gr.SB = uint32(cd.Bb / 2)
	gr.BB = uint32(cd.Bb)

	gr.minRaise = float64(gr.BB)
}

//Broadcast 广播消息
func (gr *GameRoom) Broadcast(msg interface{}) {
	for _, p := range gr.AllPlayer {
		if p != nil {
			p.SendMsg(msg)
		}
	}
}

//CanJoin 房间是否还能加入~返回坐位号
func (gr *GameRoom) IsCanJoin() bool {
	log.Debug("当前房间人数~ :%v ", gr.curPlayerNum)
	log.Debug("房间限定人数~ :%v ", gr.roomInfo.MaxPlayer)
	return gr.curPlayerNum < gr.roomInfo.MaxPlayer
}

//IsRoomPwd 房间密码是否一致
func (gr *GameRoom) IsRoomPwd(pwd string) bool {
	//log.Debug("房间的密码为~ :%v ", gr.roomInfo.Pwd)
	//log.Debug("用户的密码为~ :%v ", pwd)
	return gr.roomInfo.Pwd == pwd
}

//FindAbleChair 寻找一个空位置
func (gr *GameRoom) FindAbleChair() int32 {
	for chair, p := range gr.PlayerList {
		if p == nil {
			log.Debug("座位号下标为~ :%v", uint32(chair))
			return int32(chair)
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

	err := p.update()
	if err != nil {
		log.Error("DragInRoomChips更新失败 ~ :", err)
	}
	return Balance
}

//KickPlayer 踢出筹码小与大盲的玩家
func (gr *GameRoom) KickPlayer() {
	for _, v := range gr.PlayerList {
		if v != nil {
			if v.chips < float64(gr.BB) {
				v.SendConfigMsg(RECODE_NOTCHIPS, data, pb_msg.Enum_SvrTipType_WARN)
				log.Debug("玩家带入筹码已不足~")
				v.PlayerExitRoom()
			}
		}
	}
}

//Banker 庄家
func (gr *GameRoom) Each(pos uint32, f func(p *Player) bool) {
	//房间最大限定人数
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
func (gr *GameRoom) Blind(pos int32) *Player {

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
	fmt.Println("总筹码变动：", gr.pot)

	//广播发送玩家盲注金额
	msg := p.RspEnterRoom()
	gr.Broadcast(msg)

}

//
func (gr *GameRoom) getTimer(p *Player) {
	//timeout := time.NewTimer(time.Second * 15)

}

//readyPlay 准备阶段
func (gr *GameRoom) readyPlay() {
	gr.Each(0, func(p *Player) bool {
		//记录当前阶段玩家的数量
		gr.remain++
		log.Debug("gr.remain++ :%v", gr.remain)
		return true
	})
}

//action 玩家行动
func (gr *GameRoom) action(pos uint32) {

	//从庄家的下家开始下注
	if pos == 0 {
		pos = gr.Button%uint32(gr.RoomMaxPlayer()) + 1
	}

	gr.Each(pos, func(p *Player) bool {
		//3、行动玩家是根据庄家的下一位玩家
		gr.activePos = p.chair
		e := p.RspEnterRoom()
		action := &pb_msg.ActionPlayerChangedS2C{}
		action.RoomData = e.RoomData
		//todo 广播还是指定用户发送
		p.connAgent.WriteMsg(action)
		log.Debug("行动玩家 ~ :%v", gr.activePos)

		//1、设置每个玩家下注时间
		//2、每个玩家下注状态
		switch p.action {
		case pb_msg.Enum_ActionOptions_ACT_FOLD:
			p.playerStatus = pb_msg.Enum_PlayerStatus_STATUS_FOLD
		case pb_msg.Enum_ActionOptions_ACT_CALL:
			p.playerStatus = pb_msg.Enum_PlayerStatus_STATUS_CALL
			//if p.room.preChips

		case pb_msg.Enum_ActionOptions_ACT_RAISE:
			p.playerStatus = pb_msg.Enum_PlayerStatus_STATUS_RAISE
		case pb_msg.Enum_ActionOptions_ACT_CHECK:
			p.playerStatus = pb_msg.Enum_PlayerStatus_STATUS_CHECK
		}
		//3、下注状态结束则停止时间，进行下一个玩家下注
		//4、下注时间超时，还未下注则直接弃牌
		//5、每个玩家的下注金额根据上个用户金额选择下注
		//6、如果玩家弃牌则改变玩家状态
		//6、将玩家的下注金额添加到奖金池
		//7、玩家全部Allin则直接跳到摊牌
		//8、玩家全部弃牌，则最后一个直接获取奖金池
		gr.getTimer(p)
		return true
	})

}

//Running 房间运行
func (gr *GameRoom) Running() {

	//踢掉筹码小与大盲的玩家
	gr.KickPlayer()

	n := gr.PlayerLen()
	log.Debug("Running 当前房间玩家人数为 ~ :%v", n)

	//当前房间人数存在2人及2人以上才开始游戏
	if n < 2 {
		return
	}

	log.Debug("this room is Running! ~")

	gr.pot = 0
	gr.minRaise = 0
	gr.publicCardKeys = []int32{}
	gr.Pots = []int32{}

	gr.remain = 0
	gr.allin = 0

	gr.Status = emRoomStateRun
	gr.gameStep = pb_msg.Enum_GameStep_STEP_WAITING

	//1、产生庄家
	var dealer *Player
	button := gr.Button - 1
	gr.Each((button+1)%uint32(gr.RoomMaxPlayer()), func(p *Player) bool {
		gr.Button = uint32(p.chair)
		dealer = p
		return false
	})
	dealer.isButton = true

	//获取庄家数据，进行广播，因为重新开始会有多名玩家
	enter := dealer.RspEnterRoom()
	gr.Broadcast(enter)
	log.Debug("庄家的座位号为 : %v", dealer.chair)

	//2、洗牌
	gr.Cards.Shuffle()

	//3、产生小盲
	sb := gr.Blind(dealer.chair)
	sb.blind = pb_msg.Enum_Blind_SMALL_BLIND
	log.Debug("小盲注座位号为 : %v", sb.chair)
	//4、小盲注下注
	gr.betting(sb, float64(gr.SB))

	//5、产生大盲
	bb := gr.Blind(sb.chair)
	bb.blind = pb_msg.Enum_Blind_BIG_BLIND
	log.Debug("大盲注座位号为 : %v", bb.chair)
	//6、大盲注下注
	gr.betting(bb, float64(gr.BB))

	// Round 1：preFlop 开始发手牌,下注
	gr.gameStep = pb_msg.Enum_GameStep_STEP_PRE_FLOP

	//准备阶段
	gr.readyPlay()

	gr.Each(0, func(p *Player) bool {
		//1、生成玩家手牌,获取的是对应牌型生成二进制的数
		p.cards = algorithm.Cards{gr.Cards.Take(), gr.Cards.Take()}
		p.cardKeys = p.cards.HexInt()

		log.Debug("获取牌型 ~ :%v", p.cards.Hex())
		log.Debug("玩家手牌 ~ :%v", p.cards.HexInt())
		//2、获取手牌类型,只有两个可能,1为高牌,2为一对
		kind, _ := algorithm.De(p.cards.GetType())
		log.Debug("手牌类型 ~ :%v", kind)

		enter := p.RspEnterRoom()
		p.connAgent.WriteMsg(enter)
		return true
	})
	//行动, 下注, 如果玩家全部摊牌直接比牌
	gr.action(0)

	//b、是否本轮已经结束
	// Round 2：Flop 翻牌圈,牌桌上发3张公牌
	//gr.gameStep = pb_msg.Enum_GameStep_STEP_FLOP
	//1、生成桌面公牌
	gr.Cards = algorithm.Cards{gr.Cards.Take(), gr.Cards.Take(), gr.Cards.Take()}
	//2、赋值
	gr.publicCardKeys = gr.Cards.HexInt()
	gr.Each(0, func(p *Player) bool {
		//生成的桌面公牌赋值
		return true
	})
	log.Debug("桌面工牌 ~ :%v", gr.Cards)

	// Round 3：Turn 转牌圈,牌桌上发第4张公共牌
	//gr.gameStep = pb_msg.Enum_GameStep_STEP_TURN

	// Round 4：River 河牌圈,牌桌上发第5张公共牌
	//gr.gameStep = pb_msg.Enum_GameStep_STEP_RIVER

	// showdown 摊开底牌,开牌比大小
	//gr.gameStep = pb_msg.Enum_GameStep_STEP_SHOW_DOWN

	//6、游戏结束，停留5秒，重新开始游戏
	//gr.Status = emRoomStateOver

	//遍历房间所有用户，玩家OnLine状态为false说明用户已经断线，直接踢掉
	for _, v := range gr.AllPlayer {
		if v != nil && v.IsOnLine == false {
			//发送配置消息给前端，用户已断线
			v.SendConfigMsg(RECODE_LOSTCONNECT, data, pb_msg.Enum_SvrTipType_MSG)
			log.Debug("用户已掉线,直接踢出房间~")
			v.PlayerExitRoom()
		}
	}
	//重开遍历PlayerList列表的用户,开始游戏

}

//PlayerJoin 玩家加入房间
func (gr *GameRoom) PlayerJoin(p *Player) uint8 {

	log.Debug("Player Join Room ~")

	// 玩家带入筹码
	p.chips = gr.DragInRoomChips(p)
	log.Debug("玩家带入筹码 : %v", p.chips)
	log.Debug("玩家剩余金额 : %v", p.balance)

	gr.curPlayerNum++
	p.chair = gr.FindAbleChair()
	gr.PlayerList[p.chair] = p

	//房间总人数
	gr.AllPlayer = append(gr.AllPlayer, p)

	p.room = gr

	//新加入的玩家信息
	p.OtherPlayerJoin()

	if gr.Status != emRoomStateRun {
		// RUN
		gr.Running()
	} else {
		// 如果已经在Running，游戏已经开始，玩家则为弃牌状态，则广播给其他玩家
		p.playerStatus = pb_msg.Enum_PlayerStatus_STATUS_FOLD
		enter := p.RspEnterRoom()
		gr.Broadcast(enter)
	}

	// 返回前端房间信息
	roomData := p.RspEnterRoom()
	p.connAgent.WriteMsg(roomData)
	fmt.Println(roomData)

	return uint8(p.chair)
}

//ExitFromRoom 玩家从房间退出
func (gr *GameRoom) ExitFromRoom(p *Player) {

	//玩家离场
	p.OtherPlayerLeave()

	if p.chair == -1 {
		log.Debug("观战玩家退出房间 ~")
		//玩家退出房间, 将剩余的筹码转换为玩家金额
		p.balance = p.chips
		p.chips = 0
		log.Debug("玩家剩余余额 : %v", p.balance)

	} else {
		log.Debug("游戏玩家退出房间 ~")
		gr.curPlayerNum--
		p.balance = p.chips
		p.chips = 0
		log.Debug("玩家剩余余额 : %v", p.balance)
		gr.PlayerList[p.chair] = nil
	}
	fmt.Println("ExitFromRoom curPlayerNum ~ :", gr.curPlayerNum)

	err := p.update()
	if err != nil {
		log.Error("ExitFromRoom更新失败 ~ :", err)
	}

	for k, v := range gr.AllPlayer {
		if v != nil {
			if v == p {
				gr.AllPlayer = append(gr.AllPlayer[:k], gr.AllPlayer[k+1:]...)
				log.Debug("删除房间总人数成功 ~")
			}
		}

		if len(gr.AllPlayer) == 0 {
			fmt.Println("Room PlayerNum is 0，so delete this room! ~ ")

			gameHall.DeleteRoom(p.room.roomInfo.RoomId)
		} else {
			//给其他玩家广播该用户已下线！
			data := &pb_msg.LoginResultS2C{}
			data.PlayerInfo = new(pb_msg.PlayerInfo)
			data.PlayerInfo.Id = &p.ID
			data.PlayerInfo.Name = &p.name
			data.PlayerInfo.Face = &p.headImg
			data.PlayerInfo.Balance = &p.balance

			gr.Broadcast(data)
		}

		data := &pb_msg.ExitRoomS2C{}
		data.PlayerInfo = new(pb_msg.PlayerInfo)
		data.PlayerInfo.Id = &p.ID
		data.PlayerInfo.Name = &p.name
		data.PlayerInfo.Face = &p.headImg
		data.PlayerInfo.Balance = &p.balance

		p.connAgent.WriteMsg(data)
	}
}
