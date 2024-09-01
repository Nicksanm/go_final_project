package nextdate

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

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
