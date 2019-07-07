package config

import (
	"encoding/base64"
	"testing"

	"github.com/babolivier/ident/common/config"
	"github.com/babolivier/ident/tests"

	"golang.org/x/crypto/ed25519"
)

func TestParseConfig(t *testing.T) {
	cfg, err := config.ParseConfig([]byte(tests.ConfigYAML))
	tests.AssertEqual(t, err, nil)

	tests.AssertEqual(t, cfg.HTTP.ListenAddr, "127.0.0.1:9999")

	tests.AssertEqual(t, cfg.Database.Driver, "sqlite3")
	tests.AssertEqual(t, cfg.Database.ConnString, ":memory:")

	tests.AssertEqual(t, cfg.Ident.ServerName, "test")
	tests.AssertEqual(t, cfg.Ident.BaseURL, "http://127.0.0.1:9999")

	tests.AssertEqual(t, cfg.Ident.SigningKey.Algo, "ed25519")
	tests.AssertEqual(t, cfg.Ident.SigningKey.ID, "0")
	tests.AssertEqual(t, cfg.Ident.SigningKey.Seed, "ahphigh9jahchiequiechee4pha1Atuv")

	privKey := ed25519.NewKeyFromSeed([]byte(cfg.Ident.SigningKey.Seed))
	pubKey := privKey.Public().(ed25519.PublicKey)
	pubKeyBase64 := base64.RawStdEncoding.EncodeToString(pubKey)

	tests.AssertEqual(t, cfg.Ident.SigningKey.PrivKey, privKey)
	tests.AssertEqual(t, cfg.Ident.SigningKey.PubKey, pubKey)
	tests.AssertEqual(t, cfg.Ident.SigningKey.PubKeyBase64, pubKeyBase64)

	tests.AssertEqual(t, cfg.Ident.Invites.SubjectTemplate, "{{.SenderDisplayName}} invited you to Matrix!")
	tests.AssertEqual(t, cfg.Ident.Invites.EmailTemplate.Text, "templates/text/invite.txt")
	tests.AssertEqual(t, cfg.Ident.Invites.EmailTemplate.HTML, "templates/html/invite.html")

	tests.AssertEqual(t, cfg.Email.From, "Ident <ident@example.com>")
	tests.AssertEqual(t, cfg.Email.SMTP.Hostname, "mail.example.com")
	tests.AssertEqual(t, cfg.Email.SMTP.Port, "465")
	tests.AssertEqual(t, cfg.Email.SMTP.Username, "ident@example.com")
	tests.AssertEqual(t, cfg.Email.SMTP.Password, "somepassword")
	tests.AssertEqual(t, cfg.Email.SMTP.EnableTLS, true)
}
