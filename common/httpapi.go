package common

import (
	"net/http"

	"github.com/matrix-org/gomatrix"
	"github.com/matrix-org/util"
	"github.com/sirupsen/logrus"
)

func MakeAPI(f func(r *http.Request) util.JSONResponse) http.Handler {
	h := util.MakeJSONAPI(util.NewJSONRequestHandler(f))
	h = util.WithCORSOptions(h)
	return h
}

func InternalServerError(err error) util.JSONResponse {
	logrus.WithError(err).Error("An error happened when processing the request")

	return util.JSONResponse{
		Code: 500,
		JSON: gomatrix.RespError{
			ErrCode: "M_UNKNOWN",
			Err:     "Internal server error",
		},
	}
}
