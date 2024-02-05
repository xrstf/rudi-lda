// SPDX-FileCopyrightText: 2024 Christoph Mewes
// SPDX-License-Identifier: MIT

package rentablo

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"go.xrstf.de/rudi-lda/pkg/email"
	"go.xrstf.de/rudi-lda/pkg/metrics"
)

var (
	subjectRegex            = regexp.MustCompile(`Ihr Rentablo Investment-Report`)
	performance1WeekRegex   = regexp.MustCompile(`• ([0-9.,-]+) %\s+seit 7 Tagen`)
	performance1MonthRegex  = regexp.MustCompile(`• ([0-9.,-]+) %\s+seit einem Monat`)
	performance6MonthsRegex = regexp.MustCompile(`Seit 6 Monaten:\s+Sie:\s+([0-9-.,]+) %`)
	performance1YearRegex   = regexp.MustCompile(`Seit 12 Monaten:\s+Sie:\s+([0-9-.,]+) %`)
)

type Proc struct {
	datadir string
}

func New(datadir string) *Proc {
	return &Proc{
		datadir: datadir,
	}
}

func (p *Proc) Matches(_ context.Context, _ logrus.FieldLogger, msg *email.Message) (bool, error) {
	return subjectRegex.MatchString(msg.GetSubject()), nil
}

func (p *Proc) Process(_ context.Context, logger logrus.FieldLogger, msg *email.Message, _ *metrics.Metrics) error {
	logger.Info("Handling rentablo.")

	info, err := parseMessage(msg)
	if err != nil {
		return err
	}

	logFile := filepath.Join(p.datadir, "rentablo.csv")
	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open data file: %w", err)
	}
	defer f.Close()

	f.WriteString(fmt.Sprintf(
		"%s;%.2F;%.2F;%.2F;%.2F\n",
		info.time.Format(time.RFC3339),
		info.performance1Week,
		info.performance1Month,
		info.performance6Months,
		info.performance1Year,
	))

	return nil
}

type data struct {
	time               time.Time
	performance1Week   float64
	performance1Month  float64
	performance6Months float64
	performance1Year   float64
}

func parseMessage(msg *email.Message) (*data, error) {
	body, err := msg.GetMultipartBody("text/plain")
	if err != nil {
		return nil, fmt.Errorf("failed to parse body: %w", err)
	}
	if body == "" {
		return nil, errors.New("mail has no text/plain part")
	}

	performance1Week, err := parseValue(body, *performance1WeekRegex)
	if err != nil {
		return nil, fmt.Errorf("failed to determine 1-week performance: %w", err)
	}

	performance1Month, err := parseValue(body, *performance1MonthRegex)
	if err != nil {
		return nil, fmt.Errorf("failed to determine 1-month performance: %w", err)
	}

	performance6Months, err := parseValue(body, *performance6MonthsRegex)
	if err != nil {
		return nil, fmt.Errorf("failed to determine 6-month performance: %w", err)
	}

	performance1Year, err := parseValue(body, *performance1YearRegex)
	if err != nil {
		return nil, fmt.Errorf("failed to determine 1-year performance: %w", err)
	}

	parsed, err := determineTime(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to get date: %w", err)
	}

	return &data{
		time:               parsed,
		performance1Week:   performance1Week,
		performance1Month:  performance1Month,
		performance6Months: performance6Months,
		performance1Year:   performance1Year,
	}, nil
}

func determineTime(msg *email.Message) (time.Time, error) {
	parsed, err := msg.GetDate()
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to get date: %w", err)
	}

	for parsed.Weekday() != time.Sunday {
		parsed = parsed.AddDate(0, 0, 1)
	}

	y, m, d := parsed.Date()

	return time.Date(y, m, d, 12, 0, 0, 0, time.UTC), nil
}

func parseValue(body string, r regexp.Regexp) (float64, error) {
	matches := r.FindStringSubmatch(body)
	if matches == nil {
		fmt.Println(body)
		return 0, errors.New("regexp did not match")
	}

	value, err := toFloat(matches[1])
	if err != nil {
		return 0, fmt.Errorf("invalid number %q: %w", matches[1], err)
	}

	return value, nil
}

// toFloat turns a DE-formatted (1.234,00) number into a float.
func toFloat(val string) (float64, error) {
	val = strings.ReplaceAll(val, ".", "")
	val = strings.ReplaceAll(val, ",", ".")

	return strconv.ParseFloat(val, 64)
}
