package utils_test

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"finance_tracker/internal/utils"

	"github.com/stretchr/testify/require"
)

func parseDate(t *testing.T, date string) time.Time {
	parsed, err := time.Parse(time.DateTime, date)
	require.NoError(t, err)
	return parsed
}

func checkDateInterval(t *testing.T, expected []string, result []time.Time) {
	require.Equal(t, len(expected), len(result))
	for i, day := range result {
		require.Equal(t, expected[i], day.Format(time.DateTime))
	}
}

func TestTimeIntervalToDays(t *testing.T) {
	table := []struct {
		start    string
		end      string
		expected []string
	}{
		{
			start: "2024-01-01 12:53:10",
			end:   "2024-01-05 02:53:10",
			expected: []string{
				"2024-01-01 00:00:00",
				"2024-01-02 00:00:00",
				"2024-01-03 00:00:00",
				"2024-01-04 00:00:00",
				"2024-01-05 00:00:00",
			},
		},
		{
			start:    "2024-01-01 00:00:00",
			end:      "2024-01-05 00:00:00",
			expected: []string{"2024-01-01 00:00:00", "2024-01-02 00:00:00", "2024-01-03 00:00:00", "2024-01-04 00:00:00", "2024-01-05 00:00:00"},
		},
		{
			start:    "2024-02-01 00:00:00",
			end:      "2024-02-01 00:00:00",
			expected: []string{"2024-02-01 00:00:00"},
		},
		{
			start:    "2024-03-10 00:00:00",
			end:      "2024-03-05 00:00:00",
			expected: []string{},
		},
		{
			start:    "2024-02-29 00:00:00",
			end:      "2024-03-02 00:00:00",
			expected: []string{"2024-02-29 00:00:00", "2024-03-01 00:00:00", "2024-03-02 00:00:00"},
		},
		{
			start:    "2024-02-28 00:00:00",
			end:      "2024-03-01 00:00:00",
			expected: []string{"2024-02-28 00:00:00", "2024-02-29 00:00:00", "2024-03-01 00:00:00"},
		},
		{
			start:    "2024-02-28 00:00:00",
			end:      "2024-02-28 00:00:00",
			expected: []string{"2024-02-28 00:00:00"},
		},
	}
	for _, tc := range table {
		checkDateInterval(t, tc.expected, utils.TimeIntervalToDays(parseDate(t, tc.start), parseDate(t, tc.end)))
	}
}

func TestTimeToDayInt(t *testing.T) {
	table := map[string]string{
		"2024-05-22 14:30:00": "20240522",
		"2024-05-21 14:30:00": "20240521",
		"2024-12-31 23:59:59": "20241231",
	}
	for timeStr, expectedTime := range table {
		parsedTime, err := time.Parse(time.DateTime, timeStr)
		require.NoError(t, err)
		require.Equal(t, expectedTime, utils.TimeToDayInt(parsedTime))
		expectedTimeInt, err := strconv.ParseInt(expectedTime, 10, 64)
		require.NoError(t, err)
		require.Equal(t, expectedTimeInt, utils.TimeToDayIntNum(parsedTime))
	}
}

func TestRoundToNearest15Minutes(t *testing.T) {
	tests := []struct {
		input      string
		expected1  string
		expected10 string
		expected15 string
	}{
		{
			input:      "2024-07-16 10:02:50",
			expected1:  "2024-07-16 10:02:00",
			expected10: "2024-07-16 10:00:00",
			expected15: "2024-07-16 10:00:00",
		},
		{
			input:      "2024-07-17 10:14:30",
			expected1:  "2024-07-17 10:14:00",
			expected10: "2024-07-17 10:10:00",
			expected15: "2024-07-17 10:00:00",
		},
		{
			input:      "2024-07-17 10:15:01",
			expected1:  "2024-07-17 10:15:00",
			expected10: "2024-07-17 10:10:00",
			expected15: "2024-07-17 10:15:00",
		},
		{
			input:      "2024-07-18 10:29:10",
			expected1:  "2024-07-18 10:29:00",
			expected10: "2024-07-18 10:20:00",
			expected15: "2024-07-18 10:15:00",
		},
		{
			input:      "2024-07-19 10:45:00",
			expected1:  "2024-07-19 10:45:00",
			expected10: "2024-07-19 10:40:00",
			expected15: "2024-07-19 10:45:00",
		},
		{
			input:      "2024-07-11 10:59:59",
			expected1:  "2024-07-11 10:59:00",
			expected10: "2024-07-11 10:50:00",
			expected15: "2024-07-11 10:45:00",
		},
		{
			input:      "2024-07-12 11:01:00",
			expected1:  "2024-07-12 11:01:00",
			expected10: "2024-07-12 11:00:00",
			expected15: "2024-07-12 11:00:00",
		},
	}
	for _, test := range tests {
		input, _ := time.Parse("2006-01-02 15:04:05", test.input)
		expected15, _ := time.Parse("2006-01-02 15:04:05", test.expected15)
		expected10, _ := time.Parse("2006-01-02 15:04:05", test.expected10)
		expected1, _ := time.Parse("2006-01-02 15:04:05", test.expected1)
		require.Equal(t, expected15, utils.RoundToNearest15Minutes(input))
		require.Equal(t, expected10, utils.RoundToNearest10Minutes(input))
		require.Equal(t, expected1, utils.RoundToNearest1Minute(input))
	}
}

