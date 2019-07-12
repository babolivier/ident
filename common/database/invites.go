package database

import (
	"database/sql"

	"github.com/babolivier/ident/common/types"
)

const invitesSchema = `
-- Stores 3PID invites
CREATE TABLE IF NOT EXISTS invites (
	token TEXT PRIMARY KEY,
	medium TEXT NOT NULL,
	address TEXT NOT NULL,
	room_id TEXT NOT NULL,
	sender TEXT NOT NULL
);
`

const insertInviteSQL = `
	INSERT INTO invites (token, medium, address, room_id, sender)
	VALUES ($1, $2, $3, $4, $5)
`

const selectInviteFromTokenSQL = `
	SELECT medium, address, room_id, sender, token FROM invites
	WHERE token = $1
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
	selectInviteFromTokenStmt            *sql.Stmt
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
	if s.selectInviteFromTokenStmt, err = db.Prepare(selectInviteFromTokenSQL); err != nil {
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

func (s *invitesStatements) insertInvite(invite *types.ThreepidInvite) (err error) {
	_, err = s.insertInviteStmt.Exec(
		invite.Token, invite.Medium, invite.Address, invite.RoomID, invite.Sender,
	)
	return
}

func (s *invitesStatements) selectInviteByToken(token string) (*types.ThreepidInvite, error) {
	var invite types.ThreepidInvite

	row := s.selectInviteFromTokenStmt.QueryRow(token)
	err := row.Scan(&invite.Medium, &invite.Address, &invite.RoomID, &invite.Sender, &invite.Token)

	return &invite, err
}
