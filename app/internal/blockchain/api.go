package blockchain

import (
	"context"

	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/ton"
)

func GetApiClient(ctx context.Context) (*ton.APIClient, *liteclient.ConnectionPool, error) {
	client := liteclient.NewConnectionPool()
	err := client.AddConnectionsFromConfigFile(configBlockchainUrl)
	if err != nil {
		return nil, nil, err
	}
	api := ton.NewAPIClient(client)
	return api, client, nil
}
