package internal

import (
	"fmt"
	"net/http"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/sirupsen/logrus"
)

const (
	safeImg    = `/img/safe-symbol.svg`
	warningImg = `/img/warning-icon.svg`
	dangerImg  = `/img/danger-icon.svg`
)

func getSafe(img string) string {
	switch img {
	case safeImg:
		return `Безопасно`
	case warningImg:
		return `Небольшой Риск`
	case dangerImg:
		return `Высокий Риск`
	}
	return `Неизвестно`
}

// Site хранит информацию о сайте
type Site struct {
	Safe       string   // Безопасность сайта
	Categories []string // Категории
}

type worker struct {
	store  StoreInterface
	ticker *RandomTicker
}

func newWorker(store StoreInterface) *worker {
	return &worker{
		store:  store,
		ticker: NewRandomTicker(1*time.Second, 5*time.Second),
	}
}

func (w *worker) run() {
	for range w.ticker.C {
		go w.work()
	}
}

func (w *worker) work() {
	url, err := w.store.GetURLForWork()
	if err == nil {
		doc, err := getDoc(url)
		if err != nil {
			logrus.Error(err)
		}
		site := &Site{Safe: `Неизвестно`}
		img, exist := doc.Find(`.status>img`).Attr("src")
		if !exist {
			logrus.Errorf(`unexpected page for url: %s`, url)
			return
		}
		site.Safe = getSafe(img)
		doc.Find(`.content>ul>li`).Each(func(i int, s *goquery.Selection) {
			site.Categories = append(site.Categories, s.Find(`a`).Text())
		})
		logrus.Info(url, site.Safe, site.Categories)
	}
}

func getDoc(checkURL string) (*goquery.Document, error) {
	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	url := fmt.Sprintf(`https://siteadvisor.com/sitereport.html?url=%s`, checkURL)

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	// russian language optionality
	request.Header.Set("Accept-Language", "ru-RU,ru;q=0.8,en-US;q=0.5,en;q=0.3")
	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		logrus.Errorf("status code error: %d %s", resp.StatusCode, resp.Status)
	}

	return goquery.NewDocumentFromReader(resp.Body)
}
