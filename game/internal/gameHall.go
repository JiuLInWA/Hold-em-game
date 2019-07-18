package internal

import (
	"fmt"
	"github.com/name5566/leaf/log"
	pb_msg "server/msg/Protocal"

	"strconv"
)

//var userRoomMap map[string]*GameRoom

//GameHall 描述游戏大厅，目前一个游戏就一个大厅
type GameHall struct {
	maxPlayer uint16
	roomList  []*GameRoom
}

//Init 大厅初始化~！
func (gh *GameHall) Init() {
	gh.maxPlayer = 5000
	fmt.Printf("this is gamehall init~!!it can support %d run~\n", gh.maxPlayer)
}

//CreateRoom 创建一个游戏房间
func (gh *GameHall) CreateRoom(p *Player, r *RoomInfo) *GameRoom {

	room := &GameRoom{}
	room.Init(r)

	dataCfg := CfgDataHandle(r.CfgId)
	if int32(p.balance) < dataCfg.MinTakeIn {
		balance := strconv.FormatFloat(p.balance, 'f', 2, 64)
		data = "玩家金额为" + balance + "," + "房间限制金额为" + string(dataCfg.MinTakeIn)
		p.SendConfigMsg(RECODE_PLAYERMONEY, data, pb_msg.Enum_SvrTipType_WARN)

		log.Debug("玩家金额不足，创建房间失败 ~")
		return nil
	}

	gh.roomList = append(gh.roomList, room)
	fmt.Println("CreateRoom Total Number ~ : ", len(gh.roomList))

	return room
}

//FindAvailableRoom 寻找一个可用的房间
func (gh *GameHall) FindAvailableRoom(p *Player, r *RoomInfo) *GameRoom {
	for _, room := range gh.roomList {
		dataCfg := CfgDataHandle(r.CfgId)
		fmt.Println("QuickStart Config :", p.balance, r.ActionTimeS, r.MaxPlayer)
		if dataCfg.MinTakeIn < int32(p.balance) && room.IsRoomActionTimes(int32(r.ActionTimeS)) &&
			room.IsPlayerMaxNum(r.MaxPlayer) && room.IsCanJoin() && room.IsRoomPwd(r.Pwd) {
			return room
		}
	}
	return gh.CreateRoom(p, r)
}

//PlayerQuickStart 玩家快速匹配房间
func (gh *GameHall) PlayerQuickStart(p *Player, r *RoomInfo) uint8 {
	room := gh.FindAvailableRoom(p, r)
	if room == nil {
		return 0
	}

	return room.PlayerJoin(p)
}

//PlayerCreatRoom 玩家手动创建房间
func (gh *GameHall) PlayerCreatRoom(p *Player, r *RoomInfo) uint8 {
	room := gh.CreateRoom(p, r)
	if room == nil {
		return 0
	}

	return room.PlayerJoin(p)
}

//PlayerJoinRoom 玩家指定房间ID加入
func (gh *GameHall) PlayerJoinRoom(p *Player, rid string, pwd string) uint8 {
	for _, room := range gh.roomList {

		if room.roomInfo.RoomId != rid {
			p.SendConfigMsg(RECODE_FINDROOM, data, pb_msg.Enum_SvrTipType_WARN)
			log.Debug("请求加入的房间号不存在~")
			return 0
		}
		if !room.IsRoomPwd(pwd) {
			p.SendConfigMsg(RECODE_JOINROOMPWD, data, pb_msg.Enum_SvrTipType_WARN)
			log.Debug("加入房间密码输入错误~")
			return 0
		}

		dataCfg := CfgDataHandle(room.roomInfo.CfgId)
		if int32(p.balance) < dataCfg.MinTakeIn {
			balance := strconv.FormatFloat(p.balance, 'f', 2, 64)
			data = "玩家金额为" + balance + "," + "房间限制金额为" + string(dataCfg.MinTakeIn)
			p.SendConfigMsg(RECODE_PLAYERMONEY, data, pb_msg.Enum_SvrTipType_WARN)

			log.Debug("玩家金额不足，进入房间失败~")
			return 0
		}
		if !room.IsCanJoin() {
			p.SendConfigMsg(RECODE_PERSONNUM, data, pb_msg.Enum_SvrTipType_WARN)
			log.Debug("房间人数已满，加入房间失败~")
			return 0
		}
		return room.PlayerJoin(p)
	}
	return 0
}

//DeleteRoom 删除房间信息
func (gh *GameHall) DeleteRoom(rid string) {
	for k, v := range gh.roomList {
		if v.roomInfo.RoomId == rid {
			gh.roomList = append(gh.roomList[:k], gh.roomList[k+1:]...)
			fmt.Println("删除房间信息成功 ~")
		}
	}
}

var gameHall GameHall
