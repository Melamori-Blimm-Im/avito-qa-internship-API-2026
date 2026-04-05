[Задание 1: Скриншот с багами](./Task1/Task1.md) 

[Задание 2.1: Тесты API](https://github.com/avito-tech/tech-internship/blob/main/Tech%20Internships/QA/QA-trainee-assignment-spring-2026/QA-trainee-assignment-spring-2026.md#%D0%B7%D0%B0%D0%B4%D0%B0%D0%BD%D0%B8%D0%B5-21-%D1%82%D0%B5%D1%81%D1%82%D1%8B-api) | [Тест-кейсы](./TESTCASES.md) | [Баг-репорты](./BUGS.md) | [Расхождения API](./API_DISCREPANCIES.md) | [Allure отчет](https://melamori-blimm-im.github.io/avito-qa-internship-API-2026/)  

# API-тесты для qa-internship.avito.com

Автоматизированные тесты для API объявлений, написанные на Go.

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

[Задание 2.2: Тесты UI](https://github.com/avito-tech/tech-internship/blob/main/Tech%20Internships/QA/QA-trainee-assignment-spring-2026/QA-trainee-assignment-spring-2026.md#%D0%B7%D0%B0%D0%B4%D0%B0%D0%BD%D0%B8%D0%B5-22-%D1%82%D0%B5%D1%81%D1%82%D1%8B-ui) Так как UI задание тоже интересно, то сделала это задание тоже. Но времени на него было сильно много... [Репозиторий на задание 2.2](https://github.com/Melamori-Blimm-Im/avito-qa-internship-UI-2026)
