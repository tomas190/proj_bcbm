package dao

import "testing"

func TestMongoHelper_UserCount(t *testing.T) {
	m := MongoHelper{}
	m.UserCount()
}

func TestCacheHelper_Set(t *testing.T) {
	c := CacheHelper{}
	c.Set()
}
