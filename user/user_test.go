package user

import (
	"errors"
	"github.com/mackross/gobot"
	"github.com/mackross/gobot/bottest"
	"github.com/mackross/gobot/chat"
	"testing"
)

type mockRepo map[string]User

func newMockRepo() mockRepo {
	return mockRepo(make(map[string]User, 0))
}

func (m mockRepo) ListUsers() ([]User, error) {
	users := make([]User, 0, len(m))
	for _, u := range m {
		users = append(users, u)
	}
	return users, nil
}

func (m mockRepo) UserForID(id string) (*User, error) {
	u, ok := m[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return &u, nil
}

func (m mockRepo) SaveUser(u User) error {
	m[u.ID] = u
	return nil
}

func mockBot(t *testing.T) (*bot.Bot, *bottest.Chat) {
	c := bottest.NewChat(t)
	b := bot.NewBot(c)
	return b, c
}

func TestThatAUserWhoSpeaksIsAdded(t *testing.T) {
	b, c := mockBot(t)
	_, _ = b, c
	repo := newMockRepo()
	SetRepo(repo)

	b.AddRootHandler(NewRootHandler())

	b.HandleMessage(chat.InMsg{From: "123"})
	equals(t, 1, len(repo))
	b.HandleMessage(chat.InMsg{From: "1234"})
	equals(t, 2, len(repo))
	b.HandleMessage(chat.InMsg{From: "1234"})
	equals(t, 2, len(repo))
}

func TestThatOneUserCanBecomeAdminThroughPM(t *testing.T) {
	b, c := mockBot(t)
	_, _ = b, c
	repo := newMockRepo()
	SetRepo(repo)

	b.AddRootHandler(NewRootHandler())

	roomID := "some room"
	b.HandleMessage(chat.InMsg{From: "123", RoomID: &roomID, Body: _BECOME_ADMIN_MSG})
	equals(t, 0, len(repo.admins()))
	c.ExpectPM(chat.OutMsg{To: "1234", Body: _BECAME_ADMIN_MSG})
	b.HandleMessage(chat.InMsg{From: "1234", Body: _BECOME_ADMIN_MSG})
	equals(t, 1, len(repo.admins()))
	b.HandleMessage(chat.InMsg{From: "123", Body: _BECOME_ADMIN_MSG})
	equals(t, 1, len(repo.admins()))

	c.Check()
}
func TestThatUserCanChangeTheirName(t *testing.T) {
	b, c := mockBot(t)
	_, _ = b, c
	repo := newMockRepo()
	SetRepo(repo)

	b.AddRootHandler(NewRootHandler())

	roomID := "some room"
	b.HandleMessage(chat.InMsg{From: "123", RoomID: &roomID, Body: "Change my name"})
	c.ExpectPM(chat.OutMsg{To: "1234", Body: "What would you like to be called?"})
	b.HandleMessage(chat.InMsg{From: "1234", Body: "Change my name"})
	c.ExpectPM(chat.OutMsg{To: "1234", Body: "Batman it is."})
	b.HandleMessage(chat.InMsg{From: "123", Body: "Joker"})
	b.HandleMessage(chat.InMsg{From: "1234", Body: "Batman"})
	b.HandleMessage(chat.InMsg{From: "1234", Body: "Joker"})

	c.Check()
	equals(t, "Batman", repo["1234"].Name)
}

func (m mockRepo) admins() []User {
	users := make([]User, 0)
	for _, u := range m {
		if u.IsAdmin {
			users = append(users, u)
		}
	}
	return users
}
