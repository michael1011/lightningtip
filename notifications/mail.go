package notifications

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"os/exec"
	"strconv"
)

// Mail contains all values needed to be able to send a mail
type Mail struct {
	Recipient string `long:"recipient" Description:"Email address to which notifications get sent"`
	Sender    string `long:"sender" Description:"Email address from which notifications get sent"`

	SMTPServer string `long:"server" Description:"SMTP server with port for sending mails"`

	SMTPSSL      bool   `long:"ssl" Description:"Whether SSL should be used or not"`
	SMTPUser     string `long:"user" Description:"User for authenticating the SMTP connection"`
	SMTPPassword string `long:"password" Description:"Password for authenticating the SMTP connection"`
}

const sendmail = "/usr/bin/mail"

const newLine = "\r\n"

const subject = "You received a tip"

// SendMail sends a mail
func (mail *Mail) SendMail(amount int64, message string) {
	body := "You received a tip of " + strconv.FormatInt(amount, 10) + " satoshis"

	if message != "" {
		body += " with the message \"" + message + "\""
	}

	if mail.SMTPServer == "" {
		// "mail" command will be used for sending
		mail.sendMailCommand(body)

	} else {
		// SMTP server will be used
		mail.sendMailSMTP(body)
	}

}

// Sends a mail with the "mail" command
func (mail *Mail) sendMailCommand(body string) {
	var cmd *exec.Cmd

	if mail.Sender == "" {
		cmd = exec.Command(sendmail, "-s", subject, mail.Recipient)

	} else {
		// Append "From" header
		cmd = exec.Command(sendmail, "-s", subject, "-a", "From: "+mail.Sender, mail.Recipient)
	}

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
}

func (mail *Mail) sendMailSMTP(body string) {
	// Because the SMTP method doesn't have a dedicated field for the subject
	// it will be in the body of the message
	body = "Subject: " + subject + newLine + body

	var auth smtp.Auth

	host, _, err := net.SplitHostPort(mail.SMTPServer)

	if err != nil {
		log.Error("Failed to parse host of SMTP server: " + mail.SMTPServer)

		return
	}

	if mail.SMTPUser != "" {
		// If there are credentials they are used
		auth = smtp.PlainAuth(
			"",
			mail.SMTPUser,
			mail.SMTPPassword,
			host,
		)

	}

	if mail.SMTPSSL {
		body = "From: " + mail.Sender + newLine + "To: " + mail.Recipient + newLine + body

		tlsConfig := &tls.Config{
			ServerName:         host,
			InsecureSkipVerify: true,
		}

		con, err := tls.Dial("tcp", mail.SMTPServer, tlsConfig)

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

		err = client.Mail(mail.Sender)
		err = client.Rcpt(mail.Recipient)

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
			mail.SMTPServer,
			auth,
			mail.Sender,
			[]string{mail.Recipient},
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
