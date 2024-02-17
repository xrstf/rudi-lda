// SPDX-FileCopyrightText: 2024 Christoph Mewes
// SPDX-License-Identifier: MIT

package sunnyportal

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
	"go.xrstf.de/rudi-lda/pkg/fs"
	"go.xrstf.de/rudi-lda/pkg/metrics"
)

// Subject: Sunny Portal Info Report Fam. Mewes 1/29/2022 Daily Production: 0
// kWh / Daily Revenue: 0 EUR / Daily CO2 Reduction: 0 kg

var (
	subjectRegex    = regexp.MustCompile(`Sunny Portal Info Report`)
	productionRegex = regexp.MustCompile(`Daily Production: ([0-9.,]+) kWh`)
	revenueRegex    = regexp.MustCompile(`Daily Revenue: ([0-9.,]+) EUR`)
	co2Regex        = regexp.MustCompile(`Daily CO2 Reduction: ([0-9.,]+) kg`)
	dateRegex       = regexp.MustCompile(`([0-9]+/[0-9]+/2[0-9]+)`)
)

type Proc struct {
	datadir string
}

func New(datadir string) *Proc {
	return &Proc{
		datadir: datadir,
	}
}

func (*Proc) Name() string {
	return "sunnyportal"
}

func (p *Proc) Process(_ context.Context, logger logrus.FieldLogger, msg *email.Message, _ *metrics.Metrics) (consumed bool, updated *email.Message, err error) {
	if !subjectRegex.MatchString(msg.GetSubject()) {
		return false, msg, nil
	}

	logger.Info("Handling sunnyportal.")

	info, err := parseMessage(msg)
	if err != nil {
		return false, msg, err
	}

	logFile := filepath.Join(p.datadir, "sunnyportal.csv")
	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, fs.FilePermissions)
	if err != nil {
		return false, msg, fmt.Errorf("failed to open data file: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString(fmt.Sprintf(
		"%s;%.4F kWh;%.4F EUR;%.4F kg\n",
		info.time.Format(time.RFC3339),
		info.production,
		info.revenue,
		info.co2,
	)); err != nil {
		return false, msg, fmt.Errorf("failed to append data: %w", err)
	}

	return true, nil, nil
}

type data struct {
	time       time.Time
	production float64
	revenue    float64
	co2        float64
}

func parseMessage(msg *email.Message) (*data, error) {
	subject := msg.GetSubject()

	matches := productionRegex.FindStringSubmatch(subject)
	if matches == nil {
		return nil, errors.New("failed to determine daily production.")
	}

	production, err := toFloat(matches[1])
	if err != nil {
		return nil, fmt.Errorf("failed to determine daily production: %v", err)
	}

	matches = revenueRegex.FindStringSubmatch(subject)
	if matches == nil {
		return nil, errors.New("failed to determine daily revenue.")
	}

	revenue, err := toFloat(matches[1])
	if err != nil {
		return nil, fmt.Errorf("failed to determine daily revenue: %v", err)
	}

	matches = co2Regex.FindStringSubmatch(subject)
	if matches == nil {
		return nil, errors.New("failed to determine daily CO2 reduction.")
	}

	co2, err := toFloat(matches[1])
	if err != nil {
		return nil, fmt.Errorf("failed to determine daily CO2 reduction: %v", err)
	}

	matches = dateRegex.FindStringSubmatch(subject)
	if matches == nil {
		return nil, errors.New("failed to determine date.")
	}

	parsed, err := time.ParseInLocation("1/2/2006", matches[1], time.UTC)
	if err != nil {
		return nil, fmt.Errorf("failed to parse date %q: %v", matches[1], err)
	}

	parsed = parsed.Add(12 * time.Hour)

	return &data{
		time:       parsed,
		production: production,
		revenue:    revenue,
		co2:        co2,
	}, nil
}

// toFloat turns a US-formatted (1,234.00) number into a float.
func toFloat(val string) (float64, error) {
	val = strings.ReplaceAll(val, ",", "")

	return strconv.ParseFloat(val, 64)
}
