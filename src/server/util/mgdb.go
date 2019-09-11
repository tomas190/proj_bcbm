package util

import (
	"context"
	"github.com/name5566/leaf/log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"time"
)

// 数据库客户端
type MgoC struct {
	*mongo.Client
}

// "mongodb://localhost:27017"
func NewMgoC(url string) (*MgoC, error) {
	client, err := mongo.NewClient(options.Client().ApplyURI(url))
	if err != nil {
		log.Error("新建数据库客户端错误", err)
		return nil, err
	}

	log.Debug("数据库客户端 %+v 创建成功...", url)
	return &MgoC{client}, err
}

func (m *MgoC) Init() error {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err := m.Connect(ctx)
	if err != nil {
		log.Error("数据库连接错误", err)
		return err
	}
	err = m.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Error("ping数据库错误", err)
		return err
	}

	log.Debug("数据库连接成功...")
	return nil
}

func (m *MgoC) CUserInfo() {

}

func (m *MgoC) CUserBet() {

}
