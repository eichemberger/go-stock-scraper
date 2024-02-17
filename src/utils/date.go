package utils

import (
	"fmt"
	"time"
)

type Date struct {
	Month, Day, Year string
}

func GetDate() *Date {
	now := time.Now()

	day := fmt.Sprintf("%02d", now.Day())
	month := fmt.Sprintf("%02d", int(now.Month()))
	year := fmt.Sprintf("%d", now.Year())

	return &Date{day, month, year}
}
