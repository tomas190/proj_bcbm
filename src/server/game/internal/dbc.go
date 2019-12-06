package internal

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"proj_bcbm/src/server/constant"
	"proj_bcbm/src/server/log"
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
	ctx, _ := context.WithTimeout(context.Background(), 2*time.Second)
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

	log.Debug("数据库连接成功...")
	return nil
}

// 插入用户信息
func (m *MgoC) CUserInfo(u interface{}) error {
	collection := m.Database(constant.DBName).Collection("users")
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)

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
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)

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
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)

	count, err := collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		log.Debug("查找用户数量错误 %+v", err)
		return 0, err
	}
	return count, nil
}

func (m *MgoC) CUserSettle(bet interface{}) error {
	collection := m.Database(constant.DBName).Collection("settles")
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
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
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)

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

func (m *MgoC) RProfitPool() ProfitDB {
	collection := m.Database(constant.DBName).Collection("profits")
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)

	opt := options.FindOne()
	opt.SetSort(bson.M{"UpdateTime": -1})

	var lastProfit ProfitDB
	err := collection.FindOne(ctx, bson.M{}, opt).Decode(&lastProfit)
	if err != nil {
		log.Debug("查找最新盈余池数据失败 %+v", err)
		return lastProfit
	}

	return lastProfit
}

func (m *MgoC) UProfitPool(lose, win float64, rid uint32) error {
	collection := m.Database(constant.DBName).Collection("profits")
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)

	opt := options.FindOne()
	opt.SetSort(bson.M{"UpdateTime": -1})
	userCount, _ := m.RUserCount()

	//init the loc
	loc, _ := time.LoadLocation("Asia/Shanghai")
	//set timezone,
	now := time.Now().In(loc)

	fmt.Print(now)
	var lastProfit ProfitDB
	err := collection.FindOne(ctx, bson.M{}, opt).Decode(&lastProfit)
	if err != nil {
		log.Debug("未查找到盈余池数据 %+v", err)
		var newProfit = ProfitDB{
			UpdateTime:     time.Now(),
			UpdateTimeStr:  now.Format("2006-01-02T15:04:05"),
			PlayerThisWin:  win,
			PlayerThisLost: lose,
			PlayerAllWin:   win,
			PlayerAllLost:  lose,
			RoomID:         rid,
			Profit:         lose - win,
			PlayerNum:      uint32(userCount),
		}
		res, err := collection.InsertOne(ctx, newProfit)
		if err != nil {
			log.Debug("插入第一条盈余数据 %+v", err)
		}

		log.Debug("插入第一条盈余数据 %+v", res)
	}

	newLost := lastProfit.PlayerAllLost + lose
	newWin := lastProfit.PlayerAllWin + win
	newCount := userCount
	newProfit := newLost - newWin*(1.0+constant.HouseEdge) - float64(userCount)*constant.GiftAmount

	newRecord := ProfitDB{
		UpdateTime:     time.Now(),
		UpdateTimeStr:  now.Format("2006-01-02T15:04:05"),
		PlayerThisLost: lose,
		PlayerThisWin:  win,
		PlayerAllWin:   newWin,
		PlayerAllLost:  newLost,
		Profit:         newProfit,
		RoomID:         rid,
		PlayerNum:      uint32(newCount),
	}

	res, err := collection.InsertOne(ctx, newRecord)
	if err != nil {
		log.Debug("插入盈余数据 %+v", err)
	}
	log.Debug("插入盈余数据 %+v", res)

	return nil
}

//InsertAccessData 插入运营数据接入
func (m *MgoC) InsertAccess(data *PlayerDownBetRecode) error {
	collection := m.Database(constant.DBName).Collection("accessData")
	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)

	res, err := collection.InsertOne(ctx, data)
	if err != nil {
		log.Error("<----- 运营接入数据插入失败 ~ ----->:%+v", err)
		return err
	}

	log.Debug("运营接入数据插入成功: %+v", res)

	return nil
}

//GetDownRecodeList 获取运营数据接入
//func (m *MgoC)GetDownRecodeList(skip, limit int, selector bson.M, sortBy string) ([]PlayerDownBetRecode, int, error) {
//	collection := m.Database(constant.DBName).Collection("accessData")
//	ctx, _ := context.WithTimeout(context.Background(), 1*time.Second)
//
//	var wts []PlayerDownBetRecode
//
//	n, err := collection.Find(ctx, selector).Count()
//	if err != nil {
//		return nil, 0, err
//	}
//	err = collection.Find(selector).Sort(sortBy).Skip(skip).Limit(limit).All(&wts)
//	if err != nil {
//		return nil, 0, err
//	}
//	return wts, n, nil
//}
