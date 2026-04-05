package item

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	httpClient "avito-qa-internship/internal/client/http/item"
	"avito-qa-internship/internal/managers/item/models"
	"avito-qa-internship/internal/utils"
)

// CreateItem создаёт объявление и возвращает его UUID.
// Автоматически заполняет пустые поля дефолтными значениями.
// API возвращает: {"status": "Сохранили объявление - <UUID>"}
func CreateItem(t testing.TB, req models.CreateItemRequest) string {
	t.Helper()
	if req.SellerID == 0 {
		req.SellerID = utils.RandomSellerID()
	}
	if req.Name == "" {
		req.Name = "Тест " + utils.RandomString(8)
	}
	if req.Price == 0 {
		req.Price = 100
	}
	// API требует ненулевые значения для всех полей statistics.
	// Заполняем дефолтами если statistics не задана явно.
	if req.Statistics == (models.Statistics{}) {
		req.Statistics = models.Statistics{Likes: 1, ViewCount: 1, Contacts: 1}
	}

	resp := httpClient.HttpPostItem(t, req, http.StatusOK)
	require.NotNil(t, resp)

	body := httpClient.ReadBody(t, resp)

	var created models.CreateItemStatusResponse
	err := json.Unmarshal(body, &created)
	require.NoError(t, err, "не удалось десериализовать ответ на создание объявления: %s", string(body))
	require.NotEmpty(t, created.Status, "поле status в ответе не должно быть пустым")

	id := parseUUIDFromStatus(created.Status)
	require.NotEmpty(t, id, "не удалось извлечь UUID из status: %q", created.Status)

	CleanupDeleteItem(t, id)
	return id
}

// ItemIDFromCreateOKBody извлекает UUID из JSON-ответа успешного POST /api/1/item (поле status).
func ItemIDFromCreateOKBody(body string) string {
	var created models.CreateItemStatusResponse
	if err := json.Unmarshal([]byte(body), &created); err != nil {
		return ""
	}
	return parseUUIDFromStatus(created.Status)
}

// CleanupDeleteItem регистрирует best-effort DELETE объявления после теста (t.Cleanup).
func CleanupDeleteItem(t testing.TB, itemID string) {
	t.Helper()
	itemID = strings.TrimSpace(itemID)
	if itemID == "" {
		return
	}
	t.Cleanup(func() {
		TryDeleteItemByID(t, itemID)
	})
}

// ScheduleDeleteIfCreatedOK планирует удаление, если ответ создания — 200 и в теле есть UUID.
func ScheduleDeleteIfCreatedOK(t testing.TB, statusCode int, body string) {
	t.Helper()
	if statusCode != http.StatusOK {
		return
	}
	if id := ItemIDFromCreateOKBody(body); id != "" {
		CleanupDeleteItem(t, id)
	}
}

// TryDeleteItemByID вызывает DELETE /api/2/item/:id. Не падает при ошибке: 200/204/404 — ок;
// 405 — метод не реализован на стенде; иначе сообщение в лог.
func TryDeleteItemByID(t testing.TB, itemID string) {
	t.Helper()
	itemID = strings.TrimSpace(itemID)
	if itemID == "" {
		return
	}
	resp := httpClient.HttpDeleteItemByID(t, itemID)
	if resp == nil {
		return
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)

	switch resp.StatusCode {
	case http.StatusOK, http.StatusNoContent, http.StatusNotFound:
		return
	case http.StatusMethodNotAllowed:
		return
	default:
		t.Logf("cleanup DELETE /api/2/item/%s: статус %d (объявление могло остаться на стенде)", itemID, resp.StatusCode)
	}
}

// parseUUIDFromStatus извлекает UUID из строки вида "Сохранили объявление - <UUID>".
func parseUUIDFromStatus(status string) string {
	const prefix = "Сохранили объявление - "
	idx := strings.Index(status, prefix)
	if idx == -1 {
		return ""
	}
	return strings.TrimSpace(status[idx+len(prefix):])
}

// CreateItemRaw отправляет запрос на создание объявления и возвращает статус и тело.
// Статус-код НЕ проверяется внутри функции — проверка происходит на стороне вызывающего.
func CreateItemRaw(t testing.TB, req models.CreateItemRequest) (int, string) {
	t.Helper()
	resp := httpClient.HttpPostItemRaw(t, marshalRequest(t, req))
	require.NotNil(t, resp)
	body := httpClient.ReadBody(t, resp)
	return resp.StatusCode, string(body)
}

