package cloud

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"

	"github.com/moby/moby/client"
)

func generateRandomString(length int) string {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return ""
	}
	return hex.EncodeToString(bytes)[:length]
}

func defaultDataDirFor(t ServiceType) (string, bool) {
	switch t {
	case PostgreSQL:
		return "/var/lib/postgresql/data", true
	case MySQL, MariaDB:
		return "/var/lib/mysql", true
	case MongoDB:
		return "/data/db", true
	case Redis:
		return "/data", true
	default:
		return "", false
	}
}

func initManager(cli *client.Client) (*CloudManager, error) {
	manager := &CloudManager{
		dockerClient: cli,
		instances:    make(map[string]*ServiceInstance),
		portCounter:  5432,
		networkName:  "cloud-db-network",
	}
	if err := manager.ensureNetwork(); err != nil {
		return nil, fmt.Errorf("network creation error: %v", err)
	}
	return manager, nil
}


func isTCPPortAvailable(port string) bool {
	address := net.JoinHostPort("localhost", port)
	// Try to listen on the specified port
	listener, err := net.Listen("tcp", address)
	if err != nil {
		fmt.Printf("Port %s is already in use: %v\n", port, err)
		return false
	}
	defer listener.Close() // Close the listener immediately after success

	fmt.Printf("Port %s is available.\n", port)
	return true
}

// func isUDPPortAvailable(port string) bool {
// 	address := net.JoinHostPort("[::]", port)
// 	// Try to listen on the specified port
// 	listener, err := net.Listen("udp", address)
// 	if err != nil {
// 		fmt.Printf("Port %s is already in use: %v\n", port, err)
// 		return false
// 	}
// 	defer listener.Close() // Close the listener immediately after success

// 	fmt.Printf("Port %s is available.\n", port)
// 	return true
// }
