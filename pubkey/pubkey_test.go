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
	cfg := testutils.NewTestConfig(t)

	realKeyID := cfg.Ident.SigningKey.Algo + ":" + cfg.Ident.SigningKey.ID
	testGetKey(t, realKeyID, cfg, http.StatusOK)
	testGetKey(t, "abcdef", cfg, http.StatusNotFound)
	testGetKey(t, "abc:def", cfg, http.StatusNotFound)
}

func testGetKey(t *testing.T, keyID string, cfg *config.Config, expectedCode int) {
	resp := GetKey(keyID, cfg)

	require.Equal(t, expectedCode, resp.Code)

	if expectedCode == http.StatusOK {
		getKeyResp, ok := resp.JSON.(PublicKeyResponse)
		require.True(t, ok)
		require.Equal(t, cfg.Ident.SigningKey.PubKeyBase64, getKeyResp.PublicKey)
	}
}

func TestIsPubKeyValid(t *testing.T) {
	cfg := testutils.NewTestConfig(t)

	testIsPubKeyValid(t, cfg.Ident.SigningKey.PubKeyBase64, cfg, true)
	testIsPubKeyValid(t, "abcdef", cfg, false)
}

func testIsPubKeyValid(t *testing.T, b64 string, cfg *config.Config, expected bool) {
	resp := IsPubKeyValid(b64, cfg)

	require.Equal(t, http.StatusOK, resp.Code)

	keyValidResp, ok := resp.JSON.(PublicKeyValidResponse)
	require.True(t, ok)
	require.Equal(t, expected, keyValidResp.Valid)
}

func TestIsEphemeralPubKeyValid(t *testing.T) {
	db := testutils.NewTestDB(t)

	realPubKey := "somekey"
	err := db.SaveEphemeralPublicKey(realPubKey)
	require.Nil(t, err, err)

	testIsEphemeralPubKeyValid(t, realPubKey, db, true)
	testIsEphemeralPubKeyValid(t, "abcdef", db, false)
}

func testIsEphemeralPubKeyValid(t *testing.T, b64 string, db *database.Database, expected bool) {
	resp := IsEphemeralPubKeyValid(b64, db)

	require.Equal(t, http.StatusOK, resp.Code)

	keyValidResp, ok := resp.JSON.(PublicKeyValidResponse)
	require.True(t, ok)
	require.Equal(t, expected, keyValidResp.Valid)
}
