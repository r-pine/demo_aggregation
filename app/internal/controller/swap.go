package controller

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/r-pine/demo_aggregation/app/internal/blockchain"
	"github.com/r-pine/demo_aggregation/app/internal/controller/utils"
	"github.com/r-pine/demo_aggregation/app/internal/entity"
	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/ton/jetton"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

const (
	aPineToTon = "APINE_TO_TON"
	configUrl  = "https://ton-blockchain.github.io/testnet-global.config.json"

	aPineMaster        = "EQAjWFZaH0Xib0VGEwe3148Hg7arST5mhJwDB3YTIS0OFUxJ"
	pTonPrivateAddress = "EQABxQiQSPSCFMM12RcW2uzeujZ2s4J8X3utZmy7BJgJXssJ"
	privateAddress     = "EQBB9cr9pFiGmAQ9vpAGNdWpaDiuw88kLdxipDNKgJzdWw91"
	stonfiAddress      = "EQD40Uooo8XiU4v58X26ShDf93R3w0Fu4d0VXLUShzFerTav"
	pTonStonfiAddress  = "EQARULUYsmJq1RiZ-YiH-IJLcAZUVkVff-KBPwEmmaQGH6aC"
	aPineStonfiAddress = "EQCqU71ESTAIL9HRBf-UZEa-4ED3m7MB1JIznAz39h5pwnbo"
	dedustVaultNative  = "EQDa4VOnTYlLvDJ0gZjNYm5PXfSmmtL6Vs6A_CZEtXCNICq_"
	dedustVaultAPine   = "EQDamGXCPYxbxLsdXFaKJoZ7VYHRIIyxhka8GOLwHJC1l_LZ"
	dedustPoolAddress  = "EQC0neK6srf_hVHiJbdqqTHT8zbH-CWbBlJSbybP4TkdG6hG"
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
	data, err := c.sc.Get("states")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	var res *entity.Aggregation
	if err := json.Unmarshal([]byte(data), &res); err != nil {
		ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
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
	dedustAmountIn, privateAmountIn, stonfiAmountIn, bestOutput := blockchain.Swap(amountToFloat, *res, swapTonToApine)

	client := liteclient.NewConnectionPool()

	err = client.AddConnectionsFromConfigUrl(ctx, configUrl)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	api := ton.NewAPIClient(client)

	tokenContract := address.MustParseAddr(aPineMaster)
	master := jetton.NewJettonMasterClient(api, tokenContract)
	jettonWallet, err := master.GetJettonWallet(ctx, address.MustParseRawAddr(br.Address))
	if err != nil {
		fmt.Println(err)
		ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	userJettonAddress := jettonWallet.Address()

	var msgs []Message
	if swapTonToApine {
		// if privateAmountIn > 0 {
		// 	privateMessage := buildPrivateTonToJettonBody(privateAmountIn, br.Address, nil)
		// 	msgs = append(msgs, privateMessage)
		// }
		if stonfiAmountIn > 0 {
			stonfiMessage := buildStonfiTonToJettonBody(stonfiAmountIn, br.Address, nil)
			msgs = append(msgs, stonfiMessage)
		}
		// if dedustAmountIn > 0 {
		// 	dedustMessage := buildDedustTonToJettonBody(
		// 		dedustAmountIn,
		// 		nil,
		// 		&utils.SwapStep{
		// 			PoolAddress: dedustPoolAddress,
		// 		},
		// 		&utils.SwapParams{},
		// 	)
		// 	msgs = append(msgs, dedustMessage)
		// }

		ctx.JSON(
			http.StatusOK,
			&BodyResponse{
				Msgs:                msgs,
				SumCalculatedAmount: strconv.FormatFloat(bestOutput, 'f', 6, 64),
			},
		)
		return
	} else {
		if privateAmountIn > 0 {
			privateMessage := buildPrivateJettonToTonBody(
				privateAmountIn, br.Address, nil, userJettonAddress,
			)
			msgs = append(msgs, privateMessage)
		}
		if stonfiAmountIn > 0 {
			stonfiMessage := buildStonfiJettonToTonBody(
				stonfiAmountIn, br.Address, nil, userJettonAddress,
			)
			msgs = append(msgs, stonfiMessage)
		}
		if dedustAmountIn > 0 {
			dedustMessage := buildDedustJettonToTonBody(
				dedustAmountIn,
				br.Address,
				nil,
				&utils.SwapStep{
					PoolAddress: dedustPoolAddress,
				},
				&utils.SwapParams{},
				userJettonAddress,
			)
			msgs = append(msgs, dedustMessage)
		}
		ctx.JSON(
			http.StatusOK,
			&BodyResponse{
				Msgs:                msgs,
				SumCalculatedAmount: strconv.FormatFloat(bestOutput, 'f', 6, 64),
			},
		)
		return
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
		MustStoreUInt(uint64(time.Now().Unix()), 64).
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

func buildPrivateJettonToTonBody(
	privateAmountIn float64,
	userAddr string,
	refAddr *string,
	userJettonWalletAddress *address.Address,
) Message {
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
		MustStoreUInt(uint64(time.Now().Unix()), 64).
		MustStoreBigCoins(big.NewInt(int64(privateAmountIn))).
		MustStoreAddr(address.MustParseAddr(privateAddress)).
		MustStoreAddr(address.MustParseAddr(userAddr)).
		MustStoreBoolBit(false).
		MustStoreBigCoins(big.NewInt(0.1 * blockchain.NanoUnit)).
		MustStoreBoolBit(true).
		MustStoreRef(fwdPayload.EndCell()).
		EndCell()

	return Message{
		AmountTon:  tlb.MustFromTON(strconv.FormatFloat(0.2*blockchain.NanoUnit, 'f', 6, 64)).String(),
		DstAddress: userJettonWalletAddress.Bounce(true).String(),
		Payload:    base64.StdEncoding.EncodeToString(body.ToBOC()),
	}
}

func buildStonfiTonToJettonBody(stonfiAmountIn float64, userAddr string, refAddr *string) Message {
	fwdPayload := cell.BeginCell().
		MustStoreUInt(0x25938561, 32).
		MustStoreAddr(address.MustParseAddr(aPineStonfiAddress)).
		MustStoreBigCoins(big.NewInt(1)).
		MustStoreAddr(address.MustParseRawAddr(userAddr)).
		MustStoreBoolBit(refAddr != nil)

	if refAddr != nil {
		fwdPayload.MustStoreAddr(address.MustParseAddr(*refAddr))
	}

	fwdAmount := 0.25 * blockchain.NanoUnit

	body := cell.BeginCell().
		MustStoreUInt(0xf8a7ea5, 32).
		MustStoreUInt(uint64(time.Now().Unix()), 64).
		MustStoreBigCoins(big.NewInt(int64(stonfiAmountIn))).
		MustStoreAddr(address.MustParseAddr(stonfiAddress)).
		MustStoreAddr(address.MustParseRawAddr(userAddr)).
		MustStoreBoolBit(false).
		MustStoreBigCoins(big.NewInt(int64(fwdAmount))).
		MustStoreBoolBit(true).
		MustStoreRef(fwdPayload.EndCell()).
		EndCell()

	return Message{
		AmountTon:  tlb.MustFromTON(strconv.FormatFloat(stonfiAmountIn+fwdAmount, 'f', 6, 64)).String(),
		DstAddress: pTonStonfiAddress,
		Payload:    base64.StdEncoding.EncodeToString(body.ToBOC()),
	}
}

func buildStonfiJettonToTonBody(
	stonfiAmountIn float64,
	userAddr string,
	refAddr *string,
	userJettonWalletAddress *address.Address,
) Message {
	fwdPayload := cell.BeginCell().
		MustStoreUInt(0x25938561, 32).
		MustStoreAddr(address.MustParseAddr(pTonStonfiAddress)).
		MustStoreBigCoins(big.NewInt(1)).
		MustStoreAddr(address.MustParseRawAddr(userAddr)).
		MustStoreBoolBit(refAddr != nil)

	if refAddr != nil {
		fwdPayload.MustStoreAddr(address.MustParseAddr(*refAddr))
	}

	fwdAmount := 0.25 * blockchain.NanoUnit

	body := cell.BeginCell().
		MustStoreUInt(0xf8a7ea5, 32).
		MustStoreUInt(uint64(time.Now().Unix()), 64).
		MustStoreBigCoins(big.NewInt(int64(stonfiAmountIn))).
		MustStoreAddr(address.MustParseAddr(stonfiAddress)).
		MustStoreAddr(address.MustParseRawAddr(userAddr)).
		MustStoreBoolBit(false).
		MustStoreBigCoins(big.NewInt(int64(fwdAmount))).
		MustStoreBoolBit(true).
		MustStoreRef(fwdPayload.EndCell()).
		EndCell()

	return Message{
		AmountTon:  tlb.MustFromTON(strconv.FormatFloat(fwdAmount+0.05*blockchain.NanoUnit, 'f', 6, 64)).String(),
		DstAddress: userJettonWalletAddress.Bounce(true).String(),
		Payload:    base64.StdEncoding.EncodeToString(body.ToBOC()),
	}
}

func buildDedustTonToJettonBody(
	dedustAmountIn float64,
	limit *float64,
	next *utils.SwapStep,
	swapParams *utils.SwapParams,
) Message {
	body := cell.BeginCell().
		MustStoreUInt(0xea06185d, 32).
		MustStoreUInt(uint64(time.Now().Unix()), 64).
		MustStoreBigCoins(big.NewInt(int64(dedustAmountIn))).
		MustStoreAddr(address.MustParseAddr(dedustPoolAddress)).
		MustStoreUInt(0, 1)
	if limit != nil {
		body.MustStoreBigCoins(big.NewInt(int64(*limit)))
	} else {
		body.MustStoreBigCoins(big.NewInt(int64(1)))
	}
	if next != nil {
		body.MustStoreMaybeRef(utils.PackSwapStep(*next))
	} else {
		body.MustStoreMaybeRef(nil)
	}
	body.MustStoreRef(utils.PackSwapParams(*swapParams))

	return Message{
		AmountTon:  tlb.MustFromTON(strconv.FormatFloat(dedustAmountIn+0.2*blockchain.NanoUnit, 'f', 6, 64)).String(),
		DstAddress: dedustVaultNative,
		Payload:    base64.StdEncoding.EncodeToString(body.EndCell().ToBOC()),
	}
}

func buildDedustJettonToTonBody(
	dedustAmountIn float64,
	userAddr string,
	limit *float64,
	next *utils.SwapStep,
	swapParams *utils.SwapParams,
	userJettonWalletAddress *address.Address,
) Message {
	fwdPayload := cell.BeginCell().
		MustStoreUInt(0xe3a0d482, 32).
		MustStoreAddr(address.MustParseAddr(dedustPoolAddress)).
		MustStoreUInt(0, 1)
	if limit != nil {
		fwdPayload.MustStoreBigCoins(big.NewInt(int64(*limit)))
	} else {
		fwdPayload.MustStoreBigCoins(big.NewInt(int64(0)))
	}
	if next != nil {
		fwdPayload.MustStoreMaybeRef(utils.PackSwapStep(*next))
	} else {
		fwdPayload.MustStoreMaybeRef(nil)
	}
	fwdPayload.MustStoreRef(utils.PackSwapParams(*swapParams))

	fwdAmount := 0.25 * blockchain.NanoUnit

	body := cell.BeginCell().
		MustStoreUInt(0xf8a7ea5, 32).
		MustStoreUInt(uint64(time.Now().Unix()), 64).
		MustStoreBigCoins(big.NewInt(int64(dedustAmountIn))).
		MustStoreAddr(address.MustParseAddr(dedustVaultAPine)).
		MustStoreAddr(address.MustParseRawAddr(userAddr)).
		MustStoreBoolBit(false).
		MustStoreBigCoins(big.NewInt(int64(fwdAmount))).
		MustStoreBoolBit(true).
		MustStoreRef(fwdPayload.EndCell()).
		EndCell()

	return Message{
		AmountTon:  tlb.MustFromTON(strconv.FormatFloat(0.05*blockchain.NanoUnit+fwdAmount, 'f', 6, 64)).String(),
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
