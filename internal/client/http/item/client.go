package item

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"avito-qa-internship/internal/constants/path"
	apiRunner "avito-qa-internship/internal/helpers/api-runner"
	"avito-qa-internship/internal/managers/item/models"
)

// HttpPostItem отправляет POST /api/1/item и ожидает указанный HTTP-статус.
func HttpPostItem(t testing.TB, req models.CreateItemRequest, expectedStatus int) *http.Response {
	t.Helper()
	body, err := json.Marshal(req)
	require.NoError(t, err, "не удалось сериализовать тело запроса")

	return apiRunner.GetRunner().Create().
		Post(path.CreateItemPath).
		Body(string(body)).
		Expect(t).
		Status(expectedStatus).
		End().Response
}

// HttpPostItemRaw отправляет POST /api/1/item с произвольным телом без проверки статуса.
func HttpPostItemRaw(t testing.TB, rawBody string) *http.Response {
	t.Helper()
	return apiRunner.GetRunner().Create().
		Post(path.CreateItemPath).
		Body(rawBody).
		Expect(t).
		End().Response
}

// HttpPostItemRawWithTimeout отправляет POST /api/1/item с ограничением времени на весь обмен (ожидание заголовков и тела).
// При зависании сервера возвращает ошибку с Timeout() == true или обёртку вокруг неё.
func HttpPostItemRawWithTimeout(t testing.TB, rawBody string, timeout time.Duration) (statusCode int, body string, err error) {
	t.Helper()
	baseStr := os.Getenv("API_URL")
	if baseStr == "" {
		return 0, "", fmt.Errorf("API_URL не задан")
	}
	base, err := url.Parse(baseStr)
	if err != nil {
		return 0, "", fmt.Errorf("разбор API_URL: %w", err)
	}
	full := base.ResolveReference(&url.URL{Path: path.CreateItemPath})

	req, err := http.NewRequest(http.MethodPost, full.String(), bytes.NewReader([]byte(rawBody)))
	if err != nil {
		return 0, "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()
	b, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return resp.StatusCode, "", readErr
	}
	return resp.StatusCode, string(b), nil
}

// HttpDeleteItemByID отправляет DELETE /api/2/item/:id без проверки статуса (очистка после теста).
func HttpDeleteItemByID(t testing.TB, id string) *http.Response {
	t.Helper()
	return apiRunner.GetRunner().Create().
		Delete(path.GetDeleteItemByIDPath(id)).
		Expect(t).
		End().Response
}

// HttpGetItemByID отправляет GET /api/1/item/:id и ожидает указанный HTTP-статус.
func HttpGetItemByID(t testing.TB, id string, expectedStatus int) *http.Response {
	t.Helper()
	return apiRunner.GetRunner().Create().
		Get(path.GetItemByIDPath(id)).
		Expect(t).
		Status(expectedStatus).
		End().Response
}

// HttpGetSellerItems отправляет GET /api/1/:sellerID/item и ожидает указанный HTTP-статус.
func HttpGetSellerItems(t testing.TB, sellerID int, expectedStatus int) *http.Response {
	t.Helper()
	return apiRunner.GetRunner().Create().
		Get(path.GetSellerItemsPath(sellerID)).
		Expect(t).
		Status(expectedStatus).
		End().Response
}

// HttpGetSellerItemsByRawID отправляет GET /api/1/{rawID}/item без числового приведения.
// Используется в негативных тестах с нечисловым sellerID.
func HttpGetSellerItemsByRawID(t testing.TB, rawID string) *http.Response {
	t.Helper()
	return apiRunner.GetRunner().Create().
		Get("/api/1/" + rawID + "/item").
		Expect(t).
		End().Response
}

// HttpGetStatisticByID отправляет GET /api/1/statistic/:id и ожидает указанный HTTP-статус.
func HttpGetStatisticByID(t testing.TB, id string, expectedStatus int) *http.Response {
	t.Helper()
	return apiRunner.GetRunner().Create().
		Get(path.GetStatisticByIDPath(id)).
		Expect(t).
		Status(expectedStatus).
		End().Response
}

// readBody — вспомогательная функция для чтения тела ответа в тестах.
func readBody(t testing.TB, resp *http.Response) []byte {
	t.Helper()
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(resp.Body)
	require.NoError(t, err, "не удалось прочитать тело ответа")
	return buf.Bytes()
}

// ReadBody читает тело http.Response.
func ReadBody(t testing.TB, resp *http.Response) []byte {
	return readBody(t, resp)
}
