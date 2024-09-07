package cases

import (
	"database/sql"
	"fmt"

	"github.com/nicksanm/go_final_project/handler"

	"log"
	"os"
	"path/filepath"
	"time"
)

const (
	LimitTasks = 30
	DateFormat = "20060102"
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

func NewDatab(db *sql.DB) Datab {
	return Datab{db: db}
}

func (d *Datab) AddTask(task Task) (string, error) {
	var err error

	if task.Date == "" {
		task.Date = time.Now().Format(DateFormat)
	}
	if task.Title == "" {
		return "", fmt.Errorf("не указан заголовок задачи")
	}

	_, err = time.Parse(DateFormat, task.Date)
	if err != nil {
		return "", fmt.Errorf("неверный формат даты")
	}
	// Если дата меньше time.Now, то устанавливаем NextDate
	if task.Date < time.Now().Format(DateFormat) {
		if task.Repeat != "" {
			nextDate, err := handler.NextDate(time.Now(), task.Date, task.Repeat)
			if err != nil {
				return "", fmt.Errorf("некорректное правило повторения")
			}
			task.Date = nextDate
		} else {
			task.Date = time.Now().Format(DateFormat)
		}
	}

	// Добавляем задачу в базу данных
	query := `INSERT INTO scheduler (date, title, comment, repeat) VALUES (:date, :title, :comment, :repeat)`
	res, err := d.db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat)
	if err != nil {
		return "", fmt.Errorf("задача не добавлена")
	}
	//  Идентификатор созданной задачи
	id, err := res.LastInsertId()
	if err != nil {
		return "", fmt.Errorf("id созданной задачи не удалось вернуть")
	}
	return fmt.Sprintf("%d", id), err
}

// Получаем список ближайших задач
func (d *Datab) GetTasks(search string) ([]Task, error) {
	var task Task
	var tasks []Task
	var row *sql.Rows
	var err error
	row, err = d.db.Query(`SELECT * FROM scheduler ORDER BY date ASC LIMIT :limit`, sql.Named("limit", LimitTasks))

	if err != nil {
		return []Task{}, fmt.Errorf("ошибка запроса")
	}
	defer row.Close()
	for row.Next() {
		err := row.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
		if err = row.Err(); err != nil {
			return []Task{}, fmt.Errorf("ошибка распознавания данных")
		}
		tasks = append(tasks, task)
	}
	if len(tasks) == 0 {
		tasks = []Task{}
	}

	return tasks, nil
}

// Получение задачи по id
func (d *Datab) GetTask(id string) (Task, error) {
	var task Task
	if id == "" {
		return Task{}, fmt.Errorf("не указан id")
	}
	row := d.db.QueryRow("SELECT id, date, title, comment, repeat FROM scheduler WHERE id = ?", id)
	err := row.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat)
	if err != nil {
		return Task{}, fmt.Errorf("задача не найдена")
	}
	return task, nil
}

// Редактирование задачи
func (d *Datab) UpdateTask(task Task) error {
	// проверяем поля в базе данных (id, date, title)
	if task.ID == "" {
		return fmt.Errorf("не указан идентификатор")
	}
	if task.Date == "" {
		task.Date = time.Now().Format(DateFormat)
	}

	_, err := time.Parse(DateFormat, task.Date)
	if err != nil {
		return fmt.Errorf("неверный формат даты")
	}
	// Если дата меньше time.Now, то устанавливаем NextDate
	if task.Date < time.Now().Format(DateFormat) {
		if task.Repeat != "" {
			nextDate, err := handler.NextDate(time.Now(), task.Date, task.Repeat)
			if err != nil {

				return fmt.Errorf("некорректное правило повторения")
			}
			task.Date = nextDate
		} else {
			task.Date = time.Now().Format(DateFormat)
		}
	}

	if task.Title == "" {
		return fmt.Errorf("не указан заголовок задачи")
	}

	// Обновляем задачу в базе
	query := `UPDATE scheduler SET date=?, title=?, comment=?, repeat=? WHERE id=?`
	res, err := d.db.Exec(query, task.Date, task.Title, task.Comment, task.Repeat, task.ID)
	if err != nil {

		return fmt.Errorf("задача с таким id не найдена")
	}

	rowsAffected, err := res.RowsAffected() // определяем измененные строки
	if err != nil {

		return fmt.Errorf("не удалось посчитать измененные строки")
	}

	if rowsAffected == 0 {

		return fmt.Errorf("задача с таким id не найдена")
	}

	return nil
}

// Выполнение задачи
func (d *Datab) TaskDone(id string) error {
	var task Task

	task, err := d.GetTask(id)
	if err != nil {
		return err
	}
	// Одноразовая задача с пустым полем repeat удаляется
	if task.Repeat == "" {

		err := d.DeleteTask(id)
		if err != nil {
			return err
		}

	} else {
		next, err := handler.NextDate(time.Now(), task.Date, task.Repeat) // расчет следующей даты
		if err != nil {
			return err
		}
		task.Date = next //изменение у задачи значение Date
		err = d.UpdateTask(task)
		if err != nil {
			return err
		}
	}

	return nil
}

// Удаление задачи из базы данных
func (d *Datab) DeleteTask(id string) error {
	if id == "" {
		return fmt.Errorf("не указан id")
	}
	query := "DELETE FROM scheduler WHERE id = ?"
	res, err := d.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("не удалось удалить задачу")
	}

	rowsAffected, err := res.RowsAffected() // определяем измененные строки
	if err != nil {

		return fmt.Errorf("не удалось посчитать измененные строки")
	}

	if rowsAffected == 0 {

		return fmt.Errorf("задача с таким id не найдена")
	}
	return nil
}
