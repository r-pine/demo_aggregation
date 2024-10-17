package controller

import (
	"encoding/base64"
	"math/big"
	"math/rand/v2"
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
	pTonPrivateAddress  = "EQABxQiQSPSCFMM12RcW2uzeujZ2s4J8X3utZmy7BJgJXssJ"
	privateAddress      = "EQCp5UpUBZIbdold9sqUeU-1gFAF_8Mk-QQKIEXgbFtat8Um"
	aPineToTon          = "APINE_TO_TON"
	stonfiAddress       = "EQB3ncyBUTjZUA5EnFKR5_EnOMI9V1tTEAAPaiU71gc4TiUt"
	pTonStonfiAddress   = "EQARULUYsmJq1RiZ-YiH-IJLcAZUVkVff-KBPwEmmaQGH6aC"
	jettonStonfiAddress = ""
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
		privateMessage := buildPrivateTonToJettonBody(privateAmountIn, br.Address, nil)
		var msgs []Message
		msgs = append(msgs, privateMessage)

		ctx.JSON(
			http.StatusOK,
			&BodyResponse{
				Msgs:                msgs,
				SumCalculatedAmount: strconv.FormatFloat(bestOutput, 'f', 6, 64),
			},
		)
	}
}

func buildPrivateTonToJettonBody(privateAmountIn float64, userAddr string, refAddr *string) Message {

	gasConsumption := 13000000                        // 13000000n
	fwdAmountPrivatePool := 0.1 * blockchain.NanoUnit // 100000000
	value := fwdAmountPrivatePool + float64(gasConsumption) + (privateAmountIn * blockchain.NanoUnit)

	fwdPayload := cell.BeginCell().
		MustStoreUInt(0x25938561, 32).
		MustStoreAddr(address.MustParseRawAddr(userAddr)).
		MustStoreBigCoins(big.NewInt(int64(privateAmountIn))).
		MustStoreBigCoins(big.NewInt(1)).
		MustStoreBoolBit(refAddr != nil)
	if refAddr != nil {
		fwdPayload.MustStoreAddr(address.MustParseAddr(*refAddr))
	}

	body := cell.BeginCell().
		MustStoreUInt(0x8f637488, 32).
		MustStoreUInt(rand.Uint64(), 64).
		MustStoreBigCoins(big.NewInt(int64(privateAmountIn))).
		MustStoreAddr(address.MustParseAddr(pTonPrivateAddress)).
		MustStoreRef(fwdPayload.EndCell()).
		EndCell()
	return Message{
		AmountTon:  tlb.MustFromTON(strconv.FormatFloat(value, 'f', 6, 64)).String(),
		DstAddress: privateAddress,
		Payload:    base64.StdEncoding.EncodeToString(body.ToBOC()),
	}
}

func buildPrivateJettonToTonBody(privateAmountIn float64, userAddr string, refAddr *string) Message {
	fwdPayload := cell.BeginCell().
		MustStoreUInt(0x25938561, 32).
		MustStoreAddr(address.MustParseRawAddr(userAddr)).
		MustStoreBigCoins(big.NewInt(int64(privateAmountIn))).
		MustStoreBigCoins(big.NewInt(1)).
		MustStoreBoolBit(refAddr != nil)

	if refAddr != nil {
		fwdPayload.MustStoreAddr(address.MustParseAddr(*refAddr))
	}

	body := cell.BeginCell().
		MustStoreUInt(0xf8a7ea5, 32).
		MustStoreUInt(rand.Uint64(), 64).
		MustStoreBigCoins(big.NewInt(int64(privateAmountIn))).
		MustStoreAddr(address.MustParseAddr(privateAddress)).
		MustStoreAddr(address.MustParseAddr(userAddr)).
		MustStoreBoolBit(false).
		MustStoreBigCoins(big.NewInt(0.1 * blockchain.NanoUnit)).
		MustStoreBoolBit(true).
		MustStoreRef(fwdPayload.EndCell()).
		EndCell()

	userJettonWalletAddress := address.MustParseAddr("nil") // retrieve user jetton wallet address from userAddr

	return Message{
		AmountTon:  tlb.MustFromTON(strconv.FormatFloat(0.2*blockchain.NanoUnit, 'f', 6, 64)).String(),
		DstAddress: userJettonWalletAddress.Bounce(true).String(),
		Payload:    base64.StdEncoding.EncodeToString(body.ToBOC()),
	}
}