func TestRoundToNearest5Minutes(t *testing.T) {
	testCases := map[string]string{
		"2024-01-12 14:14:23": "2024-01-12 14:10:00",
		"2024-01-12 14:26:47": "2024-01-12 14:25:00",
		"2024-01-12 14:30:00": "2024-01-12 14:30:00",
		"2024-01-12 14:34:59": "2024-01-12 14:30:00",
		"2024-01-12 14:00:00": "2024-01-12 14:00:00",
		"2024-01-12 14:07:15": "2024-01-12 14:05:00",
		"2024-01-12 14:12:45": "2024-01-12 14:10:00",
		"2024-01-12 14:55:10": "2024-01-12 14:55:00",
		"2024-01-12 14:59:59": "2024-01-12 14:55:00",
		"2024-01-12 15:00:00": "2024-01-12 15:00:00",
		"2024-01-12 23:59:59": "2024-01-12 23:55:00",
		"2024-01-12 00:00:00": "2024-01-12 00:00:00",
		"2024-01-12 00:02:30": "2024-01-12 00:00:00",
		"2024-01-12 00:03:30": "2024-01-12 00:00:00",
		"2024-01-12 00:04:30": "2024-01-12 00:00:00",
		"2024-01-12 00:05:30": "2024-01-12 00:05:00",
		"2024-01-12 12:45:00": "2024-01-12 12:45:00",
		"2024-01-12 12:46:00": "2024-01-12 12:45:00",
		"2024-01-12 12:47:00": "2024-01-12 12:45:00",
		"2024-01-12 12:48:00": "2024-01-12 12:45:00",
		"2024-01-12 12:49:00": "2024-01-12 12:45:00",
	}

	for inputStr, expectedStr := range testCases {
		input, err := time.Parse("2006-01-02 15:04:05", inputStr)
		require.NoError(t, err)
		expected, err := time.Parse("2006-01-02 15:04:05", expectedStr)
		require.NoError(t, err)

		require.Equalf(t, expected, utils.RoundToNearest5Minutes(input), "failed for %s", expectedStr)
	}
}

func TestRemoveTimezone(t *testing.T) {
	table := map[string]string{
		"2024-05-22 14:30:00 +02:00": "2024-05-22 14:30:00",
		"2024-05-21 14:30:00 -03:00": "2024-05-21 14:30:00",
		"2024-12-31 23:59:59 +01:00": "2024-12-31 23:59:59",
		"2024-01-01 00:00:00 +05:30": "2024-01-01 00:00:00",
		"2024-06-15 12:00:00 -07:00": "2024-06-15 12:00:00",
		"2024-02-29 12:00:00 +10:00": "2024-02-29 12:00:00", // Leap year
		"2023-03-10 08:45:00 +02:00": "2023-03-10 08:45:00",
		"2024-07-04 16:30:00 -04:00": "2024-07-04 16:30:00",
		"2024-11-05 05:15:00 +01:00": "2024-11-05 05:15:00",
		"2024-08-30 20:00:00 +03:00": "2024-08-30 20:00:00",
		"2024-10-31 23:59:59 +00:00": "2024-10-31 23:59:59",
	}
	layout := "2006-01-02 15:04:05 -07:00"
	for timeStr, expectedTime := range table {
		parsedTime, err := time.Parse(layout, timeStr)
		require.NoError(t, err)
		result := utils.RemoveTimezone(parsedTime)
		require.Equal(t, expectedTime, result.Format("2006-01-02 15:04:05"))
		require.Equal(t, time.UTC, result.Location())
	}
}

