package handler

import (
	"encoding/json"
	"errors"

	"net/http"
	"time"

	cases "go_final_project/tasks"
)

type Reply struct {
	ID    string `json:"id,omitempty"`
	Error string `json:"error,omitempty"`
}

var ErrorResponses struct {
	Error string `json:"error,omitempty"`
}

// обработчик "/api/nextdate"
func NextDateHandler(w http.ResponseWriter, req *http.Request) {
	now := req.FormValue("now")
	date := req.FormValue("date")
	repeat := req.FormValue("repeat")

	nowTime, err := time.Parse(cases.DateFormat, now)
	if err != nil {
		http.Error(w, "неверный формат даты", http.StatusBadRequest)
		return
	}
	nextDate, err := cases.NextDate(nowTime, date, repeat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(nextDate))
}

// обработчик POST "POST /api/task"
func PostTaskHandler(datab cases.Datab) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var task cases.Task
		err := json.NewDecoder(req.Body).Decode(&task)
		if err != nil {
			http.Error(w, "ошибка десериализации JSON", http.StatusBadRequest)
			return
		}
		id, err := datab.AddTask(cases.Task{})
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		reply := Reply{ID: id}

		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			http.Error(w, "ошибка кодирования JSON", http.StatusInternalServerError)
			return
		}
	}
}

// обработчик GET для "GET /api/task"
func GetTaskHandler(datab cases.Datab) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var task cases.Task
		id := req.URL.Query().Get("id")
		task, err := datab.GetTask(id)
		if err != nil {
			err := errors.New("задача с таким id не найдена")
			ErrorResponses.Error = err.Error()
			json.NewEncoder(w).Encode(ErrorResponses)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		if err := json.NewEncoder(w).Encode(task); err != nil {
			http.Error(w, "ошибка кодирования JSON", http.StatusInternalServerError)
			return
		}
	}
}

// обработчик GET для "/api/tasks"
func GetTasksHandler(datab cases.Datab) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		tasks, err := datab.GetTasks()

		if err != nil {
			err := errors.New("ошибка запроса к базе данных")
			ErrorResponses.Error = err.Error()
			json.NewEncoder(w).Encode(ErrorResponses)
			return
		}

		reply := map[string][]cases.Task{
			"tasks": tasks,
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		if err := json.NewEncoder(w).Encode(reply); err != nil {
			http.Error(w, "ошибка кодирования JSON", http.StatusInternalServerError)
			return
		}
	}
}

// обработчик PUT для "PUT /api/task"
func PutTaskHandler(datab cases.Datab) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var task cases.Task
		err := json.NewDecoder(req.Body).Decode(&task)
		if err != nil {
			http.Error(w, "ошибка десериализации JSON", http.StatusBadRequest)
			return
		}
		err = datab.UpdateTask(cases.Task{})
		if err != nil {
			http.Error(w, "задача не найдена", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		if err := json.NewEncoder(w).Encode(map[string]string{}); err != nil {
			http.Error(w, "ошибка кодирования JSON", http.StatusInternalServerError)
			return
		}
	}
}

// обработчик для "/api/task/done"
func DoneTaskHandler(datab cases.Datab) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		id := req.URL.Query().Get("id")
		err := datab.TaskDone(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		if err := json.NewEncoder(w).Encode(map[string]string{}); err != nil {
			http.Error(w, "ошибка кодирования JSON", http.StatusInternalServerError)
			return
		}
	}
}

// обработчик DELETE для "DELETE /api/task"
func DeleteTaskHandler(datab cases.Datab) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		id := req.URL.Query().Get("id")
		err := datab.DeleteTask(id)
		if err != nil {
			err := errors.New("задача с таким id не найдена")
			ErrorResponses.Error = err.Error()
			json.NewEncoder(w).Encode(ErrorResponses)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		if err := json.NewEncoder(w).Encode(map[string]string{}); err != nil {
			http.Error(w, "ошибка кодирования JSON", http.StatusInternalServerError)
			return
		}
	}
}

// эндпоинт
func TaskHandler(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		GetTaskHandler(cases.Datab{})
	case http.MethodPost:
		PostTaskHandler(cases.Datab{})
	case http.MethodPut:
		PutTaskHandler(cases.Datab{})
	case http.MethodDelete:
		DeleteTaskHandler(cases.Datab{})
	default:
		http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
	}
}