// marshalRequest сериализует запрос в JSON-строку.
func marshalRequest(t testing.TB, req models.CreateItemRequest) string {
	t.Helper()
	b, err := json.Marshal(req)
	require.NoError(t, err)
	return string(b)
}

// CreateItemRawBody отправляет POST с произвольным телом без проверки статуса.
func CreateItemRawBody(t testing.TB, rawBody string) (int, string) {
	t.Helper()
	resp := httpClient.HttpPostItemRaw(t, rawBody)
	require.NotNil(t, resp)
	body := httpClient.ReadBody(t, resp)
	return resp.StatusCode, string(body)
}

// CreateItemRawBodyWithTimeout отправляет POST /api/1/item с таймаутом клиента (для сценариев зависания сервера).
func CreateItemRawBodyWithTimeout(t testing.TB, rawBody string, timeout time.Duration) (statusCode int, body string, err error) {
	t.Helper()
	return httpClient.HttpPostItemRawWithTimeout(t, rawBody, timeout)
}

// GetItem получает объявление по ID и возвращает десериализованный ответ.
func GetItem(t testing.TB, id string) models.ItemResponse {
	t.Helper()
	resp := httpClient.HttpGetItemByID(t, id, http.StatusOK)
	require.NotNil(t, resp)

	body := httpClient.ReadBody(t, resp)

	var items []models.ItemResponse
	err := json.Unmarshal(body, &items)
	require.NoError(t, err, "не удалось десериализовать ответ на получение объявления")
	require.NotEmpty(t, items, "список объявлений не должен быть пустым")

	return items[0]
}

// DeleteItemRaw вызывает DELETE /api/2/item/:id и возвращает статус и тело (без фиксированного ожидания).
func DeleteItemRaw(t testing.TB, id string) (int, string) {
	t.Helper()
	resp := httpClient.HttpDeleteItemByID(t, id)
	require.NotNil(t, resp)
	body := httpClient.ReadBody(t, resp)
	return resp.StatusCode, string(body)
}

// GetItemRaw получает объявление по ID и возвращает статус и тело ответа.
func GetItemRaw(t testing.TB, id string, expectedStatus int) (int, string) {
	t.Helper()
	resp := httpClient.HttpGetItemByID(t, id, expectedStatus)
	require.NotNil(t, resp)
	body := httpClient.ReadBody(t, resp)
	return resp.StatusCode, string(body)
}

// GetSellerItems получает список объявлений продавца.
func GetSellerItems(t testing.TB, sellerID int) []models.ItemResponse {
	t.Helper()
	resp := httpClient.HttpGetSellerItems(t, sellerID, http.StatusOK)
	require.NotNil(t, resp)

	body := httpClient.ReadBody(t, resp)

	var items []models.ItemResponse
	err := json.Unmarshal(body, &items)
	require.NoError(t, err, "не удалось десериализовать список объявлений продавца")

	return items
}

// GetSellerItemsRaw получает объявления продавца и возвращает статус и тело.
func GetSellerItemsRaw(t testing.TB, sellerID int, expectedStatus int) (int, string) {
	t.Helper()
	resp := httpClient.HttpGetSellerItems(t, sellerID, expectedStatus)
	require.NotNil(t, resp)
	body := httpClient.ReadBody(t, resp)
	return resp.StatusCode, string(body)
}

// GetSellerItemsByRawID получает объявления по произвольному строковому ID (для негативных тестов).
func GetSellerItemsByRawID(t testing.TB, rawID string) (int, string) {
	t.Helper()
	resp := httpClient.HttpGetSellerItemsByRawID(t, rawID)
	require.NotNil(t, resp)
	body := httpClient.ReadBody(t, resp)
	return resp.StatusCode, string(body)
}

// GetStatistic получает статистику объявления по ID.
func GetStatistic(t testing.TB, id string) []models.StatisticResponse {
	t.Helper()
	resp := httpClient.HttpGetStatisticByID(t, id, http.StatusOK)
	require.NotNil(t, resp)

	body := httpClient.ReadBody(t, resp)

	var stats []models.StatisticResponse
	err := json.Unmarshal(body, &stats)
	require.NoError(t, err, "не удалось десериализовать статистику")

	return stats
}

// GetStatisticRaw получает статистику и возвращает статус и тело.
func GetStatisticRaw(t testing.TB, id string, expectedStatus int) (int, string) {
	t.Helper()
	resp := httpClient.HttpGetStatisticByID(t, id, expectedStatus)
	require.NotNil(t, resp)
	body := httpClient.ReadBody(t, resp)
	return resp.StatusCode, string(body)
}
