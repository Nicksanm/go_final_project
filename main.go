package main

import (
	handler "go_final_project/handler"
	cases "go_final_project/tasks"

	"net/http"
	"os"

	_ "modernc.org/sqlite"
)

func main() {

	db := cases.CreatDb()
	defer db.Close()
	datab := cases.NewDatab(db)
	// Определяем путь к файлу базы данных через переменную окружения
	port := "7540"
	envPort := os.Getenv("TODO_PORT")
	if len(envPort) != 0 {
		port = envPort
	}
	port = ":" + port

	http.Handle("/", http.FileServer(http.Dir("./web")))

	// обработчики:
	http.HandleFunc("/api/nextdate", handler.NextDateHandler)

	http.HandleFunc("POST /api/task", handler.PostTaskHandler(datab))
	http.HandleFunc("GET /api/task", handler.GetTaskHandler(datab))
	http.HandleFunc("/api/tasks", handler.GetTasksHandler(datab))
	http.HandleFunc("PUT /api/task", handler.PutTaskHandler(datab))
	http.HandleFunc("/api/task/done", handler.DoneTaskHandler(datab))
	http.HandleFunc("DELETE /api/task", handler.DeleteTaskHandler(datab))

	err := http.ListenAndServe(port, nil)
	if err != nil {
		panic(err)
	}

}
