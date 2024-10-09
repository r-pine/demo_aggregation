package blockchain

import (
	"errors"
	"fmt"
	"strconv"

	"context"

	"github.com/r-pine/demo_aggregation/app/pkg/config"
	"github.com/r-pine/demo_aggregation/app/pkg/logging"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
)

type Aggregation struct {
	ctx context.Context
	cfg config.Config
	log logging.Logger
}

func NewAggregation(
	ctx context.Context,
	cfg config.Config,
	log logging.Logger,
) *Aggregation {
	return &Aggregation{
		ctx: ctx,
		cfg: cfg,
		log: log,
	}
}

const (
	liteserverUrl = "https://ton.org/global.config.json"
)

func (a *Aggregation) RunAggregation(contractName, contractAddress string) {

	cfg, err := liteclient.GetConfigFromUrl(a.ctx, liteserverUrl)
	if err != nil {
		a.log.Errorln(err)
		return
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
		a.log.Errorln("connection err: ", err.Error())
		return
	}
	api := ton.NewAPIClient(client, ton.ProofCheckPolicyFast).WithRetry()

	ctx := client.StickyContext(a.ctx)

	b, err := api.CurrentMasterchainInfo(ctx)
	if err != nil {
		a.log.Errorln("get block err:", err.Error())
		return
	}

	addr := address.MustParseAddr(contractAddress)

	res, err := api.WaitForBlock(b.SeqNo).GetAccount(ctx, b, addr)
	if err != nil {
		a.log.Errorln("get account err:", err.Error())
		return
	}

	switch contractName {
	case "stonfi":
		fee, reserve0, reserve1 := a.getFeeAndReservesStonFi(res)
		fmt.Println(fee, reserve0, reserve1)
	case "dedust":
		reserve0, reserve1, err := a.getReservesDedust(api, b, contractAddress)
		if err != nil {
			a.log.Errorln("getReservesDedust err:", err.Error())
			return
		}
		fees, err := a.getFeesDedust(api, b, contractAddress)
		if err != nil {
			a.log.Errorln("getFeesDedust err:", err.Error())
			return
		}
		fmt.Println(fees, reserve0, reserve1)

	case "private":
	}
	fmt.Printf("Is active: %v\n", res.IsActive)
	if res.IsActive {
		fmt.Printf("Status: %s\n", res.State.Status)

		fmt.Printf("Balance: %s TON\n", res.State.Balance.String())
	}
}

func (a *Aggregation) getFeeAndReservesStonFi(res *tlb.Account) (uint64, int64, int64) {
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
		return fee, reserve0, reserve1
	}
	return 0, 0, 0
}

func (a *Aggregation) getFeesDedust(
	api ton.APIClientWrapped,
	b *ton.BlockIDExt,
	contractAddress string,
) (int64, error) {
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
	return fee, nil
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
