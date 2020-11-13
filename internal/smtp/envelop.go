package smtp

import (
	"bytes"
	"io/ioutil"
	"net/mail"
)

type BasicEnvelope struct {
	From   MailAddress
	Rcpts  []MailAddress
	Header mail.Header
	Body   []byte
	Raw    []byte

	buffer []byte
}

func (e *BasicEnvelope) AddRecipient(rcpt MailAddress) error {
	// Check recipient here
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
	e.buffer = append(e.buffer, line...)
	return nil
}

func (e *BasicEnvelope) EndData() error {
	var err error

	e.Raw = e.buffer

	reader := bytes.NewReader(e.buffer)
	email, err := mail.ReadMessage(reader)

	if err != nil {
		return err
	}

	e.Header = email.Header
	e.Body, err = ioutil.ReadAll(email.Body)

	return err
}
