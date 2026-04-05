# API-тесты для qa-internship.avito.com

Автоматизированные тесты для API объявлений, написанные на Go.

| [Задание 1](./Task1/Task1.md) | [Тест-кейсы](./TESTCASES.md) | [Баг-репорты](./BUGS.md) | [Расхождения API](./API_DISCREPANCIES.md) | [Allure отчет](https://melamori-blimm-im.github.io/avito-qa-internship-API-2026/) |

**Host:** `https://qa-internship.avito.com`

## Эндпоинты

| Метод | Путь | Описание |
|-------|------|----------|
| POST | `/api/1/item` | Создать объявление |
| GET | `/api/1/item/:id` | Получить объявление по UUID |
| DELETE | `/api/2/item/:id` | Удалить объявление по UUID |
| GET | `/api/1/:sellerID/item` | Получить все объявления продавца |
| GET | `/api/1/statistic/:id` | Получить статистику объявления |

## Требования

- [Go](https://go.dev/doc/install) версии 1.22 или выше
- [Allure CLI](https://docs.qameta.io/allure/#_installing_a_commandline) для просмотра отчётов (опционально)

## Установка зависимостей

```bash
go mod download
```

## Запуск тестов

Запуск всех тестов:

```bash
go test -v ./tests/...
```

Запуск тестов одного сценария:

```bash
go test -v ./tests/scenarios/createItem/...
go test -v ./tests/scenarios/deleteItem/...
go test -v ./tests/scenarios/getItem/...
go test -v ./tests/scenarios/getSellerItems/...
go test -v ./tests/scenarios/getStatistic/...
go test -v ./tests/scenarios/nonFunctional/...
```

## Allure-отчёты

Тесты используют [ozontech/allure-go](https://github.com/ozontech/allure-go) - все результаты сохраняются в формате Allure JSON.

### 1. Установить Allure CLI

```bash
# macOS
brew install allure

# Linux (через npm)
npm install -g allure-commandline
```

### 2. Запустить тесты с сохранением результатов

```bash
ALLURE_OUTPUT_PATH=$(pwd) go test -v -count=1 -timeout 120s ./tests/...
```

Результаты появятся в папке `allure-results/` в корне проекта.

### 3. Сгенерировать и открыть HTML-отчёт

```bash
# Сгенерировать отчёт
allure generate allure-results -o allure-report --clean

# Открыть в браузере
allure open allure-report
```

Или одной командой (генерация + открытие):

```bash
allure serve allure-results
```

## Настройка окружения

Переменные окружения задаются в файле `.env` (уже содержит дефолтные значения):

```
API_URL=https://qa-internship.avito.com
DEBUG=false
```

Для локального переопределения создайте файл `.env.override` в корне проекта:

```
DEBUG=true
```

`.env.override` добавлен в `.gitignore` и не попадёт в репозиторий.

## Режим отладки

Для вывода полных HTTP-запросов и ответов в консоль:

```bash
DEBUG=true go test -v ./tests/...
```

Или выставьте `DEBUG=true` в `.env.override`.

## Запуск линтера

```bash
# Установить golangci-lint (один раз)
brew install golangci-lint

# Запуск
golangci-lint run
```

## Структура проекта

```
.
├── internal/
│   ├── helpers/api-runner/   # обёртка над библиотекой apitest
│   ├── constants/path/       # константы путей API
│   ├── client/http/item/     # низкоуровневые HTTP-запросы
│   ├── managers/item/        # хелперы для тестов
│   │   └── models/           # модели запросов и ответов
│   └── utils/                # загрузка env, логирование, рандом
├── tests/
│   ├── base.go               # общая инициализация
│   └── scenarios/
│       ├── createItem/       # тесты POST /api/1/item
│       ├── deleteItem/       # тесты DELETE /api/2/item/:id
│       ├── getItem/          # тесты GET /api/1/item/:id
│       ├── getSellerItems/   # тесты GET /api/1/:sellerID/item
│       ├── getStatistic/     # тесты GET /api/1/statistic/:id
│       └── nonFunctional/    # NF: Content-Type, время ответа
├── allure-results/           # JSON-результаты тестов (генерируется)
├── TESTCASES.md              # описание всех тест-кейсов
├── BUGS.md                   # найденные баги API
└── API_DISCREPANCIES.md      # расхождения реального API с Postman-коллекцией
```
