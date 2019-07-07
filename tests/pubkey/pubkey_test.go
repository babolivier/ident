package pubkey

import (
	"net/http"
	"testing"

	"github.com/babolivier/ident/common/config"
	"github.com/babolivier/ident/common/database"
	"github.com/babolivier/ident/pubkey"
	"github.com/babolivier/ident/tests"
)

func TestGetKey(t *testing.T) {
	cfg, err := tests.NewTestConfig()
	tests.AssertEqual(t, err, nil)

	realKeyID := cfg.Ident.SigningKey.Algo + ":" + cfg.Ident.SigningKey.ID
	testGetKey(t, realKeyID, cfg, http.StatusOK)
	testGetKey(t, "abcdef", cfg, http.StatusNotFound)
	testGetKey(t, "abc:def", cfg, http.StatusNotFound)
}

func testGetKey(t *testing.T, keyID string, cfg *config.Config, expectedCode int) {
	resp := pubkey.GetKey(keyID, cfg)

	tests.AssertEqual(t, resp.Code, expectedCode)

	if expectedCode == http.StatusOK {
		getKeyResp, ok := resp.JSON.(pubkey.PublicKeyResponse)
		tests.AssertEqual(t, ok, true)
		tests.AssertEqual(t, getKeyResp.PublicKey, cfg.Ident.SigningKey.PubKeyBase64)
	}
}

func TestIsPubKeyValid(t *testing.T) {
	cfg, err := tests.NewTestConfig()
	tests.AssertEqual(t, err, nil)

	testIsPubKeyValid(t, cfg.Ident.SigningKey.PubKeyBase64, cfg, true)
	testIsPubKeyValid(t, "abcdef", cfg, false)
}

func testIsPubKeyValid(t *testing.T, b64 string, cfg *config.Config, expected bool) {
	resp := pubkey.IsPubKeyValid(b64, cfg)

	tests.AssertEqual(t, resp.Code, http.StatusOK)

	keyValidResp, ok := resp.JSON.(pubkey.PublicKeyValidResponse)
	tests.AssertEqual(t, ok, true)
	tests.AssertEqual(t, keyValidResp.Valid, expected)
}

func TestIsEphemeralPubKeyValid(t *testing.T) {
	cfg, err := tests.NewTestConfig()
	tests.AssertEqual(t, err, nil)
	db, err := database.NewDatabase(cfg.Database.Driver, cfg.Database.ConnString)
	tests.AssertEqual(t, err, nil)

	realPubKey := "somekey"
	err = db.Save3PIDInvite(
		"token", "email", "test@example.com", "!room:example.com",
		"@alice:example.com", realPubKey,
	)
	tests.AssertEqual(t, err, nil)

	testIsEphemeralPubKeyValid(t, realPubKey, db, true)
	testIsEphemeralPubKeyValid(t, "abcdef", db, false)
}

func testIsEphemeralPubKeyValid(t *testing.T, b64 string, db *database.Database, expected bool) {
	resp := pubkey.IsEphemeralPubKeyValid(b64, db)

	tests.AssertEqual(t, resp.Code, http.StatusOK)

	keyValidResp, ok := resp.JSON.(pubkey.PublicKeyValidResponse)
	tests.AssertEqual(t, ok, true)
	tests.AssertEqual(t, keyValidResp.Valid, expected)
}
