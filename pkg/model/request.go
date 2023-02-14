package model

import "time"

type CreateTableReq struct {
	Name       string        `json:"name"`
	ExpireTime time.Duration `json:"expireTime"`
}

type UpdateCacheReq struct {
	Name string `json:"name"`
	// 缓存的key
	Key string `json:"key"`
	// 缓存的data
	Value string `json:"value"`
}

type AddCacheReq struct {
	// 缓存的key
	Key string `json:"key"`
	// 缓存的data
	Value string `json:"value"`
}
