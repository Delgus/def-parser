package internal

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

// API реализует взаимодействие клиента и сервиса
type API struct {
	store  StoreInterface
	regURL *regexp.Regexp
}

// NewAPI конструктор нового API
func NewAPI(store StoreInterface) *API {
	return &API{
		store:  store,
		regURL: regexp.MustCompile(`^(https?:\/\/)?(www\.)?([\w\.]+\.[a-z]+\.?)(\/[\w\.]*)*\/?$`),
	}
}

// Start обрабатывает поступившую заявку
func (h *API) Start(w http.ResponseWriter, r *http.Request) {
	// id заявки
	id, err := h.store.GetNewID()
	if err != nil {
		logrus.Error("failed get new id for request for client")
		http.Redirect(w, r, `/`, http.StatusTemporaryRedirect)
		return
	}

	dirtyUrls := r.FormValue("urls")
	if dirtyUrls == "" {
		http.Redirect(w, r, `/`, http.StatusTemporaryRedirect)
		return
	}

	// TODO: вынести в отдельный метод
	var urls []string
	fields := strings.Fields(dirtyUrls)
	for _, f := range fields {
		f := strings.TrimSpace(f)
		if f == "" {
			continue
		}
		sbs := h.regURL.FindAllStringSubmatch(f, -1)
		if len(sbs) == 0 {
			continue
		}
		urls = append(urls, sbs[0][3])
	}

	if err := h.store.SaveStatement(id, urls); err != nil {
		logrus.Error("failed save statement for client")
		http.Redirect(w, r, `/`, http.StatusTemporaryRedirect)
		return
	}

	redirectRoute := fmt.Sprintf(`/result?id=%d`, id)
	http.Redirect(w, r, redirectRoute, http.StatusTemporaryRedirect)
}

// Result сообшает клиенту о результатах заявки
func (h *API) Result(w http.ResponseWriter, r *http.Request) {
	val := r.FormValue("id")
	fmt.Fprintln(w, val)
	id, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		logrus.Errorf(`incorrect int64 id`)
		http.Redirect(w, r, `/`, http.StatusTemporaryRedirect)
		return
	}
	urls, err := h.store.GetStatement(id)
	if err != nil {
		logrus.Errorf(`can not get statement from store`)
	}
	fmt.Fprintln(w, urls)
}
