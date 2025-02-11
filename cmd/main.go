package main

import (
	"fmt"
	"log"
	"vpn-service/internal/database"
	"vpn-service/internal/telegram"
	"vpn-service/internal/vless"
	"vpn-service/internal/wireguard"
)

func main() {
	// Инициализация базы данных
	database.InitDB()
	database.RunMigration()

	// Запуск WireGuard
	err := wireguard.StartWireGuard()
	if err != nil {
		log.Fatal("Ошибка при запуске WireGuard:", err)
	}
	fmt.Println("VPN-сервер работает...")

	// Запуск V2Ray
	err = vless.StartV2Ray()
	if err != nil {
		log.Fatal("Ошибка при запуске V2Ray:", err)
	}
	fmt.Println("V2Ray сервер работает...")

	// Запуск Telegram-бота
	token := "8089259249:AAGN7uEGOGpXVY86IHTJ7h8hcL194_6ix2I"
	telegram.InitBot(token)
	telegram.StartBot()

	// Блокировка основного потока, чтобы сервер продолжал работать
	select {}
}
