package handler

import (
	"github.com/gin-gonic/gin"
)

func Register(r *gin.Engine) {
	dataHandler := &DataStruct{}
	dataHandler.Registry(r)
	tableHandler := &TableStruct{}
	tableHandler.Registry(r)
}
