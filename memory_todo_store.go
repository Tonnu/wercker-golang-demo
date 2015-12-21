package main

import (
	"errors"
	"strconv"
	"sync/atomic"
)

func NewMemoryTodoStore() *MemoryTodoStore {
	counter := int32(0)
	return &MemoryTodoStore{
		counter: &counter,
		s:       make(map[string]*Todo),
	}
}

type MemoryTodoStore struct {
	s map[string]*Todo

	counter *int32
}

func (s *MemoryTodoStore) createID() string {
	id := atomic.AddInt32(s.counter, 1)

	return strconv.Itoa(int(id))
}

func (s *MemoryTodoStore) Create(t *Todo) (id string, err error) {
	id = s.createID()
	t.ID = id

	s.s[id] = t

	return id, nil
}

func (s *MemoryTodoStore) Get(id string) (*Todo, error) {
	if t, exists := s.s[id]; exists {
		return t, nil
	}

	return nil, nil
}

func (s *MemoryTodoStore) Update(t *Todo) error {
	if _, exists := s.s[t.ID]; exists {
		s.s[t.ID] = t
		return nil
	}

	return errors.New("TODO not found")
}

func (s *MemoryTodoStore) Delete(id string) error {
	delete(s.s, id)

	return nil
}

func (s *MemoryTodoStore) GetAll() ([]*Todo, error) {
	buf := []*Todo{}
	for _, todo := range s.s {
		buf = append(buf, todo)
	}
	return buf, nil
}
