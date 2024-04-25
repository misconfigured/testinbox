CREATE TABLE emails (
    id SERIAL PRIMARY KEY,
    sender VARCHAR(255),
    recipient VARCHAR(255),
    subject TEXT,
    body_html TEXT,
    body_text TEXT,
    headers TEXT,
    received_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    sender_ip TEXT,
    spam_score NUMERIC,
    attachments_info TEXT
    dkim TEXT,
    content_ids TEXT,
    envelope JSON,
    attachments INT,
    spam_report TEXT,
    attachment_info TEXT,
    charsets JSON,
    spf TEXT;
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_emails_received_at ON emails (received_at DESC);
CREATE INDEX idx_emails_recipient_received_at ON emails (recipient, received_at DESC);