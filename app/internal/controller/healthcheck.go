package controller

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func (c *Controller) Healthcheck(ctx *gin.Context) {
	var obj struct {
		Msg string `json:"msg"`
	}
	obj.Msg = "Hello Rpine Demo Aggregation"
	ctx.JSON(http.StatusOK, &obj)
}
