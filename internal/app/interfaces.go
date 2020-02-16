package app

// StoreInterface интерфейс хранилища заявок
type StoreInterface interface {
	GetNewID() (int64, error)
	SaveStatement(int64, []string) error
	GetStatementURLs(int64) ([]string, error)
}

// CacheInterface интерфейс кэша для инфо о безопасности сайтов
type CacheInterface interface {
	Set(key string, value *Site)
	Get(key string) (*Site, bool)
	Delete(key string) error
}

// NotifierInterface реализует интерфейс для оповещения клиента о изменениях
type NotifierInterface interface {
	CreateStream(stream string)
	Publish(stream string, site *Site) error
}
