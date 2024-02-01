// SPDX-FileCopyrightText: 2024 Christoph Mewes
// SPDX-License-Identifier: MIT

package spam

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"go.xrstf.de/rudi-lda/pkg/email"
	"go.xrstf.de/rudi-lda/pkg/rudilib"
)

type Status string

const (
	Spam      Status = "spam"
	MaybeSpam Status = "maybe-spam"
	Ham       Status = "ham"
)

type Result struct {
	Status Status `json:"status"`
	Rule   string `json:"rule"`
}

func Check(ctx context.Context, scriptFile string, msg *email.Message) (*Result, error) {
	result, err := rudilib.ProcessMessage(ctx, scriptFile, msg, nil, Functions)
	if err != nil {
		return nil, fmt.Errorf("script failed: %w", err)
	}

	return parseResult(result), nil
}

func parseResult(result any) *Result {
	if result == nil {
		return nil
	}

	var buf bytes.Buffer

	if err := json.NewEncoder(&buf).Encode(result); err != nil {
		return nil
	}

	var r Result
	if err := json.NewDecoder(&buf).Decode(&r); err != nil {
		return nil
	}

	return &r
}
