package internal

type BetDB struct {
	User       UserDB  `bson:"User"`
	Area       uint32  `bson:"Area"`
	AreaStr    string  `bson:"AreaStr"`
	Chip       uint32  `bson:"Chip"`
	ChipAmount float64 `bson:"ChipAmount"`
}

type UserDB struct {
	UserID   uint32  `bson:"UserID" json:"UserID"`   // 用户id
	NickName string  `bson:"NickName" json:"UserID"` // 用户昵称
	Avatar   string  `bson:"Avatar" json:"UserID"`   // 用户头像
	Balance  float64 `bson:"Balance"json:"UserID"`   // 用户金额
}
