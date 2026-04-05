# Тест-кейсы API qa-internship.avito.com

## Общие сведения

**Host:** `https://qa-internship.avito.com`  
**Формат данных:** JSON  
**Аутентификация:** не требуется  


**Приоритеты:** P0 - критичный сценарий / смоук; P1 - основная функциональность; P2 - негативы, границы, безопасность.

**Тесты, которые намеренно падают, пока дефект не исправлен** (ожидание «как должно быть» vs факт API):

| Автотест | Приоритет | Шаги | Ожидание теста | Связь |
|----------|-----------|------|----------------|-------|
| `TestCreateItemNegativePrice` | P0 | 1. POST с `price < 0`.<br>2. Ожидать 400. | 400 | BUG-001 |
| `TestCreateItemZeroStatisticsBug` | P1 | 1. POST со всеми счётчиками 0.<br>2. Ожидать 200. | 200 (нулевая статистика допустима) | BUG-002 |
| `TestCreateItemNegativeStatisticsLikes`, `TestCreateItemNegativeStatisticsViewCount`, `TestCreateItemNegativeStatisticsContacts` | P2 | 1. POST с отрицательным likes, viewCount или contacts.<br>2. Ожидать 400. | 400 | BUG-004 |
| `TestCreateItemWhitespaceName` | P2 | 1. POST с `name` из пробелов.<br>2. Ожидать 400. | 400 | BUG-005 |
| `TestCreateItemVeryLargePrice` | P2 | 1. POST с `price=9007199254740993` (2^53+1).<br>2. Ожидать 400. | 400 | BUG-007 |
| `TestCreateItemNegativeSellerIDPost` | P2 | 1. POST с `sellerID=-111111` (не `-1`).<br>2. Ожидать 400. | 400 | BUG-008 |
| `TestCreateItemNameTooLong` | P2 | 1. POST с `name` длиной 100 000 символов.<br>2. Ожидать 400. | 400 | BUG-009 |
| `TestGetItemCreatedAtFormat` | P1 | 1. Создать объявление, GET по id.<br>2. Распарсить `createdAt` без предобработки известными layout'ами. | строка соответствует ISO 8601 / RFC3339 | BUG-010 |
| `TestCreateItemSellerIDMinusOneHangOrTimeout` | P0 | 1. POST с `sellerID=-1`, клиент с таймаутом.<br>2. Ожидать зависание до таймаута (дефект) или быстрый **400** (исправление); не ожидать **200** с созданием объявления. | таймаут или 400, не успешное создание | BUG-003 |

---

## TC-01: POST `/api/1/item`

