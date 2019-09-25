package internal

import (
	"context"
	"github.com/name5566/leaf/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"proj_bcbm/src/server/constant"
	"time"
)

// 数据库客户端
type MgoC struct {
	*mongo.Client
}

// "mongodb://localhost:27017"
func NewMgoC(url string) *MgoC {
	client, err := mongo.NewClient(options.Client().ApplyURI(url))
	if err != nil {
		log.Error("新建数据库客户端错误", err)
		return nil
	}

	log.Debug("数据库客户端 %+v 创建成功...", url)
	return &MgoC{client}
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
		log.Error("ping数据库错误 %+v", err)
		return err
	}

	u := UserDB{UserID: 100000001}
	err = m.CUserInfo(u)

	log.Debug("数据库连接成功...")
	return nil
}

// 插入用户信息
func (m *MgoC) CUserInfo(u interface{}) error {
	collection := m.Database(constant.DBName).Collection("users")
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)

	res, err := collection.InsertOne(ctx, u)
	if err != nil {
		log.Error("%+v", err)
		return err
	}
	id := res.InsertedID
	log.Debug("玩家信息已保存 %+v", id)
	return err
}

func (m *MgoC) RUserInfo(userID uint32) error {
	collection := m.Database(constant.DBName).Collection("users")
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)

	var userInfo UserDB

	filter := bson.M{"UserID": userID}
	err := collection.FindOne(ctx, filter).Decode(&userInfo)
	if err != nil {
		log.Debug("查找用户信息错误 %+v", err)
	}

	return err
}

func (m *MgoC) RUserCount() (int64, error) {
	collection := m.Database(constant.DBName).Collection("users")
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)

	count, err := collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		log.Debug("查找用户数量错误 %+v", err)
		return 0, err
	}
	return count, nil
}

func (m *MgoC) UUserInfo() {

}

func (m *MgoC) DUserInfo() {

}

func (m *MgoC) CUserSettle(bet interface{}) error {
	collection := m.Database(constant.DBName).Collection("settles")
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	res, err := collection.InsertOne(ctx, bet)
	if err != nil {
		log.Error("%+v", err)
		return err
	}
	id := res.InsertedID
	log.Debug("用户结算信息已保存 %+v", id)
	return err
}

func (m *MgoC) RUserSettle(userID uint32) ([]SettleDB, error) {
	collection := m.Database(constant.DBName).Collection("settles")
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)

	var res []SettleDB
	filter := bson.M{"User.UserID": userID}
	opt := options.Find()
	opt.SetLimit(20)
	opt.SetSort(bson.M{"_id": -1})

	cur, err := collection.Find(ctx, filter, opt)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var result SettleDB
		err := cur.Decode(&result)
		if err != nil {
			log.Debug("数据库数据解码错误 %+v", err)
		}
		res = append(res, result)
	}
	return res, nil
}
