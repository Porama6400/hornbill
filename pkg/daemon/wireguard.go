package daemon

import (
	"fmt"
	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
	"hornbill/pkg/model"
	"net"
)

type WireGuardConfig struct {
	PublicKey      wgtypes.Key
	InterfaceName  string
	PublicAddress  string
	AllowedAddress []string
}

type WireGuard struct {
	Client *wgctrl.Client
	Config WireGuardConfig
}

func NewWireGuard(config WireGuardConfig) (*WireGuard, error) {
	client, err := wgctrl.New()
	if err != nil {
		return nil, err
	}
	return &WireGuard{
		Client: client,
		Config: config,
	}, nil
}

func (w *WireGuard) GetKey() (*string, error) {
	device, err := w.Client.Device(w.Config.InterfaceName)
	if err != nil {
		return nil, err
	}

	keyString := device.PublicKey.String()
	return &keyString, nil
}

func (w *WireGuard) Configure(users []model.User) (bool, error) {
	if users == nil {
		return false, fmt.Errorf("users is nil")
	}
	peers := make([]wgtypes.PeerConfig, 0, len(users))
	for _, user := range users {
		key, err := wgtypes.ParseKey(user.Identity.PublicKey)
		if err != nil {
			return false, fmt.Errorf("key %s for user %s is invalid", user.Identity.PublicKey, user.Identity.Id)
		}
		peers = append(peers, wgtypes.PeerConfig{
			PublicKey: key,
			AllowedIPs: []net.IPNet{
				{
					IP:   user.Address,
					Mask: net.IPv4Mask(0xff, 0xff, 0xff, 0xff),
				},
			},
		})
	}
	err := w.Client.ConfigureDevice(w.Config.InterfaceName, wgtypes.Config{
		ReplacePeers: true,
		Peers:        peers,
	})
	if err != nil {
		return false, fmt.Errorf("WireGuard failed to configure device: %w", err)
	}

	return true, nil
}
