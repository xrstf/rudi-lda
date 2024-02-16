// SPDX-FileCopyrightText: 2024 Christoph Mewes
// SPDX-License-Identifier: MIT

package util

import (
	"fmt"
	"os"
	"time"
)

func Filename() string {
	now := time.Now().UTC().Format("20060102_150405")

	return fmt.Sprintf("%s_%d.eml", now, os.Getpid())
}
