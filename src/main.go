package main

import (
	"encoding/binary"
	"log"
	"net"
	"os"

	"MailTrap/src/config"
	"MailTrap/src/database"
	"MailTrap/src/http"
	"MailTrap/src/model"
	"MailTrap/src/smtp"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

var cfg = config.Get()

func main() {
	f, err := os.OpenFile("./logs/smtp.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)

	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}

	defer f.Close()

	//log.SetOutput(f)

	go http.Start()

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

	email := model.Email{
		Remote:    session.HelloHost,
		IP:        ip2int(IPs[0]),
		From:      mail.From.Email(),
		To:        mail.Rcpts[0].Email(),
		Date:      mail.Header.Get("Date"),
		EmailFrom: mail.Header.Get("From"),
		EmailTo:   mail.Header.Get("To"),
		Subject:   mail.Header.Get("Subject"),
		Body:      string(mail.Body),
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
