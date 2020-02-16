package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/delgus/def-parser/internal/app"
	cachemem "github.com/delgus/def-parser/internal/infrastructure/cache/memory"
	"github.com/delgus/def-parser/internal/infrastructure/notify/sse"
	"github.com/delgus/def-parser/internal/infrastructure/store/memory"
	"github.com/kelseyhightower/envconfig"
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
	notifier := sse.NewNotifier("/events/")
	defer notifier.Shutdown()

	// cache - хранит инфо о хостах и их безопасности
	cache := cachemem.NewCache(cfg.CacheExpiration, cfg.CacheCleanInterval)

	// parser - воркер который с определенной периодичностью отправляет запросы на siteadvisor
	parser := app.NewParser(notifier, cache, cfg.MinParseInterval, cfg.MaxParseInterval, cfg.ParseClientTimeout)

	service := app.NewService(store, parser)

	api := app.NewAPI(service)

	// web client
	http.Handle("/", http.FileServer(http.Dir("web")))

	// api
	http.HandleFunc("/api", api.Start)
	http.HandleFunc("/result", api.Sites)

	logrus.Fatal(http.ListenAndServe(fmt.Sprintf(`%s:%d`, cfg.Host, cfg.Port), nil))
}
