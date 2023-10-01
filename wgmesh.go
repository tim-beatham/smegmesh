package main

import (
	"fmt"

	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func main() {
	client, err := wgctrl.New()

	if err != nil {
		return
	}

	privateKey, err := wgtypes.GeneratePrivateKey()
	var listenPort int = 5109

	if err != nil {
		return
	}

	cfg := wgtypes.Config{
		PrivateKey: &privateKey,
		ListenPort: &listenPort,
	}

	err = client.ConfigureDevice("utun9", cfg)

	if err != nil {
		return
	}

	devices, err := client.Devices()

	if err != nil {
	return
	}

	fmt.Printf("Number of devices: %d\n", len(devices))

	for _, device := range devices {
		fmt.Printf("Device Name: %s\n", device.Name)
		fmt.Printf("Listen Port: %d\n", device.ListenPort)
		fmt.Printf("Private Key: %s\n", device.PrivateKey.String())
		fmt.Printf("Public Key: %s\n", device.PublicKey.String())
	}
}
