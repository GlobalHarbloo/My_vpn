package main

import (
	"fmt"
	"log"
	datebase "vpn-service/internal/database"
	"vpn-service/internal/server"
	"vpn-service/internal/telegram"
	"vpn-service/internal/wireguard"
)

func main() {

	datebase.InitDB()

	datebase.RunMigration()

	err := wireguard.StartWireGuard()
	if err != nil {
		log.Fatal("Ошибка запуска WireGuard:", err)
	}

	// Добавление нового пользователя
	err = wireguard.AddPeer("КЛЮЧ_КЛИЕНТА", "10.0.0.2")
	if err != nil {
		log.Fatal("Ошибка добавления клиента:", err)
	}

	// Генерация конфигурации клиента
	clientConfig := wireguard.GenerateClientConfig("PRIVATE_KEY", "SERVER_PUBLIC_KEY", "127.0.0.1:51820", "10.0.0.2")
	log.Println("Client configuration:\n", clientConfig)

	fmt.Printf("Starting vpn-server...")
	server.StartServer()

	telegram.StartBot()
}