func buildStonfiTonToJettonBody(privateAmountIn float64, userAddr string, refAddr *string) Message {
	fwdPayload := cell.BeginCell().
		MustStoreUInt(0x25938561, 32).
		MustStoreAddr(address.MustParseRawAddr(jettonStonfiAddress)).
		MustStoreBigCoins(big.NewInt(1)).
		MustStoreAddr(address.MustParseRawAddr(userAddr)).
		MustStoreBoolBit(refAddr != nil)

	if refAddr != nil {
		fwdPayload.MustStoreAddr(address.MustParseAddr(*refAddr))
	}

	fwdAmount := 0.25 * blockchain.NanoUnit

	body := cell.BeginCell().
		MustStoreUInt(0xf8a7ea5, 32).
		MustStoreUInt(rand.Uint64(), 64).
		MustStoreBigCoins(big.NewInt(int64(privateAmountIn))).
		MustStoreAddr(address.MustParseAddr(stonfiAddress)).
		MustStoreAddr(address.MustParseAddr(userAddr)).
		MustStoreBoolBit(false).
		MustStoreBigCoins(big.NewInt(int64(fwdAmount))).
		MustStoreBoolBit(true).
		MustStoreRef(fwdPayload.EndCell()).
		EndCell()

	return Message{
		AmountTon:  tlb.MustFromTON(strconv.FormatFloat(privateAmountIn+fwdAmount, 'f', 6, 64)).String(),
		DstAddress: pTonStonfiAddress,
		Payload:    base64.StdEncoding.EncodeToString(body.ToBOC()),
	}
}

func buildStonfiJettonToTonBody(privateAmountIn float64, userAddr string, refAddr *string) Message {
	fwdPayload := cell.BeginCell().
		MustStoreUInt(0x25938561, 32).
		MustStoreAddr(address.MustParseRawAddr(pTonStonfiAddress)).
		MustStoreBigCoins(big.NewInt(1)).
		MustStoreAddr(address.MustParseRawAddr(userAddr)).
		MustStoreBoolBit(refAddr != nil)

	if refAddr != nil {
		fwdPayload.MustStoreAddr(address.MustParseAddr(*refAddr))
	}

	fwdAmount := 0.25 * blockchain.NanoUnit

	body := cell.BeginCell().
		MustStoreUInt(0xf8a7ea5, 32).
		MustStoreUInt(rand.Uint64(), 64).
		MustStoreBigCoins(big.NewInt(int64(privateAmountIn))).
		MustStoreAddr(address.MustParseAddr(stonfiAddress)).
		MustStoreAddr(address.MustParseAddr(userAddr)).
		MustStoreBoolBit(false).
		MustStoreBigCoins(big.NewInt(int64(fwdAmount))).
		MustStoreBoolBit(true).
		MustStoreRef(fwdPayload.EndCell()).
		EndCell()

	userJettonWalletAddress := address.MustParseAddr("nil") // retrieve user jetton wallet address from userAddr

	return Message{
		AmountTon:  tlb.MustFromTON(strconv.FormatFloat(fwdAmount+0.05*blockchain.NanoUnit, 'f', 6, 64)).String(),
		DstAddress: userJettonWalletAddress.Bounce(true).String(),
		Payload:    base64.StdEncoding.EncodeToString(body.ToBOC()),
	}
}

func (c *Controller) getErrorResponse() *BodyResponse {
	return &BodyResponse{
		Msgs:                []Message{},
		SumCalculatedAmount: "0",
	}
}
