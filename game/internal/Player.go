package internal

import (
	"fmt"
	"github.com/name5566/leaf/gate"
	"github.com/name5566/leaf/log"
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
	chair uint8
	//全局索引
	index uint32
	//客户端延迟
	uClientDelay uint16

	name    string  //玩家昵称
	headImg string  //玩家头像
	balance float64 //玩家余额

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

	//所在房间
	room *GameRoom

	//cards Cards	//玩家牌型
	//chips uint32	//带入筹码
	//Bet uint32	//当前下注
}

func (p *Player) Init() {
	p.connAgent = nil
	p.chair = 0
	p.index = 0

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

	p.room = nil
	p.uClientDelay = 0
}

//保存用户数据
func (p *Player) Save() {

}

//StartBreathe 开始呼吸 这里提供函数去和中心
func (p *Player) StartBreathe() {
	ticker := time.NewTicker(time.Second * 3)

	func() {
		for { //循环
			<-ticker.C
			p.ActiveBreathe()
			//if b == true {
			//	return
			//}
		}
	}()
}

//主动呼吸
func (p *Player) ActiveBreathe()  {
	p.uClientDelay++

	if p.uClientDelay > 3 {
		//log.Debug("Player close websocket~!!!!")
		//DeletePlayer(p)
		//p.PlayerExitRoom()

		//已经超过9秒没有收到客户端心跳，踢掉好了
		p.connAgent.Destroy()

	}
}

//监测用户是否已经断线, 服务器写个心跳包
func (p *Player) onClientBreathe() {
	p.uClientDelay = 0
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
	fmt.Println("PlayerRegister", ID)
	oldp, ok := mapUserID2Player[ID]
	if ok {
		fmt.Println(ID, "have")
		fmt.Println("force destroy player who after login", oldp.ID)
		// A、B同一账号，A处于登陆状态，B登陆把A挤掉，发送消息给前端做处理
		msg := pb_msg.SvrMsgS2C{}
		msg.Code = RECODE_PLAYERDESTORY
		msg.TipType = pb_msg.Enum_SvrTipType_WARN
		oldp.connAgent.WriteMsg(&msg)
		log.Debug("msg ~: %v", msg)

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
	log.Debug(ID)
	if ok {
		log.Debug("!11")
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
		fmt.Println("我进去了")
	} else {
		fmt.Println("我没进去")
	}
	// 有玩家退出时也保存一遍
	//gStat.save()
}

//-----------------------------------------

//PlayerExitRoom 玩家退出房间
func (p *Player) PlayerExitRoom() {
	fmt.Println("Player from Room Exit")
	if p.room != nil {
		p.room.ExitFromRoom(p)
		p.room = nil
	}
}