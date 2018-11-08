package smtp

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net"
	"time"
)

type Session struct {
	From      string
	To        []string
	HelloHost string
	Envelope  Envelope

	server *Server
	peer   net.Conn
	reader *bufio.Reader
	writer *bufio.Writer
}

func (s *Session) Addr() net.Addr {
	return s.peer.RemoteAddr()
}

func (s *Session) Close() error {
	return s.peer.Close()
}

func (s *Session) process() {
	defer s.peer.Close()

	if onNewConnecion := s.server.OnNewConnection; onNewConnecion != nil {
		if err := onNewConnecion(s); err != nil {
			s.sendSMTPErrorOrLinef(err, "554 connection rejected")
			return
		}
	}

	s.sendf("220 %s ESMTP %s\r\n", s.server.hostname(), s.server.Motd)

	for {
		if s.server.ReadTimeout != 0 {
			s.peer.SetReadDeadline(time.Now().Add(s.server.ReadTimeout))
		}

		slice, err := s.reader.ReadSlice('\n')

		if err != nil {
			s.errorf("read error: %v", err)
			return
		}

		line := cmdLine(string(slice))

		if err := line.checkValid(); err != nil {
			s.sendlinef("500 %v", err)
			continue
		}

		switch line.Verb() {
		case "HELO", "EHLO":
			s.onHelo(line.Verb(), line.Arg())
		case "QUIT":
			s.sendlinef("221 2.0.0 Bye")
			return
		case "RSET":
			s.Envelope = nil
			s.sendlinef("250 2.0.0 OK")
		case "NOOP":
			s.sendlinef("250 2.0.0 OK")
		case "VRFY":
			s.sendlinef("252 2.1.5 Cannot VRFY user")
		case "MAIL":
			arg := line.Arg() // "From:<foo@bar.com>"
			match := mailFromRE.FindStringSubmatch(arg)

			if match == nil {
				log.Printf("invalid MAIL arg: %q", arg)
				s.sendlinef("501 5.1.7 Bad sender address syntax")
				continue
			}

			s.onMail(match[1])
		case "RCPT":
			s.onRcpt(line)
		case "DATA":
			s.onData()
		default:
			log.Printf("Client: %q, verhb: %q", line, line.Verb())
			s.sendlinef("502 5.5.2 Error: command not recognized")
		}
	}
}

func (s *Session) onHelo(greeting, host string) {
	s.HelloHost = host

	fmt.Fprintf(s.writer, "250-%s\r\n", s.server.hostname())

	var extensions []string

	if s.server.PlainAuth {
		extensions = append(extensions, "250-AUTH PLAIN")
	}

	extensions = append(extensions, "250-PIPELINING",
		"250-SIZE 10240000",
		"250-ENHANCEDSTATUSCODES",
		"250-8BITMIME",
		"250 DSN")

	for _, extension := range extensions {
		fmt.Fprintf(s.writer, "%s\r\n", extension)
	}

	s.writer.Flush()
}

func (s *Session) onMail(email string) {
	// TODO: 4.1.1.11.  If the server SMTP does not recognize or
	// cannot implement one or more of the parameters associated
	// qwith a particular MAIL FROM or RCPT TO command, it will return
	// code 555.

	if s.Envelope != nil {
		s.sendlinef("503 5.5.1 Error: nested MAIL command")
		return
	}

	onNewMail := s.server.OnNewMail

	if onNewMail == nil {
		log.Printf("smtp: Server.OnNewMail is nil; rejecting MAIL FROM")
		s.sendf("451 Server.OnNewMail not configured\r\n")
		return
	}

	s.Envelope = nil
	envelope, err := onNewMail(s, addrString(email))

	if err != nil {
		log.Printf("rejecting MAIL FROM %q: %v", email, err)
		s.sendf("451 denied\r\n")
		s.writer.Flush()
		time.Sleep(100 * time.Millisecond)
		s.peer.Close()
		return
	}

	s.Envelope = envelope
	s.sendlinef("250 2.1.0 Ok")
}

func (s *Session) onRcpt(line cmdLine) {
	// TODO: 4.1.1.11.  If the server SMTP does not recognize or
	// cannot implement one or more of the parameters associated
	// qwith a particular MAIL FROM or RCPT TO command, it will return
	// code 555.

	if s.Envelope == nil {
		s.sendlinef("503 5.5.1 Error: need MAIL command")
		return
	}

	arg := line.Arg() // "To:<foo@bar.com>"
	m := rcptToRE.FindStringSubmatch(arg)

	if m == nil {
		log.Printf("bad RCPT address: %q", arg)
		s.sendlinef("501 5.1.7 Bad sender address syntax")
		return
	}

	err := s.Envelope.AddRecipient(addrString(m[1]))

	if err != nil {
		s.sendSMTPErrorOrLinef(err, "550 bad recipient")
		return
	}

	s.sendlinef("250 2.1.0 Ok")
}

func (s *Session) onData() {
	if s.Envelope == nil {
		s.sendlinef("503 5.5.1 Error: need RCPT command")
		return
	}

	if err := s.Envelope.BeginData(); err != nil {
		s.handleError(err)
		return
	}

	s.sendlinef("354 Go ahead")

	for {
		sl, err := s.reader.ReadSlice('\n')

		if err != nil {
			s.errorf("read error: %v", err)
			return
		}

		if bytes.Equal(sl, []byte(".\r\n")) {
			break
		}

		if sl[0] == '.' {
			sl = sl[1:]
		}

		err = s.Envelope.Write(sl)

		if err != nil {
			s.sendSMTPErrorOrLinef(err, "550 ??? failed")
			return
		}
	}

	if err := s.Envelope.Close(); err != nil {
		s.handleError(err)
		return
	}

	go s.Envelope.OnFinish()

	s.sendlinef("250 2.0.0 Ok: queued")
	s.Envelope = nil
}

func (s *Session) handleError(err error) {
	if se, ok := err.(SMTPError); ok {
		s.sendlinef("%s", se)
		return
	}

	log.Printf("Error: %s", err)
	s.Envelope = nil
}

func (s *Session) errorf(format string, args ...interface{}) {
	log.Printf("Client error: "+format, args...)
}

func (s *Session) sendf(format string, args ...interface{}) {
	if s.server.WriteTimeout != 0 {
		s.peer.SetWriteDeadline(time.Now().Add(s.server.WriteTimeout))
	}

	fmt.Fprintf(s.writer, format, args...)
	s.writer.Flush()
}

func (s *Session) sendlinef(format string, args ...interface{}) {
	s.sendf(format+"\r\n", args...)
}

func (s *Session) sendSMTPErrorOrLinef(err error, format string, args ...interface{}) {
	if se, ok := err.(SMTPError); ok {
		s.sendlinef("%s", se.Error())
		return
	}

	s.sendlinef(format, args...)
}
