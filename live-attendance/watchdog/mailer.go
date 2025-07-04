package watchdog

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"net/smtp"
	"strconv"
)

// This code as been stolen from tac -> https://github.com/tac4ttack

type ConfMailer struct {
	SmtpServer string   `yaml:"smtp_server"`
	SmtpPort   int      `yaml:"smtp_port"`
	SmtpAuth   bool     `yaml:"smtp_auth"`
	SmtpUser   string   `yaml:"smtp_user"`
	SmtpPass   string   `yaml:"smtp_pass"`
	SmtpTls    bool     `yaml:"smtp_tls"`
	Helo       string   `yaml:"helo"`
	FromName   string   `yaml:"from_name"`
	FromMail   string   `yaml:"from_mail"`
	Admins     []string `yaml:"admins"`
}

type mailer struct {
	smtpServer string
	smtpPort   int
	smtpAuth   bool
	smtpUser   string
	smtpPass   string
	smtpTls    bool
	helo       string
	fromName   string
	fromMail   string
	auth       smtp.Auth
	client     *smtp.Client
	admins     []string
}

var m mailer

func Init(conf ConfMailer) {
	m = mailer{
		smtpPort:   conf.SmtpPort,
		smtpUser:   conf.SmtpUser,
		smtpPass:   conf.SmtpPass,
		smtpServer: conf.SmtpServer,
		smtpAuth:   conf.SmtpAuth,
		smtpTls:    conf.SmtpTls,
		helo:       conf.Helo,
		fromName:   conf.FromName,
		fromMail:   conf.FromMail,
		admins:     conf.Admins,
		auth:       nil,
		client:     nil,
	}
}

func Send(to []string, subject string, body string, html bool) error {
	var err error

	// Connection to the remote SMTP server
	if m.client, err = smtp.Dial(m.smtpServer + ":" + strconv.Itoa(m.smtpPort)); err != nil {
		return err
	}
	defer m.client.Close()

	// Sending HELO / EHLO message to the server
	if err = m.client.Hello(m.helo); err != nil {
		m.client.Reset()
		return err
	}

	// Start TLS command if server requires it
	if m.smtpTls {
		if err = m.client.StartTLS(
			&tls.Config{
				ServerName:         m.smtpServer,
				InsecureSkipVerify: true}); err != nil {
			m.client.Reset()
			return err
		}
	}

	// SMTP Authentication
	if m.smtpAuth {
		if err = m.client.Auth(smtp.PlainAuth("", m.smtpUser, m.smtpPass, m.smtpServer)); err != nil {
			m.client.Reset()
			return err
		}
	}

	// Set the sender
	if err = m.client.Mail(m.fromMail); err != nil {
		m.client.Reset()
		return err
	}

	// Set the recipients
	for _, r := range to {
		if err = m.client.Rcpt(r); err != nil {
			m.client.Reset()
			return nil
		}
	}

	// Prepare the body
	var wc io.WriteCloser
	if wc, err = m.client.Data(); err != nil {
		m.client.Reset()
		return err
	}
	var data string
	if html {

		data = string(
			composeHTML(to,
				m.fromName+"<"+m.fromMail+">",
				subject,
				body))
	} else {

		data = string(
			compose(to,
				m.fromName+"<"+m.fromMail+">",
				subject,
				body))
	}
	if _, err = io.WriteString(wc, data); err != nil {
		m.client.Reset()
		return err
	}
	if err = wc.Close(); err != nil {
		m.client.Reset()
		return err
	}

	// Send the QUIT command and close connection
	if err = m.client.Quit(); err != nil {
		m.client.Reset()
		return err
	}

	return nil
}

func compose(to []string, from string, subject string, body string) []byte {
	// Formatting MIME headers
	tmp := make(map[string]string)
	tmp["From"] = from
	tmp["To"] = ""
	for i := 0; i < len(to); i++ {
		if to[i] != "" {
			tmp["To"] += to[i]
			if i+1 < len(to) {
				tmp["To"] += ","
			}
		}
	}
	tmp["Subject"] = subject
	tmp["MIME-Version"] = "1.0"
	tmp["Content-Type"] = "text/plain; charset=\"utf-8\""
	tmp["Content-Transfer-Encoding"] = "base64"

	msg := ""
	// Formatting MIME headers into our message
	for i, j := range tmp {
		msg += fmt.Sprintf("%s: %s\r\n", i, j)
	}
	// Adding the email content into our mesage
	msg += "\r\n" + base64.StdEncoding.EncodeToString([]byte(body))
	return []byte(msg)
}

func composeHTML(to []string, from string, subject string, body string) []byte {
	// Formatting MIME headers
	tmp := make(map[string]string)
	tmp["From"] = from
	tmp["To"] = ""
	for i := 0; i < len(to); i++ {
		if to[i] != "" {
			tmp["To"] += to[i]
			if i+1 < len(to) {
				tmp["To"] += ","
			}
		}
	}
	tmp["Subject"] = subject
	tmp["MIME-Version"] = "1.0"
	tmp["Content-Type"] = "text/html; charset=\"UTF-8\";"
	tmp["Content-Transfer-Encoding"] = "base64"

	msg := ""
	// Formatting MIME headers into our message
	for i, j := range tmp {
		msg += fmt.Sprintf("%s: %s\r\n", i, j)
	}
	// Adding the email content into our mesage
	msg += "\r\n" + base64.StdEncoding.EncodeToString([]byte(body))
	return []byte(msg)
}
