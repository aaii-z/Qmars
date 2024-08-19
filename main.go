package main

import (
	"bufio"
	"fmt"
	"net"
	"time"
	"os"
	"strings"
)

// Function to scan a given IP and port
func scanPort(ip string, port string) bool {
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(ip, port), 1*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// Function to get local IP address
func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for _, address := range addrs {
		// Check the address type and ensure it is not a loopback address
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}

	return ""
}

// Function to generate a list of IPs in the subnet
func generateIPs(subnet string) []string {
	ips := []string{}
	for i := 40; i < 50; i++ {
		ips = append(ips, fmt.Sprintf("%s.%d", subnet, i))
	}
	return ips
}

// Function to handle incoming connections
func handleConnection(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Connection closed by the other Qmars.")
			return
		}
		fmt.Printf("Received message: %s", message)
	}
}

func main() {
	// Define the port to search for
	port := "5457"
	
	// Get the local IP address and subnet
	localIP := getLocalIP()
	if localIP == "" {
		fmt.Println("Unable to determine local IP address.")
		return
	}
	fmt.Printf("Local IP: %s\n", localIP)

	// Extract the subnet by trimming the last octet of the IP address
	subnet := localIP[:strings.LastIndex(localIP, ".")]

	// Generate list of IPs in the subnet
	ips := generateIPs(subnet)

	// Scan the subnet for open port 5457
	var foundIP string
	for i, ip := range ips {
		fmt.Printf("Scanning %s...\n", ip)
		if scanPort(ip, port) {
			fmt.Printf("Qmars%d: %s has port %s open\n", i+1, ip, port)
			foundIP = ip
			break
		}
	}

	if foundIP == "" {
		// If no devices found, start listening on port 5457
		fmt.Printf("No Qmars found. Starting server on port %s...\n", port)
		ln, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
		if err != nil {
			fmt.Println(err)
			return
		}
		defer ln.Close()

		// Start a goroutine to handle incoming connections
		go func() {
			for {
				conn, err := ln.Accept()
				if err != nil {
					fmt.Println(err)
					continue
				}
				go handleConnection(conn)
			}
		}()

		// Keep the main thread running to allow messaging
		for {
			fmt.Printf("Waiting for another Qmars to connect...\n")
		}
	} else {
		// If another Qmars is found, connect to it and start messaging
		conn, err := net.Dial("tcp", net.JoinHostPort(foundIP, port))
		if err != nil {
			fmt.Println(err)
			return
		}
		defer conn.Close()

		// Start a goroutine to handle incoming messages
		go handleConnection(conn)

		// Send messages to the connected Qmars
		for {
			fmt.Print("Enter message: ")
			reader := bufio.NewReader(os.Stdin)
			message, _ := reader.ReadString('\n')
			_, err = fmt.Fprintf(conn, message)
			if err != nil {
				fmt.Println("Error sending message:", err)
				return
			}
		}
	}
}

