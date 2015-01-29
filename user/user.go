package user

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/mackross/go-bot"
	"github.com/mackross/go-bot/chat"
)

type Repo interface {
	UserForID(id string) (*User, error)
	ListUsers() ([]User, error)
	SaveUser(u User) error
}

const (
	_BECOME_ADMIN_MSG = "admin me"
	_BECAME_ADMIN_MSG = "Adminized"
)

type User struct {
	ID      string
	Name    string
	IsAdmin bool
	Flags   []string
}

func (u *User) HasFlag(s string) bool {
	for _, f := range u.Flags {
		if f == s {
			return true
		}
	}
	return false
}

var repo Repo

func SetRepo(r Repo) {
	repo = r
}

type userRootHandler struct {
}

func NewRootHandler() bot.MessageHandler {
	return &userRootHandler{}
}

func (r *userRootHandler) HandleMessage(b *bot.Bot, m chat.InMsg) bool {
	u, err := GetUser(m)
	panicErr(err)

	if u == nil {
		u = &User{ID: m.From}
		err = repo.SaveUser(*u)
		panicErr(err)
	}
	lower := strings.ToLower(m.Body)
	if lower == "hi" {
		name := u.Name
		if len(name) == 0 {
			name = u.ID
		}
		b.Reply(m, "Hey "+name)
		return true
	}

	if lower == "you suck" {
		b.ReplyPM(m, "tut tut potty mouth")
		return true
	}

	if lower == "whoami" {
		b.ReplyPM(m, fmt.Sprintf("ID: %v\nName: %v\n", u.ID, u.Name))
		return true
	}
	if strings.HasPrefix(lower, "whois ") && u.IsAdmin {
		split := strings.Split(m.Body, " ")
		if len(split) == 2 {
			u, err := repo.UserForID(split[1])
			if err != nil || u == nil {
				b.ReplyPM(m, fmt.Sprintf("No record found for %v.", split[1]))
			} else {
				b.ReplyPM(m, fmt.Sprintf("ID: %v\nName: %v\nAdmin: %v\nFlags: %v\n", u.ID, u.Name, u.IsAdmin, u.Flags))

			}
			return true
		}
	}

	if lower == "docker ps" {
		cmd := exec.Command("docker", "ps")
		stdout, _ := cmd.StdoutPipe()
		stderr, _ := cmd.StderrPipe()

		reader := io.MultiReader(stdout, stderr)
		scanner := bufio.NewScanner(reader)

		go func() {
			cmd.Start()
			for s := scanner.Scan(); s; s = scanner.Scan() {
				t := scanner.Text()
				b.ReplyPM(m, t)
			}
			if err != nil && err != io.EOF {
				b.ReplyPM(m, fmt.Sprintf("Error occured: %v", err))
			}
		}()
		return true
	}

	if u.IsAdmin && strings.HasPrefix(lower, "toggle flag ") && len(strings.Split(lower, " ")) == 4 {
		split := strings.Split(m.Body, " ")
		u, err := repo.UserForID(split[2])
		if err != nil || u == nil {
			b.ReplyPM(m, fmt.Sprintf("No record found for %v.", split[1]))
		} else {
			if u.Flags == nil {
				u.Flags = make([]string, 0)
			}
			for i, flag := range u.Flags {
				if flag == split[3] {
					u.Flags[i], u.Flags = u.Flags[len(u.Flags)-1], u.Flags[:len(u.Flags)-1] // delete flag
					err := repo.SaveUser(*u)
					if err != nil {
						b.ReplyPM(m, fmt.Sprintf("Unable to save change to %v due to error: %v", u.ID, err))
						return true
					}
					b.ReplyPM(m, fmt.Sprintf("%v now has flags %v", u.ID, u.Flags))
					return true
				}
			}
			u.Flags = append(u.Flags, split[3])
			err := repo.SaveUser(*u)
			if err != nil {
				b.ReplyPM(m, fmt.Sprintf("Unable to save change to %v due to error: %v", u.ID, err))
				return true
			}
			b.ReplyPM(m, fmt.Sprintf("%v now has flags %v", u.ID, u.Flags))

		}
		return true
	}

	if u.IsAdmin && strings.HasPrefix(lower, "list users") && len(strings.Split(lower, " ")) == 2 || len(strings.Split(lower, " ")) == 3 {
		users, err := repo.ListUsers()
		if err != nil {
			b.ReplyPM(m, fmt.Sprintf("Unable to fetch users due to error: %v", err))
		}
		var flag *string
		if len(strings.Split(lower, " ")) == 3 {
			flag = &strings.Split(lower, " ")[2]
		}
		for _, u := range users {
			if flag != nil && !u.HasFlag(*flag) {
				continue
			}
			b.ReplyPM(m, fmt.Sprintf("ID: %v\tName: %v\tAdmin: %v\tFlags: %v\t", u.ID, u.Name, u.IsAdmin, u.Flags))
		}
	}

	if u.IsAdmin && strings.HasPrefix(lower, "toggle admin ") && len(strings.Split(lower, " ")) == 3 {
		split := strings.Split(m.Body, " ")
		u, err := repo.UserForID(split[2])
		if err != nil || u == nil {
			b.ReplyPM(m, fmt.Sprintf("No record found for %v.", split[1]))
		} else {
			u.IsAdmin = !u.IsAdmin
			err := repo.SaveUser(*u)
			if err != nil {
				b.ReplyPM(m, fmt.Sprintf("Unable to save change to %v due to error: %v", u.ID, err))
				return true
			}
			if u.IsAdmin {
				b.ReplyPM(m, fmt.Sprintf("%v is now an admin.", u.ID))
			} else {
				b.ReplyPM(m, fmt.Sprintf("%v is no longer an admin.", u.ID))
			}
		}
		return true
	}

	if m.IsPM() {
		if m.Body == _BECOME_ADMIN_MSG && len(Admins()) == 0 {
			u.IsAdmin = true
			err = repo.SaveUser(*u)
			b.Reply(m, _BECAME_ADMIN_MSG)
			panicErr(err)
			return true
		}

		if lower == "change my name" {
			b.Reply(m, "What would you like to be called?")
			from := m.From
			b.PushHandler(&questionHandler{true, &from, "", func(q *questionHandler) bool {
				u, err := GetUser(m)
				if u == nil || err != nil {
					b.Reply(m, "Sorry "+q.value+". Something went wrong try the command again from the start.")
					return true
				}
				u.Name = q.value
				repo.SaveUser(*u)
				b.Reply(m, u.Name+" it is.")
				return true
			}}, nil)
			return true
		}

	}

	return false
}

type questionHandler struct {
	pmOnly bool
	userID *string
	value  string
	done   func(q *questionHandler) bool
}

func (q *questionHandler) HandleMessage(b *bot.Bot, m chat.InMsg) bool {
	isCorrectUser := q.userID == nil || m.From == *q.userID
	isCorrectType := !q.pmOnly || m.IsPM()
	if isCorrectType && isCorrectUser {
		q.value = m.Body
		if q.done == nil || q.done(q) {
			b.PopHandler(q)
			return true
		}
	}
	return false
}

func Admins() []User {
	admins := make([]User, 0)
	users, err := repo.ListUsers()
	panicErr(err)
	for _, u := range users {
		if u.IsAdmin {
			admins = append(admins, u)
		}
	}
	return admins

}

func GetUser(m chat.InMsg) (*User, error) {
	u, err := repo.UserForID(m.From)
	if err != nil && err.Error() != "not found" {
		return nil, err
	}
	return u, nil
}

func panicErr(err error) {
	if err != nil {
		panic(err)
	}
}
