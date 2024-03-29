package config

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/babolivier/ident/common/constants"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ed25519"
)

func TestParseConfig(t *testing.T) {
	cfg, err := ParseConfig([]byte(constants.TestConfigYAML))

	require.Nil(t, err, err)

	require.Equal(t, "127.0.0.1:9999", cfg.HTTP.ListenAddr)

	require.Equal(t, "sqlite3", cfg.Database.Driver)
	require.Equal(t, ":memory:", cfg.Database.ConnString)

	require.Equal(t, "test", cfg.Ident.ServerName)
	require.Equal(t, "http://127.0.0.1:9999", cfg.Ident.BaseURL)

	require.Equal(t, "ed25519", cfg.Ident.SigningKey.Algo)
	require.Equal(t, "0", cfg.Ident.SigningKey.ID)
	require.Equal(t, "ahphigh9jahchiequiechee4pha1Atuv", cfg.Ident.SigningKey.Seed)

	privKey := ed25519.NewKeyFromSeed([]byte(cfg.Ident.SigningKey.Seed))
	pubKey := privKey.Public().(ed25519.PublicKey)
	pubKeyBase64 := base64.RawStdEncoding.EncodeToString(pubKey)

	require.Equal(t, privKey, cfg.Ident.SigningKey.PrivKey)
	require.Equal(t, pubKey, cfg.Ident.SigningKey.PubKey)
	require.Equal(t, pubKeyBase64, cfg.Ident.SigningKey.PubKeyBase64)

	require.Equal(t, "{{.SenderDisplayName}} invited you to Matrix!", cfg.Ident.Invites.SubjectTemplate)
	require.Equal(t, "/tmp/ident_invite_template_txt", cfg.Ident.Invites.EmailTemplate.Text)
	require.Equal(t, "/tmp/ident_invite_template_html", cfg.Ident.Invites.EmailTemplate.HTML)

	require.Equal(t, "Ident <ident@example.com>", cfg.Email.From)
	require.Equal(t, "mail.example.com", cfg.Email.SMTP.Hostname)
	require.Equal(t, "465", cfg.Email.SMTP.Port)
	require.Equal(t, "ident@example.com", cfg.Email.SMTP.Username)
	require.Equal(t, "somepassword", cfg.Email.SMTP.Password)
	require.True(t, cfg.Email.SMTP.EnableTLS)
}

func TestParseConfigInvalidYAML(t *testing.T) {
	yaml := "something: something_else: invalid_value" +
		"ident:\n" +
		"  signing_key:" +
		"    algo: unsupported_algo"

	_, err := ParseConfig([]byte(yaml))
	require.NotNil(t, err)
	require.True(t, strings.HasPrefix(err.Error(), "Couldn't read the configuration file"), err.Error())
}

func TestParseConfigInvalidKeyAlgo(t *testing.T) {
	yaml := "" +
		"ident:\n" +
		"  signing_key:\n" +
		"    algo: unsupported_algo"

	_, err := ParseConfig([]byte(yaml))
	require.NotNil(t, err)
	require.True(t, strings.HasPrefix(err.Error(), "Invalid signing key configuration"), err)
}