| ID | Приоритет | Описание | Шаги | Ожидаемый результат (спецификация) | Автотест |
|----|-----------|----------|------|-------------------------------------|----------|
| TC-01-01 | P0 | Создание со всеми полями, проверка через GET | 1. POST `/api/1/item` с валидным телом (sellerID, name, price, statistics).<br>2. Из ответа извлечь UUID.<br>3. GET `/api/1/item/{id}`.<br>4. Сверить поля с запросом. | 200, данные совпадают | `TestCreateItemAllFieldsPositive` |
| TC-01-02 | P1 | Минимальная цена `price=1` | 1. `CreateItem` с `price=1` и валидными остальными полями (неуспех = не 200).<br>2. Проверить непустой UUID в ответе. | 200 | `TestCreateItemMinimumPrice` |
| TC-01-03 | P1 | Два одинаковых POST - разные UUID | 1. POST с телом A.<br>2. Повторить POST с тем же телом A.<br>3. Сравнить UUID из ответов. | 200, id различаются | `TestCreateItemIdempotencyPositive` |
| TC-01-04 | P1 | Нет поля `name` | 1. POST без `name`, остальное валидно (в т.ч. statistics).<br>2. Проверить статус. | 400 | `TestCreateItemMissingNameField` |
| TC-01-05 | P1 | Нет поля `price` | 1. POST без `price`.<br>2. Проверить статус. | 400 | `TestCreateItemMissingPriceField` |
| TC-01-06 | P1 | Нет поля `sellerID` | 1. POST без `sellerID`.<br>2. Проверить статус. | 400 | `TestCreateItemMissingSellerIDField` |
| TC-01-07 | P1 | Пустое тело `{}` | 1. POST с `{}`.<br>2. Проверить статус. | 400 | `TestCreateItemMissingEmptyBody` |
| TC-01-08 | P1 | Пустой `name` | 1. POST с `"name":""`.<br>2. Проверить статус. | 400 | `TestCreateItemMissingEmptyName` |
| TC-01-09 | P2 | Невалидный JSON | 1. POST с синтаксически неверным JSON.<br>2. Проверить статус. | 400 | `TestCreateItemMissingInvalidJSON` |
| TC-01-10 | P2 | `price` / `sellerID` строкой | 1. POST с `"price"` или `"sellerID"` строкой.<br>2. Проверить статус. | 400 | `TestCreateItemMissingPriceAsString`, `TestCreateItemMissingSellerIDAsString` |
| TC-01-11 | P1 | Нет `likes` в `statistics` | 1. POST без поля `likes` внутри `statistics`.<br>2. Проверить статус. | 400 | `TestCreateItemMissingLikesInStatistics` |
| TC-01-12 | P2 | `price=-1` | 1. POST с отрицательной ценой и ненулевой statistics.<br>2. Проверить статус. | 400 | `TestCreateItemNegativePrice` (падает при BUG-001) |
| TC-01-13 | P2 | `price=0` | 1. POST с `price=0`.<br>2. Проверить статус. | 400 | `TestCreateItemZeroPrice` |
| TC-01-14 | P2 | `sellerID=0` | 1. POST с `sellerID=0`.<br>2. Проверить статус. | 400 | `TestCreateItemSellerIDZero` |
| TC-01-15 | P2 | Слишком длинный `name` (100 000 символов) | 1. POST с `name` длиной 100000 символов, остальные поля валидны.<br>2. Проверить статус. | 400 | `TestCreateItemNameTooLong` (падает при BUG-009) |
| TC-01-16 | P2 | `name` из пробелов | 1. POST с `name` из одних пробелов.<br>2. Проверить статус. | 400 | `TestCreateItemWhitespaceName` (падает при BUG-005) |
| TC-01-17 | P2 | Отрицательные счётчики statistics | 1. POST с `likes=-1`, `viewCount=-1` или `contacts=-1`.<br>2. Проверить статус. | 400 | `TestCreateItemNegativeStatisticsLikes`, `TestCreateItemNegativeStatisticsViewCount`, `TestCreateItemNegativeStatisticsContacts` (падают при BUG-004) |
| TC-01-18 | P1 | `statistics` все нули | 1. POST с `likes=0, viewCount=0, contacts=0`.<br>2. Проверить статус. | 200 | `TestCreateItemZeroStatisticsBug` (падает при BUG-002) |
| TC-01-19 | P2 | `price=2^53+1` | 1. POST с `price=9007199254740993`.<br>2. Проверить статус. | 400 | `TestCreateItemVeryLargePrice` (падает при BUG-007) |
| TC-01-20 | P2 | `sellerID=-111111` | 1. POST с отрицательным sellerID (не `-1`, см. BUG-003).<br>2. Проверить статус. | 400 | `TestCreateItemNegativeSellerIDPost` (падает при BUG-008) |
| TC-01-21 | P2 | `sellerID=2^53+1` | 1. `CreateItem` с `sellerID=9007199254740993`.<br>2. Успех подразумевает 200. | факт: 200 | `TestCreateItemSellerIDBoundaryHugePow53PlusOne` |
| TC-01-22 | P2 | Все счётчики `statistics=2^53+1` | 1. `CreateItem` с likes/viewCount/contacts = `9007199254740993`.<br>2. Успех подразумевает 200. | факт: 200 | `TestCreateItemHugeStatisticsValues` |
| TC-01-23 | P2 | `sellerID=2147483647` (INT_MAX) | 1. `CreateItem` с `sellerID=2147483647`.<br>2. Успех подразумевает 200. | 200 | `TestCreateItemSellerIDBoundaryIntMax` |
| TC-01-24 | P2 | Лишние поля в теле | 1. POST с `unknownField` в корне.<br>2. POST с `unknownField` в `statistics`.<br>3. Проверить статус. | 200, лишнее игнорируется | `TestCreateItemUnknownFieldsIgnored` |
| TC-01-25 | P1 | Формат успешного ответа POST | 1. Валидный POST через `CreateItem` (внутри - разбор поля `status` с префиксом «Сохранили объявление - » и UUID в хвосте).<br>2. Проверить длину UUID (36 символов). | фактический контракт | `TestCreateItemSuccessResponseHasStatusWithUUID` |
| TC-01-26 | P2 | Неверная структура и типы JSON | 1. POST с невалидной структурой/типами (отдельный автотест на сценарий).<br>2. Проверить статус 400. | 400 | `TestCreateItemInvalidEmptyStatisticsObject`, `TestCreateItemInvalidMissingViewCountInStats`, `TestCreateItemInvalidMissingContactsInStats`, `TestCreateItemInvalidStatisticsAsString`, `TestCreateItemInvalidNameAsNumber`, `TestCreateItemInvalidSellerIDAsFloat`, `TestCreateItemInvalidPriceAsFloat`, `TestCreateItemInvalidLikesAsString`, `TestCreateItemInvalidViewCountAsString`, `TestCreateItemInvalidContactsAsString`, `TestCreateItemInvalidLikesAsFloat`, `TestCreateItemInvalidViewCountAsFloat`, `TestCreateItemInvalidContactsAsFloat`, `TestCreateItemInvalidSellerIDNull`, `TestCreateItemInvalidNameNull`, `TestCreateItemInvalidPriceNull`, `TestCreateItemInvalidStatisticsNull`, `TestCreateItemInvalidStatisticsAsArray`, `TestCreateItemInvalidBodyAsArray`, `TestCreateItemInvalidLikesNullInStatistics`, `TestCreateItemInvalidViewCountNullInStatistics`, `TestCreateItemInvalidContactsNullInStatistics`, `TestCreateItemInvalidMissingStatisticsField` |
| TC-01-27 | P2 | SQL / XSS в `name` | 1. POST с подозрительной строкой в `name`.<br>2. GET объявления по id.<br>3. Сверить `name`. | 200, значение совпадает | `TestCreateItemSecuritySQLInName`, `TestCreateItemSecurityXSSInName` |
| TC-01-28 | P2 | Null-byte в `name` | 1. POST с `\u0000` внутри `name`.<br>2. Проверить статус (не 200). | не 200; факт: 400 или 500 | `TestCreateItemSecurityNullByteInName` |
| TC-01-29 | P2 | Эмодзи / Unicode в `name` | 1. `CreateItem` с составным эмодзи в `name`.<br>2. Успех подразумевает 200. | 200 | `TestCreateItemUnicodeCompositeEmojiInName` |
| TC-01-30 | P0 | `sellerID=-1` (зависание) | 1. POST с `sellerID=-1`, клиент с таймаутом (см. тест).<br>2. Либо ошибка таймаута (дефект BUG-003), либо **400** без ожидания таймаута (исправление). | таймаут или 400 | `TestCreateItemSellerIDMinusOneHangOrTimeout` |

