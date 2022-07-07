package utils

import (
	"fmt"
	"strings"
	"time"
)

func ToSpreadsheetDateTime(d time.Time) string {
	year := d.Year()
	month := d.Month()
	day := d.Day()

	hour := d.Hour()
	minute := d.Minute()
	second := d.Second()

	return fmt.Sprintf("%d/%d/%d %d:%d:%d", month, day, year, hour, minute, second)
}

func FromSpreadsheetDateTime(d string) time.Time {
	arr := strings.Split(d, " ")

	mdy := strings.Split(arr[0], "/")
	month := ToInt(mdy[0])
	day := ToInt(mdy[1])
	year := ToInt(mdy[2])

	hms := strings.Split(arr[1], ":")
	hour := ToInt(hms[0])
	minute := ToInt(hms[1])
	second := ToInt(hms[2])

	return time.Date(year, time.Month(month), day, hour, minute, second, 0, time.UTC)
}
