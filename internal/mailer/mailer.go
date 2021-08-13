package mailer

import (
	"bytes"
	"embed"
	"html/template"
	"time"

	"github.com/go-mail/mail/v2"
)

//go:embed "templates"
var templateFS embed.FS

// Define a Mailer struct which contains a mail.Dialer instance (used to connect to a 
// SMTP server) and the sender information for your emails (the name and address you 
// want the email to be from, such as "Alice Smith <alice@example.com>").
type Mailer struct {
	dialer *mail.Dialer
	sender string
}

func New(host string, port int, username, password, sender string) Mailer {
	// Initialize a new mail.Dialer instance with the given SMTP server settings.
	// We also configure this to use a 5-second timeout whenever we send an email.
	dialer := mail.NewDialer(host, port, username, password)
	dialer.Timeout = 5 * time.Second

	return Mailer{
		dialer: dialer,
		sender: sender,
	}
}

// Define a Send() method on the Mailer type. This takes the recipient email address 
// as the first parameter, the name of the file containing the templates, and any
// dynamic data for the templates as an interface{} parameter.
func (m Mailer) Send(recipient, templateFile string, data interface{}) error {
	tmpl, err := template.New("email").ParseFS(templateFS, "templates/"+templateFile)
	if err != nil {
		return err
	}

	// Execute the named template "subject", passing in the dynamic data and storing the 
	// result in a bytes.Buffer variable.
	subject := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return err
	}

	plainBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(plainBody, "plainBody", data)
	if err != nil {
		return err
	}

	htmlBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(htmlBody, "htmlBody", data)
	if err != nil {
		return err
	}

	// Use the mail.NewMessage() function to initialize a new mail.Message instance.
	// Then we use the SetHeader() method to set the email recipient, sender and subject 
	// headers, the SetBody() method to set the plain-text body, and the AddAlternative() 
	// method to set the HTML body. It's important to note that AddAlternative() should 
	// always be called *after* SetBody().
	msg := mail.NewMessage()
	msg.SetHeader("To", recipient)
	msg.SetHeader("From", m.sender)
	msg.SetHeader("Subject", subject.String())
	msg.SetBody("text/plain", plainBody.String())
	msg.AddAlternative("text/html", htmlBody.String())

	// Add some retry basic logic using a for loop and call the DialAndSend() method 
	// on the dialer, passing in the message to send. This // opens a connection to the SMTP server, sends the message, then closes the
	// connection. If there is a timeout, it will return a "dial tcp: i/o timeout"
	// error.
	for i := 0; i <= 3; i++ {
		err = m.dialer.DialAndSend(msg)
		if err == nil {
			return nil
		}

		time.Sleep(500 * time.Millisecond)
	}

	return err
}