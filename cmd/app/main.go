package main

import (
	"net/http"
	"time"

	"github.com/delgus/def-parser/internal"
	"github.com/delgus/def-parser/store/memory"
	"github.com/r3labs/sse"
	"github.com/sirupsen/logrus"
)

func main() {
	// data storage
	store := memory.NewMemoryStore()

	// notifier
	notifier := sse.New()

	// cache
	cache := internal.NewCache(1*time.Hour, 1*time.Hour)

	// parser
	parser := internal.NewParser(notifier, cache)

	// service
	service := internal.NewService(store, parser)

	// api
	api := internal.NewAPI(service)

	// web client
	http.Handle("/", http.FileServer(http.Dir("web")))
	http.HandleFunc("/api", api.Start)
	http.HandleFunc("/result", api.Sites)

	// sse notify handler
	http.HandleFunc("/events", notifier.HTTPHandler)

	// TODO: настраивать адрес и порт сервера из энвов
	logrus.Fatal(http.ListenAndServe(":8080", nil))
}
