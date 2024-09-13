package handler

import (
	"encoding/json"

	"net/http"

	cases "go_final_project/internal/tasks"
)

// chFile
// обработчик POST "/api/task"
func PostTaskHandler(datab cases.Datab) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var task cases.Task
		err := json.NewDecoder(req.Body).Decode(&task)
		if err != nil {
			http.Error(w, "Ошибка десериализации JSON", http.StatusBadRequest)
			return
		}
		id, err := datab.AddTask(task)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		rule := cases.Rule{ID: id}

		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		if err := json.NewEncoder(w).Encode(rule); err != nil {
			http.Error(w, "Ошибка кодирования JSON", http.StatusInternalServerError)
			return
		}
	}
}
