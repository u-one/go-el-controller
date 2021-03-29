package wisun

import "context"

//go:generate mockgen -source client.go -destination client_mock.go -package wisun

// Client is wisun client
type Client interface {
	Connect(ctx context.Context, bRouteID, bRoutePW string) error
	Close()
	Send(data []byte) ([]byte, error)
}
