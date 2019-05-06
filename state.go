package hans

import "sync"

type State struct {
	running bool
	sync.Mutex
}

func (s *State) Running() bool {
	s.Lock()
	defer s.Unlock()
	return s.running
}

func (s *State) RunningState(b bool) {
	s.Lock()
	defer s.Unlock()
	s.running = b
}
