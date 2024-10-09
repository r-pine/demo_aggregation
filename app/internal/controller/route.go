package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/r-pine/demo_aggregation/app/internal/service"
	"github.com/r-pine/demo_aggregation/app/pkg/config"
	"github.com/r-pine/demo_aggregation/app/pkg/logging"
)

type Controller struct {
	log logging.Logger
	sc  service.Service
	cfg config.Config
}

func NewController(
	log logging.Logger,
	sc service.Service,
	cfg config.Config,
) *Controller {
	return &Controller{
		log: log,
		sc:  sc,
		cfg: cfg,
	}
}

func (c *Controller) InitRoutes(r *gin.Engine) *gin.Engine {

	api := r.Group("api/")
	{
		api.GET("healthcheck", c.Healthcheck)
	}

	return r
}
