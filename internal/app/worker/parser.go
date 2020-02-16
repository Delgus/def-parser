package worker

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/delgus/def-parser/internal/app"
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
	notifier      app.NotifierInterface
	queue         app.QueueInterface
	cache         app.CacheInterface
}

// NewParser вернет новый воркер для парсинга результатов
func NewParser(notifier app.NotifierInterface, cache app.CacheInterface, queue app.QueueInterface,
	minTick, maxTick, clientTimeout time.Duration) *Parser {
	return &Parser{
		clientTimeout: clientTimeout,
		ticker:        NewRandomTicker(minTick, maxTick),
		notifier:      notifier,
		cache:         cache,
		queue:         queue,
	}
}

// получаем название хоста для дальнейшей обработки
func (p *Parser) getTask() (app.HostTask, error) {
	task, err := p.queue.Get()
	if err != nil {
		return task, err
	}
	// проверяем кэш вдруг этот сайт уже был обработан
	// если так то отправляем клиенту инфо и переходим к следующему сайту в очереди
	site, found := p.cache.Get(task.Host)
	if found {
		err := p.notifier.Publish(strconv.Itoa(task.StatementID), site)
		if err != nil {
			logrus.WithError(err).Errorf(`can not publish message to channel %d`, task.StatementID)
		}
		return p.getTask()
	}

	return task, nil
}

// Run - запуск парсера
func (p *Parser) Run() {
	for range p.ticker.C {
		go p.work()
	}
}

func (p *Parser) work() {
	task, err := p.getTask()
	if err != nil {
		if err != io.EOF {
			logrus.WithError(err).Error(`can not get task for work`)
		}
		return
	}

	// получаем документ
	doc, err := p.getDoc(task.Host)
	if err != nil {
		logrus.WithError(err).Error(`can not get site info from web advisor`)
		return
	}

	// Узнаем безопасность сайта
	img, exist := doc.Find(`.status>img`).Attr("src")
	if !exist {
		logrus.Errorf(`unexpected page for url: %s`, task.Host)
		return
	}
	site := &app.Site{Host: task.Host, Safe: getSafe(img)}
	// Записываем категории
	doc.Find(`.content>ul>li`).Each(func(i int, s *goquery.Selection) {
		site.Categories = append(site.Categories, s.Find(`a`).Text())
	})
	site.Status = app.Complete

	// Добавляем в кэш
	p.cache.Set(task.Host, site)

	// Отправляем уведомление клиенту
	if err := p.notifier.Publish(strconv.Itoa(task.StatementID), site); err != nil {
		logrus.WithError(err).Errorf(`error by publish message in stream %d`, task.StatementID)
	}
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
