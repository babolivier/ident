package database

import (
	"database/sql"
)

const invitesSchema = `
-- Stores public ephemeral keys
CREATE TABLE IF NOT EXISTS invites (
	token TEXT PRIMARY KEY,
	medium TEXT NOT NULL,
	address TEXT NOT NULL,
	room_id TEXT NOT NULL,
	sender TEXT NOT NULL,
	ephemeral_public_key TEXT NOT NULL
);
`

const insertInviteSQL = `
	INSERT INTO invites (token, medium, address, room_id, sender, ephemeral_public_key)
	VALUES ($1, $2, $3, $4, $5, $6)
`

const ephemeralPublicKeyExistsSQL = `
	SELECT COUNT(ephemeral_public_key) FROM invites WHERE ephemeral_public_key = $1
`

const selectInvitesForAddressAndMediumSQL = `
	SELECT medium, address, room_id, sender, token FROM invites
	WHERE medium = $1 AND address = $2
`

const deleteInvitesByAddressAndMediumSQL = `
	DELETE FROM invites WHERE medium = $1 AND address = $2
`

type invitesStatements struct {
	insertInviteStmt                     *sql.Stmt
	ephemeralPublicKeyExistsStmt         *sql.Stmt
	selectInvitesForAddressAndMediumStmt *sql.Stmt
	deleteInvitesByAddressAndMediumStmt  *sql.Stmt
}

func (s *invitesStatements) prepare(db *sql.DB) (err error) {
	_, err = db.Exec(invitesSchema)
	if err != nil {
		return
	}
	if s.insertInviteStmt, err = db.Prepare(insertInviteSQL); err != nil {
		return
	}
	if s.ephemeralPublicKeyExistsStmt, err = db.Prepare(ephemeralPublicKeyExistsSQL); err != nil {
		return
	}
	if s.selectInvitesForAddressAndMediumStmt, err = db.Prepare(selectInvitesForAddressAndMediumSQL); err != nil {
		return
	}
	if s.deleteInvitesByAddressAndMediumStmt, err = db.Prepare(deleteInvitesByAddressAndMediumSQL); err != nil {
		return
	}
	return

}

func (s *invitesStatements) insertInvite(
	token, medium, address, room_id, sender, ephemeral_public_key string,
) (err error) {
	_, err = s.insertInviteStmt.Exec(
		token, medium, address, room_id, sender, ephemeral_public_key,
	)
	return
}

func (s *invitesStatements) ephemeralPublicKeyExists(pubkey string) (exists bool, err error) {
	var count int
	row := s.ephemeralPublicKeyExistsStmt.QueryRow(pubkey)
	err = row.Scan(&count)
	return count != 0, err
}
