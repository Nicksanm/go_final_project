package tasks

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Task struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}
type Datab struct {
	db *sql.DB
}

func NextDate(now time.Time, date string, repeat string) (string, error) {

	if repeat == "" {
		return "", fmt.Errorf("не указана строка")
	}

	nowDate, err := time.Parse("20060102", date)

	if err != nil {
		return "", fmt.Errorf("неверный формат даты: %v", err)
	}

	parts := strings.Split(repeat, " ")

	editParts := parts[0]

	switch editParts {
	case "d":
		if len(parts) < 2 {
			return "", fmt.Errorf("не указано количество дней")
		}
		moreDays, err := strconv.Atoi(parts[1])
		if err != nil || moreDays < 1 || moreDays > 400 {
			return "", fmt.Errorf("превышен максимально допустимый интервал дней")
		}
		newDate := nowDate.AddDate(0, 0, moreDays)
		for newDate.Before(now) {
			newDate = newDate.AddDate(0, 0, moreDays)
		}
		return newDate.Format("20060102"), nil

	case "y":
		newDate := nowDate.AddDate(1, 0, 0)
		for newDate.Before(now) {
			newDate = newDate.AddDate(1, 0, 0)
		}
		return newDate.Format("20060102"), nil

	default:
		return "", fmt.Errorf("неверный ввод")

	}
}
func (d *Datab) AddTask(task Task) (string, error) {
	var err error

	if task.Date == "" {
		task.Date = time.Now().Format("20060102")
	}
	if task.Title == "" {
		return "", fmt.Errorf("не указан заголовок задачи")
	}

	_, err = time.Parse("20060102", task.Date)
	if err != nil {
		return "", fmt.Errorf("неверный формат даты")
	}
	// Если дата меньше сегодняшней, устанавливаем следующую дату по правилу
	if task.Date < time.Now().Format("20060102") {
		if task.Repeat != "" {
			nextDate, err := NextDate(time.Now(), task.Date, task.Repeat)
			if err != nil {
				return "", fmt.Errorf("некорректное правило повторения")
			}
			task.Date = nextDate
		} else {
			task.Date = time.Now().Format("20060102")
		}
	}

	// Добавляем задачу в базу данных

	query := `INSERT INTO scheduler (date, title, comment, repeat) VALUES (:date, :title, :comment, :repeat)`
	res, err := d.db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		return "", fmt.Errorf("задача не добавлена")
	}

	id, err := res.LastInsertId()
	if err != nil {
		return "", fmt.Errorf("id созданной задачи не удалось вернуть")
	}
	return fmt.Sprintf("%d", id), err
}
