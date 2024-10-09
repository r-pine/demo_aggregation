package controller

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/r-pine/demo_aggregation/app/internal/entity"
)

func (c *Controller) Aggregation(ctx *gin.Context) {
	data, err := c.sc.Get("states")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	var res *entity.Aggregation
	if err := json.Unmarshal([]byte(data), &res); err != nil {
		ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, res)
}
