package internal

import (
	"github.com/name5566/leaf/log"
	"reflect"
	"server/game"
	pb_msg "server/msg/Protocal"
)

func init() {
	//向当前模块（game 模块）注册 Test 消息的消息处理函数 handleTest
	//handler(&pb_msg.Test{},handleTest)
	handler(&pb_msg.PingC2S{}, handlePing)
	handler(&pb_msg.LoginC2S{}, handleLogin)
}

// 异步处理
func handler(m interface{}, h interface{}) {
	skeleton.RegisterChanRPC(reflect.TypeOf(m), h)
}

func handlePing(args []interface{}) {
	// 收到的 Hello 消息
	m := args[0].(*pb_msg.PingC2S)

	// 输出收到的消息的内容
	log.Debug("Hello : %v", m)

	//直接向game模块
	game.ChanRPC.Go("Ping", args[0], args[1])
}

func handleLogin(args []interface{}) {
	// 收到的 Hello 消息
	m := args[0].(*pb_msg.LoginC2S)
	// 输出收到的消息的内容
	log.Debug("handleLogin : %v", m.LoginInfo.Id)

	//直接向game模块
	game.ChanRPC.Go("UserLogin", args[0], args[1])
}
