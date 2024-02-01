// SPDX-FileCopyrightText: 2024 Christoph Mewes
// SPDX-License-Identifier: MIT

package rudilib

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"go.xrstf.de/rudi"
	"k8s.io/apimachinery/pkg/util/sets"
)

var Functions = rudi.Functions{
	"domain":   rudi.NewFunctionBuilder(domainFunc).WithDescription("returns the domain portion of the address' email").Build(),
	"user":     rudi.NewFunctionBuilder(userFunc).WithDescription("returns the user portion of the address' email").Build(),
	"matches?": rudi.NewFunctionBuilder(matchesFunc).WithDescription("returns true if the first value matches any of the given expressions").Build(),
}

func domainFunc(val any) (any, error) {
	if thing, ok := val.(map[string]any); ok {
		val = thing["address"]
	}

	if s, ok := val.(string); ok {
		parts := strings.Split(s, "@")
		return parts[len(parts)-1], nil
	}

	return nil, errors.New("cannot deal with provided value")
}

func userFunc(val any) (any, error) {
	if s, ok := val.(string); ok {
		parts := strings.Split(s, "@")
		return parts[0], nil
	}

	return nil, errors.New("cannot deal with provided value")
}

func matchesFunc(ctx rudi.Context, input string, patterns ...any) (any, error) {
	for _, pattern := range patterns {
		if stringSet, ok := pattern.(sets.Set[string]); ok {
			matches, err := matchesFunc(ctx, input, stringSetToAnys(stringSet)...)
			if err != nil {
				return nil, err
			}

			if matched, _ := matches.(bool); matched {
				return true, nil
			}
		} else {
			s, err := ctx.Coalesce().ToString(pattern)
			if err != nil {
				return nil, err
			}

			re, err := regexp.Compile(s)
			if err != nil {
				return nil, fmt.Errorf("invalid expression %q: %w", s, err)
			}

			if re.MatchString(input) {
				return true, nil
			}
		}
	}

	return false, nil
}

func stringSetToAnys(s sets.Set[string]) []any {
	result := make([]any, 0, s.Len())
	for _, value := range sets.List(s) {
		result = append(result, value)
	}
	return result
}
