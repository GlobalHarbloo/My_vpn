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
	clientAddr := conn.RemoteAddr().String()
	fmt.Println("New client connected:", clientAddr)

	writer := bufio.NewWriter(conn)
	reader := bufio.NewReader(conn)

	// Приветственное сообщение
	writer.WriteString("Welcome to the VPN server! Type your command (PING, HELLO, STATS, QUIT):\n")
	writer.Flush()

	for {
		// Читаем сообщение от клиента
		message, err := reader.ReadString('\n')
		if err != nil {
			log.Println("Error reading message from", clientAddr, ":", err)
			break
		}
		message = strings.TrimSpace(message)
		log.Println("Received message from", clientAddr, ":", message)

		// Если команда пустая, просим клиента ввести снова
		if len(message) == 0 {
			writer.WriteString("Empty command. Please try again.\n")
			writer.Flush()
			continue
		}

		// Обрабатываем команды
		switch strings.ToUpper(message) {
		case "PING":
			writer.WriteString("PONG\n")
		case "HELLO":
			writer.WriteString("Hello, client!\n")
		case "STATS":
			uptime := time.Since(startTime)
			writer.WriteString(fmt.Sprintf("Server uptime: %v\nActive clients: %d\n", uptime, atomic.LoadInt32(&activeClients)))
		case "TIME":
			writer.WriteString("Current time is: " + time.Now().Format("2006-01-02 15:04:05") + "\n")
		case "QUIT":
			writer.WriteString("Goodbye!\n")
			writer.Flush()
			log.Println("Client disconnected:", clientAddr)

			// Уменьшаем счетчик активных клиентов при отключении
			atomic.AddInt32(&activeClients, -1)
			return
		default:
			writer.WriteString("Unknown command\n")
		}
		writer.Flush()
	}
}
