package testutils

import (
	"net/http/httptest"

	"github.com/babolivier/ident/common/config"
	"github.com/babolivier/ident/common/constants"
	"github.com/babolivier/ident/common/database"

	"github.com/gorilla/mux"
)

func NewTestConfig() *config.Config {
	c, err := config.ParseConfig([]byte(constants.TestConfigYAML))
	if err != nil {
		panic(err)
	}

	return c
}

func InitTestRouting(
	setupRouting func(*mux.Router, *config.Config, *database.Database),
) (cfg *config.Config, db *database.Database, s *httptest.Server, err error) {
	cfg = NewTestConfig()

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
	router := mux.NewRouter().UseEncodedPath().PathPrefix(constants.APIPrefix).Subrouter()
	setupRouting(router, cfg, db)

	return httptest.NewServer(router)
}
