package config

import (
	"encoding/base64"
	"io/ioutil"

	"github.com/pkg/errors"
	"golang.org/x/crypto/ed25519"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Database DatabaseConfig `yaml:"database"`
	HTTP     HTTPConfig     `yaml:"http"`
	Ident    IdentConfig    `yaml:"ident"`
	Email    EmailConfig    `yaml:"email"`
}

type HTTPConfig struct {
	ListenAddr string `yaml:"listen_addr"`
}

type DatabaseConfig struct {
	Driver     string `yaml:"driver"`
	ConnString string `yaml:"conn_string"`
}

type IdentConfig struct {
	ServerName string           `yaml:"server_name"`
	BaseURL    string           `yaml:"base_url"`
	SigningKey SigningKeyConfig `yaml:"signing_key"`
	Invites    InvitesConfig    `yaml:"invites"`
}

type SigningKeyConfig struct {
	Algo         string `yaml:"algo"`
	ID           string `yaml:"id"`
	Seed         string `yaml:"seed"`
	PrivKey      ed25519.PrivateKey
	PubKey       ed25519.PublicKey
	PubKeyBase64 string
}

type InvitesConfig struct {
	EmailTemplate   TemplateConfig `yaml:"email_template"`
	SubjectTemplate string         `yaml:"subject_template"`
}

type TemplateConfig struct {
	HTML string `yaml:"html"`
	Text string `yaml:"text"`
}

type EmailConfig struct {
	From string     `yaml:"from"`
	SMTP SMTPConfig `yaml:"smtp"`
}

type SMTPConfig struct {
	Hostname string `yaml:"hostname"`
	Port     string `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

func NewConfig(filename string) (*Config, error) {
	configBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, errors.Wrap(err, "Couldn't open the configuration file")
	}

	return ParseConfig(configBytes)
}

func ParseConfig(b []byte) (*Config, error) {
	c := new(Config)

	if err := yaml.Unmarshal(b, c); err != nil {
		return nil, errors.Wrap(err, "Couldn't read the configuration file")
	}

	if c.Ident.SigningKey.Algo != "ed25519" {
		return nil, errors.New("Invalid signing key configuration: only ed25519 is currently allowed")
	}

	c.Ident.SigningKey.PrivKey = ed25519.NewKeyFromSeed([]byte(c.Ident.SigningKey.Seed))
	c.Ident.SigningKey.PubKey = c.Ident.SigningKey.PrivKey.Public().(ed25519.PublicKey)
	c.Ident.SigningKey.PubKeyBase64 = base64.RawStdEncoding.EncodeToString(c.Ident.SigningKey.PubKey)

	return c, nil

}
