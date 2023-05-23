package handler

import (
	"net/http"

	"github.com/AlpherJang/mcache/pkg/cache"
	"github.com/AlpherJang/mcache/pkg/common/errs"
	"github.com/AlpherJang/mcache/pkg/model"
	"github.com/gin-gonic/gin"
)

type DataStruct struct {
}

func NewDataStruct() *DataStruct {
	return &DataStruct{}
}

func (d *DataStruct) list(ctx *gin.Context) {
	tableName := ctx.Param("table")
	table, err := cache.GetTable(tableName)
	if err != nil {
		ctx.AbortWithError(err.Code(), err.ToError())
		return
	}
	cacheData := table.List()
	ctx.JSON(http.StatusOK, gin.H{"data": cacheData})

}

func (d *DataStruct) get(ctx *gin.Context) {
	tableName := ctx.Param("table")
	key := ctx.Param("key")
	table, err := cache.GetTable(tableName)
	if err != nil {
		ctx.AbortWithError(err.Code(), err.ToError())
		return
	}
	cacheData, err := table.Get(key)
	if err != nil {
		ctx.AbortWithError(err.Code(), err.ToError())
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": cacheData})
}

func (d *DataStruct) add(ctx *gin.Context) {
	tableName := ctx.Param("table")
	table, err := cache.GetTable(tableName)
	if err != nil {
		ctx.AbortWithError(err.Code(), err.ToError())
		return
	}
	var req model.AddCacheReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.AbortWithError(errs.ParamErr.Code(), errs.ParamErr.ToError())
		return
	}
	if success, err := table.Add(req.Key, req.Value); !success {
		ctx.AbortWithError(err.Code(), err.ToError())
		return
	}
	ctx.JSON(http.StatusOK, nil)
}

func (d *DataStruct) update(ctx *gin.Context) {
	var req model.UpdateCacheReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.AbortWithError(errs.ParamErr.Code(), errs.ParamErr.ToError())
		return
	}
	table, err := cache.GetTable(req.Name)
	if err != nil {
		ctx.AbortWithError(err.Code(), err.ToError())
		return
	}
	if err = table.Update(req.Key, req.Value); err != nil {
		ctx.AbortWithError(err.Code(), err.ToError())
		return
	}
	ctx.JSON(http.StatusOK, nil)
}

func (d *DataStruct) delete(ctx *gin.Context) {
	tableName := ctx.Param("table")
	key := ctx.Param("key")
	table, err := cache.GetTable(tableName)
	if err != nil {
		ctx.AbortWithError(errs.TableNotFoundErr.Code(), errs.TableNotFoundErr.ToError())
		return
	}
	if !table.Delete(key) {
		ctx.AbortWithError(errs.CacheDeleteErr.Code(), errs.CacheDeleteErr.ToError())
		return
	}
	ctx.JSON(http.StatusOK, nil)
}

func (d *DataStruct) Registry(r *gin.Engine) {
	data := r.Group("data")
	data.GET("/:table", d.list)
	data.GET("/:table/:key", d.get)
	data.PUT("/:table", d.add)
	data.POST("/:table", d.update)
	data.DELETE("/:table/:key", d.delete)
}
