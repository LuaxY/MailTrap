package smtp

import (
	"bufio"
	"log"
	"net"
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

	OnNewMail func(session *Session, email *BasicEnvelope)
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
