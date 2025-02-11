package vless

import (
	"database/sql"
	"fmt"
	"os"
	"os/exec"

	"encoding/json"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

// Функция для создания нового пользователя и генерации UUID для него
func CreateUserWithUUID(username string, email string, db *sql.DB) (string, error) {
	// Генерация нового UUID
	newUUID := uuid.New().String()

	// Добавление нового пользователя в базу данных
	query := `INSERT INTO users (username, email, uuid) VALUES ($1, $2, $3) RETURNING id`
	var userID int
	err := db.QueryRow(query, username, email, newUUID).Scan(&userID)
	if err != nil {
		return "", fmt.Errorf("failed to create user: %v", err)
	}
	if err := addClientToV2RayConfig(newUUID); err != nil {
		return "", fmt.Errorf("failed to add client to V2Ray config: %v", err)
	}

	// Вернем UUID нового пользователя
	return newUUID, nil
}

func addClientToV2RayConfig(clientUUID string) error {
	configFile := "etc/v2ray/config.json"

	// Читаем текущую конфигурацию
	file, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("failed to read V2Ray config file: %v", err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(file, &config); err != nil {
		return fmt.Errorf("failed to parse V2Ray config: %v", err)
	}

	// Добавляем нового клиента
	clients := config["clients"].([]interface{})
	clients = append(clients, map[string]interface{}{
		"id":      clientUUID,
		"alterId": 0,
		"email":   "user@example.com",
	})
	config["clients"] = clients

	// Записываем обратно
	newConfig, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal V2Ray config: %v", err)
	}

	if err := os.WriteFile(configFile, newConfig, 0644); err != nil {
		return fmt.Errorf("failed to write V2Ray config file: %v", err)
	}

	return restartV2Ray()
}

func restartV2Ray() error {
	cmd := exec.Command("systemctl", "restart", "v2ray")
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed restart V2Ray: %v", err)
	}
	return nil
}

func StartV2Ray() error {
	cmd := exec.Command("systemctl", "start", "v2ray")
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed start V2Ray: %v", err)
	}
	return nil
}

func stopV2Ray() error {
	cmd := exec.Command("systemctl", "stop", "v2ray")
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed stop V2Ray: %v", err)
	}
	return nil
}

// Функция для удаления пользователя по UUID
func DeleteUserByUUID(clientUUID string, db *sql.DB) error {
	// Удаление пользователя из базы данных
	query := `DELETE FROM users WHERE uuid = $1`
	_, err := db.Exec(query, clientUUID)
	if err != nil {
		return fmt.Errorf("failed to delete user: %v", err)
	}

	// Удаление клиента из конфигурации V2Ray
	if err := removeClientFromV2RayConfig(clientUUID); err != nil {
		return fmt.Errorf("failed to remove client from V2Ray config: %v", err)
	}

	return nil
}

func removeClientFromV2RayConfig(clientUUID string) error {
	configFile := "etc/v2ray/config.json"

	// Читаем текущую конфигурацию
	file, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("failed to read V2Ray config file: %v", err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(file, &config); err != nil {
		return fmt.Errorf("failed to parse V2Ray config: %v", err)
	}

	// Удаляем клиента
	clients := config["clients"].([]interface{})
	for i, client := range clients {
		if client.(map[string]interface{})["id"] == clientUUID {
			clients = append(clients[:i], clients[i+1:]...)
			break
		}
	}
	config["clients"] = clients

	// Записываем обратно
	newConfig, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal V2Ray config: %v", err)
	}

	if err := os.WriteFile(configFile, newConfig, 0644); err != nil {
		return fmt.Errorf("failed to write V2Ray config file: %v", err)
	}

	return restartV2Ray()
}
