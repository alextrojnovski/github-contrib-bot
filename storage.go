package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Storage работает с базой данных
type Storage struct {
	db *sql.DB
}

// NewStorage создает новое подключение к БД
func NewStorage(dbPath string) (*Storage, error) {
	// Открываем БД
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// Создаем таблицу, если её нет
	sqlStmt := `
	CREATE TABLE IF NOT EXISTS commits (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		date TEXT NOT NULL UNIQUE,
		count INTEGER DEFAULT 0,
		streak INTEGER DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	
	CREATE INDEX IF NOT EXISTS idx_date ON commits(date);
	`

	_, err = db.Exec(sqlStmt)
	if err != nil {
		return nil, err
	}

	log.Println("База данных инициализирована")
	return &Storage{db: db}, nil
}

// Close закрывает соединение с БД
func (s *Storage) Close() error {
	return s.db.Close()
}

// SaveToday сохраняет данные за сегодня
func (s *Storage) SaveToday(count int, streak int) error {
	today := time.Now().Format("2006-01-02")

	_, err := s.db.Exec(
		`INSERT INTO commits (date, count, streak) 
		 VALUES (?, ?, ?)
		 ON CONFLICT(date) DO UPDATE SET 
			count = excluded.count,
			streak = excluded.streak`,
		today, count, streak,
	)

	return err
}

// GetLastStreak возвращает streak за вчерашний день
func (s *Storage) GetLastStreak() (int, error) {
	var streak int
	err := s.db.QueryRow(
		`SELECT streak FROM commits 
		 ORDER BY date DESC LIMIT 1`,
	).Scan(&streak)

	if err == sql.ErrNoRows {
		return 0, nil // Нет записей
	}
	return streak, err
}

// GetYesterdayCount возвращает количество коммитов за вчера
func (s *Storage) GetYesterdayCount() (int, error) {
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	var count int
	err := s.db.QueryRow(
		"SELECT count FROM commits WHERE date = ?",
		yesterday,
	).Scan(&count)

	if err == sql.ErrNoRows {
		return 0, nil // Нет записи за вчера
	}
	return count, err
}

// GetStats возвращает статистику для сообщения
func (s *Storage) GetStats() (string, error) {
	var totalCommits int
	var maxStreak int
	var currentStreak int

	// Общее количество коммитов
	err := s.db.QueryRow("SELECT SUM(count) FROM commits").Scan(&totalCommits)
	if err != nil {
		return "", err
	}

	// Максимальный streak
	err = s.db.QueryRow("SELECT MAX(streak) FROM commits").Scan(&maxStreak)
	if err != nil {
		return "", err
	}

	// Текущий streak (берем последнюю запись)
	err = s.db.QueryRow("SELECT streak FROM commits ORDER BY date DESC LIMIT 1").Scan(&currentStreak)
	if err != nil && err != sql.ErrNoRows {
		return "", err
	}

	stats := fmt.Sprintf(
		"📊 Статистика:\n"+
			"📝 Всего коммитов: %d\n"+
			"🔥 Текущая серия: %d дней\n"+
			"🏆 Рекорд: %d дней",
		totalCommits, currentStreak, maxStreak,
	)

	return stats, nil
}
