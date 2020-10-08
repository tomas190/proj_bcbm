package internal

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"gopkg.in/mgo.v2/bson"
	"proj_bcbm/src/server/conf"
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
	ctx, _ := context.WithTimeout(context.Background(), 60*time.Second)
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
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)

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
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)

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
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)

	count, err := collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		log.Debug("查找用户数量错误 %+v", err)
		return 0, err
	}
	return count, nil
}

func (m *MgoC) CUserSettle(bet interface{}) error {
	collection := m.Database(constant.DBName).Collection("settles")
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
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
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)

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
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)

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
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)

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
		log.Debug("插入第一条盈余数据~")
	}

	newLost := lastProfit.PlayerAllLost + lose
	newWin := lastProfit.PlayerAllWin + win
	newCount := userCount
	newProfit := (newLost - (newWin * 1)) * 0.5
	log.Debug("newProfit:%v", newLost-(newWin*1))
	log.Debug("盈余数据为： %+v", newProfit)

	SurPool := &SurPool{}
	SurPool.GameId = conf.Server.GameID
	SurPool.SurplusPool = newProfit
	SurPool.PlayerTotalLoseWin = newLost - newWin
	SurPool.PlayerTotalLose = newLost
	SurPool.PlayerTotalWin = newWin
	SurPool.TotalPlayer = userCount
	SurPool.FinalPercentage = 0.5
	SurPool.PercentageToTotalWin = 1
	SurPool.CoefficientToTotalPlayer = userCount * 0
	SurPool.PlayerLoseRateAfterSurplusPool = 0.7
	SurPool.DataCorrection = 0
	m.FindSurPool(SurPool)

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

func (m *MgoC) FindSurPool(data *SurPool) {
	collection := m.Database(constant.DBName).Collection("surplus-pool")
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)

	//collection.DeleteMany(ctx, bson.M{}, nil)

	opt := options.Find()

	cur, err := collection.Find(ctx, bson.M{}, opt)
	if err != nil {
		log.Debug("获取用户數據错误 %+v", err)
		_ = m.InsertSurPool(data)
	} else {
		var sur SurPool
		for cur.Next(ctx) {
			var wts SurPool
			_ = cur.Decode(&wts)
			sur = wts
		}
		data.SurplusPool = (data.PlayerTotalLose - (data.PlayerTotalWin * sur.PercentageToTotalWin) - float64(data.TotalPlayer*sur.CoefficientToTotalPlayer) + sur.DataCorrection) * sur.FinalPercentage
		data.FinalPercentage = sur.FinalPercentage
		data.PercentageToTotalWin = sur.PercentageToTotalWin
		data.CoefficientToTotalPlayer = sur.CoefficientToTotalPlayer
		data.PlayerLoseRateAfterSurplusPool = sur.PlayerLoseRateAfterSurplusPool
		data.DataCorrection = sur.DataCorrection
		_ = m.UpdateSurPool(data)
	}

	//count, _ := collection.CountDocuments(ctx, bson.M{})
	//log.Debug("FindSurPool 数量:%v", count)
}

func (m *MgoC) InsertSurPool(data *SurPool) error {
	collection := m.Database(constant.DBName).Collection("surplus-pool")
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)

	res, err := collection.InsertOne(ctx, data)
	if err != nil {
		log.Error("<----- 插入SurPool失败 ~ ----->:%+v", err)
		return err
	}

	log.Debug("<----- 插入SurPool成功 ~ ----->: %+v,%v", res, data)
	return nil
}

func (m *MgoC) UpdateSurPool(data *SurPool) error {
	collection := m.Database(constant.DBName).Collection("surplus-pool")
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)

	res, err := collection.ReplaceOne(ctx, bson.M{}, data)
	if err != nil {
		log.Error("<----- 更新 SurPool数据失败 ~ ----->:%v", err)
		return err
	}
	log.Debug("<----- 更新SurPool数据成功 ~ ----->:%v", res)
	return nil
}

//InsertAccessData 插入运营数据接入
func (m *MgoC) InsertAccess(data *PlayerDownBetRecode) error {
	collection := m.Database(constant.DBName).Collection("accessData")
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)

	_, err := collection.InsertOne(ctx, data)
	if err != nil {
		log.Error("<----- 运营接入数据插入失败 ~ ----->:%+v", err)
		return err
	}

	log.Debug("运营接入数据插入成功: %+v", data)
	return nil
}

