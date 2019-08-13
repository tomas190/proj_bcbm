package dao

import "go.mongodb.org/mongo-driver/mongo"

type MongoHelper struct {
	mongo.Client
}

func (mgo *MongoHelper) UserCount() int {
	// fixme
	return 9
}
