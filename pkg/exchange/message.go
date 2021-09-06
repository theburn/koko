package exchange

import "encoding/json"

type RoomMessage struct {
	Event string `json:"event"`
	Body  []byte `json:"data"`

	Meta MetaInfo `json:"meta"`
}

type MetaInfo struct {
	UserId   string `json:"user_id"`
	Username string `json:"username"`
}

func (m RoomMessage) Marshal() []byte {
	p, _ := json.Marshal(m)
	return p
}

func (m RoomMessage) UnMarshal(p interface{}) {
	_ = json.Unmarshal(m.Body, p)
}

const (
	PingEvent    = "Ping"
	DataEvent    = "Data"
	WindowsEvent = "Windows"

	JoinEvent  = "Join"
	LeaveEvent = "Leave"

	ExitEvent = "Exit"

	JoinSuccessEvent = "JoinSuccess"

	ShareTyping = "Share_TYPING"
	ShareJoin   = "Share_JOIN"
	ShareLeave  = "Share_LEAVE"
)
