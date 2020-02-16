package app

type CacheInterface interface {
	Set(key string, value *Site)
	Get(key string) (*Site, bool)
	Delete(key string) error
}
