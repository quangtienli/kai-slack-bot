package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/slack-go/slack"
)

const (
	LAYOUT_YYYYMMDD = "2006-01-02"
	LAYOUT_UTC      = "2006-01-02T15:04:00Z"
)

func DatifyHour(hourStr string, dateStr string) time.Time {
	times := strings.Split(hourStr, ":")
	hour, _ := strconv.Atoi(times[0])
	minute, _ := strconv.Atoi(times[1])

	dates := strings.Split(dateStr, "-")
	year, _ := strconv.Atoi(dates[0])
	month, _ := strconv.Atoi(dates[1])
	day, _ := strconv.Atoi(dates[2])
	return time.Date(year, time.Month(month), day, hour, minute, 0, 0, time.UTC)
}

func DatifyStringDate(str string) time.Time {
	arr := strings.Split(str, "-")
	y, _ := strconv.Atoi(arr[0])
	m, _ := strconv.Atoi(arr[1])
	d, _ := strconv.Atoi(arr[2])
	return time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC)
}

func GetHourAndMinute(hour string) (uint8, uint8) {
	s := strings.Split(hour, ":")
	// h, _ := strconv.ParseFloat(s[0], 64)
	// m, _ := strconv.ParseFloat(s[1], 64)
	// return h, m
	h, _ := strconv.Atoi(s[0])
	m, _ := strconv.Atoi(s[1])
	return uint8(h), uint8(m)
}

func GetYearAndMonthAndDay(date string) (int, uint8, uint8) {
	s := strings.Split(date, "-")
	y, _ := strconv.Atoi(s[0])
	m, _ := strconv.Atoi(s[1])
	d, _ := strconv.Atoi(s[2])
	return y, uint8(m), uint8(d)
}

func AddDate(date string, amount int) string {
	d, err := time.Parse(LAYOUT_YYYYMMDD, date)
	if err != nil {
		log.Fatalf("Unable to parse date string to date time: %s\n", err.Error())
	}
	return d.Add(24 * time.Hour * time.Duration(amount)).Format(LAYOUT_YYYYMMDD)
}

func JSONString(object any) string {
	b, err := json.Marshal(object)
	if err != nil {
		return fmt.Sprintf("Unable to convert object to json string\b")
	}
	return string(b)
}

func ToDate(str string) time.Time {
	date, err := time.Parse(LAYOUT_UTC, str)
	if err != nil {
		log.Fatalf("Unable to parse date string to date time: %s\n", err.Error())
	}

	return date
}

func ToFloat(str string) float64 {
	f, err := strconv.ParseFloat(str, 64)
	if err != nil {
		log.Fatalf("Unable to parse float string to float: %s\n", err.Error())
	}

	return f
}

func ToInt(str string) int {
	i, err := strconv.Atoi(str)
	if err != nil {
		log.Fatalf("Unable to parse int string to int: %s\n", err.Error())
	}
	return i
}

func RoundFloat(f float64, p uint) float64 {
	r := math.Pow(10, float64(p))
	return math.Round(f*r) / r
}

func ToLocaleDate(str string) time.Time {
	d, err := time.Parse(LAYOUT_YYYYMMDD, str)
	if err != nil {
		log.Fatalf("Unable to parse date string to yyyy-mm-dd date: %s\n", err.Error())
	}
	return d
}

// d1, d2 in "YYYY-MM-DDTHH:MM:SSZ" format, -1 if s1 > s2, 1 if s1 < s2, otherwise 0
func CompareStringDates(s1, s2 string) int {
	d1, _ := time.Parse(LAYOUT_YYYYMMDD, s1)
	d2, _ := time.Parse(LAYOUT_YYYYMMDD, s2)
	if d1.Before(d2) {
		return -1
	} else if d1.After(d2) {
		return 1
	}
	return 0
}

func InsertAt[T slack.Block](s []T, idx int, b T) []T {
	if len(s) == idx {
		return append(s, b)
	}

	s = append(s[:idx+1], s[idx:]...)
	s[idx] = b
	return s
}
