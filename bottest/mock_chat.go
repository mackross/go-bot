package bottest

import (
	"github.com/mackross/go-bot/chat"
	"testing"
)

func NewChat(t *testing.T) *Chat {
	c := &Chat{t, nil, nil, make([]chat.OutMsg, 0), make([]chat.OutMsg, 0), make(chan chat.InMsg, 0)}
	c.setupChans()
	return c
}

type Chat struct {
	t        *testing.T
	pms      chan chat.OutMsg
	rooms    chan chat.OutMsg
	expPMs   []chat.OutMsg
	expRooms []chat.OutMsg
	messages chan chat.InMsg
}

func (c *Chat) ExpectPM(m chat.OutMsg) {
	c.expPMs = append(c.expPMs, m)
}

func (c *Chat) ExpectRoomMsg(m chat.OutMsg) {
	c.expRooms = append(c.expRooms, m)
}

func (c *Chat) Messages() <-chan chat.InMsg {
	return c.messages
}

func (c *Chat) JoinRoom(s string) error {
	return nil
}

func (c *Chat) SetStatus(s string) error {
	return nil
}

func (c *Chat) NickName() string {
	return "botty"
}

func (c *Chat) Send(m chat.OutMsg) error {
	msg := c.expRooms[0]
	c.expRooms = c.expRooms[1:]
	equals(c.t, msg.Body, m.Body)
	equals(c.t, msg.To, m.To)
	if c.rooms != nil {
		go func() {
			c.rooms <- m
		}()
	}
	return nil
}

func (c *Chat) setupChans() {
	c.pms = make(chan chat.OutMsg, 0)
	c.rooms = make(chan chat.OutMsg, 0)
}

func (c *Chat) SendPM(m chat.OutMsg) error {
	msg := c.expPMs[0]
	c.expPMs = c.expPMs[1:]
	equals(c.t, msg.Body, m.Body)
	equals(c.t, msg.To, m.To)
	if c.pms != nil {
		go func() {
			c.pms <- m
		}()
	}
	return nil
}

func (c *Chat) blockUnlessPMSent() {
	<-c.pms
}

func (c *Chat) waitForMsgToSend() {
	<-c.rooms
}

func (m *Chat) Check() {
	assert(m.t, len(m.expPMs) == 0, "did not receive pms %+v", m.expPMs)
	assert(m.t, len(m.expRooms) == 0, "did not receive pms %+v", m.expRooms)
}
