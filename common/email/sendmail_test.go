package email

import (
	"bufio"
	"bytes"
	"io"
	"mime/multipart"
	"net/mail"
	"strings"
	"testing"

	"github.com/babolivier/ident/common/testutils"

	"github.com/stretchr/testify/require"
)

// Mock of invites.StoreInviteReq. The reason we're not testing with a real StoreInviteReq is to
// avoid an import cycle (email -> invites -> email).
type req struct {
	RoomID            string
	SenderDisplayName string
	Token             string
}

func TestGenerateEmail(t *testing.T) {
	cfg := testutils.NewTestConfig(t)

	files := map[string]string{
		cfg.Ident.Invites.EmailTemplate.Text: "{{.SenderDisplayName}} - {{.RoomID}} - {{.Token}}",
		cfg.Ident.Invites.EmailTemplate.HTML: "<p>{{.SenderDisplayName}} - {{.RoomID}} - {{.Token}}</p>",
	}

	testutils.TestWithTmpFiles(t, testGenerateEmail, files)
}

func testGenerateEmail(t *testing.T) {
	// Tests that generateEmail builds a multipart message with the right structure.
	//
	// The structure to follow is:
	//
	//  multipart/mixed
	//  `- multipart/alternative
	//     |- text/plain
	//     `- multipart/related
	//        `- text/html

	// When reading the body of a part, EOF means that we've reached a boundary or the end of the payload before all of
	// the bytes in the slice are filled. We only care about having the full body, so we'll deliberately give the reader
	// more bytes than necessary. Therefore, we must consider this error as a nil error.
	allowedReadErrors := []error{nil, io.EOF}

	cfg := testutils.NewTestConfig(t)
	buf := bytes.NewBuffer(nil)
	to := "alice@example.com"
	req := &req{
		SenderDisplayName: "alice",
		RoomID:            "!someroom:example.com",
		Token:             "sometoken",
	}

	err := generateEmail(cfg, buf, to, cfg.Ident.Invites.EmailTemplate.Text, cfg.Ident.Invites.EmailTemplate.HTML, req)
	require.Nil(t, err, err)

	reader := bytes.NewReader(buf.Bytes())
	msg, err := mail.ReadMessage(reader)
	require.Nil(t, err, err)

	parsedSubject, err := loadSubjectTemplate(cfg, req)
	require.Nil(t, err, err)

	// Test email headers
	require.Equal(t, cfg.Email.From, msg.Header.Get("From"))
	require.Equal(t, to, msg.Header.Get("To"))
	require.Equal(t, parsedSubject, msg.Header.Get("Subject"))
	require.Equal(t, "1.0", msg.Header.Get("MIME-Version"))
	require.True(t, strings.HasPrefix(msg.Header.Get("Content-Type"), "multipart/mixed; boundary="))

	// Parse the multipart/mixed.
	boundary := strings.SplitN(msg.Header.Get("Content-Type"), "=", 2)[1]
	multipartReader := multipart.NewReader(msg.Body, boundary)
	part, err := multipartReader.NextPart()
	require.Nil(t, err, err)

	// Parse the multipart/alternative
	require.True(t, strings.HasPrefix(part.Header.Get("Content-Type"), "multipart/alternative; boundary="))
	alternativeBoundary := strings.SplitN(part.Header.Get("Content-Type"), "=", 2)[1]

	alternativeBytes := make([]byte, 1500)
	n, err := part.Read(alternativeBytes)
	require.Contains(t, allowedReadErrors, err, err)
	alternativeBytes = alternativeBytes[:n]

	alternativeReader := bytes.NewReader(alternativeBytes)
	alternativeMultipartReader := multipart.NewReader(alternativeReader, alternativeBoundary)

	// Parse the text/plain
	alternativePart, err := alternativeMultipartReader.NextPart()
	require.Nil(t, err, err)

	require.Equal(t, "text/plain; charset=UTF-8", alternativePart.Header.Get("Content-Type"))
	require.Equal(t, "inline", alternativePart.Header.Get("Content-Disposition"))

	textContentBytes := make([]byte, 1500)
	n, err = alternativePart.Read(textContentBytes)
	require.False(t, err != nil && err != io.EOF, err)
	textContentBytes = textContentBytes[:n]

	require.Equal(t, "alice - !someroom:example.com - sometoken", string(textContentBytes))

	// Parse the multipart/related
	alternativePart, err = alternativeMultipartReader.NextPart()
	require.Nil(t, err, err)

	require.True(t, strings.HasPrefix(alternativePart.Header.Get("Content-Type"), "multipart/related; boundary="))
	relatedBoundary := strings.SplitN(alternativePart.Header.Get("Content-Type"), "=", 2)[1]

	relatedBytes := make([]byte, 1500)
	n, err = alternativePart.Read(relatedBytes)
	require.Contains(t, allowedReadErrors, err, err)
	relatedBytes = relatedBytes[:n]

	relatedReader := bytes.NewReader(relatedBytes)
	relatedMultipartReader := multipart.NewReader(relatedReader, relatedBoundary)

	// Parse the text/html
	relatedPart, err := relatedMultipartReader.NextPart()
	require.Nil(t, err, err)

	require.Equal(t, "text/html; charset=UTF-8", relatedPart.Header.Get("Content-Type"))
	require.Equal(t, "inline", relatedPart.Header.Get("Content-Disposition"))

	htmlContentBytes := make([]byte, 1500)
	n, err = relatedPart.Read(htmlContentBytes)
	require.Contains(t, allowedReadErrors, err, err)
	htmlContentBytes = htmlContentBytes[:n]

	require.Equal(t, "<p>alice - !someroom:example.com - sometoken</p>", string(htmlContentBytes))

}

func TestLoadBodyTemplate(t *testing.T) {
	cfg := testutils.NewTestConfig(t)

	files := map[string]string{
		cfg.Ident.Invites.EmailTemplate.Text: "{{.SenderDisplayName}} - {{.RoomID}} - {{.Token}}",
	}

	testutils.TestWithTmpFiles(t, testLoadBodyTemplate, files)
}

func testLoadBodyTemplate(t *testing.T) {
	cfg := testutils.NewTestConfig(t)

	req := &req{
		SenderDisplayName: "alice",
		RoomID:            "!someroom:example.com",
		Token:             "sometoken",
	}

	buf := bytes.NewBuffer(nil)
	mw := multipart.NewWriter(buf)

	err := loadBodyTemplate(mw, cfg.Ident.Invites.EmailTemplate.Text, "text/plain", req)
	require.Nil(t, err, err)

	r := strings.NewReader(buf.String())
	scanner := bufio.NewScanner(r)

	var n int
	for scanner.Scan() {
		line := scanner.Text()

		switch n {
		case 0:
			require.Equal(t, "--"+mw.Boundary(), line)
		case 1:
			require.Equal(t, "Content-Disposition: inline", line)
		case 2:
			require.Equal(t, "Content-Type: text/plain; charset=UTF-8", line)
		case 4:
			require.Equal(t, "alice - !someroom:example.com - sometoken", line)
		}

		n++
	}
}

func TestLoadSubjectTemplate(t *testing.T) {
	cfg := testutils.NewTestConfig(t)

	req := &req{
		SenderDisplayName: "alice",
	}

	subj, err := loadSubjectTemplate(cfg, req)

	require.Nil(t, err, err)
	require.Equal(t, "alice invited you to Matrix!", subj)
}
