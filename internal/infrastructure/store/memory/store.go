package memory

import (
	"sync"
)

// Store реализует интерфейс хранилища данных
type Store struct {
	statements map[int][]string
	counter    int
	mu         sync.Mutex
}

// NewMemoryStore конструктор Store
func NewMemoryStore() (s *Store) {
	return &Store{
		statements: make(map[int][]string),
	}
}

// SaveStatement - сохранить заявку
func (s *Store) SaveStatement(urls []string) (int, error) {
	s.mu.Lock()
	s.counter++
	counter := s.counter
	s.statements[counter] = urls
	s.mu.Unlock()
	return counter, nil
}

// GetStatementURLs - получить данные о заявке
func (s *Store) GetStatementURLs(id int) ([]string, error) {
	s.mu.Lock()
	sites := s.statements[id]
	s.mu.Unlock()
	return sites, nil
}
