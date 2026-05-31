// Доменные ошибки мониторинга.
package model

import "errors"

var (
	// ErrPeriodNotValid — send_period не может быть больше calc_period.
	ErrPeriodNotValid = errors.New("send_period must be less than or equal to calc_period")
	// ErrStatsNotValid — не удалось распарсить вывод системной команды.
	ErrStatsNotValid = errors.New("stats output is not valid")
)
