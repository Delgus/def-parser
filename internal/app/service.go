package app

// Service реализует сервис для обработки входящик заявок
type Service struct {
	store  StoreInterface
	parser *Parser
}

// NewService вернет новый Service
func NewService(store StoreInterface, parser *Parser) *Service {
	return &Service{
		store:  store,
		parser: parser,
	}
}

func (s *Service) getStatementID() (int64, error) {
	return s.store.GetNewID()
}

func (s *Service) addStatement(statementID int64, domains []string) error {
	return s.store.SaveStatement(statementID, domains)
}

func (s *Service) getSites(statementID int64) ([]*Site, error) {
	urls, err := s.store.GetStatementURLs(statementID)
	if err != nil {
		return nil, err
	}
	var response []*Site
	for _, url := range urls {
		response = append(response, s.parser.getSite(url, statementID))
	}
	return response, nil
}
