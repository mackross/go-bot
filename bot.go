package bot

import (
	"fmt"
	"time"

	"github.com/mackross/go-bot/chat"
	"github.com/mackross/go-bot/cmd"
)

type MessageHandler interface {
	HandleMessage(b *Bot, m chat.InMsg) bool
}

type PoppedChildHandler interface {
	ChildPopped(b *Bot, child MessageHandler, id int)
}

type Bot struct {
	chat.Network

	cmdStack   *cmd.Stack
	handlerMap map[MessageHandler]*commandWrapper
	logging    bool
}

func (b *Bot) AddRootHandler(obj MessageHandler) {
	b.cmdStack.AddRoot(b.wrappedHandler(obj))
}

func (b *Bot) PushHandler(obj MessageHandler, parent MessageHandler) int {
	return b.cmdStack.PushCmd(b.wrappedHandler(obj), b.wrappedHandler(parent))
}

func (b *Bot) PopHandler(obj MessageHandler) {
	b.cmdStack.Pop(b.wrappedHandler(obj))
}

func (b *Bot) ParentHandler(obj MessageHandler) MessageHandler {
	wrappedParent := b.cmdStack.Parent(b.wrappedHandler(obj)).(*commandWrapper)
	if wrappedParent != nil {
		return wrappedParent.handler
	}
	return nil
}

func NewBot(n chat.Network) *Bot {
	b := &Bot{n, cmd.NewStack(), make(map[MessageHandler]*commandWrapper, 0), true}
	go func() {
		for m := range n.Messages() {
			b.HandleMessage(m)
			duration := time.Since(m.ArrivedAt)
			fmt.Printf("[Handled in %v]\n", duration)
		}
	}()
	return b
}

func (b *Bot) wrappedHandler(obj MessageHandler) cmd.Command {
	if obj == nil {
		return nil
	}
	if wrapper, ok := b.handlerMap[obj]; ok {
		return wrapper
	} else {
		wrapper = &commandWrapper{obj, b}
		b.handlerMap[obj] = wrapper
		return wrapper
	}
}

type commandWrapper struct {
	handler MessageHandler
	bot     *Bot
}

func (c *commandWrapper) Handle(s *cmd.Stack, obj interface{}) bool {
	return c.handler.HandleMessage(c.bot, obj.(chat.InMsg))
}

func (c *commandWrapper) ChildPopped(s *cmd.Stack, child cmd.Command, id int) {
	if p, ok := c.handler.(PoppedChildHandler); ok {
		p.ChildPopped(c.bot, child.(*commandWrapper).handler, id)
	}
}

func (b *Bot) HandleMessage(m chat.InMsg) {
	roomID := "PM"
	if m.RoomID != nil {
		roomID = *m.RoomID
	}
	fmt.Printf("<%v (%v)> %v\n", m.From, roomID, m.Body)
	b.cmdStack.Handle(m)
}

func (b *Bot) ReplyPM(orig chat.InMsg, body string) {
	orig.RoomID = nil
	b.Reply(orig, body)
}

func (b *Bot) Reply(orig chat.InMsg, body string) {
	if orig.RoomID != nil {
		fmt.Printf("<%v (%v)> %v\n", b.NickName(), *orig.RoomID, body)
		b.Send(chat.OutMsg{To: *orig.RoomID, Body: body})
	} else {
		fmt.Printf("<%v (PM:%v)> %v\n", b.NickName(), orig.From, body)
		b.SendPM(chat.OutMsg{To: orig.From, Body: body})
	}
}

func panicErr(err error) {
	if err != nil {
		panic(err)
	}
}
