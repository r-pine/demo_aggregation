package utils

import (
	"math/big"

	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

func packSwapStep(poolAddress *address.Address, limit *big.Int, next *cell.Cell) *cell.Cell {
	swapStep := cell.BeginCell().
		MustStoreAddr(poolAddress).
		MustStoreUInt(0, 1)

	if limit != nil {
		swapStep.MustStoreBigCoins(limit)
	} else {
		swapStep.MustStoreBigCoins(big.NewInt(0))
	}

	if next != nil {
		swapStep.MustStoreMaybeRef(next)
	} else {
		swapStep.MustStoreMaybeRef(nil)
	}
	return swapStep.EndCell()
}

func packSwapParams(deadline *uint64, recipientAddress *address.Address, referralAddress *address.Address, fulfillPayload *cell.Cell, rejectPayload *cell.Cell) *cell.Cell {
	swapParams := cell.BeginCell()
	if deadline != nil {
		swapParams.MustStoreUInt(*deadline, 32)
	} else {
		swapParams.MustStoreUInt(0, 32)
	}
	if recipientAddress != nil {
		swapParams.MustStoreAddr(recipientAddress)
	} else {
		swapParams.MustStoreAddr(nil)
	}
	if referralAddress != nil {
		swapParams.MustStoreAddr(referralAddress)
	} else {
		swapParams.MustStoreAddr(nil)
	}
	if fulfillPayload != nil {
		swapParams.MustStoreMaybeRef(fulfillPayload)
	} else {
		swapParams.MustStoreMaybeRef(nil)
	}
	if rejectPayload != nil {
		swapParams.MustStoreMaybeRef(rejectPayload)
	} else {
		swapParams.MustStoreMaybeRef(nil)
	}
	return swapParams.EndCell()
}
