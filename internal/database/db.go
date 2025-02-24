package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"
	"vpn-service/models"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

var db *sql.DB
var logger *log.Logger

func InitDB() {
	connStr := "postgres://vpn_user:S62uLkkXa1UZ@localhost/vpn_service?sslmode=disable"
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatalf("Ошибка проверки соединения с базой данных: %v", err)
	}

	// Initialize the logger
	logFile, err := os.OpenFile("database.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Ошибка открытия файла лога: %v", err)
	}
	logger = log.New(logFile, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)

	log.Println("Подключение к базе данных успешно!")
}

func GetDB() *sql.DB {
	return db
}

func RunMigration() {
	sql := `
    CREATE TABLE IF NOT EXISTS users (
        id SERIAL PRIMARY KEY,
        username VARCHAR(100) UNIQUE NOT NULL,
        email VARCHAR(255) UNIQUE NOT NULL,
        password VARCHAR(100) NOT NULL,
        telegram_id BIGINT UNIQUE,
        tariff_id INT DEFAULT 1,
        used_traffic BIGINT DEFAULT 0,
        subscription_start TIMESTAMP,
        subscription_end TIMESTAMP,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );

    CREATE TABLE IF NOT EXISTS tariffs (
        id SERIAL PRIMARY KEY,
        name VARCHAR(50) NOT NULL,
        price DECIMAL(10, 2) NOT NULL,
        traffic_limit BIGINT NOT NULL
    );

    CREATE TABLE IF NOT EXISTS payments (
        id SERIAL PRIMARY KEY,
        user_id INT REFERENCES users(id) ON DELETE CASCADE,
        amount DECIMAL(10, 2),
        status VARCHAR(50),
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    );

    CREATE TABLE IF NOT EXISTS sessions (
        id SERIAL PRIMARY KEY,
        user_id INT REFERENCES users(id) ON DELETE CASCADE,
        start_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        end_time TIMESTAMP,
        data_usage BIGINT DEFAULT 0
    );
    `

	_, err := db.Exec(sql)
	if err != nil {
		logger.Fatal("Error running migration: ", err)
	}

	logger.Println("Migrations completed successfully!")
}

func RegisterUser(username, email, password string) (*models.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %v", err)
	}

	query := `INSERT INTO users (username, email, password) VALUES ($1, $2, $3) RETURNING id, created_at`
	var user models.User
	err = db.QueryRow(query, username, email, hashedPassword).Scan(&user.ID, &user.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %v", err)
	}

	user.Username = username
	user.Email = email
	user.Password = string(hashedPassword)

	return &user, nil
}

func AuthenticateUser(identifier, password string) (*models.User, error) {
	query := `SELECT id, username, email, password, tariff_id, used_traffic, subscription_start, subscription_end, created_at FROM users WHERE username = $1 OR email = $1`
	var user models.User
	var storedPassword string

	err := db.QueryRow(query, identifier).Scan(&user.ID, &user.Username, &user.Email, &storedPassword, &user.TariffID, &user.UsedTraffic, &user.SubscriptionStart, &user.SubscriptionEnd, &user.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %v", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("invalid password: %v", err)
	}

	return &user, nil
}

func GetAllTariffs() ([]models.Tariff, error) {
	query := `SELECT id, name, price, traffic_limit FROM tariffs`
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tariffs: %v", err)
	}
	defer rows.Close()

	var tariffs []models.Tariff
	for rows.Next() {
		var tariff models.Tariff
		err := rows.Scan(&tariff.ID, &tariff.Name, &tariff.Price, &tariff.TrafficLimit)
		if err != nil {
			return nil, fmt.Errorf("failed to scan tariff: %v", err)
		}
		tariffs = append(tariffs, tariff)
	}

	return tariffs, nil
}

func CreateSession(session models.Session) error {
	query := `INSERT INTO sessions (user_id, start_time, data_usage) VALUES ($1, $2, $3)`
	_, err := db.Exec(query, session.UserID, session.StartTime, session.DataUsage)
	if err != nil {
		return fmt.Errorf("failed to create session: %v", err)
	}
	return nil
}

func EndSession(session models.Session) error {
	query := `UPDATE sessions SET end_time = $1 WHERE id = $2`
	_, err := db.Exec(query, time.Now(), session.ID)
	if err != nil {
		return fmt.Errorf("failed to end session: %v", err)
	}
	return nil
}

func LinkTelegramIDToUser(userID int, telegramID int64) error {
	query := `UPDATE users SET telegram_id = $1 WHERE id = $2`
	_, err := db.Exec(query, telegramID, userID)
	if err != nil {
		logger.Printf("Failed to link Telegram ID: %v\n", err)
		return fmt.Errorf("failed to link Telegram ID: %v", err)
	}
	return nil
}

func GetUserByToken(token string) (*models.User, error) {
	var user models.User
	query := `SELECT id, username, email, password, tariff_id, used_traffic, subscription_start, subscription_end, created_at FROM users WHERE token = $1`
	err := db.QueryRow(query, token).Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.TariffID, &user.UsedTraffic, &user.SubscriptionStart, &user.SubscriptionEnd, &user.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to find user by token: %v", err)
	}
	return &user, nil
}

func UpdatePasswordByEmail(email, newPassword string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		logger.Printf("Failed to hash password: %v\n", err)
		return fmt.Errorf("failed to hash password: %v", err)
	}

	query := `UPDATE users SET password = $1 WHERE email = $2`
	_, err = db.Exec(query, hashedPassword, email)
	if err != nil {
		logger.Printf("Failed to update password for email %v: %v\n", email, err)
		return fmt.Errorf("failed to update password: %v", err)
	}

	logger.Printf("Password updated successfully for email: %v\n", email)
	return nil
}

func ProcessPayment(payment models.Payment) error {
	query := `INSERT INTO payments (user_id, amount, status) VALUES ($1, $2, $3) RETURNING id, created_at`
	err := db.QueryRow(query, payment.UserID, payment.Amount, payment.Status).Scan(&payment.ID, &payment.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to process payment: %v", err)
	}
	return nil
}
func UpdateUsedTraffic(userID int, dataUsage int64) error {
	query := `UPDATE users SET used_traffic = used_traffic + $1 WHERE id = $2`
	_, err := db.Exec(query, dataUsage, userID)
	return err
}

func UpdateSubscription(userID int) error {
	query := `UPDATE users SET subscription_end = subscription_end + INTERVAL '30 days' WHERE id = $1`
	_, err := db.Exec(query, userID)
	return err
}
