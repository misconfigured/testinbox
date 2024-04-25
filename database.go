package main

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/sirupsen/logrus"
)

type Database struct {
	Conn *pgxpool.Pool
}

func (db *Database) InsertEmail(ctx context.Context, email Email) error {
	sql := `INSERT INTO emails (
        sender, recipient, subject, body_html, body_text, headers, sender_ip, 
        spam_score, attachments_info, dkim, content_ids, envelope, attachments, 
        spam_report, attachment_info, charsets, spf
    ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)
	RETURNING id`

	err := db.Conn.QueryRow(ctx, sql,
		email.Sender,
		email.Recipient,
		email.Subject,
		email.BodyHTML,
		email.BodyText,
		email.Headers,
		email.SenderIP,
		email.SpamScore,
		email.AttachmentsInfo,
		email.DKIM,
		email.ContentIDs,
		email.Envelope,
		email.Attachments,
		email.SpamReport,
		email.AttachmentInfo,
		email.Charsets,
		email.SPF,
	).Scan(&email.ID)

	if err != nil {
		log.WithFields(logrus.Fields{
			"action": "InsertEmail",
			"error":  err,
		}).Error("Failed to insert email into database")
		return err
	}

	log.WithFields(logrus.Fields{
		"action":  "InsertEmail",
		"emailID": email.ID,
	}).Debug("Email inserted successfully")

	return nil
}

func (db *Database) GetEmails(limit, offset int, recipient string) ([]Email, int, error) {
	var emails []Email
	var total int

	var query string
	var rows pgx.Rows
	var err error

	baseQuery := "SELECT id, sender, recipient, subject, received_at FROM emails"
	countQuery := "SELECT COUNT(*) FROM emails"

	if recipient != "" {
		query = baseQuery + " WHERE recipient = $3 ORDER BY received_at DESC LIMIT $1 OFFSET $2"
		rows, err = db.Conn.Query(context.Background(), query, limit, offset, recipient)
		countQuery += " WHERE recipient = $1"
		err = db.Conn.QueryRow(context.Background(), countQuery, recipient).Scan(&total)
	} else {
		query = baseQuery + " ORDER BY received_at DESC LIMIT $1 OFFSET $2"
		rows, err = db.Conn.Query(context.Background(), query, limit, offset)
		err = db.Conn.QueryRow(context.Background(), countQuery).Scan(&total)
	}

	if err != nil {
		log.WithFields(logrus.Fields{
			"action": "GetEmails",
			"limit":  limit,
			"offset": offset,
			"error":  err,
		}).Error("Query failed to fetch emails")
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var email Email
		if err := rows.Scan(&email.ID, &email.Sender, &email.Recipient, &email.Subject, &email.ReceivedAt); err != nil {
			log.WithFields(logrus.Fields{
				"action": "GetEmails",
				"error":  err,
			}).Error("Failed to scan row")
			return nil, 0, err
		}
		emails = append(emails, email)
	}

	if rows.Err() != nil {
		log.WithFields(logrus.Fields{
			"action": "GetEmails",
			"error":  rows.Err(),
		}).Error("Error iterating through rows")
		return nil, 0, rows.Err()
	}

	log.WithFields(logrus.Fields{
		"action": "GetEmails",
		"count":  len(emails),
	}).Debug("Emails fetched successfully")

	return emails, total, nil
}

func (db *Database) GetEmailByID(ID string) (Email, error) {
	var email Email
	err := db.Conn.QueryRow(context.Background(),
		`SELECT 
            id, sender, recipient, subject, body_html, body_text, headers, 
            received_at, sender_ip, spam_score, attachments_info, dkim, 
            content_ids, envelope, attachments, spam_report, attachment_info, 
            charsets, spf, created_at, updated_at
         FROM emails WHERE id = $1`, ID).Scan(
		&email.ID, &email.Sender, &email.Recipient, &email.Subject,
		&email.BodyHTML, &email.BodyText, &email.Headers, &email.ReceivedAt,
		&email.SenderIP, &email.SpamScore, &email.AttachmentsInfo, &email.DKIM,
		&email.ContentIDs, &email.Envelope, &email.Attachments, &email.SpamReport,
		&email.AttachmentInfo, &email.Charsets, &email.SPF, &email.CreatedAt, &email.UpdatedAt,
	)
	if err != nil {
		log.WithFields(logrus.Fields{
			"action": "getEmailByID",
			"id":     ID,
			"error":  err,
		}).Error("Failed to fetch email")
		return Email{}, err
	}
	return email, nil
}

func (db *Database) GetEmailsCount() (int, error) {
	var count int
	err := db.Conn.QueryRow(context.Background(), "SELECT COUNT(1) FROM emails").Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (db *Database) GetRecipientSuggestions(term string) ([]string, error) {
	var suggestions []string
	query := `SELECT DISTINCT recipient FROM emails WHERE recipient ILIKE $1 LIMIT 10`
	rows, err := db.Conn.Query(context.Background(), query, term+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var recipient string
		if err := rows.Scan(&recipient); err != nil {
			return nil, err
		}
		suggestions = append(suggestions, recipient)
	}

	return suggestions, rows.Err()
}
