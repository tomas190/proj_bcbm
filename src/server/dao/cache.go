package dao

import "github.com/patrickmn/go-cache"

type CacheHelper struct {
	cache.Cache
}

func (mc *CacheHelper) Set() {

}
