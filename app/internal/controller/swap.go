package controller

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/r-pine/demo_aggregation/app/internal/blockchain"
	"github.com/r-pine/demo_aggregation/app/internal/entity"
)

const (
	aPineToTon = "APINE_TO_TON"

	aPineMaster        = "EQAjWFZaH0Xib0VGEwe3148Hg7arST5mhJwDB3YTIS0OFUxJ"
	pTonPrivateAddress = "EQCzGHwSIX6VM_PCBWUNm-d_hS5JuO46UNGtCjJcxSb2mMx7"
	privateAddress     = "EQBB9cr9pFiGmAQ9vpAGNdWpaDiuw88kLdxipDNKgJzdWw91"
	stonfiAddress      = "EQB3ncyBUTjZUA5EnFKR5_EnOMI9V1tTEAAPaiU71gc4TiUt"
	pTonStonfiAddress  = "EQARULUYsmJq1RiZ-YiH-IJLcAZUVkVff-KBPwEmmaQGH6aC"
	aPineStonfiAddress = "EQCqU71ESTAIL9HRBf-UZEa-4ED3m7MB1JIznAz39h5pwnbo"
	dedustVaultNative  = "EQDa4VOnTYlLvDJ0gZjNYm5PXfSmmtL6Vs6A_CZEtXCNICq_"
	dedustVaultAPine   = "EQDamGXCPYxbxLsdXFaKJoZ7VYHRIIyxhka8GOLwHJC1l_LZ"
	dedustPoolAddress  = "EQC0neK6srf_hVHiJbdqqTHT8zbH-CWbBlJSbybP4TkdG6hG"
)

type BodyResponse struct {
	Msgs                []blockchain.Message `json:"messages"`
	SumCalculatedAmount string               `json:"sum_calculated_amount"`
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
	dedustAmountIn, privateAmountIn, stonfiAmountIn, bestOutput := blockchain.Swap(amountToFloat, *res, swapTonToApine)

	api, err := blockchain.GetApiClient(ctx)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	userJettonAddress, err := blockchain.GetUserJettonWalletAddress(ctx, api, br.Address)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
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
