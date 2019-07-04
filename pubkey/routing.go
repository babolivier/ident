package pubkey

import (
	"net/http"

	"github.com/babolivier/ident/common"
	"github.com/babolivier/ident/common/config"
	"github.com/babolivier/ident/common/database"

	"github.com/gorilla/mux"
	"github.com/matrix-org/util"
)

func SetupRouting(router *mux.Router, cfg *config.Config, db *database.Database) {
	router.Handle("/pubkey/isvalid", common.MakeAPI(func(r *http.Request) util.JSONResponse {
		return IsPubKeyValid(r.URL.Query().Get("public_key"), cfg)
	})).Methods(http.MethodGet)

	router.Handle("/pubkey/ephemeral/isvalid", common.MakeAPI(func(r *http.Request) util.JSONResponse {
		return IsEphemeralPubKeyValid(r.URL.Query().Get("public_key"), db)
	})).Methods(http.MethodGet)

	router.Handle("/pubkey/{keyId}", common.MakeAPI(func(r *http.Request) util.JSONResponse {
		vars := mux.Vars(r)
		return GetKey(vars["keyId"], cfg)
	})).Methods(http.MethodGet)
}
