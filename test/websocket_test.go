package test

import (
	"encoding/binary"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
	pb_msg "server/msg/Protocal"
	"testing"
	"time"
)

const Host = "10.63.60.96"
const TCPPort = "8888"
const WSPort = "8889"

func TestWSPing2(t *testing.T) {

	conn, _, err := websocket.DefaultDialer.Dial("ws://"+Host+":"+WSPort, nil)
	if err != nil {
		fmt.Println("[dial ws]", err)
		panic("[dial ws]")
		return
	}

	fmt.Println("conn success")
	m := pingMsg()
	fmt.Println(string(m))

	for i := 0; i < 100; i++ {
		err = conn.WriteMessage(websocket.BinaryMessage, m)
		if err != nil {
			fmt.Println("[write error] ", err)
		}
		time.Sleep(time.Second * 3)
	}
}

func TestWSPing(t *testing.T) {

	conn, _, err := websocket.DefaultDialer.Dial("ws://"+Host+":"+WSPort, nil)
	if err != nil {
		fmt.Println("[dial ws]", err)
		panic("[dial ws]")
		return
	}

	fmt.Println("conn success")
	m := pingMsg()
	fmt.Println(string(m))

	for i := 0; i < 100; i++ {
		err = conn.WriteMessage(websocket.BinaryMessage, m)
		if err != nil {
			fmt.Println("[write error] ", err)
		}
		time.Sleep(time.Second * 3)
	}
}

func pingMsg() []byte {
	//记得一定要对应消息号 在FindMsgId()函数
	message := &pb_msg.LoginC2S{

	}

	payload, err := proto.Marshal(message)
	if err != nil {
		fmt.Println("Marshal error ", err)
	}

	// 创建一个新的字节数组，也可以在payload操作
	m := make([]byte, len(payload))
	////binary.BigEndian.PutUint16(m, uint16(len(payload)))

	// 封入 id 字段
	// -------------------------
	// | id | protobuf message |
	// -------------------------
	// tagId := []byte{0x0, 0x0}
	id := uint16(0)
	tagId := make([]byte, 2)
	binary.BigEndian.PutUint16(tagId, id)
	m = append(tagId, m...)
	// 封入 payload
	copy(m[2:], payload)

	// 打印
	for i, b := range m {
		fmt.Println(i, "-", b, string(b))
	}

	return m
}
