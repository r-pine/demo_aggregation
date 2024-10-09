package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (c *Controller) Aggregation(ctx *gin.Context) {
	var obj struct {
		Msg string `json:"msg"`
	}
	obj.Msg = "Aggregation"
	c.aggregation.RunAggregation()
	ctx.JSON(http.StatusOK, &obj)
}
