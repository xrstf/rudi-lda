// SPDX-FileCopyrightText: 2024 Christoph Mewes
// SPDX-License-Identifier: MIT

package util

import "mime"

func DecodeQuotedPrintable(s string) string {
	dec := new(mime.WordDecoder)
	b, _ := dec.DecodeHeader(s)

	return string(b)
}

func EncodeQuotedPrintable(s string) string {
	return mime.QEncoding.Encode("utf-8", s)
}