**BUG-003:** не использовать обычный `CreateItem` / `HttpPostItem` с `sellerID=-1` - запрос без таймаута может повесить прогон. См. `TestCreateItemSellerIDMinusOneHangOrTimeout`.

---

## TC-02: GET `/api/1/item/:id`

| ID | Приоритет | Описание | Шаги | Ожидаемый результат | Автотест |
|----|-----------|----------|------|---------------------|----------|
| TC-02-01 | P0 | E2E: создать и получить, все поля | 1. POST создать объявление.<br>2. GET `/api/1/item/{id}`.<br>3. Сверить поля ответа с созданием. | 200 | `TestGetItemByIDPositive` |
| TC-02-02 | P1 | Несуществующий UUID | 1. GET `/api/1/item/00000000-0000-0000-0000-000000000000`.<br>2. Проверить статус. | 404 | `TestGetItemNonExistentID` |
| TC-02-03 | P2 | Невалидный формат id | 1. GET с id не в формате UUID.<br>2. Проверить статус. | 400 | `TestGetItemInvalidIDNotUUID`, `TestGetItemInvalidIDNumeric`, `TestGetItemInvalidIDWrongLengthUUID` |
| TC-02-04 | P2 | Пустой / пробельный id | 1. GET с id = пробел (URL-encoded).<br>2. Проверить статус. | 400 | `TestGetItemEmptyID` |
| TC-02-05 | P1 | Парсинг `createdAt` | 1. Создать объявление.<br>2. GET по id.<br>3. Распарсить сырую строку `createdAt` известными layout'ами (без предобработки). | соответствует ISO 8601 / RFC3339 | `TestGetItemCreatedAtFormat` (падает при BUG-010) |
| TC-02-06 | P2 | Два GET подряд | 1. Создать объявление.<br>2. GET дважды с тем же id.<br>3. Сравнить тела ответа байт-в-байт. | 200, тело идентично | `TestGetItemDoubleGETSameBody` |
| TC-02-07 | P2 | SQL-подобная строка в path | 1. GET `/api/1/item/{id}` с id вида SQL-инъекции.<br>2. Убедиться, что не 500.<br>3. Проверить статус. | 400, не 500 | `TestGetItemSQLInjectionLikeID` |

