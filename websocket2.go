package main

import (
	"encoding/binary"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
	websocket2 "golang.org/x/net/websocket"
	"net"
	"net/url"
	"reflect"
	pb_msg "server/msg/Protocal"
)

//10.63.60.96
const Host1 = "47.75.183.211"
const TCPPort1 = "8888"
const WSPort1 = "8889"

func main() {
	wsTest1()
}

func NewTCPConn1() net.Conn {
	conn, err := net.Dial("tcp", Host1+":"+TCPPort1)
	if err != nil {
		fmt.Println("[dial tcp]", err)
	}

	return conn
}

func tcpMsg1() []byte {
	m := wsMsg1()
	// 使用TCP协议传输要加入消息长度
	// 封入 len 字段
	// len 包含了 id 的长度！！！
	// -------------------------
	// |len | id | protobuf message |
	// -------------------------
	msgLen := make([]byte, 2)
	binary.BigEndian.PutUint16(msgLen, uint16(len(m)))
	m = append(msgLen, m...)

	return m
}

func wsMsg1() []byte {
	// 记得一定要对应消息号 在FindMsgId()函数
	message := &pb_msg.QuickStartC2S{
		RoomInfo: &pb_msg.RoomInfo{
			CfgId:       "1",
			MaxPlayer:   5,
			ActionTimeS: 15,
		},
	}

	payload, err := proto.Marshal(message)
	if err != nil {
		fmt.Println("Marshal error ", err)
	}

	// 创建一个新的字节数组，也可以在payload操作
	m := make([]byte, len(payload))
	binary.BigEndian.PutUint16(m, uint16(len(payload)))

	// 封入 id 字段
	// -------------------------
	// | id | protobuf message |
	// -------------------------
	// tagId := []byte{0x0, 0x0}
	id := findMsgID1(fmt.Sprintf("%v", reflect.TypeOf(message)))
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

func findMsgID1(t string) uint16 {
	// fixme 服务器中打印这个表
	msgType2ID := map[string]uint16{
		"*pb_msg.PingC2S":        0,
		"*pb_msg.PongS2C":        1,
		"*pb_msg.SvrMsgS2C":      2,
		"*pb_msg.LoginC2S":       3,
		"*pb_msg.LoginResultS2C": 4,
		"*pb_msg.QuickStartC2S":  5,
		"*pb_msg.CreateRoomC2S":  6,
		"*pb_msg.JoinRoomC2S":    7,
		"*pb_msg.EnterRoomS2C":   8,
		"*pb_msg.ExitRoomC2S":    9,
		"*pb_msg.ExitRoomS2C":    10,
	}

	if id, ok := msgType2ID[t]; ok {
		return id
	}

	return 1024
}

func tcpTest1() {
	conn := NewTCPConn1()
	m := tcpMsg1()
	// 打印
	for i, b := range m {
		fmt.Println(i, "-", b, string(b))
	}

	// 写入到连接
	_, err := conn.Write(m)
	if err != nil {
		fmt.Println("[write error] ", err)
	}
}

func wsTest1() {
	conn, _, err := websocket.DefaultDialer.Dial("ws://"+Host1+":"+WSPort1, nil)
	if err != nil {
		fmt.Println("[dial ws]", err)
		panic("[dial ws]")
		return
	}

	fmt.Println("conn success")

	m := wsMsg1()
	fmt.Println(string(m))
	err = conn.WriteMessage(websocket.BinaryMessage, m)
	if err != nil {
		fmt.Println("[write error] ", err)
	}
}

func ws2Test1() {
	c := NewWebsocketClient1(Host1+":"+WSPort1, "")
	err := c.SendMessage1(wsMsg1())
	if err != nil {
		fmt.Println("[ws2Test send message error]")
	}
}

type Client1 struct {
	Host string
	Path string
}

func NewWebsocketClient1(host, path string) *Client1 {
	return &Client1{
		Host: host,
		Path: path,
	}
}

func (c *Client1) SendMessage1(body []byte) error {
	u := url.URL{Scheme: "ws", Host: c.Host, Path: c.Path}
	fmt.Println(u.String())
	ws, err := websocket2.Dial(u.String(), "", "http://"+c.Host+"/")

	defer ws.Close() //关闭连接
	if err != nil {
		fmt.Println("[dial error]", err)
		return err
	}

	_, err = ws.Write(body)
	if err != nil {
		return err
	}

	fmt.Println("写入完成")
	return nil
}
