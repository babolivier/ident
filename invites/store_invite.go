package invites

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/babolivier/ident/common"
	"github.com/babolivier/ident/common/config"
	"github.com/babolivier/ident/common/constants"
	"github.com/babolivier/ident/common/database"
	"github.com/babolivier/ident/common/email"
	"github.com/babolivier/ident/common/types"

	"github.com/matrix-org/gomatrix"
	"github.com/matrix-org/gomatrixserverlib"
	"github.com/matrix-org/util"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ed25519"
)

type StoreInviteReq struct {
	types.ThreepidInvite
	RoomAlias         string `json:"room_alias"`
	RoomAvatarURL     string `json:"room_avatar_url"`
	RoomJoinRules     string `json:"room_join_rules"`
	RoomName          string `json:"room_name"`
	SenderDisplayName string `json:"sender_display_name"`
	SenderAvatarURL   string `json:"sender_avatar_url"`
	PrivKeyBase64     string
	BaseURL           string
}

type StoreInviteResp struct {
	Token       string      `json:"token"`
	PublicKey   string      `json:"public_key"`
	PublicKeys  []PublicKey `json:"public_keys"`
	DisplayName string      `json:"display_name"`
}

type PublicKey struct {
	PublicKey      string `json:"public_key"`
	KeyValidityURL string `json:"key_validity_url"`
}

func StoreInvite(r *http.Request, cfg *config.Config, db *database.Database) util.JSONResponse {
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

	// Load the body's JSON into an instance of StoreInviteReq.
	// Sydent supports both the `application/json` and `application/x-www-form-urlencoded` content-types,
	// but that's mainly due to an implementation bug in Synapse: https://github.com/matrix-org/synapse/issues/5634
	// Let's just follow the spec here.
	var req StoreInviteReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return common.InternalServerError(err)
	}

	// Check that the request params are valid.
	if resp := checkStoreInviteReq(&req); resp != nil {
		return *resp
	}

	// TODO: Check if there's an MXID associated with this 3PID and return here with it if so.

	// Generate the ephemeral key.
	pubKey, privKey, err := ed25519.GenerateKey(rand.New(rand.NewSource(time.Now().Unix())))
	if err != nil {
		return common.InternalServerError(err)
	}

	// Add additional info to the request instance (will be used when processing the templates)
	req.PrivKeyBase64 = base64.RawStdEncoding.EncodeToString(privKey)
	req.BaseURL = cfg.Ident.BaseURL
	req.Token = common.RandString(128)

	// Send the invite email.
	if err = email.SendMail(
		cfg, req.Address, cfg.Ident.Invites.EmailTemplate.Text, cfg.Ident.Invites.EmailTemplate.HTML, &req,
	); err != nil {
		// Log the error as the mail sending process is a bit more complex.
		logrus.WithError(err).Error("Couldn't send 3PID invite email")
		return common.InternalServerError(err)
	}

	// Save the data about the invite in the database.
	if err = db.Save3PIDInvite(&req.ThreepidInvite); err != nil {
		return common.InternalServerError(err)
	}

	// Encode the public key into base 64 to save it in the database and send it to the client.
	pubKeyBase64 := base64.RawStdEncoding.EncodeToString(pubKey)

	// Save the data about the public key in the database.
	if err = db.SaveEphemeralPublicKey(pubKeyBase64); err != nil {
		return common.InternalServerError(err)
	}

	// Send the invite data to the client.
	return util.JSONResponse{
		Code: 200,
		JSON: getStoreInviteResp(&req, cfg, pubKeyBase64),
	}
}

func checkStoreInviteReq(req *StoreInviteReq) *util.JSONResponse {
	var resp util.JSONResponse

	// Check if we support this medium.
	// TODO: Implement MSISDN.
	if req.Medium != constants.MediumEmail {
		resp = common.InvalidParamError("Unsupported medium: " + req.Medium)
		return &resp
	}

	// Check if the email address is valid.
	if req.Medium == constants.MediumEmail && !isEmailAddressValid(req.Address) {
		resp = util.JSONResponse{
			Code: 400,
			JSON: gomatrix.RespError{
				ErrCode: "M_INVALID_EMAIL",
				Err:     "Invalid email address",
			},
		}
		return &resp
	}

	if _, _, err := gomatrixserverlib.SplitID('!', req.RoomID); err != nil {
		// Check if the room ID is valid.
		resp = common.InvalidParamError("Invalid room ID")
		return &resp
	}

	// Check if the sender's user ID is valid.
	if _, _, err := gomatrixserverlib.SplitID('@', req.Sender); err != nil {
		resp = common.InvalidParamError("Invalid sender ID")
		return &resp
	}

	return nil
}

func isEmailAddressValid(email string) bool {
	var atCount int
	atCount = strings.Count(email, "@")

	// Prevent username@domain1@domain2
	// c.f. https://matrix.org/blog/2019/04/18/security-update-sydent-1-0-2
	// Also ensure that there's a localpart and a server name.
	return atCount == 1 && len(email) >= 3
}

func getStoreInviteResp(req *StoreInviteReq, cfg *config.Config, pubKeyBase64 string) *StoreInviteResp {
	// Instantiate a response.
	resp := StoreInviteResp{
		Token:       req.Token,
		PublicKey:   cfg.Ident.SigningKey.PubKeyBase64,
		PublicKeys:  make([]PublicKey, 2),
		DisplayName: redactEmail(req.Address),
	}

	// Add the public key's details.
	resp.PublicKeys[0] = PublicKey{
		PublicKey:      cfg.Ident.SigningKey.PubKeyBase64,
		KeyValidityURL: cfg.Ident.BaseURL + path.Join(constants.APIPrefix, "pubkey/isvalid"),
	}
	resp.PublicKeys[1] = PublicKey{
		PublicKey:      pubKeyBase64,
		KeyValidityURL: cfg.Ident.BaseURL + path.Join(constants.APIPrefix, "pubkey/ephemeral/isvalid"),
	}

	return &resp
}

func redactEmail(email string) string {
	split := strings.SplitN(email, "@", 2)

	if len(split) < 2 {
		return fmt.Sprintf("%c...", split[0][0])
	}

	return fmt.Sprintf("%c...@%c...", split[0][0], split[1][0])
}
