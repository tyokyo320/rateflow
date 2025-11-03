// Package timeutil provides utility functions for time parsing and formatting.
package timeutil

import (
	"fmt"
	"time"
)

const (
	// DateFormat is the standard date format (YYYY-MM-DD)
	DateFormat = "2006-01-02"

	// DateTimeFormat is the standard datetime format
	DateTimeFormat = "2006-01-02 15:04:05"

	// ISO8601Format is the ISO 8601 datetime format
	ISO8601Format = time.RFC3339

	// CompactDateFormat is for dates without separators (YYYYMMDD)
	CompactDateFormat = "20060102"
)

// ParseDate parses a date string in YYYY-MM-DD format.
func ParseDate(s string) (time.Time, error) {
	return time.Parse(DateFormat, s)
}

// ParseCompactDate parses a date string in YYYYMMDD format.
func ParseCompactDate(s string) (time.Time, error) {
	return time.Parse(CompactDateFormat, s)
}

// FormatDate formats a time as YYYY-MM-DD.
func FormatDate(t time.Time) string {
	return t.Format(DateFormat)
}

// FormatCompactDate formats a time as YYYYMMDD.
func FormatCompactDate(t time.Time) string {
	return t.Format(CompactDateFormat)
}

// ParseDateTime parses a datetime string.
func ParseDateTime(s string) (time.Time, error) {
	return time.Parse(DateTimeFormat, s)
}

// FormatDateTime formats a time as datetime.
func FormatDateTime(t time.Time) string {
	return t.Format(DateTimeFormat)
}

// StartOfDay returns the start of the day (00:00:00).
func StartOfDay(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}

// EndOfDay returns the end of the day (23:59:59).
func EndOfDay(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 23, 59, 59, 999999999, t.Location())
}

// Today returns the start of today.
func Today() time.Time {
	return StartOfDay(time.Now())
}

// Yesterday returns the start of yesterday.
func Yesterday() time.Time {
	return StartOfDay(time.Now().AddDate(0, 0, -1))
}

// IsToday checks if the given time is today.
func IsToday(t time.Time) bool {
	now := time.Now()
	return t.Year() == now.Year() && t.YearDay() == now.YearDay()
}

// IsWeekend checks if the given time is a weekend day.
func IsWeekend(t time.Time) bool {
	weekday := t.Weekday()
	return weekday == time.Saturday || weekday == time.Sunday
}

// DaysBetween returns the number of days between two dates.
func DaysBetween(start, end time.Time) int {
	duration := end.Sub(start)
	return int(duration.Hours() / 24)
}

// ParseFlexible attempts to parse a date string in various formats.
func ParseFlexible(s string) (time.Time, error) {
	formats := []string{
		DateFormat,
		CompactDateFormat,
		DateTimeFormat,
		ISO8601Format,
		time.RFC3339,
		time.RFC3339Nano,
	}

	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse time: %s", s)
}

// ToJST converts a time to Japan Standard Time.
func ToJST(t time.Time) time.Time {
	jst, _ := time.LoadLocation("Asia/Tokyo")
	return t.In(jst)
}

// NowJST returns the current time in Japan Standard Time.
func NowJST() time.Time {
	jst, _ := time.LoadLocation("Asia/Tokyo")
	return time.Now().In(jst)
}

// DateRange generates a slice of dates between start and end (inclusive).
func DateRange(start, end time.Time) []time.Time {
	var dates []time.Time
	current := StartOfDay(start)
	endDay := StartOfDay(end)

	for !current.After(endDay) {
		dates = append(dates, current)
		current = current.AddDate(0, 0, 1)
	}

	return dates
}
