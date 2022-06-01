package handler

import (
	"github.com/AlpherJang/mcache/pkg/cache"
	"github.com/AlpherJang/mcache/pkg/common/errs"
	"github.com/AlpherJang/mcache/pkg/model"
	"github.com/gin-gonic/gin"
	"net/http"
)

type DataStruct struct {
}

func (d *DataStruct) list() {

}

func (d *DataStruct) get() {

}

func (d *DataStruct) add() {

}

func (d *DataStruct) update() {

}

func (d *DataStruct) register(ctx *gin.Context) {
	var req model.CreateTableReq
	if err := ctx.ShouldBindJSON(req); err != nil {
		ctx.AbortWithError(errs.ParamErr.Code(), errs.ParamErr.Error())
		return
	}
	_ = cache.Cache(req.Name, req.ExpireTime)
	ctx.JSON(http.StatusOK, nil)
}

func (d *DataStruct) delete(ctx *gin.Context) {
	tableName := ctx.Param("table")
	key := ctx.Param("key")
	table, err := cache.GetTable(tableName)
	if err != nil {
		ctx.AbortWithError(errs.TableNotFoundErr.Code(), errs.TableNotFoundErr.Error())
		return
	}
	if !table.Delete(key) {
		ctx.AbortWithError(errs.CacheDeleteErr.Code(), errs.CacheDeleteErr.Error())
		return
	}
	ctx.JSON(http.StatusOK, nil)
}
