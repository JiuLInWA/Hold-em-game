package internal

import (
	"fmt"
	"github.com/name5566/leaf/gate"
	"github.com/name5566/leaf/log"
	"server/game/algorithm"
	pb_msg "server/msg/Protocal"
	"time"
)

// 牌型数据
type CardSuitData struct {
	HandCardKeys   []int32              // 组成牌型的手牌
	PublicCardKeys []int32              // 组成牌型的公牌
	SuitPattern    pb_msg.Enum_CardSuit // 牌型
}

//定义一个玩家
type Player struct {
	//玩家的连接代理
	connAgent gate.Agent
	//玩家ID
	ID string
	//玩家座位号
	chair int32
	//全局索引
	index uint32
	//客户端延迟
	uClientDelay uint16

	TimerNow int64

	name    string  //玩家昵称
	headImg string  //玩家头像
	balance float64 //玩家余额

	cards         algorithm.Cards          //牌型数据
	IsRaised      bool                     //本轮是否已经raise过(每个玩家每轮只有一次raise的机会)
	playerStatus  pb_msg.Enum_PlayerStatus //本轮玩家表态
	dropedBets    float64                  //玩家状态为跟牌、加注时最终要下的赌注额
	dropedBetsSum float64                  //这局中总共下注了多少
	cardKeys      []int32                  //玩家手牌
	cardSuitData  *CardSuitData            //玩家手牌和公牌能组成的牌型数据
	isWinner      bool                     //玩家是否赢家
	blind         pb_msg.Enum_Blind        //盲注类型
	isButton      bool                     //是否庄家
	isAllIn       bool                     //是否已经AllIn
	resultMoney   float64                  //本局游戏结束时收到的钱

	HandValue uint32
	actions   chan pb_msg.Enum_ActionOptions //玩家行动命令
	chips     float64                        //带入筹码
	room      *GameRoom                      //所在房间
	IsOnLine  bool                           //是否在线
}

func (p *Player) Init() {
	p.connAgent = nil
	p.chair = 0
	p.index = 0

	//TODO 用户登录创建玩家初始化设定，后面根据拿去中心数据做修改
	p.headImg = "http://1814.img.pp.sohu.com.cn/images/blog/2008/11/1/13/20/11dfe567377g213.jpg"
	p.balance = 4000

	p.IsRaised = false
	p.playerStatus = pb_msg.Enum_PlayerStatus_STATUS_WAITING
	p.dropedBets = 0
	p.dropedBetsSum = 0
	p.cardSuitData = nil
	p.isWinner = false
	p.blind = pb_msg.Enum_Blind_NO_BLIND
	p.isButton = false
	p.isAllIn = false
	p.resultMoney = 0

	p.chips = 0
	p.room = nil
	p.actions = make(chan pb_msg.Enum_ActionOptions)
	p.IsOnLine = true
	p.uClientDelay = 0
}

//StartBreathe 开始呼吸
func (p *Player) StartBreathe() {
	ticker := time.NewTicker(time.Second * 3)
	go func() {
		for { //循环
			<-ticker.C
			p.uClientDelay++
			fmt.Println("用户 ~ :", p.ID, " p.uClientDelay++ :", p.uClientDelay)
			//已经超过9秒没有收到客户端心跳，踢掉好了
			if p.uClientDelay > 3 {
				if p.room != nil {
					if p.room.activePos == p.chair {
						//TODO 直接弃牌，设为观战玩家
						p.playerStatus = pb_msg.Enum_PlayerStatus_STATUS_FOLD

					}
					//玩家离场
					p.room.curPlayerNum--
					p.room.PlayerList[p.chair] = nil
					p.IsOnLine = false
					p.OtherPlayerLeave()

					//未入座 座位为 -1
					p.chair = -1

					s := &pb_msg.StandUpS2C{}
					room := p.RspRoomData()
					s.RoomData = room
					p.room.Broadcast(s)
				}
				p.connAgent.Destroy()
				return
			}
		}
	}()
}

//onClientBreathe 客户端呼吸，长时间未执行该函数可能已经断网，将主动踢掉
func (p *Player) onClientBreathe() {
	p.uClientDelay = 0
}

//保存用户数据
func (p *Player) Save() {

}

//玩家向客户端发送消息  对于离线玩家需要把数据保存到会话数据库中
func (p *Player) SendMsg(msg interface{}) {
	if p.connAgent != nil {
		p.connAgent.WriteMsg(msg)
	}
}

