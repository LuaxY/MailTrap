package main

import (
	"MailTrap/internal/config"
	"MailTrap/internal/database"
	"MailTrap/internal/http"
	"MailTrap/internal/model"
	"MailTrap/internal/smtp"
	"bytes"
	"encoding/binary"
	"github.com/jhillyerd/enmime"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"log"
	"net"
)

var cfg = config.Get()

func main() {
	go http.Start()

	s := &smtp.Server{
		Motd:      cfg.MOTD,
		Addr:      cfg.SMTP,
		Hostname:  cfg.Hostname,
		OnNewMail: onNewMail,
	}

	err := s.ListenAndServe()

	if err != nil {
		log.Fatalf("ListenAndServe: %v", err)
	}
}

func onNewMail(session *smtp.Session, mail *smtp.BasicEnvelope) {
	host, _, err := net.SplitHostPort(session.Addr().String())

	if err != nil {
		log.Println(err)
		return
	}

	IPs, err := net.LookupIP(host)

	if err != nil {
		log.Println(err)
		return
	}

	reader := bytes.NewReader(mail.Raw)
	envelope, err := enmime.ReadEnvelope(reader)

	// TODO error

	email := model.Email{
		Remote:    session.HelloHost,
		IP:        ip2int(IPs[0]),
		From:      mail.From.Email(),
		To:        mail.Rcpts[0].Email(),
		Date:      envelope.GetHeader("Date"),
		EmailFrom: envelope.GetHeader("From"),
		EmailTo:   envelope.GetHeader("To"),
		Subject:   envelope.GetHeader("Subject"),
		Raw:       string(mail.Raw),
	}

	database.DB.Create(&email)
}

func ip2int(ip net.IP) uint32 {
	if len(ip) == 16 {
		return binary.BigEndian.Uint32(ip[12:16])
	}

	return binary.BigEndian.Uint32(ip)
}

func int2ip(nn uint32) net.IP {
	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, nn)

	return ip
}
