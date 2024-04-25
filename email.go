package main

import (
	"html/template"
	"time"
)

type Email struct {
	ID              int           `json:"id"`
	Sender          string        `json:"sender"`
	Recipient       string        `json:"recipient"`
	Subject         string        `json:"subject"`
	BodyHTML        template.HTML `json:"body_html"`
	BodyText        string        `json:"body_text"`
	Headers         string        `json:"headers"`
	ReceivedAt      time.Time     `json:"received_at"`
	SenderIP        string        `json:"sender_ip"`
	SpamScore       float64       `json:"spam_score"`
	AttachmentsInfo string        `json:"attachments_info"`
	DKIM            string        `json:"dkim"`
	ContentIDs      string        `json:"content_ids"`
	Envelope        string        `json:"envelope"`
	Attachments     int           `json:"attachments"`
	SpamReport      string        `json:"spam_report"`
	AttachmentInfo  string        `json:"attachment_info"`
	Charsets        string        `json:"charsets"`
	SPF             string        `json:"spf"`
	CreatedAt       time.Time     `json:"created_at"`
	UpdatedAt       time.Time     `json:"updated_at"`
}
