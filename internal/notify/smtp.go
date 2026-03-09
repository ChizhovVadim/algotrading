package notify

import (
	"bytes"
	"fmt"
	"net/smtp"
)

type SmtpWriter struct {
	addr    string
	auth    smtp.Auth
	from    string
	to      []string
	subject string
}

func NewSmtpWriter(
	from string,
	to string,
	password string,
	host string,
	port string,
	subject string,
) *SmtpWriter {
	return &SmtpWriter{
		addr:    host + ":" + port,
		auth:    smtp.PlainAuth("", from, password, host),
		from:    from,
		to:      []string{to},
		subject: subject,
	}
}

func (srv *SmtpWriter) Write(s string) error {
	var buff = &bytes.Buffer{}
	fmt.Fprintf(buff, "Subject: %v\r\n", srv.subject)
	buff.WriteString(s)
	var msg = buff.Bytes()
	return smtp.SendMail(srv.addr, srv.auth, srv.from, srv.to, msg)
}
