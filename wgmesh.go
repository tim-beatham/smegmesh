package main

import (
	"fmt"

	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

func main() {
	client, err := wgctrl.New()

	if err != nil {
		fmt.Println("Error creating device")
		return
	}

	privateKey, err := wgtypes.GeneratePrivateKey()
	var listenPort int = 5109

	if err != nil {
		fmt.Println("Error creating private key")
		return
	}

	cfg := wgtypes.Config{
		PrivateKey: &privateKey,
		ListenPort: &listenPort,
	}

	err = client.ConfigureDevice("utun9", cfg)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	devices, err := client.Devices()

	if err != nil {
		fmt.Println("unable to retrieve devices")
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
