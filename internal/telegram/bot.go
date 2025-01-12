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

	logger = log.New(log.Writer(), "LOG: ", log.Ldate|log.Ltime|log.Lshortfile)
	logger.Println("Bot successfully initialized.")
}

// Обработка команд
func HandleCommands(update tgbotapi.Update) {
	if update.Message == nil || update.Message.Text == "" {
		return
	}

	command, args := parseCommand(update.Message.Text)

	switch command {
	case "/start":
		sendMessage(update.Message.Chat.ID, "Welcome! Type /help for available commands.")
	case "/help":
		sendMessage(update.Message.Chat.ID, "Available commands: /register <username> <email> <password>, /login <username/email> <password>, /updatepassword <email> <new_password>")
	case "/register":
		handleRegister(update, args)
	case "/login":
		handleLogin(update, args)
	case "/updatepassword":
		handleUpdatePassword(update, args)
	default:
		sendMessage(update.Message.Chat.ID, "Unknown command. Type /help for available commands.")
	}
}

// Парсинг команды и аргументов
func parseCommand(input string) (string, []string) {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return "", nil
	}
	return parts[0], parts[1:]
}

// Отправка сообщения с обработкой ошибок
func sendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	if _, err := bot.Send(msg); err != nil {
		logger.Printf("Failed to send message: %v\n", err)
	}
}

// Обработка команды регистрации
func handleRegister(update tgbotapi.Update, args []string) {
	if len(args) < 3 {
		sendMessage(update.Message.Chat.ID, "Usage: /register <username> <email> <password>")
		return
	}

	username, email, password := args[0], args[1], args[2]
	user, err := database.RegisterUser(username, email, password)
	if err != nil {
		sendMessage(update.Message.Chat.ID, fmt.Sprintf("Error registering user: %s", err))
		return
	}

	sendMessage(update.Message.Chat.ID, fmt.Sprintf("User %s registered successfully!", user.Username))
}

// Обработка команды входа
func handleLogin(update tgbotapi.Update, args []string) {
	if len(args) < 2 {
		sendMessage(update.Message.Chat.ID, "Usage: /login <username/email> <password>")
		return
	}

	identifier, password := args[0], args[1]
	user, err := database.AuthenticateUser(identifier, password)
	if err != nil {
		sendMessage(update.Message.Chat.ID, fmt.Sprintf("Error logging in: %s", err))
		return
	}

	if err := database.LinkTelegramIDToUser(user.ID, int64(update.Message.From.ID)); err != nil {
		sendMessage(update.Message.Chat.ID, fmt.Sprintf("Error linking Telegram ID: %s", err))
		return
	}

	sendMessage(update.Message.Chat.ID, fmt.Sprintf("User %s logged in successfully!", user.Username))
}

// Обработка команды смены пароля
func handleUpdatePassword(update tgbotapi.Update, args []string) {
	if len(args) < 2 {
		sendMessage(update.Message.Chat.ID, "Usage: /updatepassword <email> <new_password>")
		return
	}

	email, newPassword := args[0], args[1]
	if err := database.UpdatePasswordByEmail(email, newPassword); err != nil {
		sendMessage(update.Message.Chat.ID, fmt.Sprintf("Error updating password: %s", err))
		return
	}

	sendMessage(update.Message.Chat.ID, "Password updated successfully!")
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
