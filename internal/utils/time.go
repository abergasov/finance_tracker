package utils

import (
	"strconv"
	"time"
)

func TimeIntervalToDays(start, end time.Time) []time.Time {
	var days []time.Time
	start = StartOfDay(start)
	end = EndOfDay(end)
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		days = append(days, d)
	}
	return days
}

func RoundToNearest1Minute(t time.Time) time.Time {
	roundedMinutes := (t.Minute() / 1) * 1
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), roundedMinutes, 0, 0, t.Location())
}

func RoundToNearest5Minutes(t time.Time) time.Time {
	roundedMinutes := (t.Minute() / 5) * 5
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), roundedMinutes, 0, 0, t.Location())
}

func RoundToNearest10Minutes(t time.Time) time.Time {
	roundedMinutes := (t.Minute() / 10) * 10
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), roundedMinutes, 0, 0, t.Location())
}

func RoundToNearest15Minutes(t time.Time) time.Time {
	roundedMinutes := (t.Minute() / 15) * 15
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), roundedMinutes, 0, 0, t.Location())
}

// RemoveTimezone removes the timezone information from a time.Time object
// without changing the hour, minute, second, and nanosecond values.
func RemoveTimezone(t time.Time) time.Time {
	// Extract the year, month, day, hour, minute, second, and nanosecond
	year, month, day := t.Date()
	hour, minute, second := t.Clock()
	nanosecond := t.Nanosecond()

	// Create a new time.Time object without timezone (local time)
	localTime := time.Date(year, month, day, hour, minute, second, nanosecond, time.UTC)

	return localTime
}

func TimeToRFC3339String(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}

func TimeToDayInt(src time.Time) string {
	return src.Format("20060102")
}

func TimeToDayHourInt(src time.Time) string {
	return src.Format("2006010215")
}

func TimeToDayHourIntNum(src time.Time) int64 {
	i, _ := strconv.ParseInt(TimeToDayHourInt(src), 10, 64) // we can skip err here
	return i
}

func TimeToDayIntNum(src time.Time) int64 {
	i, _ := strconv.ParseInt(TimeToDayInt(src), 10, 64) // we can skip err here
	return i
}

func TimeToDayMinuteInt(src time.Time) string {
	return src.Format("200601021504")
}

func TimeToDayMinuteIntNum(src time.Time) int64 {
	i, _ := strconv.ParseInt(TimeToDayMinuteInt(src), 10, 64) // we can skip err here
	return i
}

func StartOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func EndOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, int(time.Second-time.Nanosecond), t.Location())
}

// TimeToPQ converts a time.Time to a string formatted for parquet files.
func TimeToPQ(t *time.Time) string {
	if t == nil || t.IsZero() {
		return ""
	}
	return t.Format(time.DateTime)
}

// TimestampToTime expect timestamp is in seconds
func TimestampToTime(timestamp int64) time.Time {
	if timestamp == 0 {
		return time.Time{}
	}
	return time.Unix(timestamp, 0).UTC()
}

func GetBetweenInterval(start time.Time, intervalMinutes int) (from, to time.Time) {
	from = start.Add(-(time.Minute * time.Duration(intervalMinutes)))
	to = start.Add(time.Minute * time.Duration(intervalMinutes))
	return from, to
}
