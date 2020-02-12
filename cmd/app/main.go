package main

import (
	"net/http"

	"github.com/delgus/def-parser/internal"
	"github.com/delgus/def-parser/store/memory"
	"github.com/sirupsen/logrus"
)

func main() {
	// api handler
	store := memory.NewMemoryStore()
	api := internal.NewAPI(store)

	// web client
	http.Handle("/", http.FileServer(http.Dir("web")))
	http.HandleFunc("/api", api.Start)
	http.HandleFunc("/result", api.Result)
	// TODO: настраивать адрес и порт сервера из энвов
	logrus.Fatal(http.ListenAndServe(":8080", nil))
}
