package internal

import (
	"encoding/json"
	"fmt"
)

const (
	RECODE_CREATERR      = 1001
	RECODE_FINDROOM      = 1002
	RECODE_JOINROOMPWD   = 1003
	RECODE_PERSONNUM     = 1004
	RECODE_PLAYERMONEY   = 1005
	RECODE_JoinROOMERR   = 1006
	RECODE_PLAYERDESTORY = 1007
)

var recodeText = map[int32]string{
	RECODE_CREATERR:      "用户已创建房间",
	RECODE_FINDROOM:      "请求加入的房间号不存在",
	RECODE_JOINROOMPWD:   "加入房间密码错误",
	RECODE_PERSONNUM:     "房间人数已满",
	RECODE_PLAYERMONEY:   "用户金额不足",
	RECODE_JoinROOMERR:   "用户已在当前房间,不能再次进入",
	RECODE_PLAYERDESTORY: "用户已在其他地方登录",
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
			fmt.Println("cfgData :", v)
			return v
		}
	}
	return CfgData{}
}