//GetDownRecodeList 获取运营数据接入
func (m *MgoC) GetDownRecodeList(skip, limit int, selector bson.M, sortBy string) ([]PlayerDownBetRecode, int, error) {
	collection := m.Database(constant.DBName).Collection("accessData")
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)

	var wts []PlayerDownBetRecode

	opt := options.Find()
	opt.SetSort(sortBy)
	opt.SetSkip(int64(skip))
	opt.SetLimit(int64(limit))

	count, err := collection.CountDocuments(ctx, selector)
	if err != nil {
		log.Debug("获取用户数量错误 %+v", err)
	}
	log.Debug("获取用户数量 %+v", count)

	cur, err2 := collection.Find(ctx, selector, opt)
	if err2 != nil {
		log.Debug("获取用户數據错误 %+v", err2)
	}

	for cur.Next(ctx) {
		var PRecode PlayerDownBetRecode
		err := cur.Decode(&PRecode)
		if err != nil {
			//log.Debug("数据库数据解码错误 %+v", err)
		}
		wts = append(wts, PRecode)
	}

	return wts, int(count), nil
}

//GetDownRecodeList 获取盈余池数据
func (m *MgoC) GetSurPoolData(selector bson.M) (SurPool, error) {
	collection := m.Database(constant.DBName).Collection("surplus-pool")
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)

	opt := options.Find()

	var sur SurPool

	cur, err2 := collection.Find(ctx, selector, opt)
	if err2 != nil {
		log.Debug("获取用户數據错误 %+v", err2)
	}

	for cur.Next(ctx) {
		var wts SurPool
		_ = cur.Decode(&wts)
		sur = wts
	}
	return sur, nil
}

type ChipDownBet struct {
	Chip1    int32
	Chip10   int32
	Chip100  int32
	Chip500  int32
	Chip1000 int32
}

type RobotDATA struct {
	RoomId   uint32       `json:"room_id" bson:"room_id"`
	RoomTime int64        `json:"room_time" bson:"room_time"`
	RobotNum int          `json:"robot_num" bson:"robot_num"`
	AreaX1   *ChipDownBet `json:"area_x_1" bson:"area_x_1"`
	AreaX2   *ChipDownBet `json:"area_x_2" bson:"area_x_2"`
	AreaX3   *ChipDownBet `json:"area_x_3" bson:"area_x_3"`
	AreaX4   *ChipDownBet `json:"area_x_4" bson:"area_x_4"`
	AreaX5   *ChipDownBet `json:"area_x_5" bson:"area_x_5"`
	AreaX6   *ChipDownBet `json:"area_x_6" bson:"area_x_6"`
	AreaX7   *ChipDownBet `json:"area_x_7" bson:"area_x_7"`
	AreaX8   *ChipDownBet `json:"area_x_8" bson:"area_x_8"`
}

//InsertRobotData 机器人数据插入
func (m *MgoC) InsertRobotData(data *RobotDATA) error {
	collection := m.Database(constant.DBName).Collection("robotData")
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)

	_, err := collection.InsertOne(ctx, data)
	if err != nil {
		log.Error("<----- 运营接入数据插入失败 ~ ----->:%+v", err)
		return err
	}

	return nil
}

//GetRobotData 获取机器人数据
func (m *MgoC) GetRobotData() ([]RobotDATA, error) {
	collection := m.Database(constant.DBName).Collection("robotData")
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)

	opt := options.Find()

	var wts []RobotDATA

	currentTime := time.Now()
	startTime := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), 0, 0, 0, 0, currentTime.Location()).Unix()
	endTime := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), 23, 59, 59, 0, currentTime.Location()).Unix()

	selector := bson.M{}
	selector["room_time"] = bson.M{"$gte": startTime, "$lte": endTime}

	cur, err2 := collection.Find(ctx, selector, opt)
	if err2 != nil {
		log.Debug("获取機器人數據错误 %+v", err2)
	}

	for cur.Next(ctx) {
		var PRecode RobotDATA
		err := cur.Decode(&PRecode)
		if err != nil {
		}
		wts = append(wts, PRecode)
	}

	return wts, nil
}
