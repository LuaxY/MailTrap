package smtp

// TODO:
//  -- send 421 to connected clients on graceful server shutdown (s3.8)
//

import (
	"errors"
	"regexp"
	"strings"
	"unicode"
)

var (
	rcptToRE   = regexp.MustCompile(`[Tt][Oo]:<(.+)>`)
	mailFromRE = regexp.MustCompile(`[Ff][Rr][Oo][Mm]:<(.*)>`)

	fromToRE = regexp.MustCompile(`(?m)["“”']{0,1}(.*)["“”']{0,1}[ ]{0,1}[<](.*)@(.*)[>]`)
)

type cmdLine string

func (cl cmdLine) checkValid() error {
	if !strings.HasSuffix(string(cl), "\r\n") {
		return errors.New(`line doesn't end in \r\n`)
	}

	// Check for verbs defined not to have an argument
	// (RFC 5321 s4.1.1)
	switch cl.Verb() {
	case "RSET", "DATA", "QUIT":
		if cl.Arg() != "" {
			return errors.New("unexpected argument")
		}
	}

	return nil
}

func (cl cmdLine) Verb() string {
	s := string(cl)

	if idx := strings.Index(s, " "); idx != -1 {
		return strings.ToUpper(s[:idx])
	}

	return strings.ToUpper(s[:len(s)-2])
}

func (cl cmdLine) Arg() string {
	s := string(cl)

	if idx := strings.Index(s, " "); idx != -1 {
		return strings.TrimRightFunc(s[idx+1:len(s)-2], unicode.IsSpace)
	}

	return ""
}

func (cl cmdLine) String() string {
	return string(cl)
}

type SMTPError string

func (e SMTPError) Error() string {
	return string(e)
}
