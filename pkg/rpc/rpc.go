package rpc

import (
	"context"
	"github.com/AlpherJang/mcache/pkg/cache"
	"github.com/AlpherJang/mcache/pkg/proto"
	"time"
)

type CacheRpc struct {
}

// RegisterTable create a table named with user custom define
func (c *CacheRpc) RegisterTable(ctx context.Context, req *proto.RegisterTableReq) (*proto.RegisterTableResp, error) {
	_ = cache.Cache(req.GetData().GetName(), time.Duration(req.GetData().GetExpireTime()))
	return &proto.RegisterTableResp{Name: req.GetData().GetName()}, nil
}

func NewServer() *CacheRpc {
	return &CacheRpc{}
}
