package main

import (
	"database/sql"
	"go_final_project/handler"
	"go_final_project/tasks"
	"log"
	"net/http"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// Создаем базу данных
func CreatDb() *sql.DB {

	appPath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	dbFile := filepath.Join(filepath.Dir(appPath), "scheduler.db")
	_, err = os.Stat(dbFile)
	// путь к файлу базы данных через переменную окружения.
	envFile := os.Getenv("TODO_DBFILE")
	if len(envFile) > 0 {
		dbFile = envFile
	}
	log.Println("Путь к базе данных", dbFile)

	var install bool
	if err != nil {
		install = true
	}
	// если install равен true, после открытия БД требуется выполнить
	// sql-запрос с CREATE TABLE и CREATE INDEX
	db, err := sql.Open("sqlite", "scheduler.db")
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	if install {
		Table := `CREATE TABLE IF NOT EXISTS scheduler (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				date CHAR(8) NOT NULL,
				title TEXT NOT NULL,
				comment TEXT, 
				repeat VARCHAR(128) NOT NULL
				);`
		_, err = db.Exec(Table)
		if err != nil {
			log.Fatal(err)
		}

		Index := `CREATE INDEX IF NOT EXISTS scheduler_date ON scheduler(date);`
		_, err = db.Exec(Index)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("База данных создана")
	}
	return db
}

func main() {

	db := CreatDb()
	defer db.Close()

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
	http.HandleFunc("POST /api/task", handler.PostTaskHandler(tasks.Datab{}))

	err := http.ListenAndServe(port, nil)
	if err != nil {
		panic(err)
	}

}
