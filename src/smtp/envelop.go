package smtp

import (
	"log"
)

type Envelope interface {
	AddRecipient(rcpt MailAddress) error
	BeginData() error
	Write(line []byte) error
	Close() error
	OnFinish()
}

type BasicEnvelope struct {
	From   MailAddress
	Rcpts  []MailAddress
	Buffer []byte
}

func (e *BasicEnvelope) AddRecipient(rcpt MailAddress) error {
	e.Rcpts = append(e.Rcpts, rcpt)
	return nil
}

func (e *BasicEnvelope) BeginData() error {
	if len(e.Rcpts) == 0 {
		return SMTPError("554 5.5.1 Error: no valid recipients")
	}
	return nil
}

func (e *BasicEnvelope) Write(line []byte) error {
	log.Printf("Line: %q", string(line))
	return nil
}

func (e *BasicEnvelope) Close() error {
	return nil
}

// MailAddress is defined by
type MailAddress interface {
	Email() string    // email address, as provided
	Hostname() string // canonical hostname, lowercase
}