//mapGlobalPlayer 全局玩家列表
var gPlayerGlobalIndex uint32
var mapGlobalPlayer map[uint32]*Player

//以用户ID为key的玩家映射表
var mapUserID2Player map[string]*Player

//初始化全局玩家列表
func InitPlayerMap() {
	gPlayerGlobalIndex = 0
	mapGlobalPlayer = make(map[uint32]*Player)
	mapUserID2Player = make(map[string]*Player)
}

//创建一个玩家 一旦有客户端连接先创建一个玩家
func CreatePlayer() *Player {
	p := &Player{}
	p.Init()
	mapGlobalPlayer[gPlayerGlobalIndex] = p

	p.index = gPlayerGlobalIndex
	fmt.Println("CreatePlayer index ~ :", p.index)
	gPlayerGlobalIndex++
	return p
}

//PlayerRegister 以id进行玩家注册，每个玩家只能有唯一ID，如果有相同的ID注册 需要关闭之前相同ID的玩家
func PlayerRegister(ID string, neo *Player) {
	//先检查该ID玩家是否已经注册过
	log.Debug("PlayerRegister ~ :%v", ID)
	oldp, ok := mapUserID2Player[ID]
	if ok {
		// A、B同一账号，A处于登陆状态，B登陆把A挤掉，发送消息给前端做处理
		log.Debug("player have :%v", oldp.ID)
		log.Debug("用户已在其他地方登录 ~")
		log.Debug("force destroy player who after login")

		oldp.SendConfigMsg(RECODE_PLAYERDESTORY, data, pb_msg.Enum_SvrTipType_WARN)

		// B用户登录，主动断掉A用户
		oldp.connAgent.Destroy()
		DeletePlayer(oldp)
	}
	neo.ID = ID
	mapUserID2Player[ID] = neo
}

//GetPlayer 获取玩家结构
func GetPlayer(ID string) *Player {
	p, ok := mapUserID2Player[ID]
	if ok {
		log.Debug("获取用户结构成功并返回 ~")
		return p
	}
	return nil
}

// DeletePlayer 删除玩家
func DeletePlayer(p *Player) {
	p.Save()
	delete(mapGlobalPlayer, p.index)

	// 删除id表时检查一下是不是当前客户端
	saveone, ok := mapUserID2Player[p.ID]

	if ok && saveone.index == p.index {
		// 是当前客户端才删除
		delete(mapUserID2Player, p.ID)
		fmt.Println("我进去了~")
	} else {
		fmt.Println("我没进去~")
	}
	// 有玩家退出时也保存一遍
	//gStat.save()
}

//-----------------------------------------

//SendConfigMsg 返回客户端配置消息
func (p *Player) SendConfigMsg(code int32, data string, tipType pb_msg.Enum_SvrTipType) {
	msg := &pb_msg.SvrMsgS2C{}
	msg.Code = new(int32)
	msg.Data = new(string)
	msg.TipType = (*pb_msg.Enum_SvrTipType)(new(int32))
	*msg.Code = code
	*msg.Data = data
	*msg.TipType = tipType
	p.connAgent.WriteMsg(msg)
}

//PlayerExitRoom 玩家退出房间
func (p *Player) PlayerExitRoom() {
	log.Debug("Player from Room Exit ~: %v", p.ID)

	if p.room != nil {
		p.room.ExitFromRoom(p)
		p.room = nil
	} else {
		log.Debug("Exit Room , but not found Player Room")
	}
}

//GetUserRoomInfo 用户重新登陆，获取房间信息
func (p *Player) GetUserRoomInfo() *Player {
	//TODO 方法一,这样每个用户都先要遍历一遍，这样多用户登陆进来，速度会变慢
	for _, v := range gameHall.roomList {
		if v != nil {
			for _, pl := range v.AllPlayer {
				if pl != nil && pl.ID == p.ID {
					return pl
				}
			}
		}
	}
	return nil
}

