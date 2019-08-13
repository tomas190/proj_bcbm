package dao

import "github.com/patrickmn/go-cache"

type CacheHelper struct {
	cache.Cache
}

func (mc *CacheHelper) get() (interface{}, bool) {
	return mc.Get("foo")
}

func (mc *CacheHelper) set() {

}