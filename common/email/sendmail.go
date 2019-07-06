package email

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/smtp"
	"net/textproto"
	"time"

	"github.com/babolivier/ident/common/config"

	"github.com/pkg/errors"
)

func SendMail(cfg *config.Config, to, templateTXT, templateHTML string, data interface{}) (err error) {
	// Dial the SMTP server.
	tlsconfig := &tls.Config{ServerName: cfg.Email.SMTP.Hostname}
	addr := cfg.Email.SMTP.Hostname + ":" + cfg.Email.SMTP.Port
	conn, err := tls.Dial("tcp", addr, tlsconfig)
	if err != nil {
		return errors.Wrap(err, "Couldn't dial the SMTP server")
	}

	// Initiate the SMTP client.
	client, err := smtp.NewClient(conn, cfg.Email.SMTP.Hostname)
	if err != nil {
		return errors.Wrap(err, "Couldn't instantiate the SMTP client")
	}

	// Auth against the SMTP server
	auth := smtp.PlainAuth("", cfg.Email.SMTP.Username, cfg.Email.SMTP.Password, cfg.Email.SMTP.Hostname)
	if err = client.Auth(auth); err != nil {
		return errors.Wrap(err, "Couldn't authenticate against the SMTP server")
	}

	// Send the MAIL FROM command.
	if err = client.Mail(cfg.Email.From); err != nil {
		return errors.Wrap(err, "Couldn't send MAIL FROM to the SMTP server")
	}

	// Send the RCPT TO command.
	if err = client.Rcpt(to); err != nil {
		return errors.Wrap(err, "Couldn't send RCPT TO to the SMTP server")
	}

	// Send the DATA command and get the writer to write the email's body to.
	w, err := client.Data()
	if err != nil {
		return errors.Wrap(err, "Couldn't send DATA to the SMTP server")
	}

	// Generate the email body.
	if err = generateEmail(cfg, w, to, templateTXT, templateHTML, data); err != nil {
		return errors.Wrap(err, "Couldn't generate the email's body")
	}

	// Close the writer now that all of the content is written.
	if err = w.Close(); err != nil {
		return errors.Wrap(err, "Couldn't close the email body writer")
	}

	// Send the QUIT command to validate the operation with the server.
	if err = client.Quit(); err != nil {
		return errors.Wrap(err, "Couldn't send QUIT to the SMTP server")
	}

	return nil
}

func generateEmail(cfg *config.Config, w io.Writer, to, templateTXT, templateHTML string, data interface{}) (err error) {
	// Instantiate the multipart.Writer and generate the subject from the template.
	mw := multipart.NewWriter(w)
	subject, err := loadSubjectTemplate(cfg, data)
	if err != nil {
		return
	}

	// Set the email headers.
	if _, err = fmt.Fprintf(w, "Date: %s\r\n", time.Now().Format(time.RFC1123Z)); err != nil {
		return
	}
	if _, err = fmt.Fprintf(w, "From: %s\r\n", cfg.Email.From); err != nil {
		return
	}
	if _, err = fmt.Fprintf(w, "To: %s\r\n", to); err != nil {
		return
	}
	if _, err = fmt.Fprintf(w, "Subject: %s\r\n", subject); err != nil {
		return
	}
	if _, err = fmt.Fprintf(w, "MIME-Version: 1.0\r\n"); err != nil {
		return
	}
	if _, err = fmt.Fprintf(w, "Content-Type: multipart/mixed; boundary=%s\r\n", mw.Boundary()); err != nil {
		return
	}
	if _, err = fmt.Fprintf(w, "\r\n"); err != nil {
		return
	}

	// Start writing the multipart message.
	//
	// The mainly used structure and the (only?) one accepted by Gmail is:
	//
	//  multipart/mixed
	//  `- multipart/alternative
	//     |- text/plain
	//     `- multipart/related
	//        `- text/html
	//
	// c.f. https://stackoverflow.com/a/23853079 (minus images and attachments because we don't care about these)
	aw := multipart.NewWriter(w)
	_, err = mw.CreatePart(textproto.MIMEHeader{"Content-Type": {"multipart/alternative; boundary=" + aw.Boundary()}})
	if err != nil {
		return
	}

	// Generate the plain text version from the plain text template if there's one.
	if len(templateTXT) > 0 {
		if err = loadBodyTemplate(aw, templateTXT, "text/plain", data); err != nil {
			return errors.Wrap(err, "Couldn't generate the plain text part of the message")
		}
	}

	// Generate the HTML version from the HTML template if there's one.
	if len(templateHTML) > 0 {
		rw := multipart.NewWriter(w)
		_, _ = aw.CreatePart(textproto.MIMEHeader{"Content-Type": {"multipart/related; boundary=" + rw.Boundary()}})

		if err = loadBodyTemplate(rw, templateTXT, "text/html", data); err != nil {
			return errors.Wrap(err, "Couldn't generate the plain HTML of the message")
		}
	}

	return nil
}

func loadSubjectTemplate(cfg *config.Config, data interface{}) (subject string, err error) {
	buf := bytes.NewBuffer(nil)

	// Parse the template.
	tmpl, err := template.New("subject").Parse(cfg.Ident.Invites.SubjectTemplate)
	if err != nil {
		return
	}

	// Generate bytes from the template and data.
	if err = tmpl.Execute(buf, data); err != nil {
		return
	}

	return buf.String(), nil
}

func loadBodyTemplate(w *multipart.Writer, templateName, mimetype string, data interface{}) error {
	// Define the part's header.
	mimeHeader := textproto.MIMEHeader{
		"Content-Type":        {mimetype + "; charset=UTF-8"},
		"Content-Disposition": {"inline"},
	}

	// Create the part in the multipart body.
	part, err := w.CreatePart(mimeHeader)
	if err != nil {
		return err
	}

	// Open and read the template file.
	b, err := ioutil.ReadFile(templateName)
	if err != nil {
		return err
	}

	// Parse the template file.
	tmpl, err := template.New(mimetype).Parse(string(b))
	if err != nil {
		return err
	}

	// Generate bytes from the template and the data and write them to the multipart.Writer.
	return tmpl.Execute(part, data)
}
