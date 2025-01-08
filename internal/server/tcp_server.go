package server

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	startTime     time.Time
	activeClients int
	clientMutex   sync.Mutex
)

func StartServer() {
	startTime = time.Now()
	logFile, err := os.OpenFile("server.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening log file:", err)
		return
	}
	defer logFile.Close()

	log.SetOutput(logFile)
	log.Println("Server started.")

	listener, err := net.Listen("tcp", ":8080")

	if err != nil {
		fmt.Println("Error starting the server:", err)
		os.Exit(1)
	}

	defer listener.Close()
	fmt.Println("Server is listening on port 8080...")

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}

		clientMutex.Lock()
		activeClients++
		clientMutex.Unlock()

		go hanleConnection(conn)
	}
}

func hanleConnection(conn net.Conn) {
	defer conn.Close()
	clientAddr := conn.RemoteAddr().String()
	fmt.Println("New client connected:", clientAddr)

	writer := bufio.NewWriter(conn)
	reader := bufio.NewReader(conn)

	writer.WriteString("Welcone to the Vpn server! Type your command (PING, HELLO, STATS, QUIT):\n")
	writer.Flush()

	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			log.Println("Error reading message from", clientAddr, ":", err)
			break
		}
		message = strings.TrimSpace(message)
		log.Println("Received message from", clientAddr, ":", message)

		if len(message) == 0 {
			writer.WriteString("Empty command. Please try again.\n")
			writer.Flush()
			continue
		}

		switch strings.ToUpper(message) {
		case "PING":
			writer.WriteString("PONG\n")
		case "HELLO":
			writer.WriteString("Hello, client!\n")
		case "STATS":
			uptime := time.Since(startTime)
			clientMutex.Lock()
			writer.WriteString(fmt.Sprintf("Server uptime: %v\nActive clients: %d\n", uptime, activeClients))
			clientMutex.Unlock()
		case "TIME":
			writer.WriteString("Current time is: " + time.Now().Format("2006-01-02 15:04:05") + "\n")
		case "QUIT":
			writer.WriteString("Goodbye!\n")
			writer.Flush()
			log.Println("Client disconnected:", clientAddr)

			clientMutex.Lock()
			activeClients--
			clientMutex.Unlock()
			return
		default:
			writer.WriteString("Unknown command\n")
		}
		writer.Flush()
	}

}
