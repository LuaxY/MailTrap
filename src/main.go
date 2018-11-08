package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/mail"
	"os"

	"SMTP/src/config"
	"SMTP/src/model"
	"SMTP/src/smtp"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

var cfg = config.Get()

type Envelop struct {
	*smtp.BasicEnvelope
}

func (e *Envelop) AddRecipient(rcpt smtp.MailAddress) error {
	/*if !contains(cfg.AllowedAddress, rcpt.Email()) {
		return errors.New("not allowed email")
	}*/

	return e.BasicEnvelope.AddRecipient(rcpt)
}

func (e *Envelop) Write(line []byte) error {
	e.Buffer = append(e.Buffer, line...)
	return nil
}

func (e *Envelop) OnFinish() {
	r := bytes.NewReader(e.Buffer)
	m, err := mail.ReadMessage(r)

	if err != nil {
		log.Println(err)
		return
	}

	header := m.Header

	f, err := os.OpenFile("./logs/email.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)

	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}

	defer f.Close()

	fmt.Fprintln(f, "FR:", e.From)

	for _, rcpt := range e.Rcpts {
		fmt.Fprintln(f, "TO:", rcpt)
	}

	fmt.Fprintln(f, "> Date:", header.Get("Date"))
	fmt.Fprintln(f, "> From:", header.Get("From"))
	fmt.Fprintln(f, "> To:", header.Get("To"))
	fmt.Fprintln(f, "> Subject:", header.Get("Subject"))

	scanner := bufio.NewScanner(m.Body)

	for scanner.Scan() {
		fmt.Fprintln(f, ">", scanner.Text())
	}

	fmt.Fprintln(f, "")
	fmt.Fprintln(f, "---------------------------------------------------------------")
	fmt.Fprintln(f, "")

	body, err := ioutil.ReadAll(m.Body)

	if err != nil {
		log.Println(err)
		return
	}

	email := model.Email{
		Remote:    "",
		IP:        0,
		From:      e.From.Email(),
		To:        e.Rcpts[0].Email(),
		Date:      header.Get("Date"),
		EmailFrom: header.Get("From"),
		EmailTo:   header.Get("To"),
		Subject:   header.Get("Subject"),
		Body:      string(body),
	}

	db.Create(&email)
}

func onNewMail(c smtp.Connection, from smtp.MailAddress) (smtp.Envelope, error) {
	return &Envelop{BasicEnvelope: &smtp.BasicEnvelope{From: from}}, nil
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

var db *gorm.DB

func main() {
	var err error

	db, err = gorm.Open("sqlite3", cfg.Database)

	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

	db.AutoMigrate(&model.Email{})

	f, err := os.OpenFile("./logs/smtp.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)

	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}

	defer f.Close()

	log.SetOutput(f)

	s := &smtp.Server{
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
