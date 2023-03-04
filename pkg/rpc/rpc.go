package rpc

import (
	"context"
	"github.com/AlpherJang/mcache/pkg/cache"
	"github.com/AlpherJang/mcache/pkg/common/errs"
	"github.com/AlpherJang/mcache/pkg/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

type CacheRpc struct {
}

func (c *CacheRpc) GetCache(ctx context.Context, req *proto.GetCacheReq) (*proto.GetCacheResp, error) {
	table, err := cache.GetTable(req.GetTableName())
	if err != nil {
		return nil, err
	}
	cacheData, err := table.Get(req.GetCacheName())
	if err != nil {
		return nil, err
	}
	return &proto.GetCacheResp{Data: &proto.CacheInfo{
		Key:   req.GetCacheName(),
		Value: cacheData.(string),
	}}, nil
}

func (c *CacheRpc) ListCache(_ context.Context, req *proto.ListCacheReq) (*proto.ListCacheResp, error) {
	table, err := cache.GetTable(req.GetTableName())
	if err != nil {
		return nil, err
	}
	cacheData := table.ListCacheInfo()
	return &proto.ListCacheResp{List: cacheData}, nil
}

func (c *CacheRpc) DeleteCache(ctx context.Context, req *proto.DeleteCacheReq) (*emptypb.Empty, error) {
	table, err := cache.GetTable(req.GetTableName())
	if err != nil {
		return nil, errs.TableNotFoundErr
	}
	if !table.Delete(req.GetTableName()) {
		return nil, errs.CacheDeleteErr
	}
	return nil, nil
}

func (c *CacheRpc) AddCache(ctx context.Context, req *proto.AddCacheReq) (*emptypb.Empty, error) {
	table, err := cache.GetTable(req.GetTableName())
	if err != nil {
		return nil, err
	}
	if success, err := table.Add(req.GetData().GetKey(), req.GetData().GetValue()); !success {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

// RegisterTable create a table named with user custom define
func (c *CacheRpc) RegisterTable(ctx context.Context, req *proto.RegisterTableReq) (*proto.RegisterTableResp, error) {
	_ = cache.Cache(req.GetData().GetName(), req.GetData().GetExpireTime().AsDuration())
	return &proto.RegisterTableResp{Name: req.GetData().GetName()}, nil
}

func NewServer() *CacheRpc {
	return &CacheRpc{}
}
