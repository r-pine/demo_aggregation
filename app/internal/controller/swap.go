package controller

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/r-pine/demo_aggregation/app/internal/blockchain"
	"github.com/r-pine/demo_aggregation/app/internal/entity"
	"github.com/xssnick/tonutils-go/address"
)

const aPineToTon = "APINE_TO_TON"

type BodyResponse struct {
	Msgs                []blockchain.Message `json:"messages"`
	SumCalculatedAmount string               `json:"sum_calculated_amount"`
}

func (c *Controller) GetSwapPayload(ctx *gin.Context) {
	type bodyRequest struct {
		Amount    string `json:"amount"`
		Address   string `json:"address"`
		Direction string `json:"direction"`
		Query     string `json:"query,omitempty"`
		UserId    string `json:"user_id,omitempty"`
	}
	var br bodyRequest

	if err := ctx.ShouldBind(&br); err != nil {
		ctx.JSON(http.StatusBadRequest, c.getErrorResponse())
		return
	}

	go func(query, address, userId string) {

		if err := c.sc.UpdateUserWallet(query, address, userId); err != nil {
			c.log.Errorln(err)
			return
		}

	}(br.Query, br.Address, br.UserId)

	data, err := c.sc.Get("states")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	var res *entity.Aggregation
	if err := json.Unmarshal([]byte(data), &res); err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	swapTonToApine := true
	if br.Direction == aPineToTon {
		swapTonToApine = false
	}

	amountToFloat, err := strconv.ParseFloat(br.Amount, 64)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	stonfiAmountIn, dedustAmountIn, privateAmountIn, bestOutput := blockchain.Swap(amountToFloat, *res, swapTonToApine)

	api, _, err := blockchain.GetApiClient(ctx)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	var userJettonAddress *address.Address
	for {
		userJettonAddress, err = blockchain.GetUserJettonWalletAddress(ctx, api, br.Address)
		if err != nil || userJettonAddress == nil {
			c.log.Error(err)
			time.Sleep(1 * time.Second)
			continue
		} else {
			break
		}
	}

	msgs := blockchain.BuildMessageSwap(
		swapTonToApine,
		privateAmountIn,
		stonfiAmountIn,
		dedustAmountIn,
		br.Address,
		userJettonAddress,
	)
	ctx.JSON(
		http.StatusOK,
		&BodyResponse{
			Msgs:                msgs,
			SumCalculatedAmount: strconv.FormatFloat(bestOutput, 'f', 6, 64),
		},
	)
}

func (c *Controller) getErrorResponse() *BodyResponse {
	return &BodyResponse{
		Msgs:                []blockchain.Message{},
		SumCalculatedAmount: "0",
	}
}
