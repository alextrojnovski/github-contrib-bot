package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {

	token := os.Getenv("TELEGRAM_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_TOKEN не установлен")
	}

	chatIDStr := os.Getenv("TELEGRAM_CHAT_ID")
	if chatIDStr == "" {
		log.Fatal("TELEGRAM_CHAT_ID не установлен")
	}

	chatID, err := strconv.ParseInt(chatIDStr, 10, 64)
	if err != nil {
		log.Fatal("TELEGRAM_CHAT_ID должен быть числом")
	}

	username := os.Getenv("GITHUB_USERNAME")
	if username == "" {
		log.Fatal("GITHUB_USERNAME не установлен")
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal("Ошибка создания бота:", err)
	}

	log.Printf("Бот авторизован: %s", bot.Self.UserName)

	log.Printf("Проверяем коммиты для пользователя: %s", username)

	commitsCount, err := GetTodayCommitsCount(username)
	if err != nil {
		log.Printf("Ошибка при получении коммитов: %v", err)
		commitsCount = -1 // Маркер ошибки
	}

	var messageText string

	if commitsCount == -1 {
		messageText = "❌ Не удалось проверить коммиты. GitHub API временно недоступен."
	} else if commitsCount == 0 {
		messageText = "😴 Сегодня ещё нет коммитов! Напиши хотя бы пару строк кода 🔥"
	} else if commitsCount == 1 {
		messageText = "👍 1 коммит сегодня. Хорошее начало!"
	} else {
		messageText = fmt.Sprintf("🚀 %d коммитов сегодня! Продуктивный день!", commitsCount)
	}

	// ============ ОТПРАВЛЯЕМ ============
	msg := tgbotapi.NewMessage(chatID, messageText)

	_, err = bot.Send(msg)
	if err != nil {
		log.Fatal("Ошибка отправки:", err)
	}

	log.Printf("Сообщение отправлено: %s", messageText)
}
