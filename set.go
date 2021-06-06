package botnet

import "sync"

type set struct {
	mit sync.Mutex
	it  map[interface{}]struct{}
}

func (s *set) Add(i interface{}) {
	s.mit.Lock()
	defer s.mit.Unlock()

	s.init()

	s.it[i] = struct{}{}
}

func (s *set) Drop(i interface{}) {
	s.mit.Lock()
	defer s.mit.Unlock()

	s.init()

	delete(s.it, i)
}

func (s *set) Empty() bool {
	return len(s.it) == 0
}

func (s *set) Size() int {
	return len(s.it)
}

func (s *set) Slice() (ret []interface{}) {
	for key := range s.it {
		ret = append(ret, key)
	}
	return
}

func (s *set) init() {
	if s.it != nil {
		return
	}

	s.it = map[interface{}]struct{}{}
}
