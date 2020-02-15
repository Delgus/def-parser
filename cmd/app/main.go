package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/delgus/def-parser/internal"
	"github.com/delgus/def-parser/store/memory"
	"github.com/kelseyhightower/envconfig"
	"github.com/r3labs/sse"
	"github.com/sirupsen/logrus"
)

type configuration struct {
	Host               string        `envconfig:"APP_HOST" default:""`                // хост сервера
	Port               int           `envconfig:"APP_PORT" default:"8080"`            // порт сервера
	CacheExpiration    time.Duration `envconfig:"CACHE_EXPIRATION" default:"4h"`      // время хранения информации о сайте в кэше
	CacheCleanInterval time.Duration `envconfig:"CACHE_CLEAN_INTERVAL" default:"2h"`  // интервал с которым очищать старый кэш
	MinParseInterval   time.Duration `envconfig:"MIN_PARSE_INTERVAL" default:"1s"`    // минимальный интервал между запросами
	MaxParseInterval   time.Duration `envconfig:"MAX_PARSE_INTERVAL" default:"5s"`    // максимальный интервал между запросами
	ParseClientTimeout time.Duration `envconfig:"PARSE_CLIENT_TIMEOUT" default:"30s"` // таймаут для клиента парсера, сколько ждать ответа
}

func main() {
	var cfg configuration
	err := envconfig.Process("", &cfg)
	if err != nil {
		logrus.WithError(err).Fatal("can not get environments for app")
	}
	// data storage - хранит заявки клиентов на получения информации по хостам
	store := memory.NewMemoryStore()

	// notifier - оповещает клиента. использованы Server Side Events
	notifier := sse.New()

	// cache - хранит инфо о хостах и их безопасности
	cache := internal.NewCache(cfg.CacheExpiration, cfg.CacheCleanInterval)

	// parser - воркер который с определенной периодичностью отправляет запросы на siteadvisor
	parser := internal.NewParser(notifier, cache, cfg.MinParseInterval, cfg.MaxParseInterval, cfg.ParseClientTimeout)

	service := internal.NewService(store, parser)

	api := internal.NewAPI(service)

	// web client
	http.Handle("/", http.FileServer(http.Dir("web")))

	// api
	http.HandleFunc("/api", api.Start)
	http.HandleFunc("/result", api.Sites)

	// sse notify handler
	http.HandleFunc("/events", notifier.HTTPHandler)

	logrus.Fatal(http.ListenAndServe(fmt.Sprintf(`%s:%d`, cfg.Host, cfg.Port), nil))
}
