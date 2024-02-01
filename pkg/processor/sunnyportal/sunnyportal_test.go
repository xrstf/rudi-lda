// SPDX-FileCopyrightText: 2024 Christoph Mewes
// SPDX-License-Identifier: MIT

package sunnyportal

import (
	"math"
	"os"
	"testing"
	"time"

	"go.xrstf.de/rudi-lda/pkg/email"
)

func TestParseMessage(t *testing.T) {
	testcases := []struct {
		filename           string
		isSunnyportal      bool
		expectedTime       string
		expectedProduction float64
		expectedRevenue    float64
		expectedCO2        float64
	}{
		{
			filename:           "a.eml",
			isSunnyportal:      true,
			expectedTime:       "29 Jan 22 12:00 UTC",
			expectedProduction: 0,
			expectedRevenue:    0,
			expectedCO2:        0,
		},
		{
			filename:           "b.eml",
			isSunnyportal:      true,
			expectedTime:       "28 Jan 22 12:00 UTC",
			expectedProduction: 6.205,
			expectedRevenue:    0.496,
			expectedCO2:        4.344,
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

			if testcase.isSunnyportal != (info != nil) {
				t.Fatalf("Expected isSunnyportal=%v, but got a match mismatch", testcase.isSunnyportal)
			}

			if !testcase.isSunnyportal {
				return
			}

			expectedTime, err := time.ParseInLocation(time.RFC822, testcase.expectedTime, time.UTC)
			if err != nil {
				t.Fatalf("Failed to parse expected time: %v", err)
			}

			if !expectedTime.Equal(info.time) {
				t.Errorf("Expected time %v does not match result %v.", expectedTime, info.time)
			}

			if math.Abs(info.production-testcase.expectedProduction) > 0.0001 {
				t.Errorf("Expected production %v does not match result %v.", testcase.expectedProduction, info.production)
			}

			if math.Abs(info.revenue-testcase.expectedRevenue) > 0.0001 {
				t.Errorf("Expected revenue %v does not match result %v.", testcase.expectedRevenue, info.revenue)
			}

			if math.Abs(info.co2-testcase.expectedCO2) > 0.0001 {
				t.Errorf("Expected CO2 %v does not match result %v.", testcase.expectedCO2, info.co2)
			}
		})
	}
}