func TestStartOfDay(t *testing.T) {
	tests := []struct {
		input         time.Time
		expectedStart time.Time
		expectedEnd   time.Time
	}{
		{
			input:         time.Date(2024, time.November, 22, 14, 30, 45, 123456789, time.UTC),
			expectedStart: time.Date(2024, time.November, 22, 0, 0, 0, 0, time.UTC),
			expectedEnd:   time.Date(2024, time.November, 22, 23, 59, 59, 999999999, time.UTC),
		},
		{
			input:         time.Date(2024, time.January, 1, 23, 59, 59, 999999999, time.UTC),
			expectedStart: time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC),
			expectedEnd:   time.Date(2024, time.January, 1, 23, 59, 59, 999999999, time.UTC),
		},
		{
			input:         time.Date(2023, time.December, 31, 5, 0, 0, 0, time.FixedZone("UTC+5", 5*3600)),
			expectedStart: time.Date(2023, time.December, 31, 0, 0, 0, 0, time.FixedZone("UTC+5", 5*3600)),
			expectedEnd:   time.Date(2023, time.December, 31, 23, 59, 59, 999999999, time.FixedZone("UTC+5", 5*3600)),
		},
	}

	for _, test := range tests {
		require.Equal(t, test.expectedEnd, utils.EndOfDay(test.input))
		require.Equal(t, test.expectedStart, utils.StartOfDay(test.input))
	}
}

func TestTimeToRFCString(t *testing.T) {
	tests := []struct {
		input    time.Time
		expected string
	}{
		{
			input:    time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			expected: "2024-01-01T12:00:00Z",
		},
		{
			input:    time.Date(2024, 1, 1, 12, 0, 0, 0, time.FixedZone("UTC+2", 2*3600)),
			expected: "2024-01-01T12:00:00+02:00",
		},
		{
			input:    time.Time{},
			expected: "",
		},
	}

	for _, test := range tests {
		require.Equal(t, test.expected, utils.TimeToRFC3339String(test.input))
	}
}

func TestTimestampToTime(t *testing.T) {
	tests := []struct {
		input    int64
		expected time.Time
	}{
		{
			input:    time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC).Unix(),
			expected: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		},
		{
			input:    time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC).Unix(), // same instant as 12:00 +0200
			expected: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC),
		},
		{
			input:    time.Date(2024, 1, 1, 12, 0, 0, 0, time.FixedZone("UTC+2", 2*3600)).Unix(),
			expected: time.Date(2024, 1, 1, 10, 0, 0, 0, time.UTC), // should convert to UTC
		},
		{
			input:    0,
			expected: time.Time{},
		},
	}

	for _, test := range tests {
		result := utils.TimestampToTime(test.input)
		require.Equal(t, test.expected, result, "Expected %s but got %s", test.expected, result)
	}
}

func TestGetBetweenInterval(t *testing.T) {
	tests := []struct {
		start        time.Time
		interval     int
		expectedFrom time.Time
		expectedTo   time.Time
	}{
		{
			start:        time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
			interval:     30,
			expectedFrom: time.Date(2024, 1, 1, 11, 30, 0, 0, time.UTC),
			expectedTo:   time.Date(2024, 1, 1, 12, 30, 0, 0, time.UTC),
		},
		{
			start:        time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC),
			interval:     60,
			expectedFrom: time.Date(2024, 1, 1, 14, 0, 0, 0, time.UTC),
			expectedTo:   time.Date(2024, 1, 1, 16, 0, 0, 0, time.UTC),
		},
	}

	for _, test := range tests {
		from, to := utils.GetBetweenInterval(test.start, test.interval)

		require.Equal(t, test.expectedFrom, from, test.expectedFrom, from)
		require.Equal(t, test.expectedTo, to, test.expectedTo, to)
	}
}

func TestTimeToDayHourIntNum(t *testing.T) {
	cases := map[string]int64{
		"2024-01-01 12:30:00": 2024010112,
		"2024-01-01 00:00:00": 2024010100,
		"2024-01-01 23:59:59": 2024010123,
		"2024-01-02 15:45:30": 2024010215,
	}
	for timeStr, expectedInt := range cases {
		parsedTime, err := time.Parse("2006-01-02 15:04:05", timeStr)
		require.NoError(t, err)
		require.Equal(t, expectedInt, utils.TimeToDayHourIntNum(parsedTime))
		require.Equal(t, fmt.Sprintf("%d", expectedInt), utils.TimeToDayHourInt(parsedTime))
	}
}

func TestTimeToPQ(t *testing.T) {
	tests := []struct {
		name     string
		input    *time.Time
		expected string
	}{
		{"nil time", nil, ""},
		{"zero time", &time.Time{}, ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.expected, utils.TimeToPQ(test.input))
		})
	}
}

func TestTimeToDayMinuteInt(t *testing.T) {
	cases := map[string]string{
		"2024-01-01 12:30:00": "202401011230",
		"2024-01-01 00:00:00": "202401010000",
		"2024-01-01 23:59:59": "202401012359",
		"2024-01-02 15:45:30": "202401021545",
	}
	for timeStr, expected := range cases {
		parsedTime, err := time.Parse("2006-01-02 15:04:05", timeStr)
		require.NoError(t, err)
		require.Equal(t, expected, utils.TimeToDayMinuteInt(parsedTime))
	}
}
