package notifications

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"os/exec"
	"strconv"
)

// TODO: make subject and body of mail configurable
type Mail struct {
	Email string `long:"email" Description:"Email address to which notifications get sent"`

	SmtpServer string `long:"server" Description:"SMTP server with port for sending mails"`
	SmtpSender string `long:"sender" Description:"Email address from which notifications get sent"`

	SmtpSSL      bool   `long:"ssl" Description:"Whether SSL should be used or not"`
	SmtpUser     string `long:"user" Description:"User for authenticating the SMTP connection"`
	SmtpPassword string `long:"password" Description:"Password for authenticating the SMTP connection"`
}

const sendmail = "/usr/bin/mail"

const newLine = "\r\n"

const subject = "You received a tip"

func (mail *Mail) SendMail(amount int64, message string) {
	body := "You received a tip of " + strconv.FormatInt(amount, 10) + " satoshis"

	if message != "" {
		body += " with the message \"" + message + "\""
	}

	if mail.SmtpServer == "" {
		// "mail" command will be used for sending
		cmd := exec.Command(sendmail, "-s", subject, mail.Email)

		writer, err := cmd.StdinPipe()

		if err == nil {
			err = cmd.Start()

			if err == nil {
				_, err = writer.Write([]byte(body))

				if err == nil {
					err = writer.Close()

					if err == nil {
						err = cmd.Wait()

						if err == nil {
							logSent()

							return
						}

					}

				}

			}

		}

		logSendingFailed(err)

	} else {
		// SMTP server will be used
		mail.sendMailSmtp(body)
	}

}

func (mail *Mail) sendMailSmtp(body string) {
	// Because the SMTP method does'n have a dedicated field for the subject
	// it will be in the body of the message
	body = "Subject: " + subject + newLine + body

	var auth smtp.Auth

	host, _, err := net.SplitHostPort(mail.SmtpServer)

	if err != nil {
		log.Error("Failed to parse host of SMTP server: " + mail.SmtpServer)

		return
	}

	if mail.SmtpUser != "" {
		// If there are credentials they are used
		auth = smtp.PlainAuth(
			"",
			mail.SmtpUser,
			mail.SmtpPassword,
			host,
		)

	}

	if mail.SmtpSSL {
		body = "From: " + mail.SmtpSender + newLine + "To: " + mail.Email + newLine + body

		tlsConfig := &tls.Config{
			ServerName:         host,
			InsecureSkipVerify: true,
		}

		con, err := tls.Dial("tcp", mail.SmtpServer, tlsConfig)

		defer con.Close()

		if err != nil {
			log.Error("Failed to connect to connect to SMTP server: " + fmt.Sprint(err))

			return
		}

		client, err := smtp.NewClient(con, host)

		defer client.Close()

		if err != nil {
			log.Error("Failed to create SMTP client: " + fmt.Sprint(err))

			return
		}

		err = client.Auth(auth)

		if err != nil {
			log.Error("Failed to authenticate SMTP client: " + fmt.Sprint(err))

			return
		}

		err = client.Mail(mail.SmtpSender)
		err = client.Rcpt(mail.Email)

		writer, err := client.Data()

		defer writer.Close()

		if err == nil {
			_, err = writer.Write([]byte(body))
		}

		if err == nil {
			logSent()

		} else {
			logSendingFailed(err)
		}

	} else {
		err := smtp.SendMail(
			mail.SmtpServer,
			auth,
			mail.SmtpSender,
			[]string{mail.Email},
			[]byte(body),
		)

		if err == nil {
			logSent()

		} else {
			logSendingFailed(err)
		}

	}

}

func logSent() {
	log.Debug("Sent email")
}

func logSendingFailed(err error) {
	log.Error("Failed to send mail: " + fmt.Sprint(err))
}
