package wisun

//go:generate mockgen -source client.go -destination client_mock.go -package wisun

// Client is wisun client
type Client interface {
	Version() error
	SetBRoutePassword(password string) error
	SetBRouteID(id string) error
	Scan() (PanDesc, error)
	LL64(addr string) (string, error)
	SRegS2(channel string) error
	SRegS3(panID string) error
	Join(desc PanDesc) (bool, error)
	Close()
	SendTo(ipv6addr string, data []byte) ([]byte, error)
}

// PanDesc is...
type PanDesc struct {
	Addr     string
	IPV6Addr string
	Channel  string
	PanID    string
}