---

## TC-02A: DELETE `/api/2/item/:id`

На стенде удаление объявления доступно по **v2** пути; `DELETE /api/1/item/:id` отдаёт **405**.

| ID | Приоритет | Описание | Шаги | Ожидаемый результат | Автотест |
|----|-----------|----------|------|---------------------|----------|
| TC-02A-01 | P1 | Удаление существующего объявления | 1. POST `/api/1/item` создать объявление.<br>2. DELETE `/api/2/item/{id}`.<br>3. GET `/api/1/item/{id}`. | DELETE: 200 или 204; GET: 404 | `TestDeleteItemExisting` |
| TC-02A-02 | P2 | Удаление несуществующего UUID | 1. DELETE `/api/2/item/00000000-0000-0000-0000-000000000000`. | 404 | `TestDeleteItemNonExistentID` |

---

## TC-03: GET `/api/1/:sellerID/item`

| ID | Приоритет | Описание | Шаги | Ожидаемый результат | Автотест |
|----|-----------|----------|------|---------------------|----------|
| TC-03-01 | P0 | Два объявления продавца в списке | 1. Создать 2 объявления с одним sellerID.<br>2. GET `/api/1/{sellerID}/item`.<br>3. Проверить наличие обоих id. | 200 | `TestGetSellerItemsPositive` |
| TC-03-02 | P1 | Все элементы с нужным `sellerId` | 1. Создать несколько объявлений с одним sellerID.<br>2. GET список.<br>3. Для каждого элемента проверить `sellerId`. | 200 | `TestGetSellerItemsAllBelongToSeller` |
| TC-03-03 | P2 | Продавец без объявлений | 1. Взять sellerID из диапазона 1e9–1.9e9 (вне типичных тестовых id), чтобы реже пересекаться с чужими данными на стенде.<br>2. GET `/api/1/{sellerID}/item`.<br>3. Проверить статус и десериализацию в пустой массив. | 200, тело `[]` | `TestGetSellerItemsEmptySeller` |
| TC-03-04 | P1 | Нечисловой `sellerID` | 1. GET с нечисловым сегментом в пути.<br>2. Проверить статус. | 400 | `TestGetSellerItemsInvalidSellerIDLatin`, `TestGetSellerItemsInvalidSellerIDCyrillic` |
| TC-03-05 | P2 | Отрицательный `sellerID` в пути | 1. GET `/api/1/-111111/item`.<br>2. Проверить статус. | спецификация: 400; факт: 200 | `TestGetSellerItemsNegativeSellerID` (+ BUG-006) |
| TC-03-06 | P2 | `sellerID=2^53+1` в пути | 1. GET `/api/1/9007199254740993/item`.<br>2. Проверить статус. | факт: 200 | `TestGetSellerItemsVeryLargeSellerID` |
| TC-03-07 | P2 | Дробный `sellerID` | 1. GET `/api/1/111111.5/item`.<br>2. Проверить статус. | 400 | `TestGetSellerItemsFractionalSellerID` |
| TC-03-08 | P2 | 50 объявлений в выдаче | 1. Создать 50 объявлений с одним sellerID.<br>2. GET список.<br>3. Считать объявления этого продавца. | не менее 50 записей | `TestGetSellerItemsFiftyItemsAllReturned` |
| TC-03-09 | P1 | Изоляция продавцов | 1. Создать объявление у продавца A и у продавца B (разные `sellerID`).<br>2. GET `/api/1/{sellerA}/item`.<br>3. Проверить: в списке есть `id` объявления A, нет `id` объявления B; у всех элементов `sellerId` = sellerA. | 200 | `TestGetSellerItemsExcludesOtherSellers` |

