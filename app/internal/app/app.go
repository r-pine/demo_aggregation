package app

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/r-pine/demo_aggregation/app/internal/blockchain"
	"github.com/r-pine/demo_aggregation/app/internal/controller"
	"github.com/r-pine/demo_aggregation/app/internal/db/redis"
	sc "github.com/r-pine/demo_aggregation/app/internal/service"
	st "github.com/r-pine/demo_aggregation/app/internal/storage"
	"github.com/r-pine/demo_aggregation/app/pkg/config"
	"github.com/r-pine/demo_aggregation/app/pkg/logging"
	"golang.org/x/sync/errgroup"
)

func RunApplication() {
	// Init Logger
	logging.Init()
	log := logging.GetLogger()
	log.Infoln("Connect logger successfully!")

	// Init Config
	cfg := config.GetConfig()
	log.Infoln("Connect config successfully!")

	g, ctx := errgroup.WithContext(context.Background())

	rcClient := redis.NewRedisClient(ctx, cfg, log)
	rc, err := rcClient.ConnectToRedis()
	if err != nil {
		log.Fatalf("redis connect to redis failed: %v", err)
		return
	}
	log.Infoln("Connect redis successfully!")

	storage := st.NewStorage(ctx, rc)
	log.Infoln("Connect storage successfully!")

	service := sc.NewService(ctx, storage)
	log.Infoln("Connect service successfully!")

	aggregation := blockchain.NewAggregation(*cfg, log, service)
	log.Infoln("Connect aggregation successfully!")

	gin.SetMode(cfg.AppConfig.GinMode)
	ginRouter := gin.New()
	httpController := controller.NewController(log, *service, *cfg, aggregation)
	handlers := httpController.InitRoutes(ginRouter)
	log.Infoln("Connect handlers successfully!")

	// go aggregation.Run(ctx)

	g.Go(func() (err error) {
		aggregation.Run(ctx)
		return nil
	})

	g.Go(func() (err error) {
		return http.ListenAndServe(
			cfg.AppConfig.HttpAddr,
			handlers,
		)
	})

	err = g.Wait()
	if err != nil {
		log.Fatalln(err)
	}
	// server.RunServer(log, handlers, cfg.AppConfig.HttpAddr)
}
