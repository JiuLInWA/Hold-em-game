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
//192.168.43.1
const Host = "10.63.60.96"
const TCPPort = "8888"
const WSPort = "8889"

func main() {
	wsTest()
}

func NewTCPConn() net.Conn {
	conn, err := net.Dial("tcp", Host+":"+TCPPort)
	if err != nil {
		fmt.Println("[dial tcp]", err)
	}

	return conn
}

func tcpMsg() []byte {
	m := wsMsg()
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

func wsMsg() []byte {
	// 记得一定要对应消息号 在FindMsgId()函数
	message := &pb_msg.LoginC2S{

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
	id := findMsgID(fmt.Sprintf("%v", reflect.TypeOf(message)))
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

func findMsgID(t string) uint16 {
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

func tcpTest() {
	conn := NewTCPConn()
	m := tcpMsg()
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

func wsTest() {
	conn, _, err := websocket.DefaultDialer.Dial("ws://"+Host+":"+WSPort, nil)
	if err != nil {
		fmt.Println("[dial ws]", err)
		panic("[dial ws]")
		return
	}

	fmt.Println("conn success")

	m := wsMsg()
	fmt.Println(string(m))
	err = conn.WriteMessage(websocket.BinaryMessage, m)
	if err != nil {
		fmt.Println("[write error] ", err)
	}
}



func ws2Test() {
	c := NewWebsocketClient(Host+":"+WSPort, "")
	err := c.SendMessage(wsMsg())
	if err != nil {
		fmt.Println("[ws2Test send message error]")
	}
}

type Client struct {
	Host string
	Path string
}

func NewWebsocketClient(host, path string) *Client {
	return &Client{
		Host: host,
		Path: path,
	}
}

func (c *Client) SendMessage(body []byte) error {
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




