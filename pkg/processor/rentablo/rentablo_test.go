// SPDX-FileCopyrightText: 2024 Christoph Mewes
// SPDX-License-Identifier: MIT

package rentablo

import (
	"math"
	"os"
	"testing"
	"time"

	"go.xrstf.de/rudi-lda/pkg/email"
)

func TestParseMessage(t *testing.T) {
	testcases := []struct {
		filename                   string
		isRentablo                 bool
		expectedTime               string
		expected1WeekPerformance   float64
		expected1MonthPerformance  float64
		expected6MonthsPerformance float64
		expected1YearPerformance   float64
	}{
		{
			filename:                   "a.eml",
			isRentablo:                 true,
			expectedTime:               "30 Jan 22 12:00 UTC",
			expected1WeekPerformance:   0.05,
			expected1MonthPerformance:  -7.62,
			expected6MonthsPerformance: 0.9,
			expected1YearPerformance:   17.1,
		},
		{
			filename:                   "b.eml",
			isRentablo:                 true,
			expectedTime:               "23 Jan 22 12:00 UTC",
			expected1WeekPerformance:   -4.99,
			expected1MonthPerformance:  -5.94,
			expected6MonthsPerformance: 1.3,
			expected1YearPerformance:   13.4,
		},
		{
			filename:                   "c.eml",
			isRentablo:                 true,
			expectedTime:               "16 Jan 22 12:00 UTC",
			expected1WeekPerformance:   -1.02,
			expected1MonthPerformance:  -0.84,
			expected6MonthsPerformance: 6.6,
			expected1YearPerformance:   21.8,
		},
		{
			filename:                   "d.eml",
			isRentablo:                 true,
			expectedTime:               "25 Sep 22 12:00 UTC",
			expected1WeekPerformance:   -2.71,
			expected1MonthPerformance:  -7.81,
			expected6MonthsPerformance: -9.1,
			expected1YearPerformance:   -8.5,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.filename, func(t *testing.T) {
			body, err := os.ReadFile("testdata/" + testcase.filename)
			if err != nil {
				t.Fatalf("Failed to read testdata: %v", err)
			}

			msg, err := email.ParseMessage(body)
			if err != nil {
				t.Fatalf("Failed to parse mail body: %v", err)
			}

			info, err := parseMessage(msg)
			if err != nil {
				t.Fatalf("Failed to parse mail: %v", err)
			}

			if testcase.isRentablo != (info != nil) {
				t.Fatalf("Expected isRentablo=%v, but got a match mismatch", testcase.isRentablo)
			}

			if !testcase.isRentablo {
				return
			}

			expectedTime, err := time.ParseInLocation(time.RFC822, testcase.expectedTime, time.UTC)
			if err != nil {
				t.Fatalf("Failed to parse expected time: %v", err)
			}

			if !expectedTime.Equal(info.time) {
				t.Errorf("Expected time %v does not match result %v.", expectedTime, info.time)
			}

			if math.Abs(info.performance1Week-testcase.expected1WeekPerformance) > 0.0001 {
				t.Errorf("Expected 1-week performance %v does not match result %v.", testcase.expected1WeekPerformance, info.performance1Week)
			}

			if math.Abs(info.performance1Month-testcase.expected1MonthPerformance) > 0.0001 {
				t.Errorf("Expected 1-month performance %v does not match result %v.", testcase.expected1MonthPerformance, info.performance1Month)
			}

			if math.Abs(info.performance6Months-testcase.expected6MonthsPerformance) > 0.0001 {
				t.Errorf("Expected 6-months performance %v does not match result %v.", testcase.expected6MonthsPerformance, info.performance6Months)
			}

			if math.Abs(info.performance1Year-testcase.expected1YearPerformance) > 0.0001 {
				t.Errorf("Expected 1-year performance %v does not match result %v.", testcase.expected1YearPerformance, info.performance1Year)
			}
		})
	}
}
