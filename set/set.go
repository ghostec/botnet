package set

import "sync"

type Set struct {
	mit sync.Mutex
	it  map[interface{}]struct{}
}

func New() *Set {
	return &Set{
		it: map[interface{}]struct{}{},
	}
}

func (s *Set) Add(i interface{}) {
	s.mit.Lock()
	defer s.mit.Unlock()

	s.it[i] = struct{}{}
}

func (s *Set) Drop(i interface{}) {
	s.mit.Lock()
	defer s.mit.Unlock()

	delete(s.it, i)
}

func (s *Set) Empty() bool {
	return len(s.it) == 0
}

func (s *Set) Size() int {
	return len(s.it)
}

func (s *Set) Slice() (ret []interface{}) {
	for key := range s.it {
		ret = append(ret, key)
	}
	return
}
