package blockchain

import (
	"encoding/base64"
	"math/big"
	"strconv"
	"time"

	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/tvm/cell"
)

type Message struct {
	AmountTon  string `json:"amount"`
	DstAddress string `json:"address"`
	Payload    string `json:"payload"`
}

func BuildMessageSwap(
	swapTonToApine bool,
	privateAmountIn, stonfiAmountIn, dedustAmountIn float64,
	userRawAddr string,
	userJettonAddress *address.Address,
) []Message {
	var msgs []Message
	if swapTonToApine {
		if privateAmountIn > 0 {
			privateMessage := buildPrivateTonToJettonBody(privateAmountIn, userRawAddr, nil)
			msgs = append(msgs, privateMessage)
		}
		if stonfiAmountIn > 0 {
			stonfiMessage := buildStonfiTonToJettonBody(stonfiAmountIn, userRawAddr, nil)
			msgs = append(msgs, stonfiMessage)
		}
		if dedustAmountIn > 0 {
			dedustMessage := buildDedustTonToJettonBody(dedustAmountIn)
			msgs = append(msgs, dedustMessage)
		}
		return msgs
	} else {
		if privateAmountIn > 0 {
			privateMessage := buildPrivateJettonToTonBody(
				privateAmountIn, userRawAddr, nil, userJettonAddress,
			)
			msgs = append(msgs, privateMessage)
		}
		if stonfiAmountIn > 0 {
			stonfiMessage := buildStonfiJettonToTonBody(
				stonfiAmountIn, userRawAddr, nil, userJettonAddress,
			)
			msgs = append(msgs, stonfiMessage)
		}
		if dedustAmountIn > 0 {
			dedustMessage := buildDedustJettonToTonBody(
				dedustAmountIn,
				userRawAddr,
				userJettonAddress,
			)
			msgs = append(msgs, dedustMessage)
		}
		return msgs
	}
}

func buildPrivateTonToJettonBody(privateAmountIn float64, userAddr string, refAddr *string) Message {
	gasConsumption := 13000000
	fwdAmountPrivatePool := 0.1 * NanoUnit
	value := fwdAmountPrivatePool + float64(gasConsumption) + privateAmountIn
	prAmIn := privateAmountIn / NanoUnit
	prAmInStr := strconv.FormatFloat(prAmIn, 'f', 6, 64)
	fwdPayload := cell.BeginCell().
		MustStoreUInt(0x25938561, 32).
		MustStoreAddr(address.MustParseRawAddr(userAddr)).
		MustStoreBigCoins(tlb.MustFromTON(prAmInStr).Nano()).
		MustStoreBigCoins(big.NewInt(1)).
		MustStoreBoolBit(refAddr != nil)
	if refAddr != nil {
		fwdPayload.MustStoreAddr(address.MustParseAddr(*refAddr))
	}

	body := cell.BeginCell().
		MustStoreUInt(0x8f637488, 32).
		MustStoreUInt(uint64(time.Now().Unix()), 64).
		MustStoreBigCoins(tlb.MustFromTON(prAmInStr).Nano()).
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
		MustStoreAddr(address.MustParseRawAddr(userAddr)).
		MustStoreBoolBit(false).
		MustStoreBigCoins(big.NewInt(0.1 * NanoUnit)).
		MustStoreBoolBit(true).
		MustStoreRef(fwdPayload.EndCell()).
		EndCell()

	return Message{
		AmountTon:  tlb.MustFromTON(strconv.FormatFloat(0.2*NanoUnit, 'f', 6, 64)).String(),
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

	fwdAmount := 0.25 * NanoUnit

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

	fwdAmount := 0.25 * NanoUnit

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
		AmountTon:  tlb.MustFromTON(strconv.FormatFloat(fwdAmount+(0.05*NanoUnit), 'f', 6, 64)).String(),
		DstAddress: userJettonWalletAddress.Bounce(true).String(),
		Payload:    base64.StdEncoding.EncodeToString(body.ToBOC()),
	}
}

func buildDedustTonToJettonBody(dedustAmountIn float64) Message {
	body := cell.BeginCell().
		MustStoreUInt(0xea06185d, 32).
		MustStoreUInt(uint64(time.Now().Unix()), 64).
		MustStoreBigCoins(big.NewInt(int64(dedustAmountIn))).
		MustStoreAddr(address.MustParseAddr(dedustPoolAddress)).
		MustStoreUInt(0, 1).
		MustStoreBigCoins(big.NewInt(0)).
		MustStoreMaybeRef(nil).
		MustStoreRef(cell.BeginCell().MustStoreUInt(0, 32).MustStoreAddr(nil).MustStoreAddr(nil).MustStoreMaybeRef(nil).MustStoreMaybeRef(nil).EndCell())
	return Message{
		AmountTon:  tlb.MustFromTON(strconv.FormatFloat(dedustAmountIn+0.2*NanoUnit, 'f', 6, 64)).String(),
		DstAddress: dedustVaultNative,
		Payload:    base64.StdEncoding.EncodeToString(body.EndCell().ToBOC()),
	}
}

func buildDedustJettonToTonBody(
	dedustAmountIn float64,
	userAddr string,
	userJettonWalletAddress *address.Address,
) Message {
	fwdPayload := cell.BeginCell().
		MustStoreUInt(0xe3a0d482, 32).
		MustStoreAddr(address.MustParseAddr(dedustPoolAddress)).
		MustStoreUInt(0, 1).
		MustStoreBigCoins(big.NewInt(0)).
		MustStoreMaybeRef(nil).
		MustStoreRef(cell.BeginCell().MustStoreUInt(0, 32).MustStoreAddr(nil).MustStoreAddr(nil).MustStoreMaybeRef(nil).MustStoreMaybeRef(nil).EndCell())

	fwdAmount := 0.25 * NanoUnit

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
		AmountTon:  tlb.MustFromTON(strconv.FormatFloat(0.05*NanoUnit+fwdAmount, 'f', 6, 64)).String(),
		DstAddress: userJettonWalletAddress.Bounce(true).String(),
		Payload:    base64.StdEncoding.EncodeToString(body.ToBOC()),
	}
}
