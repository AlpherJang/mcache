package model

import "time"

type CreateTableReq struct {
	Name       string        `json:"name"`
	ExpireTime time.Duration `json:"expireTime"`
}
