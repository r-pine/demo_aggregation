package blockchain

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"context"

	"github.com/r-pine/demo_aggregation/app/internal/entity"
	sc "github.com/r-pine/demo_aggregation/app/internal/service"
	"github.com/r-pine/demo_aggregation/app/pkg/config"
	"github.com/r-pine/demo_aggregation/app/pkg/logging"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
)

type Aggregation struct {
	cfg     config.Config
	log     logging.Logger
	service *sc.Service
}

func NewAggregation(
	cfg config.Config,
	log logging.Logger,
	service *sc.Service,
) *Aggregation {
	return &Aggregation{
		cfg:     cfg,
		log:     log,
		service: service,
	}
}

func (a *Aggregation) Run(ctx context.Context) {
	contracts := map[string]string{
		"stonfi":  a.cfg.AppConfig.StonfiPoolAddress,
		"private": a.cfg.AppConfig.PrivatePoolAddress,
		"dedust":  a.cfg.AppConfig.DedustPoolAddress,
	}
	for {
		aggrs := map[string]entity.Platform{}

		for k, v := range contracts {
			aggr, err := a.getAccountData(ctx, k, v)
			if err != nil {
				a.log.Errorln(err)
				return
			}
			if aggr == nil {
				return
			}
			aggrs[k] = *aggr
		}
		for k := range contracts {
			if _, ok := aggrs[k]; !ok {
				return
			}
		}

		aggrsStr, err := a.aggregationsToJsonStr(&entity.Aggregation{Dex: aggrs})
		if err != nil {
			a.log.Errorln(err)
			continue
		}
		if err := a.service.Set("states", aggrsStr); err != nil {
			a.log.Errorln(err)
			continue
		}
		time.Sleep(time.Duration(a.cfg.AppConfig.Delay) * time.Second)
	}
}

func (a *Aggregation) getAccountData(
	ctx context.Context,
	contractName, contractAddress string,
) (*entity.Platform, error) {

	api, client, err := GetApiClient(ctx)
	if err != nil {
		return nil, err
	}

	ctx = client.StickyContext(ctx)

	b, err := api.CurrentMasterchainInfo(ctx)
	if err != nil {
		return nil, err
	}

	addr := address.MustParseAddr(contractAddress)

	res, err := api.WaitForBlock(b.SeqNo).GetAccount(ctx, b, addr)
	if err != nil {
		return nil, err
	}

	var (
		fee      int
		reserve0 int64
		reserve1 int64
	)

	switch contractName {
	case "stonfi":
		fee, reserve1, reserve0 = a.getFeeAndReservesStonFi(res)
	case "dedust":
		reserve0, reserve1, err = a.getReservesDedust(
			ctx, api, b, contractAddress,
		)
		if err != nil {
			return nil, err
		}

		fee, err = a.getFeesDedust(ctx, api, b, contractAddress)
		if err != nil {
			return nil, err
		}
	case "private":
		fee, reserve0, reserve1 = a.getFeeAndReservesPrivate(res)
	}

	pl := entity.Platform{
		Name: contractName,
		Address: entity.Address{
			Bounce:   res.State.Address.Bounce(true).String(),
			UnBounce: res.State.Address.Bounce(false).String(),
		},
		Fee:      float64(fee),
		Reserve0: float64(reserve0),
		Reserve1: float64(reserve1),
		IsActive: res.IsActive,
		Status:   string(res.State.Status),
		Balance:  res.State.Balance.String(),
	}
	return &pl, nil
}

func (a *Aggregation) getFeeAndReservesPrivate(res *tlb.Account) (int, int64, int64) {
	if res.Data != nil {
		slice := res.Data.BeginParse()
		_ = slice.MustLoadInt(1)
		_ = slice.MustLoadUInt(32)
		_ = slice.MustLoadAddr()
		lpFee := slice.MustLoadUInt(8)
		protocolFee := slice.MustLoadUInt(8)
		_ = slice.MustLoadUInt(8)
		_ = slice.MustLoadAddr()
		_ = slice.MustLoadAddr()
		_ = slice.MustLoadBigCoins()
		ref := slice.MustLoadRef()
		_ = ref.MustLoadBigCoins()
		_ = ref.MustLoadBigCoins()
		_ = ref.MustLoadAddr()
		reserve0 := ref.MustLoadBigCoins().Int64()
		reserve1 := ref.MustLoadBigCoins().Int64()
		fee := lpFee + protocolFee
		return int(fee), reserve0, reserve1
	}
	return 0, 0, 0
}

func (a *Aggregation) getFeeAndReservesStonFi(res *tlb.Account) (int, int64, int64) {
	if res.Data != nil {
		slice := res.Data.BeginParse()
		_ = slice.MustLoadAddr()
		lpFee := slice.MustLoadUInt(8)
		protocolFee := slice.MustLoadUInt(8)
		_ = slice.MustLoadUInt(8)
		_ = slice.MustLoadAddr()
		_ = slice.MustLoadAddr()
		_ = slice.MustLoadBigCoins()
		ref := slice.MustLoadRef()
		_ = ref.MustLoadBigCoins()
		_ = ref.MustLoadBigCoins()
		_ = ref.MustLoadAddr()
		reserve0 := ref.MustLoadBigCoins().Int64()
		reserve1 := ref.MustLoadBigCoins().Int64()
		fee := lpFee + protocolFee
		return int(fee), reserve0, reserve1
	}
	return 0, 0, 0
}

func (a *Aggregation) getFeesDedust(
	ctx context.Context,
	api ton.APIClientWrapped,
	b *ton.BlockIDExt,
	contractAddress string,
) (int, error) {
	result, err := api.RunGetMethod(
		ctx, b, address.MustParseAddr(contractAddress), "get_trade_fee",
	)
	if err != nil {
		return 0, errors.New("run get_trade_fee method err:" + err.Error())
	}
	fees := result.AsTuple()
	fee, err := strconv.ParseInt(fmt.Sprintf("%v", fees[0]), 10, 64)
	if err != nil {
		return 0, errors.New("run ParseInt f1 err:" + err.Error())
	}
	return int(fee), nil
}

func (a *Aggregation) getReservesDedust(
	ctx context.Context,
	api ton.APIClientWrapped,
	b *ton.BlockIDExt,
	contractAddress string,
) (int64, int64, error) {
	result, err := api.RunGetMethod(
		ctx, b, address.MustParseAddr(contractAddress), "get_reserves",
	)
	if err != nil {
		return 0, 0, errors.New("run get_reserves method err")
	}
	reservers := result.AsTuple()
	var reserve0 int64
	var reserve1 int64
	for i, r := range reservers {
		if i == 0 {
			reserve0, err = strconv.ParseInt(fmt.Sprintf("%v", r), 10, 64)
			if err != nil {
				return reserve0, reserve1, errors.New("ParseInt reserve0 err:" + err.Error())
			}
		}
		if i == 1 {
			reserve1, err = strconv.ParseInt(fmt.Sprintf("%v", r), 10, 64)
			if err != nil {
				return reserve0, reserve1, errors.New("ParseInt reserve1 err:" + err.Error())
			}
		}
	}
	return reserve0, reserve1, nil
}

func (a *Aggregation) aggregationsToJsonStr(aggr *entity.Aggregation) (string, error) {
	data, err := json.Marshal(aggr)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
