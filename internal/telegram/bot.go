package telegram

import (
	"fmt"
	"log"
	"strings"
	"vpn-service/internal/database"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

var bot *tgbotapi.BotAPI
var logger *log.Logger

// Инициализация бота
func InitBot(token string) {
	var err error
	bot, err = tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal("Failed to create bot: ", err)
	}

	// Настройка логгера
	logger = log.New(log.Writer(), "LOG: ", log.Ldate|log.Ltime|log.Lshortfile)
	logger.Println("Bot successfully initialized.")
}

// Обработчик команд
func HandleCommands(update tgbotapi.Update) {
	if update.Message == nil {
		return
	}

	// Ответ на команду /start
	if update.Message.Text == "/start" {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Welcome! Type /help for available commands.")
		bot.Send(msg)
		return
	}

	// Ответ на команду /help
	if update.Message.Text == "/help" {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Available commands: /register, /login, /updatepassword")
		bot.Send(msg)
		return
	}

	// Ответ на команду /register
	if strings.HasPrefix(update.Message.Text, "/register") {
		parts := strings.SplitN(update.Message.Text, " ", 4)
		if len(parts) < 4 {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Usage: /register <username> <email> <password>")
			bot.Send(msg)
			return
		}

		username, email, password := parts[1], parts[2], parts[3]
		user, err := database.RegisterUser(username, email, password)
		if err != nil {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Error registering user: %s", err))
			bot.Send(msg)
			return
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("User %s registered successfully!", user.Username))
		bot.Send(msg)
		return
	}

	// Ответ на команду /login
	if strings.HasPrefix(update.Message.Text, "/login") {
		parts := strings.SplitN(update.Message.Text, " ", 3)
		if len(parts) < 3 {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Usage: /login <username/email> <password>")
			bot.Send(msg)
			return
		}

		identifier, password := parts[1], parts[2]
		user, err := database.AuthenticateUser(identifier, password)
		if err != nil {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Error logging in: %s", err))
			bot.Send(msg)
			return
		}

		// Сохранение Telegram ID
		err = database.LinkTelegramIDToUser(user.ID, int64(update.Message.From.ID))
		if err != nil {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Error linking Telegram ID: %s", err))
			bot.Send(msg)
			return
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("User %s logged in successfully!", user.Username))
		bot.Send(msg)
		return
	}

	// Ответ на команду /updatepassword
	if strings.HasPrefix(update.Message.Text, "/updatepassword") {
		parts := strings.SplitN(update.Message.Text, " ", 3)
		if len(parts) < 3 {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Usage: /updatepassword <email> <new_password>")
			bot.Send(msg)
			return
		}

		email, newPassword := parts[1], parts[2]
		err := database.UpdatePasswordByEmail(email, newPassword)
		if err != nil {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Error updating password: %s", err))
			bot.Send(msg)
			return
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Password updated successfully!")
		bot.Send(msg)
		return
	}
}

// Запуск бота
func StartBot() {
	if bot == nil {
		log.Fatal("Bot is not initialized.")
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Fatal("Error getting updates: ", err)
	}

	for update := range updates {
		HandleCommands(update)
	}
}
