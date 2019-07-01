package internal

import (
	"fmt"
	"github.com/name5566/leaf/log"
	pb_msg "server/msg/Protocal"

	"strconv"
)

//GameHall 描述游戏大厅，目前一个游戏就一个大厅
type GameHall struct {
	maxPlayer uint16
	roomList  []*GameRoom
}

//Init 大厅初始化~！
func (gh *GameHall) Init() {
	gh.maxPlayer = 2000

	fmt.Printf("this is gamehall init~!!it can support %d run~\n", gh.maxPlayer)
}

//CreateRoom 创建一个游戏房间
func (gh *GameHall) CreateRoom(p *Player, r *RoomInfo) *GameRoom {

	room := &GameRoom{}
	room.Init(p, r)
	gh.roomList = append(gh.roomList, room)
	fmt.Println("CreateRoom Total Number ~ : ", len(gh.roomList))
	return room
}

//FindAvailableRoom 寻找一个可用的房间
func (gh *GameHall) FindAvailableRoom(p *Player, r *RoomInfo) *GameRoom {
	for _, room := range gh.roomList {
		dataCfg := CfgDataHandle(r.CfgId)
		fmt.Println("2~22 :", p.balance, r.ActionTimeS, r.MaxPlayer)
		if dataCfg.MinTakeIn < int32(p.balance) && room.IsRoomActionTimes(int32(r.ActionTimeS)) &&
			room.IsPlayerMaxNum(r.MaxPlayer) && room.IsCanJoin() {
			return room
		}
	}
	return gh.CreateRoom(p, r)
}

//PlayerQuickStart 玩家快速匹配房间
func (gh *GameHall) PlayerQuickStart(p *Player, r *RoomInfo) uint8 {
	room := gh.FindAvailableRoom(p, r)

	return room.PlayerJoin(p)
}

//PlayerCreatRoom 玩家手动创建房间
func (gh *GameHall) PlayerCreatRoom(p *Player, r *RoomInfo) uint8 {

	room := gh.CreateRoom(p, r)

	return room.PlayerJoin(p)
}

//PlayerJoinRoom 玩家指定房间ID加入
func (gh *GameHall) PlayerJoinRoom(p *Player, rid string, pwd string) uint8 {
	for _, room := range gh.roomList {
		if room.roomInfo.RoomId == rid {
			if room.roomInfo.Pwd == pwd {
				dataCfg := CfgDataHandle(room.roomInfo.CfgId)
				if dataCfg.MinTakeIn < int32(p.balance) {
					if room.IsCanJoin() {
						return room.PlayerJoin(p)
					} else {
						msg := pb_msg.SvrMsgS2C{}
						msg.Code = RECODE_PERSONNUM
						msg.TipType = pb_msg.Enum_SvrTipType_WARN
						p.connAgent.WriteMsg(&msg)

						log.Debug("房间人数已满,加入房间失败~")
						return 0
					}
				} else {
					balance := strconv.FormatFloat(p.balance, 'f', 2, 64)
					msg := pb_msg.SvrMsgS2C{}
					msg.Code = RECODE_PLAYERMONEY
					msg.Data = "玩家金额为" + balance + "," + "房间限制金额为" + string(dataCfg.MinTakeIn)
					msg.TipType = pb_msg.Enum_SvrTipType_WARN
					p.connAgent.WriteMsg(&msg)

					log.Debug("玩家金额不足，进入房间失败~")
					return 0
				}
			} else {
				msg := pb_msg.SvrMsgS2C{}
				msg.Code = RECODE_JOINROOMPWD
				msg.TipType = pb_msg.Enum_SvrTipType_WARN
				p.connAgent.WriteMsg(&msg)

				log.Debug("加入房间密码输入错误~")
				return 0
			}
		} else {
			msg := pb_msg.SvrMsgS2C{}
			msg.Code = RECODE_FINDROOM
			msg.TipType = pb_msg.Enum_SvrTipType_WARN
			p.connAgent.WriteMsg(&msg)

			log.Debug("请求加入的房间号不存在~")
			return 0
		}
	}
	return 0
}

//DeleteRoom 删除房间信息
func (gh *GameHall) DeleteRoom(rid string) {
	for k, v := range gh.roomList {
		if v.roomInfo.RoomId == rid {
			gh.roomList = append(gh.roomList[:k], gh.roomList[k+1:]...)
		}
	}
}

var gameHall GameHall
