package invites

import (
	"net/http"

	"github.com/babolivier/ident/common"
	"github.com/babolivier/ident/common/config"
	"github.com/babolivier/ident/common/database"

	"github.com/gorilla/mux"
	"github.com/matrix-org/util"
)

func SetupRouting(router *mux.Router, cfg *config.Config, db *database.Database) {
	router.Handle("/store-invite", common.MakeAPI(func(r *http.Request) util.JSONResponse {
		return StoreInvite(r, cfg, db)
	})).Methods(http.MethodOptions, http.MethodPost)

}
