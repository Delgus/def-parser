package app

type status string

const (
	// Progress - сайт находится в обработке
	Progress status = "progress"
	// Complete - сайт успешно обработан
	Complete status = "complete"
)

// Site хранит информацию о сайте
type Site struct {
	Host       string   `json:"host"`
	Status     status   `json:"status"`
	Safe       string   `json:"safe"`       // Безопасность сайта
	Categories []string `json:"categories"` // Категории
}

// HostTask - структура для проверки сайта
type HostTask struct {
	StatementID int
	Host        string
}

// StoreInterface интерфейс хранилища заявок
type StoreInterface interface {
	SaveStatement([]string) (int, error)
	GetStatementURLs(int) ([]string, error)
}

// CacheInterface интерфейс кэша для инфо о безопасности сайтов
type CacheInterface interface {
	Set(key string, value *Site)
	Get(key string) (*Site, bool)
	Delete(key string) error
}

// QueueInterface реадизует очередь FIFO
type QueueInterface interface {
	Add(HostTask)
	Get() (HostTask, error)
}

// NotifierInterface реализует интерфейс для оповещения клиента о изменениях
type NotifierInterface interface {
	CreateStream(stream string)
	Publish(stream string, site *Site) error
}
