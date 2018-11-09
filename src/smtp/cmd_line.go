package smtp

import (
	"errors"
	"strings"
	"unicode"
)

type cmdLine string

func (cl cmdLine) checkValid() error {
	if !strings.HasSuffix(cl.string(), "\r\n") {
		return errors.New("line doesn't end in \r\n")
	}

	// Check for verbs defined not to have an argument
	// (RFC 5321 s4.1.1)
	switch cl.verb() {
	case "RSET", "DATA", "QUIT":
		if cl.arg() != "" {
			return errors.New("unexpected argument")
		}
	}

	return nil
}

func (cl cmdLine) verb() string {
	s := cl.string()

	if idx := strings.Index(s, " "); idx != -1 {
		return strings.ToUpper(s[:idx])
	}

	return strings.ToUpper(s[:len(s)-2])
}

func (cl cmdLine) arg() string {
	s := cl.string()

	if idx := strings.Index(s, " "); idx != -1 {
		return strings.TrimRightFunc(s[idx+1:len(s)-2], unicode.IsSpace)
	}

	return ""
}

func (cl cmdLine) string() string {
	return string(cl)
}