//RspRoomData 返回房间数据
func (p *Player) RspRoomData() *pb_msg.RoomData {

	r := &pb_msg.RoomData{}
	r.RoomInfo = new(pb_msg.RoomInfo)
	r.RoomInfo.RoomId = &p.room.roomInfo.RoomId
	r.RoomInfo.CfgId = &p.room.roomInfo.CfgId
	r.RoomInfo.MaxPlayer = &p.room.roomInfo.MaxPlayer
	r.RoomInfo.ActionTimeS = &p.room.roomInfo.ActionTimeS
	r.RoomInfo.Pwd = &p.room.roomInfo.Pwd
	r.IsStepEnd = &p.room.isStepEnd
	r.GameStep = &p.room.gameStep
	r.MinRaise = &p.room.minRaise
	r.ActivePos = &p.room.activePos
	r.NextStepTs = &p.room.nextStepTs
	r.Pot = &p.room.pot
	r.PublicCardKeys = p.room.publicCardKeys

	for _, v := range p.room.AllPlayer {
		if v != nil {
			data := &pb_msg.PlayerData{}
			data.PlayerInfo = new(pb_msg.PlayerInfo)
			data.PlayerInfo.Id = &v.ID
			data.PlayerInfo.Name = &v.name
			data.PlayerInfo.Face = &v.headImg
			data.PlayerInfo.Balance = &v.chips
			data.Position = &v.chair
			data.IsRaised = &v.IsRaised
			data.PlayerStatus = &v.playerStatus
			data.DropedBets = &v.dropedBets
			data.DropedBetsSum = &v.dropedBetsSum
			if p.ID == v.ID {
				data.CardKeys = v.cardKeys
			}
			data.CardSuitData = new(pb_msg.CardSuitData)
			data.CardSuitData.HandCardKeys = v.cardSuitData.HandCardKeys
			data.CardSuitData.PublicCardKeys = v.cardSuitData.PublicCardKeys
			data.CardSuitData.SuitPattern = &v.cardSuitData.SuitPattern
			data.IsWinner = &v.isWinner
			data.Blind = &v.blind
			data.IsButton = &v.isButton
			data.IsAllIn = &v.isAllIn
			data.ResultMoney = &v.resultMoney
			r.PlayerDatas = append(r.PlayerDatas, data)
		}
	}
	return r
}

//RspEnterRoom 返回客户端进入房间数据
func (p *Player) RspEnterRoom() *pb_msg.EnterRoomS2C {
	//需要返回玩家自己本身消息，和同房间其他玩家基础信息
	er := &pb_msg.EnterRoomS2C{}
	r := p.RspRoomData()
	er.RoomData = r
	return er
}

//NewPlayerJoin 新加入房间玩家信息
func (p *Player) OtherPlayerJoin() {
	pl := &pb_msg.OtherPlayerJoinS2C{}
	pl.PlayerData = new(pb_msg.PlayerData)
	pl.PlayerData.PlayerInfo = new(pb_msg.PlayerInfo)
	pl.PlayerData.PlayerInfo.Id = &p.ID
	pl.PlayerData.PlayerInfo.Name = &p.name
	pl.PlayerData.PlayerInfo.Face = &p.headImg
	pl.PlayerData.PlayerInfo.Balance = &p.chips
	pl.PlayerData.Position = &p.chair
	pl.PlayerData.IsRaised = &p.IsRaised
	pl.PlayerData.PlayerStatus = &p.playerStatus
	pl.PlayerData.DropedBets = &p.dropedBets
	pl.PlayerData.DropedBetsSum = &p.dropedBetsSum
	pl.PlayerData.CardKeys = p.cardKeys
	pl.PlayerData.CardSuitData = new(pb_msg.CardSuitData)
	p.cardSuitData = new(CardSuitData)
	pl.PlayerData.CardSuitData.HandCardKeys = p.cardSuitData.HandCardKeys
	pl.PlayerData.CardSuitData.PublicCardKeys = p.cardSuitData.PublicCardKeys
	pl.PlayerData.CardSuitData.SuitPattern = &p.cardSuitData.SuitPattern
	pl.PlayerData.IsWinner = &p.isWinner
	pl.PlayerData.Blind = &p.blind
	pl.PlayerData.IsButton = &p.isButton
	pl.PlayerData.IsAllIn = &p.isAllIn
	pl.PlayerData.ResultMoney = &p.resultMoney

	//广播新进入玩家信息
	p.room.BroadCastExcept(pl, p)
}

//OtherPlayerLeave 其他玩家离场(观战也属于)
func (p *Player) OtherPlayerLeave() {
	leave := &pb_msg.OtherPlayerLeaveS2C{}
	leave.Position = &p.chair
	leave.Pot = &p.room.pot

	p.room.BroadCastExcept(leave, p)
}

