package internal

import (
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
)

// API реализует взаимодействие клиента и сервиса
type API struct {
	store     StoreInterface
	validator *validator
	worker    *worker
}

// NewAPI конструктор нового API
func NewAPI(store StoreInterface) *API {
	api := &API{
		store:     store,
		validator: newValidator(),
		worker:    newWorker(store),
	}
	go api.worker.run()
	return api
}

// Start обрабатывает поступившую заявку
func (a *API) Start(w http.ResponseWriter, r *http.Request) {
	// id заявки
	id, err := a.store.GetNewID()
	if err != nil {
		logrus.Error("failed get new id for request for client")
		http.Redirect(w, r, `/`, http.StatusTemporaryRedirect)
		return
	}

	// парсим полученные названия доменов
	domains, err := a.validator.parseDomain(r.FormValue("urls"))
	if err != nil {
		logrus.Errorf("failed parse domains from client: %v", err)
		http.Redirect(w, r, `/`, http.StatusTemporaryRedirect)
		return
	}

	// сохраняем заявку в хранилище
	if err := a.store.SaveStatement(id, domains); err != nil {
		logrus.Error("failed save statement for client")
		http.Redirect(w, r, `/`, http.StatusTemporaryRedirect)
		return
	}

	// перенаправляем клиента на страницу с результатами
	redirectRoute := fmt.Sprintf(`/result?id=%d`, id)
	http.Redirect(w, r, redirectRoute, http.StatusTemporaryRedirect)
}

// Result сообшает клиенту о результатах заявки
func (a *API) Result(w http.ResponseWriter, r *http.Request) {
	// получаем id заявки
	id, err := a.validator.parseID(r.FormValue("id"))
	if err != nil {
		logrus.Errorf(`incorrect int64 id: %v`, err)
		http.Redirect(w, r, `/`, http.StatusTemporaryRedirect)
		return
	}

	// получаем заявку из хранилища
	urls, err := a.store.GetStatement(id)
	if err != nil {
		logrus.Errorf(`can not get statement from store`)
	}
	fmt.Fprintln(w, urls)
}
