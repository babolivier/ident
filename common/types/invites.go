package types

type ThreepidInvite struct {
	Medium  string `json:"medium"`
	Address string `json:"address"`
	RoomID  string `json:"room_id"`
	Sender  string `json:"sender"`
	Token   string
}
