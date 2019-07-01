package internal


import (
	"C"
	"fmt"
	"github.com/name5566/leaf/log"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"server/conf"
	pb_msg "server/msg/Protocal"
	"time"
)

// 连接数据库集合的函数 传入集合 默认连接IM数据库
func connect(cName string) (*mgo.Session, *mgo.Collection) {
	// 此处连接正式线上数据库  下面是模拟的直接连接
	mongoDBDialInfo := &mgo.DialInfo{
		Addrs:    []string{conf.Server.MongoDBIP},
		Timeout:  60 * time.Second,
		Database: "IM",
		Username: "im",
		Password: "123456",
	}
	session, err := mgo.DialWithInfo(mongoDBDialInfo)
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}
	log.Debug("数据库连接成功~")
	//打开数据库
	session.SetMode(mgo.Monotonic, true)

	return session, session.DB("IM").C(cName)
}

// 查找用户信息
func FindUserInfoData(m *pb_msg.LoginC2S) (pb_msg.PlayerInfo, error) {
	//根据发送ID查表
	s, c := connect("user_info")
	defer s.Close()

	ud := &pb_msg.PlayerInfo{}

	err := c.Find(bson.M{"id": m.LoginInfo.Id}).One(ud)
	if err != nil {
		fmt.Println("not Found UserInfoData")
		playInfo, err := InsertUserInfoData(m)
		log.Debug("InsertUserInfoData 插入用户信息成功~")
		return playInfo, err
	} else {
		fmt.Println("Find UserInfoData")
		err1 := c.Find(bson.M{"id": m.LoginInfo.Id}).One(ud)
		if err1 != nil {
			panic(err1)
		}
		return *ud, err1
	}
	return *ud, err
}


// 玩家基础信息
type PlayerInfo struct {
	Id      string  // 玩家ID
	Name    string  // 玩家名字
	HeadImg string  // 玩家头像
	Balance float64 // 账户余额
}

// 插入用户信息
func InsertUserInfoData(m *pb_msg.LoginC2S) (pb_msg.PlayerInfo, error) {
	s, c := connect("user_info")
	defer s.Close()

	playerInfo := &pb_msg.PlayerInfo{
		Id:      m.LoginInfo.Id,
		Name:    m.LoginInfo.Id,
		HeadImg: "https://www.andreyapopov.com/Portfolio/Conceptual/1",
		Balance: 4000,
	}

	p := new(PlayerInfo)
	p.Id = playerInfo.Id
	p.Name = playerInfo.Name
	p.HeadImg = playerInfo.HeadImg
	p.Balance = playerInfo.Balance

	err := c.Insert(p)
	return *playerInfo, err
}
