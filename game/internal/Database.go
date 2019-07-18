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

var (
	session *mgo.Session
)

const (
	dbName = "IM"
	userDB = "user_info"
)

// 连接数据库集合的函数 传入集合 默认连接IM数据库
func initMongoDB() {
	// 此处连接正式线上数据库  下面是模拟的直接连接
	mongoDBDialInfo := &mgo.DialInfo{
		Addrs:    []string{conf.Server.MongoDBAddr},
		Timeout:  60 * time.Second,
		Database: conf.Server.MongoDBAuth,
		Username: conf.Server.MongoDBUser,
		Password: conf.Server.MongoDBPwd,
	}

	var err error
	session, err = mgo.DialWithInfo(mongoDBDialInfo)
	if err != nil {
		log.Fatal("数据库连接失败: %v ", err)
	}
	log.Debug("数据库连接成功~")

	//打开数据库
	session.SetMode(mgo.Monotonic, true)

}

func connect(dbName, cName string) (*mgo.Session, *mgo.Collection) {
	s := session.Copy()
	c := s.DB(dbName).C(cName)
	return s, c
}

// 玩家基础信息
type PlayerInfo struct {
	Id      string  // 玩家ID
	Name    string  // 玩家名字
	Face    string  // 玩家头像
	Balance float64 // 账户余额
}

// 查找用户信息
func FindUserInfo(m *pb_msg.LoginC2S) (*PlayerInfo, error) {
	//根据发送ID查表
	s, c := connect(dbName, userDB)
	defer s.Close()

	ud := &PlayerInfo{}

	err := c.Find(bson.M{"id": m.LoginInfo.Id}).One(ud)
	if err != nil {
		fmt.Println("not Found UserInfoData ~ ")
		playInfo, err := InsertUserInfo(m)
		log.Debug("InsertUserInfoData 插入用户信息成功~")
		return playInfo, err
	}

	fmt.Println("Find UserInfoData ~ ")
	return ud, err
}

// 插入用户信息
func InsertUserInfo(m *pb_msg.LoginC2S) (*PlayerInfo, error) {
	s, c := connect(dbName, userDB)
	defer s.Close()

	p := &PlayerInfo{
		Id:      *m.LoginInfo.Id,
		Name:    *m.LoginInfo.Id,
		Face:    "https://timgsa.baidu.com/timg?image&quality=80&size=b9999_10000&sec=1563442244793&di=945b2375a80b622e2ff3d83e6ac2153b&imgtype=0&src=http%3A%2F%2F1814.img.pp.sohu.com.cn%2Fimages%2Fblog%2F2008%2F11%2F1%2F13%2F20%2F11dfe567377g213.jpg",
		Balance: 4000,
	}

	err := c.Insert(p)
	return p, err
}

func (p *Player) update() error {
	s, c := connect(dbName, userDB)
	defer s.Close()

	data := bson.M{"id": p.ID}

	ud := &PlayerInfo{
		Face:    p.headImg,
		Balance: p.balance,
	}
	err := c.Update(data, ud)
	return err
}
