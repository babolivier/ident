package database

import "database/sql"

const ephemeralPublicKeysSchema = `
-- Stores public ephemeral keys
CREATE TABLE IF NOT EXISTS ephemeral_public_keys (
	ephemeral_public_key TEXT PRIMARY KEY
);
`

const insertEphemeralPublicKeySQL = `
	INSERT INTO ephemeral_public_keys (ephemeral_public_key)
	VALUES ($1)
`

const ephemeralEphemeralPublicKeyExistsSQL = `
	SELECT COUNT(ephemeral_public_key) FROM ephemeral_public_keys WHERE ephemeral_public_key = $1
`

type ephemeralPublicKeysStatements struct {
	insertEphemeralPublicKeyStmt *sql.Stmt
	ephemeralPublicKeyExistsStmt *sql.Stmt
}

func (s *ephemeralPublicKeysStatements) prepare(db *sql.DB) (err error) {
	_, err = db.Exec(ephemeralPublicKeysSchema)
	if err != nil {
		return
	}
	if s.insertEphemeralPublicKeyStmt, err = db.Prepare(insertEphemeralPublicKeySQL); err != nil {
		return
	}
	if s.ephemeralPublicKeyExistsStmt, err = db.Prepare(ephemeralEphemeralPublicKeyExistsSQL); err != nil {
		return
	}
	return

}

func (s *ephemeralPublicKeysStatements) insertEphemeralPublicKey(pubkey string) (err error) {
	_, err = s.insertEphemeralPublicKeyStmt.Exec(pubkey)
	return
}

func (s *ephemeralPublicKeysStatements) ephemeralPublicKeyExists(pubkey string) (exists bool, err error) {
	var count int
	row := s.ephemeralPublicKeyExistsStmt.QueryRow(pubkey)
	err = row.Scan(&count)
	return count != 0, err
}
