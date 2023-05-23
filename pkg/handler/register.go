package handler

import (
	"github.com/gin-gonic/gin"
)

func Register(r *gin.Engine) {
	NewDataStruct().Registry(r)
	NewTableStruct().Registry(r)
}
