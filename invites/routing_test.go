package invites

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path"
	"strings"
	"testing"

	"github.com/babolivier/ident/common/config"
	"github.com/babolivier/ident/common/constants"
	"github.com/babolivier/ident/common/database"
	"github.com/babolivier/ident/common/testutils"
	"github.com/babolivier/ident/common/types"

	"github.com/matrix-org/gomatrix"
	"github.com/matrix-org/gomatrixserverlib"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ed25519"
)

// TODO: Add a test for "/store-invite". This requires a way to setup a mocked SMTP server (or a real one) in the CI,
//  otherwise the request will 500. Alternatively, we could keep the mail sending on invite optional and disable it
//  if no SMTP configuration is provided.

func TestSignED25519(t *testing.T) {
	testutils.TestWithTestServer(t, testSignED25519, SetupRouting)
}

func testSignED25519(t *testing.T, cfg *config.Config, db *database.Database, s *httptest.Server) {
	url := s.URL + path.Join(constants.APIPrefix, "sign-ed25519")
	contentType := "application/json"

	var respError gomatrix.RespError

	// Test that an empty request results in an error due to a missing MXID.
	req := make(map[string]interface{})

	resp, err := http.Post(url, contentType, structToIOReader(t, &req))
	require.Nil(t, err, err)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	httpRespToStruct(t, resp, &respError)
	require.Equal(t, "M_MISSING_PARAMS", respError.ErrCode)
	require.Equal(t, "Missing params: mxid", respError.Err)

	// Test that a request with only a MXID results in an error due to a missing token.
	req["mxid"] = "@alice:example.com"

	resp, err = http.Post(url, contentType, structToIOReader(t, &req))
	require.Nil(t, err, err)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	httpRespToStruct(t, resp, &respError)
	require.Equal(t, "M_MISSING_PARAMS", respError.ErrCode)
	require.Equal(t, "Missing params: token", respError.Err)

	// Test that a request with only a MXID and a token results in an error due to a missing private key.
	req["token"] = "RXmPapVujPyUjNWEvKzzctnIiMfZdctsfujPvXWypSlWFYkFnEewtdBCwzPjCwrOgTTOYXigmTvwzynnRChaSyvenHzEcYInatrpxMBbjUDMlaabMfXkrYIKcRmqxUYE"

	resp, err = http.Post(url, contentType, structToIOReader(t, &req))
	require.Nil(t, err, err)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	httpRespToStruct(t, resp, &respError)
	require.Equal(t, "M_MISSING_PARAMS", respError.ErrCode)
	require.Equal(t, "Missing params: private_key", respError.Err)

	// Test that a request with a private key of invalid length results in an error.
	req["private_key"] = "somekey"

	resp, err = http.Post(url, contentType, structToIOReader(t, &req))
	require.Nil(t, err, err)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)

	httpRespToStruct(t, resp, &respError)
	require.Equal(t, "M_INVALID_PARAM", respError.ErrCode)
	require.True(t, strings.HasPrefix(respError.Err, "Decoded the base64 representation of the private key"))

	// Test that a valid request results in a valid response containing a valid signature.
	sender := "@bob:example.com"
	err = db.Save3PIDInvite(&types.ThreepidInvite{
		Token:   req["token"].(string),
		Medium:  constants.MediumEmail,
		Address: "alice@example.com",
		RoomID:  "!someroom:example.com",
		Sender:  sender,
	})
	require.Nil(t, err, err)

	// base64 representation of an actual ed25519 private key generated with ed25519.NewKeyFromSeed
	req["private_key"] = "SG9oNmdlaTZnbzJHb2hwaGVpM3JlaXhvd3VvOHNob2ncvshRmehHQt+rOxXedcu2zX3MlupUtkPFQMVaJGBriA"

	var signED25519Resp SignED25519Resp
	resp, err = http.Post(url, contentType, structToIOReader(t, &req))
	require.Nil(t, err, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	b := httpRespToStruct(t, resp, &signED25519Resp)
	require.Equal(t, sender, signED25519Resp.Sender)
	require.Equal(t, req["mxid"], signED25519Resp.MXID)
	require.Equal(t, req["token"], signED25519Resp.Token)

	decodedPrivateKey, err := base64.RawStdEncoding.DecodeString(req["private_key"].(string))
	require.Nil(t, err, err)

	// If err is nil, then the signature is correct.
	err = gomatrixserverlib.VerifyJSON(
		cfg.Ident.ServerName,
		gomatrixserverlib.KeyID("ed25519:0"),
		ed25519.PrivateKey(decodedPrivateKey).Public().(ed25519.PublicKey),
		b,
	)
	require.Nil(t, err, err)
}

func structToIOReader(t *testing.T, req interface{}) io.Reader {
	jsonBytes, err := json.Marshal(req)
	require.Nil(t, err, err)

	return bytes.NewReader(jsonBytes)
}

func httpRespToStruct(t *testing.T, resp *http.Response, instance interface{}) []byte {
	require.NotNil(t, resp.Body)
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	require.Nil(t, err, err)

	err = json.Unmarshal(b, instance)
	require.Nil(t, err, err)

	return b
}
