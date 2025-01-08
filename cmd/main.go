package main

import (
	"fmt"
	"log"
	datebase "vpn-service/internal/database"
	"vpn-service/internal/telegram"
	"vpn-service/internal/wireguard"
)

func main() {

	datebase.InitDB()

	datebase.RunMigration()

	err := wireguard.StartWireGuard()
	if err != nil {
		log.Fatal("Ошибка при запуске WireGuard:", err)
	}

	fmt.Println("VPN-сервер работает...")

	telegram.StartBot()
}
