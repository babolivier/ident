package common

import (
	"net/http"

	"github.com/matrix-org/util"
)

const APIPrefix = "/_matrix/identity/api/v1"

func MakeAPI(f func(r *http.Request) util.JSONResponse) http.Handler {
	h := util.MakeJSONAPI(util.NewJSONRequestHandler(f))
	h = util.WithCORSOptions(h)
	return h
}
