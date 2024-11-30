package helper

import "gopkg.in/gomail.v2"

func SendEmail(to string, subject string, body string) error {
	msg := gomail.NewMessage()

	msg.SetHeader("From", "Fixchirp <info@fixchirp.com>")
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/html", body)

	dialer := gomail.NewDialer("smtp.email.com", 465, "email@email.com", "secret")

	if err := dialer.DialAndSend(msg); err != nil {
		return err
	}

	return nil
}
