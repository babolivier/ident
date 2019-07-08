package config

import (
	"encoding/base64"
	"testing"

	"golang.org/x/crypto/ed25519"

	"github.com/stretchr/testify/require"
)

var configYAML = `
ident:
  server_name: test
  base_url: "http://127.0.0.1:9999"
  signing_key:
    algo: ed25519
    id: 0
    seed: ahphigh9jahchiequiechee4pha1Atuv
  invites:
    email_template:
      text: "templates/text/invite.txt"
      html: "templates/html/invite.html"
    subject_template: "{{.SenderDisplayName}} invited you to Matrix!"

http:
  listen_addr: "127.0.0.1:9999"

database:
  driver: sqlite3
  conn_string: ":memory:"

email:
  from: "Ident <ident@example.com>"
  smtp:
    hostname: mail.example.com
    port: 465
    username: "ident@example.com"
    password: somepassword
    enable_tls: on
`

func TestParseConfig(t *testing.T) {
	cfg, err := ParseConfig([]byte(configYAML))

	require.Equal(t, err, nil)

	require.Equal(t, cfg.HTTP.ListenAddr, "127.0.0.1:9999")

	require.Equal(t, cfg.Database.Driver, "sqlite3")
	require.Equal(t, cfg.Database.ConnString, ":memory:")

	require.Equal(t, cfg.Ident.ServerName, "test")
	require.Equal(t, cfg.Ident.BaseURL, "http://127.0.0.1:9999")

	require.Equal(t, cfg.Ident.SigningKey.Algo, "ed25519")
	require.Equal(t, cfg.Ident.SigningKey.ID, "0")
	require.Equal(t, cfg.Ident.SigningKey.Seed, "ahphigh9jahchiequiechee4pha1Atuv")

	privKey := ed25519.NewKeyFromSeed([]byte(cfg.Ident.SigningKey.Seed))
	pubKey := privKey.Public().(ed25519.PublicKey)
	pubKeyBase64 := base64.RawStdEncoding.EncodeToString(pubKey)

	require.Equal(t, cfg.Ident.SigningKey.PrivKey, privKey)
	require.Equal(t, cfg.Ident.SigningKey.PubKey, pubKey)
	require.Equal(t, cfg.Ident.SigningKey.PubKeyBase64, pubKeyBase64)

	require.Equal(t, cfg.Ident.Invites.SubjectTemplate, "{{.SenderDisplayName}} invited you to Matrix!")
	require.Equal(t, cfg.Ident.Invites.EmailTemplate.Text, "templates/text/invite.txt")
	require.Equal(t, cfg.Ident.Invites.EmailTemplate.HTML, "templates/html/invite.html")

	require.Equal(t, cfg.Email.From, "Ident <ident@example.com>")
	require.Equal(t, cfg.Email.SMTP.Hostname, "mail.example.com")
	require.Equal(t, cfg.Email.SMTP.Port, "465")
	require.Equal(t, cfg.Email.SMTP.Username, "ident@example.com")
	require.Equal(t, cfg.Email.SMTP.Password, "somepassword")
	require.Equal(t, cfg.Email.SMTP.EnableTLS, true)
}
