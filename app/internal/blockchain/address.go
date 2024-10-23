package blockchain

import (
	"context"

	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/ton/jetton"
)

func GetUserJettonWalletAddress(
	ctx context.Context,
	api *ton.APIClient,
	userRawAddress string,
) (*address.Address, error) {
	tokenContract := address.MustParseAddr(aPineMaster)
	master := jetton.NewJettonMasterClient(api, tokenContract)
	jettonWallet, err := master.GetJettonWallet(ctx, address.MustParseRawAddr(userRawAddress))
	if err != nil {
		return nil, err
	}
	userJettonAddress := jettonWallet.Address()
	return userJettonAddress, nil
}
