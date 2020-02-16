# Парсер WebAdvisor

Серверная часть написана на go 1.13

Клиентская часть находится в папке web, одностраничник, написана на HTML5 + Bootstrap + Jquery, отдается сервером golang, но отделить ее небольшая проблема.

### Запуск через docker-compose
```
git clone https://github.com/Delgus/def-parser.git
cd def-parser
docker-compose up --build
```
на [http://localhost:8080](http://localhost:8080) откроется приложение


### Сбилдить и запустить локально
```
make build
./app
```

### Переменные окружения
```env
APP_HOST=                 # хост сервера
APP_PORT=8080             # порт сервера
CACHE_EXPIRATION=4h       # время хранения информации о сайте в кэше
CACHE_CLEAN_INTERVAL=2h   # интервал с которым очищать старый кэш
MIN_PARSE_INTERVAL=1s     # минимальный интервал между запросами к WebAdvisor
MAX_PARSE_INTERVAL=5s     # максимальный интервал между запросами к WebAdvisor
PARSE_CLIENT_TIMEOUT=30s  # таймаут для клиента парсера, сколько ждать ответа от WebAdvisor
``` 


## API

### Отправка списка url на обработку
POST `/api`

Params: 
    `url` - обязательный 

Response:
```json
{
    "statement_id": 1
}
```

### Получение результата по запросу
POST `/result`

Params:
    `statement_id` - обязательный

Response:
```json
{
    "sites":[
        {
            "host":"delgus.com",
            "status":"complete", 
            "safe":"Безопасно",
            "categories":["Технические и деловые форумы"]
        },
        {
            "host":"github.com",
            "status":"complete",
            "safe":"Безопасно",
            "categories":["Технические и деловые форумы"]
        }
    ]
}
```
status:
 `complete` - если уже обработан сайт,  
 `progress` - если сайт еще в обработке

### Event Stream
Для непрерывной доставки до клиента результата по обработке используются Server Side Events.  
Необходимо подписаться на ресурс `/events/{statement_id}`  
```js
  eventSource = new EventSource(`/events/2`);
  eventSource.onmessage = function (event) {
      // update page
  }
```
Сообщения приходят в формате JSON
```json
{
    "host":"delgus.com",
    "status":"complete",
    "safe":"Безопасно",
    "categories":["Технические и деловые форумы"]
},
```

## Клиентская часть написана на Jquery и Bootstrap

SPA на jquery.
Не бейте меня ногами, я больше backend, чем front)))
