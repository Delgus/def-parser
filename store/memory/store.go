package memory

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/delgus/def-parser/internal"
)

// Store реализует интерфейс хранилища данных
type Store struct {
	statements  map[int64]map[string]*internal.Site
	urlsForWork []string
	counter     int64
	mu          sync.Mutex
}

// NewMemoryStore конструктор Store
func NewMemoryStore() (s *Store) {
	return &Store{
		statements: make(map[int64]map[string]*internal.Site),
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
	for i := range urls {
		if s.statements[id] == nil {
			s.statements[id] = make(map[string]*internal.Site)
		}
		s.statements[id][urls[i]] = new(internal.Site)
	}
	s.urlsForWork = append(s.urlsForWork, urls...)
	s.mu.Unlock()
	return nil
}

// GetStatement - получить данные о заявке
func (s *Store) GetStatement(id int64) (map[string]*internal.Site, error) {
	s.mu.Lock()
	siteMap := s.statements[id]
	s.mu.Unlock()
	return siteMap, nil
}

// GetURLForWork - получить url для обработки
func (s *Store) GetURLForWork() (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.urlsForWork) == 0 {
		return "", fmt.Errorf(`not found url for work`)
	}
	url := s.urlsForWork[0]
	s.urlsForWork = s.urlsForWork[1:]
	return url, nil
}
