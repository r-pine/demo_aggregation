package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/r-pine/demo_aggregation/app/internal/blockchain"
)

func (c *Controller) Healthcheck(ctx *gin.Context) {
	api, _, err := blockchain.GetApiClient(ctx)
	if err != nil || api == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"status":  "fail",
			"message": "Rpine Demo Aggregation failed to obtain blockchain API client",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Service Rpine Demo Aggregation is healthy",
	})
}
