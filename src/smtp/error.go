package smtp

type SMTPError string

func (err SMTPError) Error() string {
	return string(err)
}
