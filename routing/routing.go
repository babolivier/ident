package routing

import (
	"net/http"

	"github.com/babolivier/ident/common"
	"github.com/babolivier/ident/common/config"
	"github.com/babolivier/ident/common/constants"
	"github.com/babolivier/ident/common/database"
	"github.com/babolivier/ident/invites"
	"github.com/babolivier/ident/pubkey"

	"github.com/gorilla/mux"
	"github.com/matrix-org/gomatrix"
	"github.com/matrix-org/util"
)

func NewRouter(cfg *config.Config, db *database.Database) *mux.Router {
	// Create the router and register the handler for the status check route.
	router := mux.NewRouter().UseEncodedPath().PathPrefix(constants.APIPrefix).Subrouter()
	router.Handle("", common.MakeAPI(func(r *http.Request) util.JSONResponse {
		return util.JSONResponse{
			Code: 200,
			JSON: struct{}{},
		}
	})).Methods(http.MethodGet)

	pubkey.SetupRouting(router, cfg, db)
	invites.SetupRouting(router, cfg, db)

	router.NotFoundHandler = common.MakeAPI(func(r *http.Request) util.JSONResponse {
		return util.JSONResponse{
			Code: 404,
			JSON: gomatrix.RespError{
				ErrCode: "M_NOT_FOUND",
				Err:     "Unrecognised request",
			},
		}
	})

	return router
}
