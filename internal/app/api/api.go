package api

import (
	"encoding/json"
	"net/http"

	"github.com/delgus/def-parser/internal/app"
	"github.com/sirupsen/logrus"
)

// API реализует взаимодействие клиента и сервиса
type API struct {
	validator *validator
	service   *Service
}

// ErrorResponse структура для формирования json о ошибке
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// NewAPI конструктор нового API
func NewAPI(service *Service) *API {
	return &API{
		validator: newValidator(),
		service:   service,
	}
}

// Start обрабатывает поступившую заявку и возвращает ее ID
func (a *API) Start(w http.ResponseWriter, r *http.Request) {
	// парсим полученные названия доменов
	urls := r.FormValue("urls")
	domains, err := a.validator.parseDomain(r.FormValue("urls"))
	if err != nil {
		logrus.WithError(err).Errorf("failed parse domains from client. dirty data: %v", urls)
		writeResponse(w, ErrorResponse{
			Error:   "bad request",
			Message: "bad urls",
		})
		return
	}

	// сохраняем заявку
	id, err := a.service.addStatement(domains)
	if err != nil {
		logrus.WithError(err).Errorf("failed save statement urls: %v", domains)
		writeResponse(w, ErrorResponse{
			Error:   "internal",
			Message: "failed save statement",
		})
		return
	}

	writeResponse(w, struct {
		StatemenID int `json:"statement_id"`
	}{StatemenID: id})
}

// Sites возвращает список сайтов со статусом обработки и информацией если она есть
func (a *API) Sites(w http.ResponseWriter, r *http.Request) {
	// получаем id заявки
	id, err := a.validator.parseID(r.FormValue("statement_id"))
	if err != nil {
		logrus.WithError(err).Errorf(`incorrect int statement_id %v`, id)
		writeResponse(w, ErrorResponse{
			Error:   "bad request",
			Message: "bad statement_id",
		})
		return
	}

	// получаем все сайты заявки
	sites, err := a.service.getSites(id)
	if err != nil {
		logrus.WithError(err).Errorf(`can not get sites for statement id %d`, id)
		writeResponse(w, ErrorResponse{
			Error:   "internal error",
			Message: "can not get sites for statement",
		})
		return
	}

	writeResponse(w, struct {
		Sites []*app.Site `json:"sites"`
	}{Sites: sites})
}

func writeResponse(w http.ResponseWriter, resp interface{}) {
	bytes, err := json.Marshal(resp)
	if err != nil {
		logrus.WithError(err).Errorf("failed to marshal response structure: %v", resp)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(bytes)
	if err != nil {
		logrus.WithError(err).Error("failed to send response to client")
	}
}
