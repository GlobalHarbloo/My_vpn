package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

var db *sql.DB
var logger *log.Logger

func InitDB() {
	// Настройка логгера
	logFile, err := os.OpenFile("server.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal("Failed to open log file: ", err)
	}
	logger = log.New(logFile, "LOG: ", log.Ldate|log.Ltime|log.Lshortfile)

	// Подключение к базе данных
	connStr := "user=postgres password=root dbname=vpn_service sslmode=disable"
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		logger.Fatal("Error opening database: ", err)
	}

	err = db.Ping()
	if err != nil {
		logger.Fatal("Error connecting to database: ", err)
	}

	logger.Println("Successfully connected to PostgreSQL!")
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

type User struct {
	ID                int
	Username          string
	Password          string
	Email             string
	TariffID          int
	UsedTraffic       int64
	SubscriptionStart sql.NullTime // Используем sql.NullTime для возможного NULL значения
	SubscriptionEnd   sql.NullTime
	CreatedAt         string
}

type Tariff struct {
	ID           int
	Name         string
	Price        float64
	TrafficLimit int64
}

// Функция авторизации пользователя
// Функция авторизации пользователя с проверкой хешированного пароля
// Функция авторизации пользователя с проверкой хешированного пароля
func AuthenticateUser(identifier, password string) (*User, error) {
	query := `
        SELECT id, username, email, COALESCE(tariff_id, 1) AS tariff_id, used_traffic, 
               subscription_start, subscription_end, created_at, password
        FROM users
        WHERE (username = $1 OR email = $1)
    `
	var user User
	var storedPassword string

	// Запросим пользователя и его хешированный пароль
	err := db.QueryRow(query, identifier).Scan(
		&user.ID, &user.Username, &user.Email, &user.TariffID, &user.UsedTraffic,
		&user.SubscriptionStart, &user.SubscriptionEnd, &user.CreatedAt, &storedPassword,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("error authenticating user: %v", err)
	}

	// Сравниваем пароли
	err = bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("incorrect password")
	}

	return &user, nil
}

// Привязка Telegram ID к пользователю
func LinkTelegramIDToUser(userID int, telegramID int64) error {
	query := `UPDATE users SET telegram_id = $1 WHERE id = $2`
	_, err := db.Exec(query, telegramID, userID)
	if err != nil {
		logger.Printf("Failed to link Telegram ID: %v\n", err)
		return fmt.Errorf("failed to link Telegram ID: %v", err)
	}
	return nil
}

// Поиск пользователя по Telegram ID
func GetUserByTelegramID(telegramID int64) (*User, error) {
	var user User

	query := `SELECT id, username, email, password, tariff_id, used_traffic, subscription_start, subscription_end, created_at 
	FROM users WHERE telegram_id = $1`
	err := db.QueryRow(query, telegramID).Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.TariffID,
		&user.UsedTraffic, &user.SubscriptionStart, &user.SubscriptionEnd, &user.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.Printf("User not found by Telegram ID: %v\n", telegramID)
			return nil, fmt.Errorf("user not found")
		}
		logger.Printf("Error fetching user by Telegram ID: %v\n", err)
		return nil, err
	}

	return &user, nil
}

// Обновление пароля по email
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

func GetAllTariffs() ([]Tariff, error) {
	var tariffs []Tariff

	query := `SELECT id, name, price, traffic_limit FROM tariffs`
	rows, err := db.Query(query)
	if err != nil {
		logger.Printf("Error fetching tariffs: %v\n", err)
		return nil, fmt.Errorf("error fetching tariffs: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var tariff Tariff
		err = rows.Scan(&tariff.ID, &tariff.Name, &tariff.Price, &tariff.TrafficLimit)
		if err != nil {
			logger.Printf("Error scanning tariff row: %v\n", err)
			return nil, fmt.Errorf("error scanning tariff row: %v", err)
		}
		tariffs = append(tariffs, tariff)
	}

	if err = rows.Err(); err != nil {
		logger.Printf("Error in rows iteration: %v\n", err)
		return nil, fmt.Errorf("error in rows iteration: %v", err)
	}

	return tariffs, nil
}

// RegisterUser регистрирует нового пользователя в базе данных
func RegisterUser(username, email, password string) (*User, error) {
	// Проверка уникальности имени пользователя и email
	var exists bool
	checkQuery := `
		SELECT EXISTS(
			SELECT 1 
			FROM users 
			WHERE username = $1 OR email = $2
		)
	`
	err := db.QueryRow(checkQuery, username, email).Scan(&exists)
	if err != nil {
		logger.Printf("Error checking user uniqueness: %v\n", err)
		return nil, fmt.Errorf("error checking user uniqueness: %v", err)
	}
	if exists {
		return nil, fmt.Errorf("username or email already exists")
	}

	// Хеширование пароля
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		logger.Printf("Failed to hash password: %v\n", err)
		return nil, fmt.Errorf("failed to hash password: %v", err)
	}

	// Вставка нового пользователя в базу данных
	insertQuery := `
		INSERT INTO users (username, email, password, tariff_id, created_at)
		VALUES ($1, $2, $3, 1, CURRENT_TIMESTAMP)  -- тариф по умолчанию 1
		RETURNING id, username, email, tariff_id, used_traffic, subscription_start, subscription_end, created_at
	`
	var user User
	err = db.QueryRow(insertQuery, username, email, hashedPassword).Scan(
		&user.ID, &user.Username, &user.Email, &user.TariffID, &user.UsedTraffic,
		&user.SubscriptionStart, &user.SubscriptionEnd, &user.CreatedAt,
	)
	if err != nil {
		logger.Printf("Error registering user: %v\n", err)
		return nil, fmt.Errorf("error registering user: %v", err)
	}

	logger.Printf("User registered successfully: %s\n", username)
	return &user, nil
}
