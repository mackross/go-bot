package cmd

import (
	"sync"
)

type Command interface {
	Handle(s *Stack, obj interface{}) bool
}

type CommandWithChildren interface {
	ChildPopped(s *Stack, cmd Command, id int)
}

type Stack struct {
	sync.RWMutex
	commands map[int]Command
	rootIDs  []int
	stackIDs []int
	parents  map[int]int
	lastID   int
}

func NewStack() *Stack {
	return &Stack{
		sync.RWMutex{},
		make(map[int]Command, 0),
		make([]int, 0),
		make([]int, 0),
		make(map[int]int),
		0,
	}
}

func (s *Stack) Current() []Command {
	s.RLock()
	defer s.RUnlock()

	return s.current()
}

func (s *Stack) Roots() []Command {
	s.RLock()
	defer s.RUnlock()

	return s.roots()
}

func (s *Stack) AddRoot(c Command) {
	s.Lock()
	defer s.Unlock()

	s.addRoot(c)
}

func (s *Stack) PushCmd(c Command, parent Command) int {
	s.Lock()
	defer s.Unlock()

	return s.pushCmd(c, parent)
}

func (s *Stack) Pop(cmd Command) {
	s.Lock()
	defer s.Unlock()

	s.pop(cmd, true)
}

func (s *Stack) Parent(cmd Command) Command {
	s.RLock()
	defer s.RUnlock()

	return s.parent(cmd)
}

func (s *Stack) Handle(obj interface{}) {
	s.RLock()
	cmds := s.Current()
	roots := s.Roots()
	s.RUnlock()

	for i, j := 0, len(cmds)-1; i < j; i, j = i+1, j-1 {
		cmds[i], cmds[j] = cmds[j], cmds[i]
	}
	for _, cmd := range cmds {
		if cmd.Handle(s, obj) {
			return
		}
	}

	for _, cmd := range roots {
		cmd.Handle(s, obj)
	}

}

func (s *Stack) current() []Command {
	cmds := make([]Command, 0, len(s.stackIDs))
	for _, id := range s.stackIDs {
		cmd, _ := s.commands[id]
		cmds = append(cmds, cmd)
	}
	return cmds
}

func (s *Stack) roots() []Command {
	cmds := make([]Command, 0, len(s.rootIDs))
	for _, id := range s.rootIDs {
		cmd, _ := s.commands[id]
		cmds = append(cmds, cmd)
	}
	return cmds
}

func (s *Stack) pop(cmd Command, notify bool) {
	id, ok := 0, false
	if id, ok = s.findCmdID(cmd); !ok {
		panic("cannot find id")
	}

	s.popChildren(id)

	parent := s.parent(cmd)

	s.removeID(id)

	if p, ok := parent.(CommandWithChildren); notify && ok {
		p.ChildPopped(s, cmd, id)
	}
}

func (s *Stack) popChildren(id int) {
	for i, p := range s.parents {
		if p == id {
			s.pop(s.commands[i], false)
		}
	}
}

func (s *Stack) removeID(id int) {
	s.stackIDs = remove(s.stackIDs, id)
	s.rootIDs = remove(s.rootIDs, id)
	delete(s.parents, id)
	delete(s.commands, id)
}

func (s *Stack) addRoot(c Command) {
	id := s.insertCmd(c)
	s.rootIDs = append(s.rootIDs, id)
}

func (s *Stack) pushCmd(c Command, parent Command) int {
	id := s.insertCmd(c)
	s.stackIDs = append(s.stackIDs, id)
	if parentID, ok := s.findCmdID(parent); ok {
		s.parents[id] = parentID
	}
	return id
}

func (s *Stack) insertCmd(c Command) int {
	id := s.newID()
	s.commands[id] = c
	return id
}

func (s *Stack) parent(cmd Command) Command {
	if id, ok := s.findCmdID(cmd); ok {
		return s.commands[s.parents[id]]
	}
	return nil
}

func (s *Stack) newID() int {
	s.lastID++
	return s.lastID
}

func (s *Stack) findCmdID(cmd Command) (int, bool) {
	for id, c := range s.commands {
		if c == cmd {
			return id, true
		}
	}
	return 0, false
}

func remove(orig []int, value int) []int {
	ret := make([]int, 0, len(orig))
	for i := len(orig) - 1; i >= 0; i-- {
		if orig[i] != value {
			ret = append(ret, orig[i])
		}
	}
	return ret
}
