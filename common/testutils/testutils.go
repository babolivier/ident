package testutils

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/babolivier/ident/common"
	"github.com/babolivier/ident/common/config"
	"github.com/babolivier/ident/common/constants"
	"github.com/babolivier/ident/common/database"

	"github.com/gorilla/mux"
	"github.com/matrix-org/gomatrix"
	"github.com/matrix-org/util"
	"github.com/stretchr/testify/require"
)

var testConfig *config.Config
var testDB *database.Database

func NewTestConfig(t *testing.T) *config.Config {
	if testConfig != nil {
		return testConfig
	}

	var err error
	testConfig, err = config.ParseConfig([]byte(constants.TestConfigYAML))
	require.Nil(t, err, err)

	return testConfig
}

func NewTestDB(t *testing.T) *database.Database {
	cfg := NewTestConfig(t)
	db, err := database.NewDatabase(cfg.Database.Driver, cfg.Database.ConnString)
	require.Nil(t, err, err)
	return db
}

func InitTestRouting(
	t *testing.T, setupRouting func(*mux.Router, *config.Config, *database.Database),
) (cfg *config.Config, db *database.Database, s *httptest.Server, err error) {
	cfg = NewTestConfig(t)
	db = NewTestDB(t)

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

	router.NotFoundHandler = common.MakeAPI(func(r *http.Request) util.JSONResponse {
		return util.JSONResponse{
			Code: 404,
			JSON: gomatrix.RespError{
				ErrCode: "M_NOT_FOUND",
				Err:     "Unrecognised request",
			},
		}
	})

	return httptest.NewServer(router)
}

func TestWithTmpFiles(t *testing.T, testFunc func(t *testing.T), files map[string]string) {
	for name, content := range files {
		err := ioutil.WriteFile(name, []byte(content), 0655)
		require.Nil(t, err, err)

		defer os.Remove(name)
	}

	testFunc(t)
}
