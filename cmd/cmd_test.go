package cmd

import (
	"testing"
)

type BasicCmd struct {
	popped     int
	lastPopped int
}

func (h *BasicCmd) Handle(s *Stack, obj interface{}) bool {
	return false
}

func (c *BasicCmd) ChildPopped(s *Stack, cmd Command, id int) {
	c.lastPopped = id
	c.popped++
}

func TestChildCmds(t *testing.T) {
	s := NewStack()

	rootCmd := &BasicCmd{}
	childCmd1 := &BasicCmd{}
	childCmd2 := &BasicCmd{}
	childCmd3 := &BasicCmd{}

	s.AddRoot(rootCmd)
	cmd1ID := s.PushCmd(childCmd1, rootCmd)
	cmd2ID := s.PushCmd(childCmd2, rootCmd)

	equals(t, len(s.Roots()), 1)
	equals(t, len(s.Current()), 2)

	s.PushCmd(childCmd3, childCmd2)
	// add current cmd should increase count
	equals(t, len(s.Current()), 3)

	// parent child relationships setup
	equals(t, s.Parent(childCmd1), rootCmd)
	equals(t, s.Parent(childCmd2), rootCmd)
	equals(t, s.Parent(childCmd3), childCmd2)
	equals(t, s.Parent(rootCmd), nil)

	s.Pop(childCmd1)
	// root was notified of cmd1 popping
	equals(t, 1, rootCmd.popped)
	equals(t, rootCmd.lastPopped, cmd1ID)

	// cmd2 and cmd3 left
	equals(t, len(s.Current()), 2)

	s.Pop(childCmd2)
	// popping cmd2 should remove cmd3
	equals(t, len(s.Current()), 0)

	// root was notified of cmd2 popping
	equals(t, 2, rootCmd.popped)
	equals(t, rootCmd.lastPopped, cmd2ID)

	// cmd2 shouldnt be notified of cmd3 popping
	equals(t, childCmd2.popped, 0)

}

type HandlerCmd struct {
	HandleNextMessage       bool
	PopSelfAfterNextMessage bool
	Handled                 int
	CouldHandle             int
}

func (h *HandlerCmd) Handle(s *Stack, obj interface{}) bool {
	h.CouldHandle++
	if h.PopSelfAfterNextMessage {
		s.Pop(h)
	}
	if h.HandleNextMessage {
		h.Handled++
		return true
	}
	return false
}

func TestHandler(t *testing.T) {

	cmd1 := &HandlerCmd{}
	cmd2 := &HandlerCmd{}
	cmd3 := &HandlerCmd{}
	cmd4 := &HandlerCmd{}

	s := NewStack()

	s.AddRoot(cmd1)
	s.AddRoot(cmd4)
	s.PushCmd(cmd2, cmd1)
	s.PushCmd(cmd3, nil)

	s.Handle("test")

	equals(t, cmd1.CouldHandle, 1)
	equals(t, cmd4.CouldHandle, 1)
	equals(t, cmd2.CouldHandle, 1)
	equals(t, cmd3.CouldHandle, 1)

	cmd3.HandleNextMessage = true
	s.Handle("test")

	equals(t, cmd1.CouldHandle, 1)
	equals(t, cmd4.CouldHandle, 1)
	equals(t, cmd2.CouldHandle, 1)
	equals(t, cmd3.CouldHandle, 2)
	equals(t, cmd1.Handled, 0)
	equals(t, cmd2.Handled, 0)
	equals(t, cmd3.Handled, 1)

	cmd1.HandleNextMessage = true
	cmd3.HandleNextMessage = false
	s.Handle("test")
	equals(t, cmd1.CouldHandle, 2)
	equals(t, cmd4.CouldHandle, 2)
	equals(t, cmd2.CouldHandle, 2)
	equals(t, cmd3.CouldHandle, 3)
	equals(t, cmd1.Handled, 1)
	equals(t, cmd2.Handled, 0)
	equals(t, cmd3.Handled, 1)

	cmd1.PopSelfAfterNextMessage = true
	s.Handle("test")

	equals(t, cmd1.CouldHandle, 3)
	equals(t, cmd4.CouldHandle, 3)
	equals(t, cmd2.CouldHandle, 3)
	equals(t, cmd3.CouldHandle, 4)
	equals(t, cmd1.Handled, 2)
	equals(t, cmd2.Handled, 0)
	equals(t, cmd3.Handled, 1)

	s.Handle("test")

	equals(t, cmd1.CouldHandle, 3)
	equals(t, cmd4.CouldHandle, 4)
	equals(t, cmd2.CouldHandle, 3)
	equals(t, cmd3.CouldHandle, 5)
	equals(t, cmd1.Handled, 2)
	equals(t, cmd2.Handled, 0)
	equals(t, cmd3.Handled, 1)
}
