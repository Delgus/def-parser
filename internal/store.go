package internal

// StoreInterface реализует хранилище данных
type StoreInterface interface {
	GetNewID() (int64, error)
	SaveStatement(int64, []string) error
	GetStatement(int64) (map[string]*Site, error)
	GetURLForWork() (string, error)
}
