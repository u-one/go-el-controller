package wisun

//go:generate mockgen -source wisun.go -destination wisun_mock.go -package wisun

// Serial is serial interface
type Serial interface {
	Send(cmd string) ([]string, error)
}

// Wisun is wisun client
type WiSun struct {
}

// Send send command
func (w WiSun) Send(cmd string) ([]string, error) {
	return []string{}, nil
}
