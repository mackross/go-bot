package bot

import (
	"github.com/mackross/go-bot/bottest"
	"github.com/mackross/go-bot/chat"
	"testing"
)

func TestGreeterCmd(t *testing.T) {
	mockChat := bottest.NewChat(t)
	bot := NewBot(mockChat)

	towlie := NewGreeter("Towlie", "Howdy Ho!")
	tonyMontana := NewGreeter("Tony", "Say hello to my little friend")

	bot.AddRootHandler(towlie)
	bot.AddRootHandler(tonyMontana)

	mockChat.ExpectPM(chat.OutMsg{To: "12345", Body: "Howdy Ho!"})
	mockChat.ExpectPM(chat.OutMsg{To: "777", Body: "Say hello to my little friend"})

	bot.HandleMessage(chat.InMsg{From: "12345", Body: "Hello Towlie"})
	bot.HandleMessage(chat.InMsg{From: "12345", Body: "Hello Nobody"})
	bot.HandleMessage(chat.InMsg{From: "777", Body: "Hello Tony"})

	mockChat.Check()

}

func NewGreeter(botName string, greeting string) *greetingHander {
	return &greetingHander{botName, greeting}
}

type greetingHander struct {
	name     string
	greeting string
}

func (g *greetingHander) HandleMessage(b *Bot, msg chat.InMsg) bool {
	if msg.Body == "Hello "+g.name {
		b.ReplyPM(msg, g.greeting)
		return true
	}
	return false
}

type poppableHandler struct {
	popOnMsg    bool
	childPopped int
}

func (p *poppableHandler) HandleMessage(b *Bot, msg chat.InMsg) bool {
	if p.popOnMsg {
		b.PopHandler(p)
		return true
	}
	return false
}

func (p *poppableHandler) ChildPopped(b *Bot, child MessageHandler, id int) {
	p.childPopped++
}

func TestParentPop(t *testing.T) {
	mockChat := bottest.NewChat(t)
	bot := NewBot(mockChat)

	p1 := &poppableHandler{}
	p2 := &poppableHandler{}
	p3 := &poppableHandler{}

	bot.PushHandler(p1, nil)
	bot.PushHandler(p2, p1)
	bot.PushHandler(p3, p2)

	bot.HandleMessage(chat.InMsg{})

	equals(t, p1.childPopped, 0)
	equals(t, p2.childPopped, 0)
	equals(t, p3.childPopped, 0)

	p2.popOnMsg = true

	bot.HandleMessage(chat.InMsg{})

	equals(t, p1.childPopped, 1)
	equals(t, p2.childPopped, 0)
	equals(t, p3.childPopped, 0)
}
