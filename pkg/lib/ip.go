package lib

import (
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
)

// GetOutboundIP: gets the oubound IP of this packet
func GetOutboundIP() (net.IP, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP, nil
}

const IP_SERVICE = "https://api.ipify.org?format=json"

type IpResponse struct {
	Ip string `json:"ip"`
}

func (i *IpResponse) GetIP() net.IP {
	return net.ParseIP(i.Ip)
}

// GetPublicIP: get the nodes public IP address. For when a node is behind NAT
func GetPublicIP() (net.IP, error) {
	req, err := http.NewRequest(http.MethodGet, IP_SERVICE, nil)

	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)

	if err != nil {
		return nil, err
	}

	resBody, err := io.ReadAll(res.Body)

	if err != nil {
		return nil, err
	}

	var jsonResponse IpResponse

	err = json.Unmarshal([]byte(resBody), &jsonResponse)

	if err != nil {
		return nil, err
	}

	return jsonResponse.GetIP(), nil
}
