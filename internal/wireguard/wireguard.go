package wireguard

import (
	"fmt"
	"log"
	"os/exec"
)

// Функция для старта WireGuard
func StartWireGuard() error {
	// Запуск туннеля WireGuard через команду wg
	cmd := exec.Command("C:\\Program Files\\WireGuard\\wg.exe", "setconf", "vpn-service", "C:\\Users\\glebs\\Desktop\\vpn-service\\internal\\wireguard\\wg0.conf")
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Ошибка при запуске WireGuard: %s\nВывод: %s", err, out)
		return fmt.Errorf("ошибка при запуске WireGuard: %v\nВывод: %s", err, out)
	}
	log.Println("WireGuard успешно запущен.")

	return nil
}

// Функция для остановки WireGuard
func StopWireGuard() error {
	// Остановка туннеля
	cmd := exec.Command("wg", "down", "wg0") // Убедитесь, что интерфейс остановлен
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("ошибка при остановке WireGuard: %v", err)
	}
	log.Println("WireGuard сервер остановлен.")
	return nil
}
