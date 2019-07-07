package pubkey

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"testing"

	"github.com/babolivier/ident/common"
	"github.com/babolivier/ident/common/config"
	"github.com/babolivier/ident/pubkey"
	"github.com/babolivier/ident/tests"
)

func TestGetPubKey(t *testing.T) {
	cfg, _, s, err := tests.InitTestRouting(pubkey.SetupRouting)
	tests.AssertEqual(t, err, nil)

	defer s.Close()

	realKeyID := cfg.Ident.SigningKey.Algo + ":" + cfg.Ident.SigningKey.ID
	testGetPubKey(t, s.URL, realKeyID, cfg, http.StatusOK)
	testGetPubKey(t, s.URL, "abcdef", cfg, http.StatusNotFound)
	testGetPubKey(t, s.URL, "abc:def", cfg, http.StatusNotFound)
}

func testGetPubKey(t *testing.T, serverURL, keyID string, cfg *config.Config, expectedCode int) {
	url := serverURL + path.Join(common.APIPrefix, "pubkey", keyID)

	resp, err := http.Get(url)
	tests.AssertEqual(t, err, nil)
	tests.AssertEqual(t, resp.StatusCode, expectedCode)

	if resp.StatusCode == http.StatusOK && resp.Body != nil {
		defer resp.Body.Close()

		b, err := ioutil.ReadAll(resp.Body)
		tests.AssertEqual(t, err, nil)

		var pubKeyResp pubkey.PublicKeyResponse
		err = json.Unmarshal(b, &pubKeyResp)
		tests.AssertEqual(t, err, nil)

		tests.AssertEqual(t, pubKeyResp.PublicKey, cfg.Ident.SigningKey.PubKeyBase64)
	}
}

func TestPubKeyIsValid(t *testing.T) {
	cfg, _, s, err := tests.InitTestRouting(pubkey.SetupRouting)
	tests.AssertEqual(t, err, nil)

	defer s.Close()

	realB64 := cfg.Ident.SigningKey.PubKeyBase64
	testPubKeyIsValid(t, s.URL, realB64, false, true)
	testPubKeyIsValid(t, s.URL, "abcdef", false, false)
}

func TestPubKeyEphemeralIsValid(t *testing.T) {
	_, db, s, err := tests.InitTestRouting(pubkey.SetupRouting)
	tests.AssertEqual(t, err, nil)

	defer s.Close()

	realPubKey := "somekey"
	err = db.Save3PIDInvite(
		"token", "email", "test@example.com", "!room:example.com",
		"@alice:example.com", realPubKey,
	)
	tests.AssertEqual(t, err, nil)

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

	u, err := url.Parse(serverURL + path.Join(common.APIPrefix, route))
	tests.AssertEqual(t, err, nil)

	query := u.Query()
	query.Add("public_key", b64)

	u.RawQuery = query.Encode()

	resp, err := http.Get(u.String())
	tests.AssertEqual(t, err, nil)
	tests.AssertEqual(t, resp.StatusCode, http.StatusOK)

	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	tests.AssertEqual(t, err, nil)

	var pubKeyValidResp pubkey.PublicKeyValidResponse
	err = json.Unmarshal(b, &pubKeyValidResp)
	tests.AssertEqual(t, err, nil)

	tests.AssertEqual(t, pubKeyValidResp.Valid, expected)
}
