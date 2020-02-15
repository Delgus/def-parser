package internal

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

// Parser is awesome
type Parser struct {
	ticker   *RandomTicker
	notifier *sse.Server
	urls     []string
	cache    *Cache
	mu       sync.Mutex
}

func (p *Parser) addURL(url string) {
	p.mu.Lock()
	p.urls = append(p.urls, url)
	p.mu.Unlock()
}

func (p *Parser) getSite(url string) *Site {
	site, found := p.cache.Get(url)
	if !found {
		p.addURL(url)
		return &Site{
			Host:       url,
			Status:     progress,
			Categories: []string{},
		}
	}
	return site
}

func (p *Parser) getURL() (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.urls) == 0 {
		return "", fmt.Errorf(`not found url for work`)
	}
	url := p.urls[0]
	p.urls = p.urls[1:]
	return url, nil
}

// NewParser is awesome
func NewParser(notifier *sse.Server, cache *Cache) *Parser {
	p := &Parser{
		ticker:   NewRandomTicker(1*time.Second, 5*time.Second),
		notifier: notifier,
		cache:    cache,
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
	url, err := p.getURL()
	if err != nil {
		return
	}

	// get html from siteadvisor
	doc, err := getDoc(url)
	if err != nil {
		logrus.Error(err)
	}

	// Узнаем безопасность сайта
	img, exist := doc.Find(`.status>img`).Attr("src")
	if !exist {
		logrus.Errorf(`unexpected page for url: %s`, url)
		return
	}
	site := &Site{Host: url, Safe: getSafe(img)}
	// Записываем категории
	doc.Find(`.content>ul>li`).Each(func(i int, s *goquery.Selection) {
		site.Categories = append(site.Categories, s.Find(`a`).Text())
	})
	site.Status = complete

	// Добавляем в кэш
	p.cache.Set(url, site)

	// Отправляем уведомление клиенту
	siteBytes, err := json.Marshal(site)
	if err != nil {
		logrus.WithError(err).Errorf(`can not convert site struct %v to json`, site)
		return
	}
	p.notifier.CreateStream(url)
	p.notifier.Publish(url, &sse.Event{
		Data: siteBytes,
	})
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
