package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

var db *sql.DB

func InitDB() {
	var err error
	connStr := ("user=postgres password=root dbname=vpn_service sslmode=disable")
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal("Error opening database: ", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal("Error connecting to database: ", err)
	}

	fmt.Println("Successfelly connected to PostgreSQL!")

}

func GetDB() *sql.DB {
	return db
}

func RunMigration() {
	sql := `
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(100) UNIQUE NOT NULL,
			password VARCHAR(100) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
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
			data_usage INT DEFAULT 0
		);
	`
	_, err := db.Exec(sql)
	if err != nil {
		log.Fatal("Error running migration: ", err)
	}

	fmt.Println("Migrationd completed successfully!")
}

func RegisterUser(username string) error {

	db := GetDB()

	_, err := db.Exec(`INSERT INTO users (username, password) VALUES ($1, 'default')`, username)
	if err != nil {
		log.Println("Error register user: ", err)
		return err
	}
	return nil
}
