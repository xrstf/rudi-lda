// SPDX-FileCopyrightText: 2024 Christoph Mewes
// SPDX-License-Identifier: MIT

package metrics

import (
	"encoding/json"
	"fmt"
	"os"
)

type Metrics struct {
	Total     int            `json:"total"`
	Valid     int            `json:"valid"`
	Discarded int            `json:"discarded"`
	Folders   map[string]int `json:"folders"`
	SpamRules map[string]int `json:"spamRules"`
}

func Load(filename string) (*Metrics, error) {
	m := Metrics{
		Folders:   map[string]int{},
		SpamRules: map[string]int{},
	}

	if filename == "" {
		return &m, nil
	}

	f, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return &m, nil
		}

		return &m, fmt.Errorf("failed to open '%s': %v", filename, err)
	}
	defer f.Close()

	if err := json.NewDecoder(f).Decode(&m); err != nil {
		return &m, fmt.Errorf("failed to decode '%s': %v", filename, err)
	}

	return &m, nil
}

func Save(filename string, m *Metrics) error {
	if filename == "" {
		return nil
	}

	f, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create '%s': %v", filename, err)
	}
	defer f.Close()

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(m); err != nil {
		return fmt.Errorf("failed to encode metrics: %v", err)
	}

	return nil
}
