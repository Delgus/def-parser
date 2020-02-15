package internal

// StoreInterface реализует хранилище данных
type StoreInterface interface {
	GetNewID() (int64, error)
	SaveStatement(int64, []string) error
	GetStatementURLs(int64) ([]string, error)
}
