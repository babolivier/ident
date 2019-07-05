package database

import (
	"database/sql"

	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	db      *sql.DB
	invites invitesStatements
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

	return &Database{db, invites}, nil
}

func (d *Database) Save3PIDInvite(token, medium, address, room_id, sender, ephemeral_public_key string) error {
	return d.invites.insertInvite(token, medium, address, room_id, sender, ephemeral_public_key)
}

func (d *Database) EphemeralPublicKeyExists(pubkey string) (bool, error) {
	return d.invites.ephemeralPublicKeyExists(pubkey)
}
