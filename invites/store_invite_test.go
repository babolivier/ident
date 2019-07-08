package invites

import (
	"path"
	"strings"
	"testing"

	"github.com/babolivier/ident/common/constants"
	"github.com/babolivier/ident/common/testutils"

	"github.com/matrix-org/gomatrix"
	"github.com/stretchr/testify/require"
)

func TestCheckReqValid(t *testing.T) {
	req := &StoreInviteReq{
		Medium:  constants.MediumEmail,
		Address: "test@example.com",
		RoomID:  "!someroom:example.com",
		Sender:  "@alice:example.com",
	}

	resp := checkReq(req)
	require.Nil(t, resp)
}

func TestCheckReqUnsupportedMedium(t *testing.T) {
	req := &StoreInviteReq{
		Medium:  constants.MediumMSISDN, // TODO: Change this when MSISDN is supported.
		Address: "test@example.com",
		RoomID:  "!someroom:example.com",
		Sender:  "@alice:example.com",
	}

	resp := checkReq(req)
	require.NotNil(t, resp)
	require.Equal(t, "M_INVALID_PARAMS", resp.JSON.(gomatrix.RespError).ErrCode)
	require.True(t, strings.HasSuffix(resp.JSON.(gomatrix.RespError).Err, constants.MediumMSISDN))
}

func TestCheckReqBadEmail(t *testing.T) {
	req := &StoreInviteReq{
		Medium:  constants.MediumEmail,
		Address: "testexample.com",
		RoomID:  "!someroom:example.com",
		Sender:  "@alice:example.com",
	}

	resp := checkReq(req)
	require.NotNil(t, resp)
	require.Equal(t, "M_INVALID_EMAIL", resp.JSON.(gomatrix.RespError).ErrCode)
	require.Equal(t, "Invalid email address", resp.JSON.(gomatrix.RespError).Err)

	req.Address = "test@example.com@otherdomain.com"
	resp = checkReq(req)
	require.NotNil(t, resp)
	require.Equal(t, "M_INVALID_EMAIL", resp.JSON.(gomatrix.RespError).ErrCode)
	require.Equal(t, "Invalid email address", resp.JSON.(gomatrix.RespError).Err)
}

func TestCheckReqBadRoomID(t *testing.T) {
	req := &StoreInviteReq{
		Medium:  constants.MediumEmail,
		Address: "test@example.com",
		RoomID:  "someroom:example.com",
		Sender:  "@alice:example.com",
	}

	resp := checkReq(req)
	require.NotNil(t, resp)
	require.Equal(t, "M_INVALID_PARAMS", resp.JSON.(gomatrix.RespError).ErrCode)
	require.Equal(t, "Invalid room ID", resp.JSON.(gomatrix.RespError).Err)

	req.RoomID = "!someroomexample.com"
	resp = checkReq(req)
	require.NotNil(t, resp)
	require.Equal(t, "M_INVALID_PARAMS", resp.JSON.(gomatrix.RespError).ErrCode)
	require.Equal(t, "Invalid room ID", resp.JSON.(gomatrix.RespError).Err)
}

func TestCheckReqBadSender(t *testing.T) {
	req := StoreInviteReq{
		Medium:  constants.MediumEmail,
		Address: "test@example.com",
		RoomID:  "!someroom:example.com",
		Sender:  "alice:example.com",
	}

	resp := checkReq(&req)
	require.NotNil(t, resp)
	require.Equal(t, "M_INVALID_PARAMS", resp.JSON.(gomatrix.RespError).ErrCode)
	require.Equal(t, "Invalid sender ID", resp.JSON.(gomatrix.RespError).Err)

	req.Sender = "@aliceexample.com"
	resp = checkReq(&req)
	require.NotNil(t, resp)
	require.Equal(t, "M_INVALID_PARAMS", resp.JSON.(gomatrix.RespError).ErrCode)
	require.Equal(t, "Invalid sender ID", resp.JSON.(gomatrix.RespError).Err)
}

func TestIsEmailAddressValid(t *testing.T) {
	require.True(t, isEmailAddressValid("test@example.com"))
	require.False(t, isEmailAddressValid("testexample.com"))
	require.False(t, isEmailAddressValid("test@example.com@otherdomain.com"))
}

func TestGetResp(t *testing.T) {
	cfg := testutils.NewTestConfig()

	req := &StoreInviteReq{
		Token:   "sometoken",
		Address: "alice@example.com",
	}

	pubKey := "somekey"

	resp := getResp(req, cfg, pubKey)
	require.NotNil(t, resp)
	require.Equal(t, req.Token, resp.Token)
	require.Equal(t, cfg.Ident.SigningKey.PubKeyBase64, resp.PublicKey)
	require.Len(t, resp.PublicKeys, 2)
	require.Equal(t, resp.PublicKeys[0].PublicKey, cfg.Ident.SigningKey.PubKeyBase64)
	require.Equal(t, resp.PublicKeys[0].KeyValidityURL, cfg.Ident.BaseURL+path.Join(constants.APIPrefix, "pubkey/isvalid"))
	require.Equal(t, resp.PublicKeys[1].PublicKey, pubKey)
	require.Equal(t, resp.PublicKeys[1].KeyValidityURL, cfg.Ident.BaseURL+path.Join(constants.APIPrefix, "pubkey/ephemeral/isvalid"))
	require.Equal(t, "a...@e...", resp.DisplayName)
}

func TestRedactEmail(t *testing.T) {
	require.Equal(t, "a...@e...", redactEmail("alice@example.com"))
	// We don't really care about the result here, just that it doesn't panic.
	require.Equal(t, "a...", redactEmail("aliceexample.com"))
	require.Equal(t, "a...@e...", redactEmail("alice@example.com@otherdomain.com"))
}
