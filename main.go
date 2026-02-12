package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	// ============ ЗАГРУЖАЕМ КОНФИГ ============
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

	username := os.Getenv("GH_USERNAME")
	if username == "" {
		log.Fatal("GH_USERNAME не установлен")
	}

	// ============ ИНИЦИАЛИЗИРУЕМ БАЗУ ДАННЫХ ============
	storage, err := NewStorage("commits.db")
	if err != nil {
		log.Fatal("Ошибка инициализации БД:", err)
	}
	defer storage.Close()

	// ============ СОЗДАЕМ БОТА ============
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal("Ошибка создания бота:", err)
	}

	log.Printf("Бот авторизован: %s", bot.Self.UserName)

	// ============ ПРОВЕРЯЕМ КОММИТЫ ============
	log.Printf("Проверяем коммиты для пользователя: %s", username)

	commitsCount, err := GetTodayCommitsCount(username)
	if err != nil {
		log.Printf("Ошибка при получении коммитов: %v", err)
		commitsCount = -1
	}

	// ============ ВЫЧИСЛЯЕМ STREAK ============
	var streak int

	if commitsCount > 0 {
		// Сегодня были коммиты
		yesterdayCount, err := storage.GetYesterdayCount()
		if err != nil {
			log.Printf("Ошибка получения вчерашних коммитов: %v", err)
		}

		lastStreak, err := storage.GetLastStreak()
		if err != nil {
			log.Printf("Ошибка получения последнего streak: %v", err)
		}

		if yesterdayCount > 0 {
			// Вчера тоже были коммиты -> продолжаем серию
			streak = lastStreak + 1
		} else {
			// Вчера не было -> начинаем новую серию
			streak = 1
		}
	} else {
		// Сегодня нет коммитов -> серия прервана
		streak = 0
	}

	// ============ СОХРАНЯЕМ ДАННЫЕ ============
	err = storage.SaveToday(commitsCount, streak)
	if err != nil {
		log.Printf("Ошибка сохранения в БД: %v", err)
	}

	// ============ ФОРМИРУЕМ СООБЩЕНИЕ ============
	var messageText string

	if commitsCount == -1 {
		messageText = "❌ Не удалось проверить коммиты. GitHub API временно недоступен."
	} else if commitsCount == 0 {
		messageText = "😴 Сегодня ещё нет коммитов! Серия прервана.\n"
		if streak > 0 {
			messageText += "🔥 Было " + strconv.Itoa(streak) + " дней подряд!"
		}
	} else if commitsCount == 1 {
		messageText = fmt.Sprintf("👍 1 коммит сегодня! ", commitsCount)
		if streak > 0 {
			messageText += fmt.Sprintf("🔥 Текущая серия: %d дней", streak)
		}
	} else {
		messageText = fmt.Sprintf("🚀 %d коммитов сегодня! ", commitsCount)
		if streak > 0 {
			messageText += fmt.Sprintf("🔥 Серия: %d дней", streak)
		}
	}

	// Добавляем статистику
	stats, err := storage.GetStats()
	if err == nil {
		messageText += "\n\n" + stats
	}

	// ============ ОТПРАВЛЯЕМ ============
	msg := tgbotapi.NewMessage(chatID, messageText)

	_, err = bot.Send(msg)
	if err != nil {
		log.Fatal("Ошибка отправки:", err)
	}

	log.Printf("Сообщение отправлено. Streak: %d, Коммиты: %d", streak, commitsCount)
}
