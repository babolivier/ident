package tests

import (
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/babolivier/ident/common"
	"github.com/babolivier/ident/common/config"
	"github.com/babolivier/ident/common/database"

	"github.com/gorilla/mux"
)

var ConfigYAML = `
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

func NewTestConfig() (*config.Config, error) {
	return config.ParseConfig([]byte(ConfigYAML))
}

func InitTestRouting(
	setupRouting func(*mux.Router, *config.Config, *database.Database),
) (cfg *config.Config, db *database.Database, s *httptest.Server, err error) {
	cfg, err = NewTestConfig()
	if err != nil {
		return
	}

	db, err = database.NewDatabase(cfg.Database.Driver, cfg.Database.ConnString)
	if err != nil {
		return
	}

	s = NewTestServer(cfg, db, setupRouting)
	return
}

func NewTestServer(
	cfg *config.Config, db *database.Database,
	setupRouting func(*mux.Router, *config.Config, *database.Database),
) *httptest.Server {
	// Create the router and register the handler for the status check route.
	router := mux.NewRouter().UseEncodedPath().PathPrefix(common.APIPrefix).Subrouter()
	setupRouting(router, cfg, db)

	return httptest.NewServer(router)
}

func AssertEqual(t *testing.T, a, b interface{}) {
	if !reflect.DeepEqual(a, b) {
		t.Fatalf("Assertion failed: %v != %v", a, b)
	}
}
