package pubkey

import (
	"strings"

	"github.com/babolivier/ident/common/config"
	"github.com/babolivier/ident/common/database"

	"github.com/matrix-org/gomatrix"
	"github.com/matrix-org/util"
	"github.com/sirupsen/logrus"
)

type PublicKeyResponse struct {
	PublicKey string `json:"public_key"`
}

type PublicKeyValidResponse struct {
	Valid bool `json:"valid"`
}

func GetKey(keyID string, cfg *config.Config) util.JSONResponse {
	notFoundJSONResponse := util.JSONResponse{
		Code: 404,
		JSON: gomatrix.RespError{
			ErrCode: "M_NOT_FOUND",
			Err:     "The public key was not found",
		},
	}

	split := strings.SplitN(keyID, ":", 2)

	// If the key ID isn't in the format algo:id, then we don't know it.
	if len(split) != 2 {
		return notFoundJSONResponse
	}

	// Check if the key's metadata matches with our signing key.
	if split[0] != cfg.Ident.SigningKey.Algo || split[1] != cfg.Ident.SigningKey.ID {
		return notFoundJSONResponse
	}

	return util.JSONResponse{
		Code: 200,
		JSON: PublicKeyResponse{
			PublicKey: string(cfg.Ident.SigningKey.PubKeyBase64),
		},
	}
}

func IsPubKeyValid(keyBase64 string, cfg *config.Config) util.JSONResponse {
	return util.JSONResponse{
		Code: 200,
		JSON: PublicKeyValidResponse{
			Valid: keyBase64 == cfg.Ident.SigningKey.PubKeyBase64,
		},
	}
}

func IsEphemeralPubKeyValid(keyBase64 string, db *database.Database) util.JSONResponse {
	exists, err := db.EphemeralPublicKeyExists(keyBase64)

	if err != nil {
		logrus.WithError(err).Error("Error trying to check the existence of an ephemeral key")
		return util.JSONResponse{
			Code: 500,
			JSON: gomatrix.RespError{
				ErrCode: "M_UNKNOWN",
			},
		}
	}

	return util.JSONResponse{
		Code: 200,
		JSON: PublicKeyValidResponse{
			Valid: exists,
		},
	}
}
