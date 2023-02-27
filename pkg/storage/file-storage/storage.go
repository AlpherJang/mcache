package filestorage

import (
	"context"

	"github.com/AlpherJang/mcache/pkg/storage/oper"
)

type storage struct {
	localPath   string
	needPreLoad bool
}

func NewStorage(localPath string, needPreLoad bool) Storage {
	return &storage{
		localPath:   localPath,
		needPreLoad: needPreLoad,
	}
}

// Located 数据落盘
func (s *storage) Located(ctx context.Context) error {
	return nil
}

// PreLoad 载入日志
func (s *storage) PreLoad(ctx context.Context) error {
	return nil
}

// Append 追加日志
func (s *storage) Append(ctx context.Context, data oper.Operate) error {
	return nil
}
