package api

import (
	"strconv"

	"github.com/delgus/def-parser/internal/app"
)

// Service реализует сервис для обработки входящик заявок
type Service struct {
	store    app.StoreInterface
	cache    app.CacheInterface
	queue    app.QueueInterface
	notifier app.NotifierInterface
}

// NewService вернет новый Service
func NewService(store app.StoreInterface, cache app.CacheInterface, queue app.QueueInterface, notifier app.NotifierInterface) *Service {
	return &Service{
		store:    store,
		cache:    cache,
		queue:    queue,
		notifier: notifier,
	}
}

func (s *Service) getStatementID() (int64, error) {
	return s.store.GetNewID()
}

func (s *Service) addStatement(statementID int64, domains []string) error {
	return s.store.SaveStatement(statementID, domains)
}

func (s *Service) getSites(statementID int64) ([]*app.Site, error) {
	urls, err := s.store.GetStatementURLs(statementID)
	if err != nil {
		return nil, err
	}
	var response []*app.Site
	for _, url := range urls {
		// ищем в кэше
		site, found := s.cache.Get(url)
		// если не найден - отправляем в очередь на обработку
		if !found {
			s.queue.Add(app.HostTask{Host: url, StatementID: statementID})
			// создаем соединение с клиентом
			s.notifier.CreateStream(strconv.FormatInt(statementID, 10))
			site = &app.Site{
				Host:       url,
				Status:     app.Progress,
				Categories: []string{},
			}
		}
		response = append(response, site)
	}
	return response, nil
}
