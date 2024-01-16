package utils

import (
	"testing"
	"time"
)

func TestJsonDate_UnmarshalJSON(t *testing.T) {
	// Define test cases with JSON date strings and expected time.Time values.
	testCases := []struct {
		jsonDateStr     string
		expectedTimeStr string
		expectError     bool
	}{
		{"\"2023-10-17\"", "2023-10-17 00:00:00 +0000 UTC", false},
		{"\"2022-05-10\"", "2022-05-10 00:00:00 +0000 UTC", false},
		{"\"invalid-date\"", "", true},
	}

	for _, tc := range testCases {
		var jd JsonDate

		err := jd.UnmarshalJSON([]byte(tc.jsonDateStr))

		if tc.expectError {
			if err == nil {
				t.Errorf("Expected an error for input %s, but got none", tc.jsonDateStr)
			}
		} else {
			if err != nil {
				t.Errorf("Unexpected error for input %s: %v", tc.jsonDateStr, err)
			}

			// Parse the expected time string.
			expectedTime, _ := time.Parse("2006-01-02 15:04:05 -0700 MST", tc.expectedTimeStr)

			// Compare the parsed JsonDate time with the expected time.
			if !jd.Time.Equal(expectedTime) {
				t.Errorf("Parsed time %v does not match the expected time %v for input %s",
					jd.Time, expectedTime, tc.jsonDateStr)
			}
		}
	}
}

func TestNewConfiguration(t *testing.T) {
	cfg := newConfiguration()

	backtesterDefaultStartDate := time.Date(
		2020,
		1,
		1,
		0,
		0,
		0,
		0,
		time.UTC,
	)
	if !cfg.BacktestStartDate.Equal(backtesterDefaultStartDate) {
		t.Errorf(
			"BacktesterStartDate is not correct on TestNewConfiguration, want %v, got %v",
			backtesterDefaultStartDate,
			cfg.BacktestStartDate,
		)
	}

	today := time.Now().Add(time.Hour * 24)
	backtesterDefaultEndDate := time.Date(
		today.Year(),
		today.Month(),
		today.Day(),
		0,
		0,
		0,
		0,
		time.UTC,
	)
	if !cfg.BacktestEndDate.Equal(backtesterDefaultEndDate) {
		t.Errorf(
			"BacktestEndDate is not correct on TestNewConfiguration, want %v, got %v",
			backtesterDefaultEndDate,
			cfg.BacktestStartDate,
		)
	}
}
