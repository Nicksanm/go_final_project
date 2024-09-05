package cases

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	LimitTasks = 30
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
	// Если дата меньше time.Now, то устанавливаем NextDate
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
		task.Date = time.Now().Format("20060102")
	}

	_, err := time.Parse("20060102", task.Date)
	if err != nil {
		return fmt.Errorf("неверный формат даты")
	}
	// Если дата меньше time.Now, то устанавливаем NextDate
	if task.Date < time.Now().Format("20060102") {
		if task.Repeat != "" {
			nextDate, err := NextDate(time.Now(), task.Date, task.Repeat)
			if err != nil {

				return fmt.Errorf("некорректное правило повторения")
			}
			task.Date = nextDate
		} else {
			task.Date = time.Now().Format("20060102")
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
		next, err := NextDate(time.Now(), task.Date, task.Repeat) // расчет следующей даты
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
