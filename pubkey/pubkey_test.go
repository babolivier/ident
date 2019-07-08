package pubkey

import (
	"net/http"
	"testing"

	"github.com/babolivier/ident/common/config"
	"github.com/babolivier/ident/common/database"
	"github.com/babolivier/ident/common/testutils"

	"github.com/stretchr/testify/require"
)

func TestGetKey(t *testing.T) {
	cfg, err := testutils.NewTestConfig()
	require.Equal(t, err, nil)

	realKeyID := cfg.Ident.SigningKey.Algo + ":" + cfg.Ident.SigningKey.ID
	testGetKey(t, realKeyID, cfg, http.StatusOK)
	testGetKey(t, "abcdef", cfg, http.StatusNotFound)
	testGetKey(t, "abc:def", cfg, http.StatusNotFound)
}

func testGetKey(t *testing.T, keyID string, cfg *config.Config, expectedCode int) {
	resp := GetKey(keyID, cfg)

	require.Equal(t, resp.Code, expectedCode)

	if expectedCode == http.StatusOK {
		getKeyResp, ok := resp.JSON.(PublicKeyResponse)
		require.Equal(t, ok, true)
		require.Equal(t, getKeyResp.PublicKey, cfg.Ident.SigningKey.PubKeyBase64)
	}
}

func TestIsPubKeyValid(t *testing.T) {
	cfg, err := testutils.NewTestConfig()
	require.Equal(t, err, nil)

	testIsPubKeyValid(t, cfg.Ident.SigningKey.PubKeyBase64, cfg, true)
	testIsPubKeyValid(t, "abcdef", cfg, false)
}

func testIsPubKeyValid(t *testing.T, b64 string, cfg *config.Config, expected bool) {
	resp := IsPubKeyValid(b64, cfg)

	require.Equal(t, resp.Code, http.StatusOK)

	keyValidResp, ok := resp.JSON.(PublicKeyValidResponse)
	require.Equal(t, ok, true)
	require.Equal(t, keyValidResp.Valid, expected)
}

func TestIsEphemeralPubKeyValid(t *testing.T) {
	cfg, err := testutils.NewTestConfig()
	require.Equal(t, err, nil)
	db, err := database.NewDatabase(cfg.Database.Driver, cfg.Database.ConnString)
	require.Equal(t, err, nil)

	realPubKey := "somekey"
	err = db.Save3PIDInvite(
		"token", "email", "test@example.com", "!room:example.com",
		"@alice:example.com", realPubKey,
	)
	require.Equal(t, err, nil)

	testIsEphemeralPubKeyValid(t, realPubKey, db, true)
	testIsEphemeralPubKeyValid(t, "abcdef", db, false)
}

func testIsEphemeralPubKeyValid(t *testing.T, b64 string, db *database.Database, expected bool) {
	resp := IsEphemeralPubKeyValid(b64, db)

	require.Equal(t, resp.Code, http.StatusOK)

	keyValidResp, ok := resp.JSON.(PublicKeyValidResponse)
	require.Equal(t, ok, true)
	require.Equal(t, keyValidResp.Valid, expected)
}
