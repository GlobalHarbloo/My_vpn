package wireguard

import (
	"fmt"
	"log"
	"net"
	"os/exec"

	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

const (
	interfaceName = "wg0"
	configPath    = "/etc/wireguard/wg0.conf"
)

func StartWireGuard() error {
	cmd := exec.Command("wg-quick", "up", interfaceName)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error starting WireGuard: %v", err)
	}
	log.Println("WireGuard server started.")
	return nil
}

func StopWireGuard() error {
	cmd := exec.Command("wg-quick", "down", interfaceName)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error stoping WireGuard: %v", err)
	}
	log.Println("WireGuard server stopped")
	return nil
}

func AddPeer(publicKey, allowedIP string) error {
	client, err := wgctrl.New()
	if err != nil {
		return fmt.Errorf("Error creating wgctrl client: %v", err)
	}
	defer client.Close()

	_, err = client.Device(interfaceName)
	if err != nil {
		return fmt.Errorf("Error fetching device info: %v", err)
	}

	parsedPublicKey, err := wgtypes.ParseKey(publicKey)
	if err != nil {
		return fmt.Errorf("error parsing public key: %v", err)
	}

	ip := net.ParseIP(allowedIP)
	if ip == nil {
		return fmt.Errorf("error parsing allowed IP: %v", allowedIP)
	}

	newPeer := wgtypes.PeerConfig{
		PublicKey: parsedPublicKey,
		AllowedIPs: []net.IPNet{
			{
				IP:   net.ParseIP(allowedIP),
				Mask: net.CIDRMask(32, 32),
			},
		},
	}

	deviceConfig := wgtypes.Config{
		Peers: []wgtypes.PeerConfig{newPeer},
	}

	if err := client.ConfigureDevice(interfaceName, deviceConfig); err != nil {
		return fmt.Errorf("Error configurung device: %v", err)
	}

	log.Println("Клиент успешно добавлен:", allowedIP)
	return nil
}

func GenerateClientConfig(privateKey, serverPublicKey, endpoint string, clientIP string) string {
	return fmt.Sprintf(`[Interface]
PrivateKey = %s
Address = %s/32
DNS = 8.8.8.8

[Peer]
PublicKey = %s
Endpoint = %s
AllowedIPs = 0.0.0.0/0
PersistentKeepalive = 25
`, privateKey, clientIP, serverPublicKey, endpoint)
}
