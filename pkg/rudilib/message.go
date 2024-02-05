// SPDX-FileCopyrightText: 2024 Christoph Mewes
// SPDX-License-Identifier: MIT

package rudilib

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.xrstf.de/rudi"
	"go.xrstf.de/rudi-contrib/set"
	"go.xrstf.de/rudi/pkg/coalescing"

	"go.xrstf.de/rudi-lda/pkg/email"
)

func ProcessMessage(ctx context.Context, scriptFile string, msg *email.Message, extraVars rudi.Variables, extraFuncs rudi.Functions) (any, error) {
	program, err := loadProgram(scriptFile)
	if err != nil {
		return nil, fmt.Errorf("invalid script: %w", err)
	}

	if program == nil {
		return nil, nil
	}

	data, err := msg.ToJSON()
	if err != nil {
		return nil, fmt.Errorf("cannot turn e-mail into raw data: %w", err)
	}

	_, result, err := program.Run(
		ctx,
		data,
		rudi.NewVariables().SetMany(extraVars),
		getFunctions().Add(extraFuncs),
		coalescing.NewStrict(),
	)
	if err != nil {
		return nil, fmt.Errorf("script failed: %w", err)
	}

	return result, nil
}

func loadProgram(scriptFile string) (rudi.Program, error) {
	content, err := os.ReadFile(scriptFile)
	if err != nil {
		return nil, err
	}

	code := string(content)
	code = strings.TrimSpace(code)

	if len(code) == 0 {
		return nil, nil
	}

	filename := filepath.Base(scriptFile)

	return rudi.Parse(filename, code)
}

func getFunctions() rudi.Functions {
	funcs := rudi.
		NewSafeBuiltInFunctions().
		Add(rudi.NewUnsafeBuiltInFunctions()).
		Add(Functions).
		Add(set.Functions)

	return funcs
}
