// SPDX-FileCopyrightText: 2024 Christoph Mewes
// SPDX-License-Identifier: MIT

package spam

import (
	"go.xrstf.de/rudi"
)

var Functions = rudi.Functions{
	"spam":       rudi.NewFunctionBuilder(spamFuncWithRule, spamFunc).WithDescription("marks an e-mail as spam").Build(),
	"ham":        rudi.NewFunctionBuilder(hamFuncWithRule, hamFunc).WithDescription("marks an e-mail as ham").Build(),
	"maybe-spam": rudi.NewFunctionBuilder(maybeSpamFuncWithRule, maybeSpamFunc).WithDescription("marks an e-mail as maybe spam").Build(),
}

func spamFunc() (any, error) {
	return spamFuncWithRule("")
}

func spamFuncWithRule(rule string) (any, error) {
	return resultFunc(Spam, rule)
}

func hamFunc() (any, error) {
	return hamFuncWithRule("")
}

func hamFuncWithRule(rule string) (any, error) {
	return resultFunc(Ham, rule)
}

func maybeSpamFunc() (any, error) {
	return maybeSpamFuncWithRule("")
}

func maybeSpamFuncWithRule(rule string) (any, error) {
	return resultFunc(MaybeSpam, rule)
}

func resultFunc(status Status, rule string) (any, error) {
	// Rudi 0.7 does not support custom structs, only map[string]any
	return map[string]any{
		"status": status,
		"rule":   rule,
	}, nil
}
