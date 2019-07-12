package invites

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/babolivier/ident/common"
	"github.com/babolivier/ident/common/config"
	"github.com/babolivier/ident/common/database"

	"github.com/matrix-org/gomatrix"
	"github.com/matrix-org/gomatrixserverlib"
	"github.com/matrix-org/util"
	"golang.org/x/crypto/ed25519"
)

type SignED25519Req struct {
	MXID       string                         `json:"mxid"`
	Token      string                         `json:"token"`
	PrivateKey gomatrixserverlib.Base64String `json:"private_key"`
}

type SignED25519Resp struct {
	MXID       string      `json:"mxid"`
	Sender     string      `json:"sender"`
	Token      string      `json:"token"`
	Signatures interface{} `json:"signatures,omitempty"`
}

func SignED25519(r *http.Request, cfg *config.Config, db *database.Database) util.JSONResponse {
	// Check if we have a request body.
	if r.Body == nil {
		return util.JSONResponse{
			Code: 400,
			JSON: gomatrix.RespError{
				ErrCode: "M_MISSING_PARAMS",
				Err:     "Missing request body",
			},
		}
	}

	defer r.Body.Close()

	// Load the body's JSON into an instance of SignED25519Req.
	var req SignED25519Req
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return common.InternalServerError(err)
	}

	if checkResp := checkSignED25519Req(&req); checkResp != nil {
		return *checkResp
	}

	// Query the database for the invite and check if it returned with a non-nil invite.
	invite, err := db.Get3PIDInviteByToken(req.Token)
	if err != nil {
		return common.InternalServerError(err)
	}

	if invite == nil {
		return util.JSONResponse{
			Code: 404,
			JSON: gomatrix.RespError{
				ErrCode: "M_UNRECOGNIZED",
				Err:     "Unrecognised token",
			},
		}
	}

	// Sign the data.
	resp := SignED25519Resp{
		MXID:   req.MXID,
		Sender: invite.Sender,
		Token:  invite.Token,
	}

	unsignedRespBytes, err := json.Marshal(&resp)
	if err != nil {
		return common.InternalServerError(err)
	}

	// Using ed25519:0 as the key ID here isn't part of the spec (yet), however discussion in #matrix-spec concluded
	// that the ID used here is of little importance, that the implementation is free to use whichever it wants, and
	// that "ed25519:0" is a good default value.
	signedRespBytes, err := gomatrixserverlib.SignJSON(
		cfg.Ident.ServerName,
		gomatrixserverlib.KeyID("ed25519:0"),
		ed25519.PrivateKey(req.PrivateKey),
		unsignedRespBytes,
	)
	if err != nil {
		return common.InternalServerError(err)
	}

	// Unmarshal the bytes containing the signature into the response. Not the best thing performance-wise,
	// but apparently giving bytes to the JSON handler via the JSONResponse results in it trying to respond with
	// a base64 representation of these bytes. At least we keep it as memory efficient as possible by reusing the
	// response instance we created before signing.
	err = json.Unmarshal(signedRespBytes, &resp)
	if err != nil {
		return common.InternalServerError(err)
	}

	// Return the signed data.
	return util.JSONResponse{
		Code: 200,
		JSON: resp,
	}
}

func checkSignED25519Req(req *SignED25519Req) *util.JSONResponse {
	var resp util.JSONResponse

	if len(req.MXID) == 0 {
		resp = common.MissingParamsError("mxid")
		return &resp
	}

	if len(req.Token) == 0 {
		resp = common.MissingParamsError("token")
		return &resp
	}

	if len(req.PrivateKey) == 0 {
		resp = common.MissingParamsError("private_key")
		return &resp
	}

	if len(req.PrivateKey) != ed25519.PrivateKeySize {
		resp = common.InvalidParamError(fmt.Sprintf(
			"Decoded the base64 representation of the private key into %d bytes, expected %d",
			len(req.PrivateKey), ed25519.PrivateKeySize,
		))
		return &resp
	}

	return nil
}
