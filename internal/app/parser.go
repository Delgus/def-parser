package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/r3labs/sse"
	"github.com/sirupsen/logrus"
)

const (
	advisorURL = `https://siteadvisor.com/sitereport.html?url=%s`

	safeImg    = `/img/safe-symbol.svg`
	warningImg = `/img/warning-icon.svg`
	dangerImg  = `/img/danger-icon.svg`
)

// получить уровень безопасности исходя из картинки
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

// Parser реализует воркер для отправки запросов и обработки ответов
type Parser struct {
	clientTimeout time.Duration
	ticker        *RandomTicker
	notifier      *sse.Server
	urls          []url
	cache         CacheInterface
	mu            sync.Mutex
}

type url struct {
	statementID int64
	host        string
}

// добавляет хост в очередь для дальнейшей обработки
func (p *Parser) addURL(host string, id int64) {
	p.mu.Lock()
	p.urls = append(p.urls, url{host: host, statementID: id})
	p.mu.Unlock()
}

// получение инфо о сайте
func (p *Parser) getSite(url string, id int64) *Site {
	// ищем в кэше
	site, found := p.cache.Get(url)

	// если не найден - отправляем в обработку
	if !found {
		p.addURL(url, id)
		p.notifier.CreateStream(fmt.Sprintf(`%d`, id))
		return &Site{
			Host:       url,
			Status:     progress,
			Categories: []string{},
		}
	}
	return site
}

// получаем название хоста для дальнейшей обработки
func (p *Parser) getURL() (url, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.urls) == 0 {
		return url{}, fmt.Errorf(`not found url for work`)
	}
	url := p.urls[0]
	p.urls = p.urls[1:]
	return url, nil
}

// NewParser вернет новый воркер для парсинга результатов
func NewParser(notifier *sse.Server, cache CacheInterface, minTick, maxTick, clientTimeout time.Duration) *Parser {
	p := &Parser{
		clientTimeout: clientTimeout,
		ticker:        NewRandomTicker(minTick, maxTick),
		notifier:      notifier,
		cache:         cache,
	}
	go p.run()
	return p
}

func (p *Parser) run() {
	for range p.ticker.C {
		go p.work()
	}
}

func (p *Parser) work() {
	// получаем хост из очереди для обработки
	url, err := p.getURL()
	if err != nil {
		return
	}

	// получаем документ
	doc, err := p.getDoc(url.host)
	if err != nil {
		logrus.Error(err)
	}

	// Узнаем безопасность сайта
	img, exist := doc.Find(`.status>img`).Attr("src")
	if !exist {
		logrus.Errorf(`unexpected page for url: %s`, url.host)
		return
	}
	site := &Site{Host: url.host, Safe: getSafe(img)}
	// Записываем категории
	doc.Find(`.content>ul>li`).Each(func(i int, s *goquery.Selection) {
		site.Categories = append(site.Categories, s.Find(`a`).Text())
	})
	site.Status = complete

	// Добавляем в кэш
	p.cache.Set(url.host, site)

	// Отправляем уведомление клиенту
	siteBytes, err := json.Marshal(site)
	if err != nil {
		logrus.WithError(err).Errorf(`can not convert site struct %v to json`, site)
		return
	}
	p.notifier.Publish(fmt.Sprintf(`%d`, url.statementID), &sse.Event{
		Data: siteBytes,
	})
}

func (p *Parser) getDoc(checkURL string) (*goquery.Document, error) {
	client := &http.Client{Timeout: p.clientTimeout}

	url := fmt.Sprintf(advisorURL, checkURL)
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	// заголовок устанавливающий русский язык для ответа
	request.Header.Set("Accept-Language", "ru-RU,ru;q=0.8,en-US;q=0.5,en;q=0.3")

	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("status code error: %d %s", resp.StatusCode, resp.Status)
	}

	return goquery.NewDocumentFromReader(resp.Body)
}
