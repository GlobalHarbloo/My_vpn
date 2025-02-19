package models

import "time"

type User struct {
	ID                int       `json:"id"`
	Username          string    `json:"username"`
	Email             string    `json:"email"`
	Password          string    `json:"password"`
	TariffID          int       `json:"tariff_id"`
	UsedTraffic       int64     `json:"used_traffic"`
	SubscriptionStart time.Time `json:"subscription_start"`
	SubscriptionEnd   time.Time `json:"subscription_end"`
	CreatedAt         time.Time `json:"created_at"`
}

type Credentials struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

type Tariff struct {
	ID           int     `json:"id"`
	Name         string  `json:"name"`
	Price        float64 `json:"price"`
	TrafficLimit int64   `json:"traffic_limit"`
}

type Payment struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Amount    float64   `json:"amount"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type Session struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	DataUsage int64     `json:"data_usage"`
}
