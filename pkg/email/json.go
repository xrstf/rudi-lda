// SPDX-FileCopyrightText: 2024 Christoph Mewes
// SPDX-License-Identifier: MIT

package email

import (
	"bytes"
	"encoding/json"
	"net/mail"
	"time"
)

type JSONMessage struct {
	From        map[string]any `json:"from"`
	To          map[string]any `json:"to"`
	ReplyTo     map[string]any `json:"replyTo"`
	DeliveredTo map[string]any `json:"deliveredTo"`
	Subject     string         `json:"subject"`
	Date        time.Time      `json:"date"`
	Body        string         `json:"body"`
}

func addressToJSON(addr *mail.Address) map[string]any {
	return map[string]any{
		"name":    addr.Name,
		"address": addr.Address,
	}
}

func (m *Message) ToJSON() (any, error) {
	rm := JSONMessage{}
	if f := m.GetFrom(); f != nil {
		rm.From = addressToJSON(f)
	}
	if f := m.GetTo(); f != nil {
		rm.To = addressToJSON(f)
	}

	rm.Subject = m.GetSubject()
	rm.ReplyTo = addressToJSON(m.GetReplyTo())
	rm.DeliveredTo = addressToJSON(m.GetDeliveredTo())
	rm.Body = m.Body

	date, err := m.GetDate()
	if err != nil {
		return nil, err
	}

	rm.Date = date

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(rm); err != nil {
		return nil, err
	}

	var result any
	if err := json.NewDecoder(&buf).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}
