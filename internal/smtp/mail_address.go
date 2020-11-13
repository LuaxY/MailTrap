package smtp

import (
	"strings"
)

type MailAddress string

func (ml MailAddress) Email() string {
	return string(ml)
}

func (ml MailAddress) Hostname() string {
	e := string(ml)

	if idx := strings.Index(e, "@"); idx != -1 {
		return strings.ToLower(e[idx+1:])
	}

	return ""
}
