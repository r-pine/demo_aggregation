package blockchain

import (
	"context"

	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/ton"
)

func GetApiClient(ctx context.Context) (*ton.APIClient, error) {
	client := liteclient.NewConnectionPool()

	err := client.AddConnectionsFromConfigUrl(ctx, configBlockchainUrl)
	if err != nil {
		return nil, err
	}
	api := ton.NewAPIClient(client)
	return api, nil
}
