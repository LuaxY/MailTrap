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
	Envelope  *BasicEnvelope

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

	/*if onNewConnecion := s.server.OnNewConnection; onNewConnecion != nil {
		if err := onNewConnecion(s); err != nil {
			s.sendError(err, "554 connection rejected")
			return
		}
	}*/

	s.send("220 %s ESMTP %s", s.server.Hostname, s.server.Motd)

	for {
		if s.server.ReadTimeout != 0 {
			s.peer.SetReadDeadline(time.Now().Add(s.server.ReadTimeout))
		}

		slice, err := s.reader.ReadSlice('\n')

		if err != nil {
			s.error(err)
			return
		}

		line := cmdLine(string(slice))

		if err := line.checkValid(); err != nil {
			s.send("500 %v", err)
			continue
		}

		switch line.verb() {
		case "HELO", "EHLO":
			s.onHelo(line.verb(), line.arg())
		case "QUIT":
			s.send("221 2.0.0 Bye")
			return
		case "RSET":
			s.Envelope = nil
			s.send("250 2.0.0 OK")
		case "NOOP":
			s.send("250 2.0.0 OK")
		case "VRFY":
			s.send("252 2.1.5 Cannot VRFY user")
		case "MAIL":
			arg := line.arg()
			matches := mailFromRE.FindStringSubmatch(arg)

			if matches == nil {
				log.Printf("invalid MAIL arg: %q", arg)
				s.send("501 5.1.7 Bad sender address syntax")
				continue
			}

			s.onMail(matches[1])
		case "RCPT":
			s.onRcpt(line)
		case "DATA":
			s.onData()
		default:
			log.Printf("Client: %q, verhb: %q", line, line.verb())
			s.send("502 5.5.2 Error: command not recognized")
		}
	}
}

func (s *Session) onHelo(greeting, host string) {
	s.HelloHost = host

	fmt.Fprintf(s.writer, "250-%s\r\n", s.server.Hostname)

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
		s.send("503 5.5.1 Error: nested MAIL command")
		return
	}

	// Check email from here

	/*if err != nil {
		log.Printf("rejecting MAIL FROM %q: %v", email, err)
		s.send("451 denied")
		s.writer.Flush()
		time.Sleep(100 * time.Millisecond)
		s.peer.Close()
		return
	}*/

	s.Envelope = &BasicEnvelope{From: MailAddress(email)}

	s.send("250 2.1.0 Ok")
}

func (s *Session) onRcpt(line cmdLine) {
	// TODO: 4.1.1.11.  If the server SMTP does not recognize or
	// cannot implement one or more of the parameters associated
	// qwith a particular MAIL FROM or RCPT TO command, it will return
	// code 555.

	if s.Envelope == nil {
		s.send("503 5.5.1 Error: need MAIL command")
		return
	}

	arg := line.arg()
	matches := rcptToRE.FindStringSubmatch(arg)

	if matches == nil {
		log.Printf("bad RCPT address: %q", arg)
		s.send("501 5.1.7 Bad sender address syntax")
		return
	}

	err := s.Envelope.AddRecipient(MailAddress(matches[1]))

	if err != nil {
		s.error(SMTPError("550 bad recipient"))
		return
	}

	s.send("250 2.1.0 Ok")
}

func (s *Session) onData() {
	if s.Envelope == nil {
		s.send("503 5.5.1 Error: need RCPT command")
		return
	}

	if err := s.Envelope.BeginData(); err != nil {
		s.error(err)
		return
	}

	s.send("354 Go ahead")

	for {
		data, err := s.reader.ReadSlice('\n')

		if err != nil {
			s.error(err)
			return
		}

		if bytes.Equal(data, []byte(".\r\n")) {
			break
		}

		if data[0] == '.' {
			data = data[1:]
		}

		err = s.Envelope.Write(data)

		if err != nil {
			s.error(SMTPError("550 ??? failed"))
			return
		}
	}

	if err := s.Envelope.EndData(); err != nil {
		s.error(err)
		return
	}

	go s.server.OnNewMail(s, s.Envelope)

	s.send("250 2.0.0 Ok: queued")
}

func (s *Session) send(format string, args ...interface{}) {
	if s.server.WriteTimeout != 0 {
		s.peer.SetWriteDeadline(time.Now().Add(s.server.WriteTimeout))
	}

	fmt.Fprintf(s.writer, format+"\r\n", args...)
	s.writer.Flush()
}

func (s *Session) error(err error) {
	if se, ok := err.(SMTPError); ok {
		s.send("%s", se)
		return
	}

	log.Printf("Error: %s", err)
}
