package service

import (
	"fmt"

	"github.com/r-pine/demo_aggregation/app/internal/requests"
)

var updateWalletURL string = "https://indbot.rpine.xyz/wallet/update/"

func (o *Service) UpdateUserWallet(queryString, walletAddress, userId string) error {
	if queryString != "" && walletAddress != "" && userId != "" {
		url := fmt.Sprintf("%s?%s&address=%s&user_id=%s", updateWalletURL, queryString, walletAddress, userId)
		response, err := requests.Get(url)
		if err != nil {
			return err
		}
		if response != nil && response.StatusCode == 200 {
			return nil
		}
		defer response.Body.Close()
		return nil
	}
	return nil
}
