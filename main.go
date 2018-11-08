package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net/mail"
	"os"

	"SMTP/config"
	"SMTP/smtp"
)

var cfg = config.Get()

type env struct {
	*smtpd.BasicEnvelope
	from   smtpd.MailAddress
	buffer []byte
}

func (e *env) AddRecipient(rcpt smtpd.MailAddress) error {
	/*if !contains(cfg.AllowedAddress, rcpt.Email()) {
		return errors.New("not allowed email")
	}*/

	return e.BasicEnvelope.AddRecipient(rcpt)
}

func (e *env) Write(line []byte) error {
	e.buffer = append(e.buffer, line...)
	return nil
}

func (e *env) OnFinish() {
	r := bytes.NewReader(e.buffer)
	m, err := mail.ReadMessage(r)

	if err != nil {
		log.Println(err)
		return
	}

	f, err := os.OpenFile("email.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)

	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}

	defer f.Close()

	fmt.Fprintln(f, "FR:", e.from)

	for _, rcpt := range e.Rcpts {
		fmt.Fprintln(f, "TO:", rcpt)
	}

	header := m.Header

	fmt.Fprintln(f, "> Date:", header.Get("Date"))
	fmt.Fprintln(f, "> From:", header.Get("From"))
	fmt.Fprintln(f, "> To:", header.Get("To"))
	fmt.Fprintln(f, "> Subject:", header.Get("Subject"))

	scanner := bufio.NewScanner(m.Body)

	for scanner.Scan() {
		fmt.Fprintln(f, ">", scanner.Text())
	}

	/*body, err := ioutil.ReadAll(m.Body)

	if err != nil {
		log.Println(err)
		return
	}*/

	fmt.Fprintln(f, "")
	fmt.Fprintln(f, "---------------------------------------------------------------")
	fmt.Fprintln(f, "")
}

func onNewMail(c smtpd.Connection, from smtpd.MailAddress) (smtpd.Envelope, error) {
	return &env{from: from, BasicEnvelope: new(smtpd.BasicEnvelope)}, nil
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func main() {
	f, err := os.OpenFile("smtp.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)

	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}

	defer f.Close()

	log.SetOutput(f)

	s := &smtpd.Server{
		Motd:      cfg.Motd,
		Addr:      cfg.Addr,
		Hostname:  cfg.Hostname,
		OnNewMail: onNewMail,
	}

	err = s.ListenAndServe()

	if err != nil {
		log.Fatalf("ListenAndServe: %v", err)
	}
}
