package filestorage

import "context"

type Storage interface {
	Located(ctx context.Context) error
	PreLoad(ctx context.Context) error
}
