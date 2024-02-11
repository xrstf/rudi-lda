// SPDX-FileCopyrightText: 2024 Christoph Mewes
// SPDX-License-Identifier: MIT

package email

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net/mail"
	"net/textproto"
	"regexp"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"go.xrstf.de/rudi-lda/pkg/util"
)

type Message struct {
	Header mail.Header
	Body   string
	raw    []byte
}

func (m *Message) LogFields() logrus.Fields {
	fields := logrus.Fields{
		"subject": m.GetSubject(),
	}

	if from := m.GetFrom(); from != nil {
		fields["fromName"] = util.DecodeQuotedPrintable(from.Name)
		fields["fromAddress"] = from.Address
	}

	if to := m.GetTo(); to != nil {
		fields["toName"] = util.DecodeQuotedPrintable(to.Name)
		fields["toAddress"] = to.Address
	}

	return fields
}

func (m *Message) Raw() []byte {
	return m.raw
}

func (m *Message) GetDate() (time.Time, error) {
	date := m.Header.Get("Date")
	formats := []string{
		time.RFC822, time.RFC822Z,
		time.RFC1123, time.RFC1123Z,
		"Mon, 02 Jan 2006 15:04:05 MST (-0700)", "Mon, 02 Jan 2006 15:04:05 -0700 (MST)",
		"02 Jan 06 15:04:05 MST", "02 Jan 06 15:04:05 -0700",
		"02 Jan 2006 15:04:05 MST", "02 Jan 2006 15:04:05 -0700",
		"_2 Jan 06 15:04:05 MST", "_2 Jan 06 15:04:05 -0700",
		"_2 Jan 2006 15:04:05 MST", "_2 Jan 2006 15:04:05 -0700",
	}

	for _, format := range formats {
		parsed, err := time.Parse(format, date)
		if err == nil {
			return parsed, nil
		}
	}

	return time.Time{}, fmt.Errorf("cannot parse %q as a time", date)
}

func (m *Message) GetSubject() string {
	return util.DecodeQuotedPrintable(m.Header.Get("Subject"))
}

func (m *Message) GetFrom() *mail.Address {
	return m.getAddress("From")
}

func (m *Message) GetTo() *mail.Address {
	return m.getAddress("To")
}

func (m *Message) GetReplyTo() *mail.Address {
	return m.getAddress("Reply-To")
}

func (m *Message) GetDeliveredTo() *mail.Address {
	// this header is (on my server) set to the username first
	// and Go's AddressList() function does not like that, so
	// we must hand-parse the value
	header := textproto.MIMEHeader(m.Header)

	values, exist := header["Delivered-To"]
	if !exist {
		return nil
	}

	for _, value := range values {
		parsed, err := mail.ParseAddress(value)
		if err != nil {
			continue
		}

		return parsed
	}

	return nil
}

func (m *Message) GetTos() []*mail.Address {
	list, err := m.Header.AddressList("To")
	if err != nil {
		return nil
	}

	return list
}

func (m *Message) getAddress(header string) *mail.Address {
	list, err := m.Header.AddressList(header)
	if err != nil || len(list) == 0 {
		return nil
	}

	return list[0]
}

func ParseMessage(rawMessage []byte) (*Message, error) {
	msg, err := mail.ReadMessage(bytes.NewReader(rawMessage))
	if err != nil {
		return nil, fmt.Errorf("failed to parse mail body: %v", err)
	}

	body, _ := io.ReadAll(msg.Body)

	return &Message{
		Header: msg.Header,
		Body:   string(body),
		raw:    rawMessage,
	}, nil
}

func (m *Message) BodyContainsAnyOf(needles ...string) bool {
	return stringContainsAnyOf(m.Body, needles...)
}

func (m *Message) SubjectContainsAnyOf(needles ...string) bool {
	return m.HeaderContainsAnyOf("Subject", needles...)
}

func (m *Message) HeaderContainsAnyOf(header string, needles ...string) bool {
	return stringContainsAnyOf(util.DecodeQuotedPrintable(m.Header.Get(header)), needles...)
}

func stringContainsAnyOf(haystack string, needles ...string) bool {
	for _, needle := range needles {
		if strings.Contains(haystack, needle) {
			return true
		}
	}

	return false
}

func (m *Message) BodyMatchesAnyOf(needles ...string) bool {
	return stringMatchesAnyOf(m.Body, needles...)
}

