package main

import (
	"fmt"
	"log"
	"net/http"
	"vpn-service/handlers"
	"vpn-service/internal/database"
	"vpn-service/internal/telegram"
	"vpn-service/internal/vless"
	"vpn-service/internal/wireguard"

	"github.com/gorilla/mux"
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

	router := mux.NewRouter()

	// Маршруты API
	router.HandleFunc("/register", handlers.RegisterUser).Methods("POST")
	router.HandleFunc("/login", handlers.LoginUser).Methods("POST")
	router.HandleFunc("/profile", handlers.GetProfile).Methods("GET")
	router.HandleFunc("/tariffs", handlers.GetTariffs).Methods("GET")
	router.HandleFunc("/subscribe", handlers.Subscribe).Methods("POST")
	router.HandleFunc("/connect", handlers.ConnectVPN).Methods("POST")
	router.HandleFunc("/disconnect", handlers.DisconnectVPN).Methods("POST")

	log.Println("Server started on port 8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
