package controller

import (
	"encoding/base64"
	"math/big"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/r-pine/demo_aggregation/app/internal/blockchain"
	"github.com/r-pine/demo_aggregation/app/internal/entity"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

const (
	pTonPrivateAddress = ""
	privateAddress     = ""
	aPineToTon         = "APINE_TO_TON"
)

type Message struct {
	AmountTon  string `json:"amount_ton"`
	DstAddress string `json:"dst_address"`
	Payload    string `json:"payload"`
}

type BodyResponse struct {
	Msgs                []Message `json:"messages"`
	SumCalculatedAmount string    `json:"sum_calculated_amount"`
}

func (c *Controller) GetSwapPayload(ctx *gin.Context) {
	type bodyRequest struct {
		Amount    string `json:"amount"`
		Address   string `json:"address"`
		Direction string `json:"direction"`
	}
	var br bodyRequest

	if err := ctx.ShouldBind(&br); err != nil {
		ctx.JSON(http.StatusBadRequest, c.getErrorResponse())
		return
	}

	// Замените эту часть на получение данных из вашего источника
	res := entity.Aggregation{
		Dex: map[string]entity.Platform{
			"stonfi": {
				Name:     "stonfi",
				Reserve0: 200000000000,
				Reserve1: 6666666666666660,
				Fee:      30,
			},
			"dedust": {
				Name:     "dedust",
				Reserve0: 77905132253,
				Reserve1: 6666666666666660,
				Fee:      25,
			},
			"private": {
				Name:     "private",
				Reserve0: 200000000000,
				Reserve1: 6666666666666660,
				Fee:      20,
			},
		},
	}

	swapTonToApine := true
	if br.Direction == aPineToTon {
		swapTonToApine = false
	}

	amountToFloat, err := strconv.ParseFloat(br.Amount, 64)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if swapTonToApine {
		_, privateAmountIn, bestOutput := blockchain.Swap(amountToFloat, res, swapTonToApine)
		privateBody := buildPrivateTonToJettonBody(privateAmountIn, br.Address)
	}
}

func buildPrivateTonToJettonBody(privateAmountIn float64, userAddr string) Message {

	gasConsumption := 13000000                        // 13000000n
	fwdAmountPrivatePool := 0.1 * blockchain.NanoUnit // 100000000
	value := fwdAmountPrivatePool + float64(gasConsumption) + (privateAmountIn * blockchain.NanoUnit)

	fwdPayload := cell.BeginCell().
		MustStoreUInt(0x25938561, 32).
		MustStoreAddr(address.MustParseRawAddr(userAddr)).
		MustStoreBigCoins(big.NewInt(int64(privateAmountIn))).
		MustStoreBigCoins(big.NewInt(1)).
		MustStoreBoolBit(false).
		MustStoreBoolBit(true).
		MustStoreRef(
			cell.BeginCell().
				MustStoreAddr(address.MustParseRawAddr(userAddr)).
				MustStoreAddr(nil).
				EndCell(),
		).EndCell()

	body := cell.BeginCell().
		MustStoreUInt(0x8f637488, 32).
		MustStoreUInt(0, 64).
		MustStoreBigCoins(big.NewInt(int64(privateAmountIn))).
		MustStoreAddr(address.MustParseRawAddr(pTonPrivateAddress)).
		MustStoreRef(fwdPayload).
		EndCell()
	return Message{
		AmountTon:  tlb.MustFromTON(strconv.FormatFloat(value, 'f', 6, 64)).String(),
		DstAddress: privateAddress,
		Payload:    base64.StdEncoding.EncodeToString(body.ToBOC()),
	}
}

func (c *Controller) getErrorResponse() *BodyResponse {
	return &BodyResponse{
		Msgs:                []Message{},
		SumCalculatedAmount: "0",
	}
}
