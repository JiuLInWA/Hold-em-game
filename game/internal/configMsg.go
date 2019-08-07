package internal

import (
	"encoding/json"
	"fmt"
)

//配置消息，全局变量
var data string

const (
	RECODE_CREATERR      = 1001
	RECODE_FINDROOM      = 1002
	RECODE_JOINROOMPWD   = 1003
	RECODE_PERSONNUM     = 1004
	RECODE_PLAYERMONEY   = 1005
	RECODE_JoinROOMERR   = 1006
	RECODE_PLAYERDESTORY = 1007
	RECODE_NOTCHIPS      = 1008
	RECODE_LOSTCONNECT   = 1009
	RECODE_TIMEOUTFOLD   = 1010
)

var recodeText = map[int32]string{
	RECODE_CREATERR:      "用户已创建房间",
	RECODE_FINDROOM:      "请求加入的房间号不存在",
	RECODE_JOINROOMPWD:   "加入房间密码错误,",
	RECODE_PERSONNUM:     "房间人数已满,不能进入房间",
	RECODE_PLAYERMONEY:   "用户金额不足,不能进入房间",
	RECODE_JoinROOMERR:   "用户已在当前房间,不能再次进入",
	RECODE_PLAYERDESTORY: "用户已在其他地方登录",
	RECODE_NOTCHIPS:      "玩家带入筹码已不足",
	RECODE_LOSTCONNECT:   "用户已掉线，直接踢出房间",
	RECODE_TIMEOUTFOLD:   "玩家行动超时，直接弃牌",
}


func jsonData() {
	reCode, err := json.Marshal(recodeText)
	if err != nil {
		fmt.Println("json.Marshal err:", err)
		return
	}

	data := string(reCode)
	fmt.Println("S2C jsonData String ~", data)
}

type CfgData struct {
	Id        string
	Bb        int32
	MinTakeIn int32
	MaxTakeIn int32
}

// 房间配置限定金额
func CfgDataHandle(cfgId string) CfgData {
	cfgSlice := []CfgData{
		{Id: "0", Bb: 10, MinTakeIn: 1000, MaxTakeIn: 10000},
		{Id: "1", Bb: 20, MinTakeIn: 2000, MaxTakeIn: 20000},
		{Id: "2", Bb: 30, MinTakeIn: 3000, MaxTakeIn: 30000},
		{Id: "3", Bb: 40, MinTakeIn: 4000, MaxTakeIn: 40000},
		{Id: "4", Bb: 50, MinTakeIn: 5000, MaxTakeIn: 50000},
	}
	for _, v := range cfgSlice {
		if v.Id == cfgId {
			//fmt.Println("cfgData :", v)
			return v
		}
	}
	return CfgData{}
}
