package rpc

import (
	"context"
	"github.com/AlpherJang/mcache/pkg/cache"
	"github.com/AlpherJang/mcache/pkg/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

// RegisterTable create a table named with user custom define
func (c *CacheRpc) RegisterTable(ctx context.Context, req *proto.RegisterTableReq) (*proto.RegisterTableResp, error) {
	_ = cache.Cache(req.GetData().GetName(), req.GetData().GetExpireTime().AsDuration())
	return &proto.RegisterTableResp{Name: req.GetData().GetName()}, nil
}

func (c *CacheRpc) ListTable(_ context.Context, req *proto.ListTableReq) (*proto.ListTableResp, error) {
	tables, _ := cache.ListTable(cache.NewTableNameFilter(req.GetTableName()))
	return &proto.ListTableResp{TableList: tables}, nil
}

func (c *CacheRpc) DropTable(_ context.Context, req *proto.DropTableReq) (*emptypb.Empty, error) {
	cache.DropTable(req.GetName())
	return nil, nil
}
