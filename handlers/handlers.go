package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"vpn-service/internal/database"
	"vpn-service/models"
)

// Функция регистрации пользователя
func RegisterUser(w http.ResponseWriter, r *http.Request) {
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	createdUser, err := database.RegisterUser(user.Username, user.Email, user.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(createdUser)
}

// Функция входа в систему
func LoginUser(w http.ResponseWriter, r *http.Request) {
	var creds models.Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	user, err := database.AuthenticateUser(creds.Identifier, creds.Password)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// Получение информации о пользователе (требует токен в заголовке)
func GetProfile(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "Missing token", http.StatusUnauthorized)
		return
	}

	// Удаляем "Bearer " из заголовка
	token = strings.TrimPrefix(token, "Bearer ")

	user, err := database.GetUserByToken(token)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// Получение тарифов
func GetTariffs(w http.ResponseWriter, r *http.Request) {
	tariffs, err := database.GetAllTariffs()
	if err != nil {
		http.Error(w, "Failed to fetch tariffs", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tariffs)
}

// Оплата подписки
func Subscribe(w http.ResponseWriter, r *http.Request) {
	var payment models.Payment
	if err := json.NewDecoder(r.Body).Decode(&payment); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	err := database.ProcessPayment(payment)
	if err != nil {
		http.Error(w, "Payment failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(payment)
}

// Подключение к VPN
func ConnectVPN(w http.ResponseWriter, r *http.Request) {
	var session models.Session
	if err := json.NewDecoder(r.Body).Decode(&session); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	err := database.CreateSession(session)
	if err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session)
}

// Отключение от VPN
func DisconnectVPN(w http.ResponseWriter, r *http.Request) {
	var session models.Session
	if err := json.NewDecoder(r.Body).Decode(&session); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	err := database.EndSession(session)
	if err != nil {
		http.Error(w, "Failed to end session", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session)
}
