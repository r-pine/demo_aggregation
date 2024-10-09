package blockchain

import (
	"fmt"
	"sort"

	"context"

	"github.com/r-pine/demo_aggregation/app/pkg/config"
	"github.com/r-pine/demo_aggregation/app/pkg/logging"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/liteclient"
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

const liteserverUrl = "https://ton.org/global.config.json"

func (a *Aggregation) RunAggregation() {

	cfg, err := liteclient.GetConfigFromUrl(a.ctx, liteserverUrl)
	if err != nil {
		a.log.Fatalln(err)
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
		a.log.Fatalln("connection err: ", err.Error())
		return
	}
	api := ton.NewAPIClient(client, ton.ProofCheckPolicyFast).WithRetry()

	ctx := client.StickyContext(a.ctx)

	b, err := api.CurrentMasterchainInfo(ctx)
	if err != nil {
		a.log.Fatalln("get block err:", err.Error())
		return
	}

	addr := address.MustParseAddr("UQAULcjDZ4TK9huUxR4Vl_Tfa8JRooU3bhvPrmHJHZIPGTTX")

	res, err := api.WaitForBlock(b.SeqNo).GetAccount(ctx, b, addr)
	if err != nil {
		a.log.Fatalln("get account err:", err.Error())
		return
	}

	fmt.Printf("Is active: %v\n", res.IsActive)
	if res.IsActive {
		fmt.Printf("Status: %s\n", res.State.Status)
		fmt.Printf("Balance: %s TON\n", res.State.Balance.String())
		if res.Data != nil {
			fmt.Printf("Data: %s\n", res.Data.Dump())
		}
	}

	lastHash := res.LastTxHash
	lastLt := res.LastTxLT

	fmt.Printf("\nTransactions:\n")
	for {
		if lastLt == 0 {
			break
		}

		list, err := api.ListTransactions(ctx, addr, 15, lastLt, lastHash)
		if err != nil {
			a.log.Printf("send err: %s", err.Error())
			return
		}
		lastHash = list[0].PrevTxHash
		lastLt = list[0].PrevTxLT

		sort.Slice(list, func(i, j int) bool {
			return list[i].LT > list[j].LT
		})

		for _, t := range list {
			fmt.Println(t.String())
		}
	}
}
