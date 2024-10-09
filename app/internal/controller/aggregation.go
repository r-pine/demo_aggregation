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
	contracts := map[string]string{
		"stonfi": "EQCcD96ywHvlXBjuf4ihiGyH66QChHesNyoJSQ6WKKqob3Lh",
		// "private": "EQCcD96ywHvlXBjuf4ihiGyH66QChHesNyoJSQ6WKKqob3Lh",
		"dedust": "EQBPo45inIbFXiUt8I8xrakPRB1aXZ-wzNOJfIhfQgd2rJ-z",
	}
	for k, v := range contracts {
		go c.aggregation.RunAggregation(k, v)
	}

	ctx.JSON(http.StatusOK, &obj)
}
