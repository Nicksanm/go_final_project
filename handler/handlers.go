package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"

	tasks "go_final_project/tasks"
	"net/http"
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

type Reply struct {
	ID    string `json:"id,omitempty"`
	Error string `json:"error,omitempty"`
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

// обработчик "/api/nextdate"
func NextDateHandler(w http.ResponseWriter, req *http.Request) {
	now := req.FormValue("now")
	date := req.FormValue("date")
	repeat := req.FormValue("repeat")

	nowTime, err := time.Parse("20060102", now)
	if err != nil {
		http.Error(w, "неверный формат даты", http.StatusBadRequest)
		return
	}
	nextDate, err := NextDate(nowTime, date, repeat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(nextDate))
}

// Обработчик POST "POST /api/task"
func PostTaskHandler(datab tasks.Datab) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var task Task
		err := json.NewDecoder(req.Body).Decode(&task)
		if err != nil {
			http.Error(w, "Ошибка десериализации JSON", http.StatusBadRequest)
			return
		}
		id, err := datab.AddTask(tasks.Task{})
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		reply := Reply{ID: id}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			http.Error(w, "Ошибка кодирования JSON", http.StatusInternalServerError)
			return
		}
	}
}
