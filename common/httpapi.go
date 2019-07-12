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
	logrus.WithError(err).Error("An error happened")

	return util.JSONResponse{
		Code: 500,
		JSON: gomatrix.RespError{
			ErrCode: "M_UNKNOWN",
			Err:     "Internal server error",
		},
	}
}

func MissingParamsError(paramName string) util.JSONResponse {
	return util.JSONResponse{
		Code: 400,
		JSON: gomatrix.RespError{
			ErrCode: "M_MISSING_PARAMS",
			Err:     "Missing params: " + paramName,
		},
	}
}

func InvalidParamError(errmsg string) util.JSONResponse {
	return util.JSONResponse{
		Code: 400,
		JSON: gomatrix.RespError{
			ErrCode: "M_INVALID_PARAM",
			Err:     errmsg,
		},
	}
}
