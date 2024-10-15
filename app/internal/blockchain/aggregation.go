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
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
)

type Aggregation struct {
	ctx     context.Context
	cfg     config.Config
	log     logging.Logger
	service *sc.Service
}

func NewAggregation(
	ctx context.Context,
	cfg config.Config,
	log logging.Logger,
	service *sc.Service,
) *Aggregation {
	return &Aggregation{
		ctx:     ctx,
		cfg:     cfg,
		log:     log,
		service: service,
	}
}

const (
	liteserverUrl = "https://ton.org/global.config.json"
)

func (a *Aggregation) Run() {
	contracts := map[string]string{
		"stonfi":  "EQCcD96ywHvlXBjuf4ihiGyH66QChHesNyoJSQ6WKKqob3Lh",
		"private": "EQCcD96ywHvlXBjuf4ihiGyH66QChHesNyoJSQ6WKKqob3Lh",
		"dedust":  "EQBPo45inIbFXiUt8I8xrakPRB1aXZ-wzNOJfIhfQgd2rJ-z",
	}
	for {
		aggrs := map[string]entity.Platform{}
		for k, v := range contracts {
			aggr, err := a.getAccountData(k, v)
			if err != nil {
				a.log.Errorln(err)
				continue
			}
			aggrs[k] = *aggr
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
	contractName, contractAddress string,
) (*entity.Platform, error) {

	cfg, err := liteclient.GetConfigFromUrl(a.ctx, liteserverUrl)
	if err != nil {
		return nil, err
	}

	cfg.Liteservers = append(cfg.Liteservers, liteclient.LiteserverConfig{
		IP:   a.cfg.AppConfig.LiteserverPineIP,
		Port: a.cfg.AppConfig.LiteserverPinePort,
		ID: liteclient.ServerID{
			Type: a.cfg.AppConfig.LiteserverPineType,
			Key:  a.cfg.AppConfig.LiteserverPineKey,
		},
	})

	client := liteclient.NewConnectionPool()

	err = client.AddConnectionsFromConfig(a.ctx, cfg)
	if err != nil {
		return nil, err
	}
	api := ton.NewAPIClient(client, ton.ProofCheckPolicyFast).WithRetry()

	ctx := client.StickyContext(a.ctx)

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
		fee, reserve0, reserve1 = a.getFeeAndReservesStonFi(res)
	case "dedust":
		reserve0, reserve1, err = a.getReservesDedust(
			api, b, contractAddress,
		)
		if err != nil {
			return nil, err
		}

		fee, err = a.getFeesDedust(api, b, contractAddress)
		if err != nil {
			return nil, err
		}
	case "private":
		// TODO:
		// fee, reserve0, reserve1 = a.getFeeAndReservesPrivate(res)
		fee, reserve0, reserve1 = a.getFeeAndReservesStonFi(res)
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

// storage::is_locked = ds~load_int(1);
// storage::expires_at = ds~load_uint(32);
// storage::admin_address = ds~load_msg_addr();
// storage::lp_fee = ds~load_uint(8);
// storage::protocol_fee = ds~load_uint(8);
// storage::ref_fee = ds~load_uint(8);
// storage::token0_address = ds~load_msg_addr();
// storage::token1_address = ds~load_msg_addr();
// storage::total_supply_lp = ds~load_coins();

// cell dc_0 = ds~load_ref(); slice ds_0 = dc_0.begin_parse();
// storage::collected_token0_protocol_fee = ds_0~load_coins();
// storage::collected_token1_protocol_fee = ds_0~load_coins();
// storage::protocol_fee_address = ds_0~load_msg_addr();
// storage::reserve0 = ds_0~load_coins();
// storage::reserve1 = ds_0~load_coins();
func (a *Aggregation) getFeeAndReservesPrivate(res *tlb.Account) (int64, int64, int64) {
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
		return int64(fee), reserve0, reserve1
	}
	return 0, 0, 0
}

func (a *Aggregation) getFeeAndReservesStonFi(res *tlb.Account) (int, int64, int64) {
	// adminAddr := slice.MustLoadAddr()
	// lpFee := slice.MustLoadUInt(8)
	// protocolFee := slice.MustLoadUInt(8)
	// refFee := slice.MustLoadUInt(8)
	// token0 := slice.MustLoadAddr()
	// token1 := slice.MustLoadAddr()
	// totalSupply := slice.MustLoadBigCoins()
	// ref := slice.MustLoadRef()
	// collectedToken0ProtocolFee := ref.MustLoadBigCoins()
	// collectedToken1ProtocolFee := ref.MustLoadBigCoins()
	// protocolFeeAddress := ref.MustLoadAddr()
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
	api ton.APIClientWrapped,
	b *ton.BlockIDExt,
	contractAddress string,
) (int, error) {
	result, err := api.RunGetMethod(
		a.ctx, b, address.MustParseAddr(contractAddress), "get_trade_fee",
	)
	if err != nil {
		return 0, errors.New("run get_trade_fee method err:" + err.Error())
	}
	fees := result.AsTuple()
	fee, err := strconv.ParseInt(fmt.Sprintf("%v", fees[0]), 10, 64)
	if err != nil {
		return 0, errors.New("run ParseInt f1 err:" + err.Error())
	}
	// f2, err := strconv.ParseFloat(fmt.Sprintf("%v", fees[1]), 64)
	// if err != nil {
	// 	return 0, errors.New("run ParseInt f2 err:" + err.Error())
	// }
	// fee := f1 / f2
	return int(fee), nil
}

func (a *Aggregation) getReservesDedust(
	api ton.APIClientWrapped,
	b *ton.BlockIDExt,
	contractAddress string,
) (int64, int64, error) {
	result, err := api.RunGetMethod(
		a.ctx, b, address.MustParseAddr(contractAddress), "get_reserves",
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
