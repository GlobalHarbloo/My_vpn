package server

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync/atomic"
	"time"
)

var (
	startTime     time.Time
	activeClients int32
)

func StartServer() {
	startTime = time.Now()

	// Открываем лог-файл для записи
	logFile, err := os.OpenFile("server.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening log file:", err)
		return
	}
	defer logFile.Close()

	// Настроим логирование в файл
	log.SetOutput(logFile)
	log.Println("Server started.")

	// Запускаем сервер на порту 8080
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Error starting the server:", err)
		os.Exit(1)
	}
	defer listener.Close()

	fmt.Println("Server is listening on port 8080...")

	for {
		// Принимаем новые соединения
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Error accepting connection:", err)
			continue
		}

		// Увеличиваем счетчик активных клиентов
		atomic.AddInt32(&activeClients, 1)

		// Обрабатываем соединение в горутине
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	conn.SetDeadline(time.Now().Add(5 * time.Minute))

	writer := bufio.NewWriter(conn)
	reader := bufio.NewReader(conn)

	writer.WriteString("Welcome to the VPN server! Type HELP for commands.\n")
	writer.Flush()

	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			log.Println("Client disconnected or error:", err)
			break
		}

		message = strings.TrimSpace(message)
		switch strings.ToUpper(message) {
		case "PING":
			writer.WriteString("PONG\n")
		case "HELP":
			writer.WriteString("Available commands: PING, STATS, QUIT\n")
		case "STATS":
			writer.WriteString(fmt.Sprintf("Active clients: %d\n", atomic.LoadInt32(&activeClients)))
		case "QUIT":
			writer.WriteString("Goodbye!\n")
			writer.Flush()
			return
		default:
			writer.WriteString("Unknown command. Type HELP for a list of commands.\n")
		}
		writer.Flush()
	}
}
