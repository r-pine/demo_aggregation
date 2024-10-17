package utils

import (
	"math/big"

	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

type SwapStep struct {
	PoolAddress string    `json:"pool_address"`
	Limit       *big.Int  `json:"limit,omitempty"` // Use pointer to represent optional value
	Next        *SwapStep `json:"next,omitempty"`  // Use pointer for recursive structure
}

// SwapParams represents the parameters for a swap.
type SwapParams struct {
	Deadline         *uint64    `json:"deadline,omitempty"`         // Use pointer for optional value
	RecipientAddress *string    `json:"recipientAddress,omitempty"` // Use pointer for optional value
	ReferralAddress  *string    `json:"referralAddress,omitempty"`  // Use pointer for optional value
	FulfillPayload   *cell.Cell `json:"fulfillPayload,omitempty"`   // Use pointer for optional value
	RejectPayload    *cell.Cell `json:"rejectPayload,omitempty"`    // Use pointer for optional value
}

func PackSwapStep(step SwapStep) *cell.Cell {
	swapStep := cell.BeginCell().
		MustStoreAddr(address.MustParseAddr(step.PoolAddress)).
		MustStoreUInt(0, 1)

	if step.Limit != nil {
		swapStep.MustStoreBigCoins(step.Limit)
	} else {
		swapStep.MustStoreBigCoins(big.NewInt(0))
	}

	if step.Next != nil {
		swapStep.MustStoreMaybeRef(PackSwapStep(*step.Next))
	} else {
		swapStep.MustStoreMaybeRef(nil)
	}
	return swapStep.EndCell()
}

func PackSwapParams(params SwapParams) *cell.Cell {
	swapParams := cell.BeginCell()
	if params.Deadline != nil {
		swapParams.MustStoreUInt(*params.Deadline, 32)
	} else {
		swapParams.MustStoreUInt(0, 32)
	}
	if params.RecipientAddress != nil {
		swapParams.MustStoreAddr(address.MustParseAddr(*params.RecipientAddress))
	} else {
		swapParams.MustStoreAddr(nil)
	}
	if params.ReferralAddress != nil {
		swapParams.MustStoreAddr(address.MustParseAddr(*params.ReferralAddress))
	} else {
		swapParams.MustStoreAddr(nil)
	}
	if params.FulfillPayload != nil {
		swapParams.MustStoreMaybeRef(params.FulfillPayload)
	} else {
		swapParams.MustStoreMaybeRef(nil)
	}
	if params.RejectPayload != nil {
		swapParams.MustStoreMaybeRef(params.RejectPayload)
	} else {
		swapParams.MustStoreMaybeRef(nil)
	}
	return swapParams.EndCell()
}
