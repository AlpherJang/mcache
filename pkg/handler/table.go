package handler

import (
	"net/http"

	"github.com/AlpherJang/mcache/pkg/cache"
	"github.com/AlpherJang/mcache/pkg/common/errs"
	"github.com/AlpherJang/mcache/pkg/model"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type TableStruct struct {
}

func NewTableStruct() *TableStruct {
	return &TableStruct{}
}

func (t *TableStruct) register(ctx *gin.Context) {
	var req model.CreateTableReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Errorf("err info is %s", err.Error())
		ctx.AbortWithError(errs.ParamErr.Code(), errs.ParamErr.ToError())
		return
	}
	_ = cache.Cache(req.Name, req.ExpireTime)
	ctx.JSON(http.StatusOK, gin.H{"msg": "operator success"})
}

func (t *TableStruct) listTable(ctx *gin.Context) {
	tables, _ := cache.ListTable()
	ctx.JSON(http.StatusOK, gin.H{"data": tables})
}

func (t *TableStruct) Registry(r *gin.Engine) {
	table := r.Group("table")
	table.PUT("", t.register)
	table.GET("/", t.listTable)
}
