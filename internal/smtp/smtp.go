package smtp

import (
	"regexp"
)

// TODO: send 421 to connected clients on graceful server shutdown (s3.8)

var (
	rcptToRE   = regexp.MustCompile(`[Tt][Oo]:<(.+)>`)
	mailFromRE = regexp.MustCompile(`[Ff][Rr][Oo][Mm]:<(.*)>`)

	fromToRE = regexp.MustCompile(`(?m)["“”']{0,1}(.*)["“”']{0,1}[ ]{0,1}[<](.*)@(.*)[>]`)
)
