package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/google/uuid"
)

type ConnectRequest struct {
	Email string `json:"email"`
	UUID  string `json:"uuid"`
}

type ConnectResponse struct {
	ClientConfig string `json:"clientConfig"`
}

// Генерация нового UUID
func generateUUID() string {
	return uuid.NewString()
}

// Сохранение UUID в локальный файл
func saveUUIDToFile(uuid string) error {
	return os.WriteFile("uuid.txt", []byte(uuid), 0644)
}

// Чтение UUID из локального файла
func readUUIDFromFile() (string, error) {
	data, err := os.ReadFile("uuid.txt")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Получение UUID (создаёт новый, если отсутствует)
func getOrCreateUUID() (string, error) {
	uuid, err := readUUIDFromFile()
	if err == nil {
		return uuid, nil
	}

	// Если файл отсутствует, создаём новый UUID
	uuid = generateUUID()
	err = saveUUIDToFile(uuid)
	if err != nil {
		return "", fmt.Errorf("ошибка сохранения UUID: %v", err)
	}

	return uuid, nil
}

func ConnectToServer(serverURL, email string) (string, error) {
	// Получаем или создаём UUID
	uuid, err := getOrCreateUUID()
	if err != nil {
		return "", fmt.Errorf("ошибка получения UUID: %v", err)
	}

	// Формируем запрос
	requestData := ConnectRequest{
		Email: email,
		UUID:  uuid,
	}
	requestBody, err := json.Marshal(requestData)
	if err != nil {
		return "", fmt.Errorf("ошибка формирования запроса: %v", err)
	}

	// Отправляем запрос на сервер
	resp, err := http.Post(serverURL+"/connect", "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("ошибка отправки запроса: %v", err)
	}
	defer resp.Body.Close()

	// Обрабатываем ответ
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("сервер вернул ошибку: %s", string(body))
	}

	var connectResponse ConnectResponse
	if err := json.NewDecoder(resp.Body).Decode(&connectResponse); err != nil {
		return "", fmt.Errorf("ошибка декодирования ответа: %v", err)
	}

	// Возвращаем клиентскую конфигурацию
	return connectResponse.ClientConfig, nil
}
