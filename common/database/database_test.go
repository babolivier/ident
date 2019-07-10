package database

import (
	"testing"

	"github.com/babolivier/ident/common/constants"
	"github.com/babolivier/ident/common/types"

	"github.com/stretchr/testify/require"
)

func TestInsertInvite(t *testing.T) {
	db, err := NewDatabase("sqlite3", ":memory:")
	require.Nil(t, err, err)

	in := &types.ThreepidInvite{
		Token:   "sometoken",
		Medium:  constants.MediumEmail,
		Address: "alice@example.com",
		RoomID:  "!someroom:example.com",
		Sender:  "@bob:example.com",
	}

	err = db.Save3PIDInvite(in)
	require.Nil(t, err, err)

	out, err := db.invites.selectInviteFromToken(in.Token)
	require.Nil(t, err, err)

	require.Equal(t, in.Token, out.Token)
	require.Equal(t, in.Medium, out.Medium)
	require.Equal(t, in.Address, out.Address)
	require.Equal(t, in.RoomID, out.RoomID)
	require.Equal(t, in.Sender, out.Sender)
}

func TestSaveEphemeralPublicKey(t *testing.T) {
	db, err := NewDatabase("sqlite3", ":memory:")
	require.Nil(t, err, err)

	key := "abcdef"

	err = db.SaveEphemeralPublicKey(key)
	require.Nil(t, err, err)

	exists, err := db.EphemeralPublicKeyExists(key)
	require.Nil(t, err, err)

	require.True(t, exists)
}