---

## TC-04: GET `/api/1/statistic/:id`

| ID | Приоритет | Описание | Шаги | Ожидаемый результат | Автотест |
|----|-----------|----------|------|---------------------|----------|
| TC-04-01 | P0 | Статистика созданного объявления | 1. POST создать объявление.<br>2. GET `/api/1/statistic/{id}`.<br>3. Проверить поля likes, viewCount, contacts. | 200 | `TestGetStatisticPositive` |
| TC-04-02 | P1 | E2E POST → GET item → GET statistic | 1. POST.<br>2. GET item.<br>3. GET statistic.<br>4. Проверить согласованность. | 200 | `TestGetStatisticFullE2E` |
| TC-04-03 | P1 | Несуществующий UUID | 1. GET statistic для `00000000-0000-0000-0000-000000000000`.<br>2. Проверить статус. | 404 | `TestGetStatisticNonExistentID` |
| TC-04-04 | P2 | Пустой id | 1. GET с пробелом вместо id.<br>2. Проверить статус. | 400 | `TestGetStatisticEmptyID` |
| TC-04-05 | P2 | Невалидный формат id | 1. GET с id не UUID.<br>2. Проверить статус. | 400 | `TestGetStatisticInvalidIDNotUUID`, `TestGetStatisticInvalidIDNumeric` |
| TC-04-06 | P2 | Повторный GET statistic | 1. Создать объявление с известным viewCount.<br>2. GET statistic дважды.<br>3. Сравнить viewCount. | без изменения счётчика | `TestGetStatisticRepeatedGETViewCountStable` |

---

## Нефункциональные проверки (рекомендации)

| ID | Приоритет | Проверка | Шаги | Ожидание | Автотест |
|----|-----------|----------|------|----------|----------|
| NF-01 | P2 | `Content-Type: application/json` | 1. POST `/api/1/item` (200).<br>2. GET item, GET seller items, GET statistic (200).<br>3. GET item несуществующего UUID (404).<br>4. Разобрать `Content-Type` через `mime.ParseMediaType`; media-type = `application/json` (допускается `charset` и др. параметры). | во всех ответах JSON | `TestNF01ContentTypeApplicationJSON` |
| NF-02 | P2 | Время ответа POST | 1. Прогревочный POST `/api/1/item`.<br>2. Замер: от начала запроса до полного чтения тела успешного POST. | &lt; 2 с | `TestNF02PostResponseTimeUnder2Seconds` |
| NF-03 | P2 | Время ответа GET | 1. Создать объявление.<br>2. Прогревочный GET item.<br>3. Замер полного ответа для GET item, GET seller items, GET statistic. | каждый GET &lt; 1 с | `TestNF03GetResponseTimeUnder1Second` |
