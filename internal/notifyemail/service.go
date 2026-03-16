package notifyemail

import (
	"bytes"
	"fmt"
	"net/smtp"
)

type Service struct {
	addr    string
	auth    smtp.Auth
	from    string
	to      []string
	subject string
}

func New(
	from string,
	to string,
	password string,
	host string,
	port string,
	subject string,
) *Service {
	return &Service{
		addr:    host + ":" + port,
		auth:    smtp.PlainAuth("", from, password, host),
		from:    from,
		to:      []string{to},
		subject: subject,
	}
}

func (s *Service) Notify(message string) error {
	var buff = &bytes.Buffer{}
	fmt.Fprintf(buff, "Subject: %v\r\n", s.subject)
	buff.WriteString(message)
	return smtp.SendMail(s.addr, s.auth, s.from, s.to, buff.Bytes())
}
