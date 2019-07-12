package database

import (
	"database/sql"

	"github.com/babolivier/ident/common/types"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	db                  *sql.DB
	invites             invitesStatements
	ephemeralPublicKeys ephemeralPublicKeysStatements
}

func NewDatabase(driver string, connString string) (*Database, error) {
	var db *sql.DB
	var err error

	if db, err = sql.Open(driver, connString); err != nil {
		return nil, err
	}

	invites := invitesStatements{}
	if err = invites.prepare(db); err != nil {
		return nil, err
	}

	ephemeralPublicKeys := ephemeralPublicKeysStatements{}
	if err = ephemeralPublicKeys.prepare(db); err != nil {
		return nil, err
	}

	return &Database{db, invites, ephemeralPublicKeys}, nil
}

func (d *Database) Save3PIDInvite(invite *types.ThreepidInvite) error {
	return d.invites.insertInvite(invite)
}

func (d *Database) Get3PIDInviteByToken(token string) (*types.ThreepidInvite, error) {
	invite, err := d.invites.selectInviteByToken(token)

	// Don't return an error on empty result set, instead return a nil invite.
	if err == sql.ErrNoRows {
		invite = nil
		err = nil
	}

	return invite, err
}

func (d *Database) SaveEphemeralPublicKey(pubkey string) error {
	return d.ephemeralPublicKeys.insertEphemeralPublicKey(pubkey)
}

func (d *Database) EphemeralPublicKeyExists(pubkey string) (bool, error) {
	return d.ephemeralPublicKeys.ephemeralPublicKeyExists(pubkey)
}
