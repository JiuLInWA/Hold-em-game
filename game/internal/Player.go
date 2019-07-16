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
	isSelf        bool                     //是否玩家自己
	resultMoney   float64                  //本局游戏结束时收到的钱

	chips    float64   //带入筹码
	room     *GameRoom //所在房间
	IsOnLine bool      //是否在线
}

func (p *Player) Init() {
	p.connAgent = nil
	p.chair = 0
	p.index = 0

	//TODO 用户登录创建玩家初始化设定，后面根据拿去中心数据做修改
	p.name = "Hold-em"
	p.headImg = "https://encrypted-tbn0.gstatic.com/images?q=tbn:ANd9GcRB45_5R6pdUp4xVFZ83dcA7BJkiSYjW8h6Z92uJo9WBkhbAMgN"
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
	p.isSelf = false
	p.resultMoney = 0
	p.chips = 0
	p.room = nil
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
					enter := p.RspEnterRoom()
					p.room.Broadcast(enter)
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
var mapGlobalPlayer map[uint32]*Player
var gPlayerGlobalIndex uint32

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
	fmt.Println("PlayerRegister ~ :", ID)
	oldp, ok := mapUserID2Player[ID]
	if ok {
		fmt.Println(ID, "have")
		fmt.Println("force destroy player who after login", oldp.ID)
		// A、B同一账号，A处于登陆状态，B登陆把A挤掉，发送消息给前端做处理
		msg := pb_msg.SvrMsgS2C{}
		msg.Code = RECODE_PLAYERDESTORY
		msg.TipType = pb_msg.Enum_SvrTipType_WARN
		oldp.connAgent.WriteMsg(&msg)

		log.Debug("用户已在其他地方登录 ~")

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
				if pl != nil {
					if pl.ID == p.ID {
						return pl
					}
				}
			}
		}
	}
	return nil
}

//RspEnterRoom 返回客户端房间数据
func (p *Player) RspEnterRoom() *pb_msg.EnterRoomS2C {

	//需要返回玩家自己本身消息，和同房间其他玩家基础信息
	er := &pb_msg.EnterRoomS2C{}
	er.RoomData = new(pb_msg.RoomData)
	er.RoomData.RoomInfo = new(pb_msg.RoomInfo)
	er.RoomData.RoomInfo.RoomId = p.room.roomInfo.RoomId
	er.RoomData.RoomInfo.CfgId = p.room.roomInfo.CfgId
	er.RoomData.RoomInfo.MaxPlayer = p.room.roomInfo.MaxPlayer
	er.RoomData.RoomInfo.ActionTimeS = p.room.roomInfo.ActionTimeS
	er.RoomData.RoomInfo.Pwd = p.room.roomInfo.Pwd
	er.RoomData.IsStepEnd = p.room.isStepEnd
	er.RoomData.GameStep = p.room.gameStep
	er.RoomData.MinRaise = p.room.minRaise
	er.RoomData.ActivePos = int32(p.room.activePos)
	er.RoomData.NextStepTs = p.room.nextStepTs
	er.RoomData.Pot = p.room.pot
	er.RoomData.PublicCardKeys = p.room.publicCardKeys

	for _, v := range p.room.AllPlayer {
		if v != nil {
			data := &pb_msg.PlayerData{}
			data.PlayerInfo = new(pb_msg.PlayerInfo)
			data.PlayerInfo.Id = v.ID
			data.PlayerInfo.Name = v.name
			data.PlayerInfo.HeadImg = v.headImg
			data.PlayerInfo.Balance = v.balance
			data.Position = int32(v.chair)
			data.IsRaised = v.IsRaised
			data.PlayerStatus = v.playerStatus
			data.DropedBets = v.dropedBets
			data.DropedBetsSum = v.dropedBetsSum
			data.CardKeys = v.cardKeys
			data.CardSuitData = new(pb_msg.CardSuitData)
			v.cardSuitData = new(CardSuitData)
			data.CardSuitData.HandCardKeys = v.cardSuitData.HandCardKeys
			data.CardSuitData.PublicCardKeys = v.cardSuitData.PublicCardKeys
			data.CardSuitData.SuitPattern = v.cardSuitData.SuitPattern
			data.IsWinner = v.isWinner
			data.Blind = v.blind
			data.IsButton = v.isButton
			data.IsAllIn = v.isAllIn
			data.IsSelf = v.isSelf
			data.ResultMoney = v.resultMoney
			er.RoomData.PlayerDatas = append(er.RoomData.PlayerDatas, data)
		}
	}
	return er
}

//NewPlayerJoin 新加入房间玩家信息
func (p *Player) OtherPlayerJoin() {
	pl := &pb_msg.OtherPlayerJoinS2C{}
	pl.PlayerData = new(pb_msg.PlayerData)
	pl.PlayerData.PlayerInfo = new(pb_msg.PlayerInfo)
	pl.PlayerData.PlayerInfo.Id = p.ID
	pl.PlayerData.PlayerInfo.Name = p.name
	pl.PlayerData.PlayerInfo.HeadImg = p.headImg
	pl.PlayerData.PlayerInfo.Balance = p.balance
	pl.PlayerData.Position = p.chair
	pl.PlayerData.IsRaised = p.IsRaised
	pl.PlayerData.PlayerStatus = p.playerStatus
	pl.PlayerData.DropedBets = p.dropedBets
	pl.PlayerData.DropedBetsSum = p.dropedBetsSum
	pl.PlayerData.CardKeys = p.cardKeys
	pl.PlayerData.CardSuitData = new(pb_msg.CardSuitData)
	p.cardSuitData = new(CardSuitData)
	pl.PlayerData.CardSuitData.HandCardKeys = p.cardSuitData.HandCardKeys
	pl.PlayerData.CardSuitData.PublicCardKeys = p.cardSuitData.PublicCardKeys
	pl.PlayerData.CardSuitData.SuitPattern = p.cardSuitData.SuitPattern
	pl.PlayerData.IsWinner = p.isWinner
	pl.PlayerData.Blind = p.blind
	pl.PlayerData.IsButton = p.isButton
	pl.PlayerData.IsAllIn = p.isAllIn
	pl.PlayerData.IsSelf = p.isSelf
	pl.PlayerData.ResultMoney = p.resultMoney

	//广播新进入玩家信息
	p.room.Broadcast(pl)
}

//OtherPlayerLeave 其他玩家离场(观战也属于)
func (p *Player) OtherPlayerLeave() {
	leave := &pb_msg.OtherPlayerLeaveS2C{}
	leave.Position = p.chair
	leave.Pot = p.room.pot

	p.room.Broadcast(leave)
}

//SitDownTable 玩家坐下座位
func (p *Player) SitDownTable(gr *GameRoom, pos int32) {

	gr.curPlayerNum++
	p.chair = pos
	gr.PlayerList[p.chair] = p

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
	e := p.RspEnterRoom()
	s := pb_msg.SitDownS2C{}
	s.RoomData = e.RoomData
	p.connAgent.WriteMsg(s)
}

//StandUpBattle 玩家站起观战
func (p *Player) StandUpBattle() {
	//玩家离场
	p.room.curPlayerNum--
	p.room.PlayerList[p.chair] = nil
	p.OtherPlayerLeave()

	//未入座 座位为 -1，视为观战
	p.chair = -1
	e := p.RspEnterRoom()
	s := pb_msg.SitDownS2C{}
	s.RoomData = e.RoomData
	p.room.Broadcast(s)
}
