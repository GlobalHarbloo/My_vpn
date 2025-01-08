package telegram

import (
	"encoding/json"
	"log"
	"net/http"
	"vpn-service/internal/database"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func StartBot() {
	token := "7427268402:AAEDRhYBSdBHFZ4AvBJUq0iSlhnDA3dLH88"
	bot, err := tgbotapi.NewBotAPI(token)

	if err != nil {
		log.Fatal("Failed to create bot: ", err)
	}
	log.Printf("Authorized on account %s", bot.Self.UserName)

	// Устанавливаем webhook через ngrok
	ngrokURL := "https://fda3-51-75-145-220.ngrok-free.app"
	_, err = bot.SetWebhook(tgbotapi.NewWebhook(ngrokURL + "/path"))
	if err != nil {
		log.Fatal("Failed to set webhook: ", err)
	}

	// Запускаем сервер для обработки webhook
	http.HandleFunc("/path", func(w http.ResponseWriter, r *http.Request) {
		var update tgbotapi.Update
		// Парсим тело запроса в структуру update
		err := json.NewDecoder(r.Body).Decode(&update)
		if err != nil {
			log.Printf("Error decoding update: %v", err)
			http.Error(w, "Failed to parse update", http.StatusInternalServerError)
			return
		}

		// Обрабатываем полученные обновления
		log.Printf("Update received: %+v", update)

		if update.Message == nil {
			return
		}

		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "start":
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Welcome to the VPN service bot!")
				bot.Send(msg)
			case "register":
				handleRegister(update, bot)
			}
			log.Printf("Command is: %s", update.Message.Command())
		}
	})

	// Запускаем HTTP-сервер для получения запросов от Telegram
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleRegister(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	username := update.Message.Text

	err := database.RegisterUser(username)
	if err != nil {
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Failed to register user!"))
		return
	}
	bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "User registered successfully!"))
}