//SitDownTable 玩家坐下座位
func (p *Player) SitDownTable(pos int32) {

	p.room.curPlayerNum++
	p.chair = pos
	p.room.PlayerList[p.chair] = p

	//新加入的玩家信息
	p.OtherPlayerJoin()

	if p.room.Status != emRoomStateRun {
		// RUN
		p.room.Running()
	} else {
		// 如果已经在Running，游戏已经开始，玩家则为弃牌状态，则广播给其他玩家
		p.playerStatus = pb_msg.Enum_PlayerStatus_STATUS_FOLD

		// 返回前端房间信息
		s := &pb_msg.SitDownS2C{}
		room := p.RspRoomData()
		s.RoomData = room
		p.room.Broadcast(s)
	}
}

//StandUpBattle 玩家站起观战
func (p *Player) StandUpBattle() {
	//判断玩家是否当前行动玩家，如果是则直接弃牌站起
	if p.room.activePos == p.chair {
		//TODO 直接弃牌，设为观战玩家
		p.playerStatus = pb_msg.Enum_PlayerStatus_STATUS_FOLD

	}
	//玩家离场
	p.room.curPlayerNum--
	p.room.PlayerList[p.chair] = nil
	p.OtherPlayerLeave()

	//未入座 座位为 -1，视为观战
	p.chair = -1

	s := &pb_msg.StandUpS2C{}
	room := p.RspRoomData()
	s.RoomData = room
	p.room.Broadcast(s)
}

func (p *Player) GetAction(timeout time.Duration) {
	//超时处理
	//var after <-chan time.Time
	//after = time.After(timeout)
	after := time.NewTicker(timeout)

	log.Debug("行动时间为1 :%v", time.Now().Format("2006-01-02 15:04:05"))

	for {
		select {
		case x := <-p.actions:
			fmt.Println("xxxxxxxxxxxxxxxx:", x)

			log.Debug("玩家的行动操作为 :%v", x)
			switch x {
			case pb_msg.Enum_ActionOptions_ACT_FOLD:
				p.playerStatus = pb_msg.Enum_PlayerStatus_STATUS_FOLD
				p.room.remain--
			case pb_msg.Enum_ActionOptions_ACT_CALL:
				p.playerStatus = pb_msg.Enum_PlayerStatus_STATUS_CALL
				//玩家筹码
				p.chips -= p.room.preChips
				//本轮下注金额
				p.dropedBets += p.room.preChips
				//本局玩家下注总金额
				p.dropedBetsSum += p.room.preChips
				//总筹码
				p.room.pot += p.room.preChips
			case pb_msg.Enum_ActionOptions_ACT_RAISE:
				p.playerStatus = pb_msg.Enum_PlayerStatus_STATUS_RAISE
				p.chips -= p.room.preChips
				p.dropedBets += p.room.preChips
				//上个玩家下注金额
				p.room.preChips = p.dropedBets
				p.dropedBetsSum += p.room.preChips
				p.room.pot += p.room.preChips
			case pb_msg.Enum_ActionOptions_ACT_CHECK:
				p.playerStatus = pb_msg.Enum_PlayerStatus_STATUS_CHECK
			}
			if p.chips == 0 {
				p.room.allin++
			}
			//玩家本局下注的总筹码数
			p.room.Chips[p.chair] += uint32(p.room.preChips)

			return
		case <-after.C:
			log.Debug("行动时间为2 :%v", time.Now().Format("2006-01-02 15:04:05"))

			p.SendConfigMsg(RECODE_TIMEOUTFOLD, data, pb_msg.Enum_SvrTipType_MSG)
			log.Debug("玩家行动超时,直接弃牌 :%v", p.ID)

			p.playerStatus = pb_msg.Enum_PlayerStatus_STATUS_FOLD
			p.room.remain--
			return
		}
	}
}

//func (gr *GameRoom) handleRoomEvent(a gate.Agent, p *Player, e interface{}) {
//	switch t := e.(type) {
//	case *pb_msg.PlayerActionC2S:
//		gr.HandleGameEvent(a, p, e)
//	default:
//		log.Error("GameRoom 事件无法处理~", t)
//	}
//}
//
//func (gr *GameRoom) HandleGameEvent(a gate.Agent, p *Player, e interface{}) {
//	switch t := e.(type) {
//	case *pb_msg.PlayerActionC2S:
//		gr.PlayerPlay(a, p, e.(*pb_msg.PlayerActionC2S))
//	default:
//		log.Error("PlayerGame 事件无法处理~", t)
//	}
//}
