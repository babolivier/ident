package pubkey

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"testing"

	"github.com/babolivier/ident/common/config"
	"github.com/babolivier/ident/common/constants"
	"github.com/babolivier/ident/common/testutils"

	"github.com/stretchr/testify/require"
)

func TestGetPubKey(t *testing.T) {
	cfg, _, s, err := testutils.InitTestRouting(t, SetupRouting)
	require.Nil(t, err, err)

	defer s.Close()

	realKeyID := cfg.Ident.SigningKey.Algo + ":" + cfg.Ident.SigningKey.ID
	testGetPubKey(t, s.URL, realKeyID, cfg, http.StatusOK)
	testGetPubKey(t, s.URL, "abcdef", cfg, http.StatusNotFound)
	testGetPubKey(t, s.URL, "abc:def", cfg, http.StatusNotFound)
}

func testGetPubKey(t *testing.T, serverURL, keyID string, cfg *config.Config, expectedCode int) {
	u := serverURL + path.Join(constants.APIPrefix, "pubkey", keyID)

	resp, err := http.Get(u)
	require.Nil(t, err, err)
	require.Equal(t, expectedCode, resp.StatusCode)

	if resp.StatusCode == http.StatusOK {
		require.NotNil(t, resp.Body)

		defer resp.Body.Close()

		b, err := ioutil.ReadAll(resp.Body)
		require.Nil(t, err, err)

		var pubKeyResp PublicKeyResponse
		err = json.Unmarshal(b, &pubKeyResp)
		require.Nil(t, err, err)

		require.Equal(t, cfg.Ident.SigningKey.PubKeyBase64, pubKeyResp.PublicKey)
	}
}

func TestPubKeyIsValid(t *testing.T) {
	cfg, _, s, err := testutils.InitTestRouting(t, SetupRouting)
	require.Nil(t, err, err)

	defer s.Close()

	realB64 := cfg.Ident.SigningKey.PubKeyBase64
	testPubKeyIsValid(t, s.URL, realB64, false, true)
	testPubKeyIsValid(t, s.URL, "abcdef", false, false)
}

func TestPubKeyEphemeralIsValid(t *testing.T) {
	_, db, s, err := testutils.InitTestRouting(t, SetupRouting)
	require.Nil(t, err, err)

	defer s.Close()

	realPubKey := "somekey"
	err = db.SaveEphemeralPublicKey(realPubKey)
	require.Nil(t, err, err)

	testPubKeyIsValid(t, s.URL, realPubKey, true, true)
	testPubKeyIsValid(t, s.URL, "abcdef", true, false)
}

func testPubKeyIsValid(t *testing.T, serverURL, b64 string, ephemeral, expected bool) {
	var route string
	if ephemeral {
		route = "pubkey/ephemeral/isvalid"
	} else {
		route = "pubkey/isvalid"
	}

	u, err := url.Parse(serverURL + path.Join(constants.APIPrefix, route))
	require.Nil(t, err, err)

	query := u.Query()
	query.Add("public_key", b64)

	u.RawQuery = query.Encode()

	resp, err := http.Get(u.String())
	require.Nil(t, err, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	require.Nil(t, err, err)

	var pubKeyValidResp PublicKeyValidResponse
	err = json.Unmarshal(b, &pubKeyValidResp)
	require.Nil(t, err, err)

	require.Equal(t, expected, pubKeyValidResp.Valid)
}
