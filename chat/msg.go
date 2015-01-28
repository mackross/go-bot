package chat

import (
	"time"
)

type InMsg struct {
	ID        string
	RoomID    *string // when nil it is a private message
	From      string
	Body      string
	ArrivedAt time.Time
	SentAt    time.Time
}

func (m *InMsg) IsPM() bool {
	return m.RoomID == nil
}

type OutMsg struct {
	To     string // roomID or ChatUserID
	Body   string
	Color  *string
	Notify *bool
	HTML   *bool
}

type Network interface {
	NickName() string
	SendPM(m OutMsg) error
	Send(m OutMsg) error
	JoinRoom(id string) error
	SetStatus(s string) error
	Messages() <-chan InMsg
}