func (m *Message) SubjectMatchesAnyOf(needles ...string) bool {
	return m.HeaderMatchesAnyOf("Subject", needles...)
}

func (m *Message) HeaderMatchesAnyOf(header string, needles ...string) bool {
	return stringMatchesAnyOf(util.DecodeQuotedPrintable(m.Header.Get(header)), needles...)
}

func stringMatchesAnyOf(haystack string, needles ...string) bool {
	for _, needle := range needles {
		if regexp.MustCompile(needle).MatchString(haystack) {
			return true
		}
	}

	return false
}

func (m *Message) IsToName(names ...string) bool {
	return m.IsHeaderName("To", names...)
}

func (m *Message) IsFromName(names ...string) bool {
	return m.IsHeaderName("From", names...)
}

func (m *Message) IsHeaderName(header string, names ...string) bool {
	list, err := m.Header.AddressList(header)
	if err != nil || len(list) == 0 {
		return false
	}

	for _, name := range names {
		if strings.EqualFold(util.DecodeQuotedPrintable(list[0].Name), name) {
			return true
		}
	}

	return false
}

func (m *Message) IsToAddress(addresses ...string) bool {
	return m.IsHeaderAddress("To", addresses...)
}

func (m *Message) IsFromAddress(addresses ...string) bool {
	return m.IsHeaderAddress("From", addresses...)
}

func (m *Message) IsHeaderAddress(header string, addresses ...string) bool {
	list, err := m.Header.AddressList(header)
	if err != nil || len(list) == 0 {
		return false
	}

	for _, address := range addresses {
		if strings.EqualFold(list[0].Address, address) {
			return true
		}
	}

	return false
}

func (m *Message) IsFromDomain(domains ...string) bool {
	return m.IsHeaderDomain("From", domains...)
}

func (m *Message) IsReplyToDomain(domains ...string) bool {
	return m.IsHeaderDomain("Reply-To", domains...)
}

func (m *Message) IsHeaderDomain(header string, domains ...string) bool {
	list, err := m.Header.AddressList(header)
	if err != nil || len(list) == 0 {
		return false
	}

	address := strings.ToLower(list[0].Address)

	for _, domain := range domains {
		if strings.HasSuffix(address, "@"+strings.ToLower(domain)) {
			return true
		}
	}

	return false
}

// based on https://github.com/kirabou/parseMIMEemail.go/blob/master/parseMIMEmail.go
func (m *Message) GetMultipartBody(contentType string) (string, error) {
	mediaType, params, err := mime.ParseMediaType(m.Header.Get("Content-Type"))
	if err != nil {
		return "", fmt.Errorf("failed to parse media type: %w", err)
	}

	if !strings.HasPrefix(mediaType, "multipart/") {
		return "", nil
	}

	// Instantiate a new io.Reader dedicated to MIME multipart parsing
	// using multipart.NewReader()
	reader := multipart.NewReader(strings.NewReader(m.Body), params["boundary"])
	if reader == nil {
		return "", nil
	}

	// Go through each of the MIME part of the message Body with NextPart(),
	// and read the content of the MIME part with ioutil.ReadAll()
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("failed going through the MIME parts: %w", err)
		}

		mediaType, _, err := mime.ParseMediaType(part.Header.Get("Content-Type"))
		if err != nil {
			return "", fmt.Errorf("failed to parse MIME headers: %w", err)
		}

		// not the content type we are looking for
		if mediaType != contentType {
			continue
		}

		rawPartBody, err := io.ReadAll(part)
		if err != nil {
			return "", fmt.Errorf("failed to read MIME body: %w", err)
		}

		encoding := strings.ToUpper(part.Header.Get("Content-Transfer-Encoding"))
		body := string(rawPartBody)

		switch {
		case encoding == "BASE64":
			decoded, err := base64.StdEncoding.DecodeString(body)
			if err != nil {
				return "", fmt.Errorf("failed to base64 decode MIME body: %w", err)
			}
			body = string(decoded)

		case encoding == "QUOTED-PRINTABLE":
			decoded, err := io.ReadAll(quotedprintable.NewReader(strings.NewReader(body)))
			if err != nil {
				return "", fmt.Errorf("failed to decode quoted-printable MIME body: %w", err)
			}
			body = string(decoded)
		}

		return body, nil
	}

	return "", nil
}
