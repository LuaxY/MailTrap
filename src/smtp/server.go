package smtp

import (
	"bufio"
	"log"
	"net"
	"os/exec"
	"strings"
	"time"
)

// Server is an SMTP server.
type Server struct {
	Motd         string
	Addr         string
	Hostname     string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration

	PlainAuth bool // advertise plain auth (assumes you're on SSL)

	// OnNewConnection, if non-nil, is called on new connections.
	// If it returns non-nil, the connection is closed.
	OnNewConnection func(c *Session) error

	// OnNewMail must be defined and is called when a new message beings.
	// (when a MAIL FROM line arrives)
	OnNewMail func(c *Session, from MailAddress) (Envelope, error)
}

func (srv *Server) ListenAndServe() error {
	addr := srv.Addr

	if addr == "" {
		addr = ":25"
	}

	ln, e := net.Listen("tcp", addr)

	if e != nil {
		return e
	}

	return srv.Serve(ln)
}

func (srv *Server) Serve(ln net.Listener) error {
	defer ln.Close()

	for {
		conn, err := ln.Accept()

		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				log.Printf("smtpd: Accept error: %v", err)
				continue
			}

			return err
		}

		session, err := srv.newSession(conn)

		if err != nil {
			continue
		}

		go session.process()
	}

	panic("not reached")
}

func (srv *Server) newSession(conn net.Conn) (s *Session, err error) {
	s = &Session{
		server: srv,
		peer:   conn,
		reader: bufio.NewReader(conn),
		writer: bufio.NewWriter(conn),
	}

	return
}

func (srv *Server) hostname() string {
	if srv.Hostname != "" {
		return srv.Hostname
	}

	out, err := exec.Command("hostname").Output()

	if err != nil {
		return ""
	}

	return strings.TrimSpace(string(out))
}
