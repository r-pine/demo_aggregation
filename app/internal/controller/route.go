package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/r-pine/demo_aggregation/app/internal/blockchain"
	"github.com/r-pine/demo_aggregation/app/internal/service"
	"github.com/r-pine/demo_aggregation/app/pkg/config"
	"github.com/r-pine/demo_aggregation/app/pkg/logging"
)

type Controller struct {
	log         logging.Logger
	sc          service.Service
	cfg         config.Config
	aggregation *blockchain.Aggregation
}

func NewController(
	log logging.Logger,
	sc service.Service,
	cfg config.Config,
	aggregation *blockchain.Aggregation,
) *Controller {
	return &Controller{
		log:         log,
		sc:          sc,
		cfg:         cfg,
		aggregation: aggregation,
	}
}

func (c *Controller) InitRoutes(r *gin.Engine) *gin.Engine {

	api := r.Group("api/")
	{
		api.GET("healthcheck", c.Healthcheck)
		api.GET("aggregation", c.Aggregation)
		api.POST("swap-payload", c.GetSwapPayload)
	}

	return r
}
