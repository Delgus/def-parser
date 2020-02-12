package memory

import (
	"sync"
	"sync/atomic"
)

// Store реализует интерфейс хранилища данных
type Store struct {
	statements map[int64][]string
	// sites    map[string]int
	counter int64
	mu      sync.Mutex
}

// NewMemoryStore конструктор Store
func NewMemoryStore() (s *Store) {
	return &Store{
		statements: make(map[int64][]string),
	}
}

// GetNewID - получить новый id для заявки
func (s *Store) GetNewID() (int64, error) {
	atomic.AddInt64(&s.counter, 1)
	return s.counter, nil
}

// SaveStatement - сохранить заявку
func (s *Store) SaveStatement(id int64, urls []string) error {
	s.mu.Lock()
	s.statements[id] = urls
	s.mu.Unlock()
	return nil
}

// GetStatement - получить данные о заявке
func (s *Store) GetStatement(id int64) ([]string, error) {
	s.mu.Lock()
	urls := s.statements[id]
	s.mu.Unlock()
	return urls, nil
}
