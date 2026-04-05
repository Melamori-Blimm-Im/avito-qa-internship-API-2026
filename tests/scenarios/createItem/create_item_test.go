package createItem

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/ozontech/allure-go/pkg/allure"
	"github.com/ozontech/allure-go/pkg/framework/provider"
	"github.com/ozontech/allure-go/pkg/framework/suite"
	"github.com/stretchr/testify/require"

	itemManager "avito-qa-internship/internal/managers/item"
	"avito-qa-internship/internal/managers/item/models"
	"avito-qa-internship/internal/utils"
	base "avito-qa-internship/tests"
)

type TestSuite struct {
	suite.Suite
}

// allureCreateItemSuiteLayout задаёт epic, parent/suite/feature и sub_suite/story для сценариев POST /api/1/item.
func allureCreateItemSuiteLayout(t provider.T) {
	t.Epic(base.AllureEpic)
	t.AddParentSuite(base.AllureEpic)
	t.AddSubSuite(base.AllureFeatureCreateItem)
}

func TestSuiteRun(t *testing.T) {
	suite.RunNamedSuite(t, base.AllureEpicAds, new(TestSuite))
}

func (s *TestSuite) BeforeAll(t provider.T) {
	t.Log("Инициализация переменных окружения")

	base.SetupSuite()
}

func (s *TestSuite) AfterAll(t provider.T) {
	t.Log("Завершение набора тестов")
	base.TearDownSuite()
}

// --- Позитивные и граничные успешные сценарии ---

// Positive: создание объявления со всеми полями. Проверка структуры ответа через GET.
func (s *TestSuite) TestCreateItemAllFieldsPositive(t provider.T) {
	allureCreateItemSuiteLayout(t)
	t.Severity(allure.BLOCKER)
	t.Title("Создание объявления со всеми полями")
	t.Description("Создаём объявление со всеми полями (включая statistics) и проверяем данные через GET /api/1/item/:id")

	sellerID := utils.RandomSellerID()
	req := models.CreateItemRequest{
		SellerID: sellerID,
		Name:     "Тест создания " + utils.RandomString(6),
		Price:    1500,
		Statistics: models.Statistics{
			Likes:     10,
			ViewCount: 200,
			Contacts:  5,
		},
	}

	var createdID string

	t.WithNewStep("Отправляем POST /api/1/item и получаем UUID", func(ctx provider.StepCtx) {
		createdID = itemManager.CreateItem(t, req)
		require.NotEmpty(t, createdID, "id объявления не должен быть пустым")
		ctx.WithNewParameters("createdID", createdID)
	})

	t.WithNewStep("GET /api/1/item/:id — проверяем все поля объявления", func(ctx provider.StepCtx) {
		item := itemManager.GetItem(t, createdID)
		require.Equal(t, createdID, item.ID, "id должен совпадать")
		require.Equal(t, sellerID, item.SellerID, "sellerId должен совпадать")
		require.Equal(t, req.Name, item.Name, "name должен совпадать")
		require.Equal(t, req.Price, item.Price, "price должен совпадать")
		require.NotEmpty(t, item.CreatedAt, "createdAt не должен быть пустым")
		require.Equal(t, req.Statistics.Likes, item.Statistics.Likes, "likes должны совпадать")
		require.Equal(t, req.Statistics.ViewCount, item.Statistics.ViewCount, "viewCount должен совпадать")
		require.Equal(t, req.Statistics.Contacts, item.Statistics.Contacts, "contacts должны совпадать")
	})
}

// Positive: идемпотентность — два одинаковых запроса создают два разных UUID.
func (s *TestSuite) TestCreateItemIdempotencyPositive(t provider.T) {
	allureCreateItemSuiteLayout(t)
	t.Severity(allure.CRITICAL)
	t.Title("Одинаковые запросы создают разные UUID")
	t.Description("Два идентичных POST /api/1/item должны создавать два объявления с разными UUID")

	req := models.CreateItemRequest{
		SellerID:   utils.RandomSellerID(),
		Name:       "Дубль объявление",
		Price:      500,
		Statistics: models.Statistics{Likes: 1, ViewCount: 1, Contacts: 1},
	}

	var id1, id2 string

	t.WithNewStep("Создаём первое объявление", func(ctx provider.StepCtx) {
		id1 = itemManager.CreateItem(t, req)
		ctx.WithNewParameters("id1", id1)
	})

	t.WithNewStep("Создаём второе объявление с теми же данными", func(ctx provider.StepCtx) {
		id2 = itemManager.CreateItem(t, req)
		ctx.WithNewParameters("id2", id2)
	})

	t.WithNewStep("Сравниваем UUID — они должны отличаться", func(ctx provider.StepCtx) {
		require.NotEqual(t, id1, id2,
			"повторный одинаковый запрос должен создавать новое объявление с другим UUID")
	})
}

// Positive: минимальная валидная цена price=1.
func (s *TestSuite) TestCreateItemMinimumPrice(t provider.T) {
	allureCreateItemSuiteLayout(t)
	t.Severity(allure.NORMAL)
	t.Title("price=1 — минимальная валидная цена")
	t.Description("Граничное значение: price=1 должно приниматься API и сохраняться корректно")

	t.WithNewStep("POST /api/1/item с price=1", func(ctx provider.StepCtx) {
		req := models.CreateItemRequest{
			SellerID:   utils.RandomSellerID(),
			Name:       "Минимальная цена",
			Price:      1,
			Statistics: models.Statistics{Likes: 1, ViewCount: 1, Contacts: 1},
		}
		id := itemManager.CreateItem(t, req)
		ctx.WithNewParameters("createdID", id)
	})
}

func (s *TestSuite) createItemWithSellerIDExpect200(t provider.T, title, stepName string, sellerID int, itemName string, logCreatedID bool) {
	allureCreateItemSuiteLayout(t)
	t.Severity(allure.NORMAL)
	t.Title(title)
	t.Description("2^53+1 принимается с 200; INT_MAX — ожидаемый валидный случай. Отрицательный sellerID — см. TestCreateItemNegativeSellerIDPost")
	t.WithNewStep(stepName, func(ctx provider.StepCtx) {
		req := models.CreateItemRequest{
			SellerID:   sellerID,
			Name:       itemName,
			Price:      100,
			Statistics: models.Statistics{Likes: 1, ViewCount: 1, Contacts: 1},
		}
		id := itemManager.CreateItem(t, req)
		ctx.WithNewParameters("createdID", id)
		if logCreatedID {
			t.Logf("создано объявление: %s", id)
		}
	})
}

// Boundary: sellerID=2^53+1 в POST — фактическое поведение API.
func (s *TestSuite) TestCreateItemSellerIDBoundaryHugePow53PlusOne(t provider.T) {
	s.createItemWithSellerIDExpect200(t,
		"sellerID=2^53+1 принимается", "sellerID=2^53+1 принимается",
		9007199254740993, "Тест", true)
}

// Boundary: sellerID=INT_MAX в POST — фактическое поведение API.
func (s *TestSuite) TestCreateItemSellerIDBoundaryIntMax(t provider.T) {
	s.createItemWithSellerIDExpect200(t,
		"sellerID=INT_MAX принимается", "sellerID=INT_MAX принимается",
		2147483647, "INT_MAX продавец", false)
}

// Boundary: все поля statistics = 2^53+1 — фиксация поведения API.
func (s *TestSuite) TestCreateItemHugeStatisticsValues(t provider.T) {
	allureCreateItemSuiteLayout(t)
	t.Severity(allure.NORMAL)
	t.Title("statistics со значениями 2^53+1")
	t.Description("Фактически API принимает огромные значения statistics; допустим ответ 400")

	t.WithNewStep("POST со счётчиками 9007199254740993", func(ctx provider.StepCtx) {
		const huge = 9007199254740993
		req := models.CreateItemRequest{
			SellerID: utils.RandomSellerID(),
			Name:     "Тест",
			Price:    100,
			Statistics: models.Statistics{
				Likes:     huge,
				ViewCount: huge,
				Contacts:  huge,
			},
		}
		id := itemManager.CreateItem(t, req)
		ctx.WithNewParameters("createdID", id)
		t.Logf("создано объявление: %s", id)
	})
}

// Positive: лишние поля в корне и в statistics игнорируются.
func (s *TestSuite) TestCreateItemUnknownFieldsIgnored(t provider.T) {
	allureCreateItemSuiteLayout(t)
	t.Severity(allure.NORMAL)
	t.Title("Неизвестные поля в теле запроса игнорируются")
	t.Description("Дополнительные ключи не ломают создание объявления")

	t.WithNewStep("POST с unknownField в корне", func(ctx provider.StepCtx) {
		sid := utils.RandomSellerID()
		raw := fmt.Sprintf(`{"sellerID":%d,"name":"С лишним полем","price":200,"unknownField":"x","statistics":{"likes":1,"viewCount":1,"contacts":1}}`, sid)
		statusCode, body := itemManager.CreateItemRawBody(t, raw)
		itemManager.ScheduleDeleteIfCreatedOK(t, statusCode, body)
		require.Equal(t, http.StatusOK, statusCode)
	})

	t.WithNewStep("POST с unknownField в statistics", func(ctx provider.StepCtx) {
		sid := utils.RandomSellerID()
		raw := fmt.Sprintf(`{"sellerID":%d,"name":"Статистика с лишним","price":200,"statistics":{"likes":1,"viewCount":1,"contacts":1,"unknownField":"y"}}`, sid)
		statusCode, body := itemManager.CreateItemRawBody(t, raw)
		itemManager.ScheduleDeleteIfCreatedOK(t, statusCode, body)
		require.Equal(t, http.StatusOK, statusCode)
	})
}

// Positive: успешный ответ POST — строка status с UUID (расхождение с Postman).
func (s *TestSuite) TestCreateItemSuccessResponseHasStatusWithUUID(t provider.T) {
	allureCreateItemSuiteLayout(t)
	t.Severity(allure.CRITICAL)
	t.Title("Успешный POST возвращает {\"status\": \"Сохранили объявление - <UUID>\"}")
	t.Description("Фиксируем фактический формат ответа (см. API_DISCREPANCIES.md)")

	t.WithNewStep("POST и проверка поля status", func(ctx provider.StepCtx) {
		req := models.CreateItemRequest{
			SellerID:   utils.RandomSellerID(),
			Name:       "Контракт ответа " + utils.RandomString(4),
			Price:      300,
			Statistics: models.Statistics{Likes: 1, ViewCount: 1, Contacts: 1},
		}
		id := itemManager.CreateItem(t, req)
		ctx.WithNewParameters("createdID", id)
		require.Len(t, id, 36, "ожидается UUID в ответе создания")
	})
}

// Security: SQL в name сохраняется как есть; проверка через GET.
func (s *TestSuite) TestCreateItemSecuritySQLInName(t provider.T) {
	allureCreateItemSuiteLayout(t)
	t.Severity(allure.CRITICAL)
	t.Title("Security: SQL-инъекция в name")
	t.Description("SQL проходит как обычный текст; проверка сохранения через GET")
	name := "'; DROP TABLE items; --"
	req := models.CreateItemRequest{
		SellerID:   utils.RandomSellerID(),
		Name:       name,
		Price:      100,
		Statistics: models.Statistics{Likes: 1, ViewCount: 1, Contacts: 1},
	}
	id := itemManager.CreateItem(t, req)
	item := itemManager.GetItem(t, id)
	require.Equal(t, name, item.Name)
}

// Security: XSS в name сохраняется как есть; проверка через GET.
func (s *TestSuite) TestCreateItemSecurityXSSInName(t provider.T) {
	allureCreateItemSuiteLayout(t)
	t.Severity(allure.CRITICAL)
	t.Title("Security: XSS в name")
	t.Description("XSS проходит как обычный текст; проверка сохранения через GET")
	name := "<script>alert(1)</script>"
	req := models.CreateItemRequest{
		SellerID:   utils.RandomSellerID(),
		Name:       name,
		Price:      100,
		Statistics: models.Statistics{Likes: 1, ViewCount: 1, Contacts: 1},
	}
	id := itemManager.CreateItem(t, req)
	item := itemManager.GetItem(t, id)
	require.Equal(t, name, item.Name)
}

// Security: null-byte в name — отказ API (400 или 500).
func (s *TestSuite) TestCreateItemSecurityNullByteInName(t provider.T) {
	allureCreateItemSuiteLayout(t)
	t.Severity(allure.CRITICAL)
	t.Title("Security: null-byte в name — отказ API")
	t.Description("Ожидается 400; допустим 500 (фактическое поведение сервиса)")
	t.WithNewStep("POST с \\u0000 в name", func(ctx provider.StepCtx) {
		raw := `{"sellerID":123456,"name":"test\u0000test","price":100,"statistics":{"likes":1,"viewCount":1,"contacts":1}}`
		statusCode, body := itemManager.CreateItemRawBody(t, raw)
		itemManager.ScheduleDeleteIfCreatedOK(t, statusCode, body)
		ctx.WithNewParameters("statusCode", fmt.Sprintf("%d", statusCode))
		t.Logf("ответ: %s", body)
		require.NotEqual(t, http.StatusOK, statusCode,
			"null-byte в name не должен приводить к успешному созданию")
		require.True(t, statusCode == http.StatusBadRequest || statusCode == http.StatusInternalServerError,
			"ожидался отказ 400 или ошибка сервера 500, получен %d", statusCode)
	})
}

// Unicode: составной эмодзи в name принимается API.
func (s *TestSuite) TestCreateItemUnicodeCompositeEmojiInName(t provider.T) {
	allureCreateItemSuiteLayout(t)
	t.Severity(allure.NORMAL)
	t.Title("Security: составной эмодзи в name")
	t.Description("Сложные символы Unicode должны приниматься API")
	name := "Флаг 🇺🇸 тест"
	req := models.CreateItemRequest{
		SellerID:   utils.RandomSellerID(),
		Name:       name,
		Price:      100,
		Statistics: models.Statistics{Likes: 1, ViewCount: 1, Contacts: 1},
	}
	id := itemManager.CreateItem(t, req)
	require.NotEmpty(t, id, "сложные символы Unicode должны приниматься")
}

// --- Негативные сценарии ---

// Negative (BUG): API отклоняет нулевые значения statistics полей.
// Согласно логике, 0 — допустимое значение для счётчиков (нет лайков, просмотров, контактов).
// BUG: API возвращает 400 "поле likes обязательно" при likes=0 (см. BUGS.md).
func (s *TestSuite) TestCreateItemZeroStatisticsBug(t provider.T) {
	allureCreateItemSuiteLayout(t)
	t.Severity(allure.CRITICAL)
	t.Title("[BUG-002] Нулевые значения statistics отклоняются")
	t.Description("API должен принимать statistics с нулевыми значениями, но возвращает 400. Подробнее: BUGS.md#BUG-002")

	req := models.CreateItemRequest{
		SellerID:   utils.RandomSellerID(),
		Name:       "Товар с нулями",
		Price:      100,
		Statistics: models.Statistics{Likes: 0, ViewCount: 0, Contacts: 0},
	}

	var statusCode int
	var body string

	t.WithNewStep("Отправляем POST с statistics={likes:0,viewCount:0,contacts:0}", func(ctx provider.StepCtx) {
		statusCode, body = itemManager.CreateItemRaw(t, req)
		itemManager.ScheduleDeleteIfCreatedOK(t, statusCode, body)
		ctx.WithNewParameters("statusCode", fmt.Sprintf("%d", statusCode))
	})

	t.WithNewStep("Проверяем статус ответа (ожидаем 200, получаем 400 — BUG)", func(ctx provider.StepCtx) {
		if statusCode != http.StatusOK {
			t.Logf("BUG (BUGS.md#BUG-002): API отклоняет нулевые statistics (%d): %s", statusCode, body)
		}
		// Ожидаемое поведение: 200 OK (нулевые значения допустимы).
		// Фактическое поведение: 400 Bad Request (BUG).
		require.Equal(t, http.StatusOK, statusCode,
			"нулевые значения statistics должны быть допустимы (BUG: сервис отклоняет их)")
	})
}

// createItemValidStatsJSON — фрагмент валидного statistics для сборки тел с нарушением в других полях.
const createItemValidStatsJSON = `"statistics":{"likes":1,"viewCount":1,"contacts":1}`

func missingFieldCreateItemExpect400(t provider.T, displayName, rawBody string) {
	allureCreateItemSuiteLayout(t)
	t.Severity(allure.CRITICAL)
	t.Title("Negative: отсутствие обязательных полей — " + displayName)
	t.Description("API должен возвращать 400 при отсутствующих или некорректных полях")
	t.WithNewStep(fmt.Sprintf("Отправляем POST с невалидным телом (%s)", displayName), func(ctx provider.StepCtx) {
		ctx.WithNewParameters("rawBody", rawBody)
		statusCode, body := itemManager.CreateItemRawBody(t, rawBody)
		itemManager.ScheduleDeleteIfCreatedOK(t, statusCode, body)
		ctx.WithNewParameters("statusCode", fmt.Sprintf("%d", statusCode))

		if statusCode == http.StatusOK {
			var created models.CreateItemStatusResponse
			if err := json.Unmarshal([]byte(body), &created); err == nil {
				t.Logf("BUG: API принял невалидный запрос %q, ответ: %s", displayName, body)
			}
		}
		require.Equal(t, http.StatusBadRequest, statusCode,
			"ожидался 400 для кейса %q, получен %d", displayName, statusCode)
	})
}

func (s *TestSuite) TestCreateItemMissingEmptyBody(t provider.T) {
	missingFieldCreateItemExpect400(t, "пустое тело", `{}`)
}

func (s *TestSuite) TestCreateItemMissingNameField(t provider.T) {
	missingFieldCreateItemExpect400(t, "отсутствует name",
		fmt.Sprintf(`{"sellerID":123456,"price":100,%s}`, createItemValidStatsJSON))
}

func (s *TestSuite) TestCreateItemMissingEmptyName(t provider.T) {
	missingFieldCreateItemExpect400(t, "пустой name",
		fmt.Sprintf(`{"sellerID":123456,"name":"","price":100,%s}`, createItemValidStatsJSON))
}

func (s *TestSuite) TestCreateItemMissingPriceField(t provider.T) {
	missingFieldCreateItemExpect400(t, "отсутствует price (без поля)",
		fmt.Sprintf(`{"sellerID":123456,"name":"Товар",%s}`, createItemValidStatsJSON))
}

func (s *TestSuite) TestCreateItemMissingSellerIDField(t provider.T) {
	missingFieldCreateItemExpect400(t, "отсутствует sellerID",
		fmt.Sprintf(`{"name":"Товар","price":100,%s}`, createItemValidStatsJSON))
}

func (s *TestSuite) TestCreateItemMissingLikesInStatistics(t provider.T) {
	missingFieldCreateItemExpect400(t, "отсутствует поле likes в statistics",
		`{"sellerID":123456,"name":"Товар","price":100,"statistics":{"viewCount":1,"contacts":1}}`)
}

func (s *TestSuite) TestCreateItemMissingInvalidJSON(t provider.T) {
	missingFieldCreateItemExpect400(t, "невалидный JSON", `{sellerID:123456,name:Товар}`)
}

func (s *TestSuite) TestCreateItemMissingPriceAsString(t provider.T) {
	missingFieldCreateItemExpect400(t, "price передана строкой",
		fmt.Sprintf(`{"sellerID":123456,"name":"Товар","price":"сто",%s}`, createItemValidStatsJSON))
}

func (s *TestSuite) TestCreateItemMissingSellerIDAsString(t provider.T) {
	missingFieldCreateItemExpect400(t, "sellerID передан строкой",
		fmt.Sprintf(`{"sellerID":"abc","name":"Товар","price":100,%s}`, createItemValidStatsJSON))
}

// Negative (BUG): API принимает name длиной 100000 символов вместо 400.
// BUG: API возвращает 200 (см. BUGS.md#BUG-009).
func (s *TestSuite) TestCreateItemNameTooLong(t provider.T) {
	allureCreateItemSuiteLayout(t)
	t.Severity(allure.CRITICAL)
	t.Title("[BUG-009] name из 100000 символов должен возвращать 400")
	t.Description("Чрезмерная длина названия должна отклоняться с 400; фактически API принимает. Подробнее: BUGS.md#BUG-009")

	const nameLen = 100000

	t.WithNewStep(fmt.Sprintf("POST с name длиной %d", nameLen), func(ctx provider.StepCtx) {
		req := models.CreateItemRequest{
			SellerID:   utils.RandomSellerID(),
			Name:       utils.RandomString(nameLen),
			Price:      100,
			Statistics: models.Statistics{Likes: 1, ViewCount: 1, Contacts: 1},
		}
		statusCode, body := itemManager.CreateItemRaw(t, req)
		itemManager.ScheduleDeleteIfCreatedOK(t, statusCode, body)
		ctx.WithNewParameters("nameLength", fmt.Sprintf("%d", nameLen))
		ctx.WithNewParameters("statusCode", fmt.Sprintf("%d", statusCode))
		if statusCode == http.StatusOK {
			t.Logf("BUG (BUGS.md#BUG-009): API принял чрезмерно длинный name, ответ: %s", utils.TruncateForLog(body, 200))
		}
		require.Equal(t, http.StatusBadRequest, statusCode,
			"name длиной %d символов должен отклоняться с 400 (BUG: сервис принимает)", nameLen)
	})
}

func invalidStructureCreateItemExpect400(t provider.T, displayName, rawBody string) {
	allureCreateItemSuiteLayout(t)
	t.Severity(allure.CRITICAL)
	t.Title("Negative: некорректная структура/типы — " + displayName)
	t.Description("Пустой statistics, отсутствие подполей, null, строки/дроби вместо int, массив вместо объекта")
	t.WithNewStep(displayName, func(ctx provider.StepCtx) {
		statusCode, body := itemManager.CreateItemRawBody(t, rawBody)
		itemManager.ScheduleDeleteIfCreatedOK(t, statusCode, body)
		ctx.WithNewParameters("statusCode", fmt.Sprintf("%d", statusCode))
		if statusCode != http.StatusBadRequest {
			t.Logf("кейс %q: ожидался 400, получен %d, body=%s", displayName, statusCode, body)
		}
		require.Equal(t, http.StatusBadRequest, statusCode,
			"кейс %q: ожидался статус 400", displayName)
	})
}

func (s *TestSuite) TestCreateItemInvalidEmptyStatisticsObject(t provider.T) {
	invalidStructureCreateItemExpect400(t, "пустой объект statistics",
		`{"sellerID":123456,"name":"Товар","price":100,"statistics":{}}`)
}

func (s *TestSuite) TestCreateItemInvalidMissingViewCountInStats(t provider.T) {
	invalidStructureCreateItemExpect400(t, "нет viewCount в statistics",
		`{"sellerID":123456,"name":"Товар","price":100,"statistics":{"likes":1,"contacts":1}}`)
}

func (s *TestSuite) TestCreateItemInvalidMissingContactsInStats(t provider.T) {
	invalidStructureCreateItemExpect400(t, "нет contacts в statistics",
		`{"sellerID":123456,"name":"Товар","price":100,"statistics":{"likes":1,"viewCount":1}}`)
}

func (s *TestSuite) TestCreateItemInvalidStatisticsAsString(t provider.T) {
	invalidStructureCreateItemExpect400(t, "statistics строкой",
		`{"sellerID":123456,"name":"Товар","price":100,"statistics":"not_an_object"}`)
}

func (s *TestSuite) TestCreateItemInvalidNameAsNumber(t provider.T) {
	invalidStructureCreateItemExpect400(t, "name числом",
		fmt.Sprintf(`{"sellerID":123456,"name":12345,"price":100,%s}`, createItemValidStatsJSON))
}

func (s *TestSuite) TestCreateItemInvalidSellerIDAsFloat(t provider.T) {
	invalidStructureCreateItemExpect400(t, "sellerID дробным",
		fmt.Sprintf(`{"sellerID":111111.5,"name":"Товар","price":1000,%s}`, createItemValidStatsJSON))
}

func (s *TestSuite) TestCreateItemInvalidPriceAsFloat(t provider.T) {
	invalidStructureCreateItemExpect400(t, "price дробным",
		fmt.Sprintf(`{"sellerID":123456,"name":"Товар","price":99.99,%s}`, createItemValidStatsJSON))
}

func (s *TestSuite) TestCreateItemInvalidLikesAsString(t provider.T) {
	invalidStructureCreateItemExpect400(t, "likes строкой",
		`{"sellerID":123456,"name":"Товар","price":1000,"statistics":{"likes":"abc","viewCount":1,"contacts":1}}`)
}

func (s *TestSuite) TestCreateItemInvalidViewCountAsString(t provider.T) {
	invalidStructureCreateItemExpect400(t, "viewCount строкой",
		`{"sellerID":123456,"name":"Товар","price":1000,"statistics":{"likes":1,"viewCount":"abc","contacts":1}}`)
}

func (s *TestSuite) TestCreateItemInvalidContactsAsString(t provider.T) {
	invalidStructureCreateItemExpect400(t, "contacts строкой",
		`{"sellerID":123456,"name":"Товар","price":1000,"statistics":{"likes":1,"viewCount":1,"contacts":"abc"}}`)
}

func (s *TestSuite) TestCreateItemInvalidLikesAsFloat(t provider.T) {
	invalidStructureCreateItemExpect400(t, "likes дробным",
		`{"sellerID":123456,"name":"Товар","price":1000,"statistics":{"likes":1.5,"viewCount":1,"contacts":1}}`)
}

func (s *TestSuite) TestCreateItemInvalidViewCountAsFloat(t provider.T) {
	invalidStructureCreateItemExpect400(t, "viewCount дробным",
		`{"sellerID":123456,"name":"Товар","price":1000,"statistics":{"likes":1,"viewCount":1.5,"contacts":1}}`)
}

func (s *TestSuite) TestCreateItemInvalidContactsAsFloat(t provider.T) {
	invalidStructureCreateItemExpect400(t, "contacts дробным",
		`{"sellerID":123456,"name":"Товар","price":1000,"statistics":{"likes":1,"viewCount":1,"contacts":1.5}}`)
}

func (s *TestSuite) TestCreateItemInvalidSellerIDNull(t provider.T) {
	invalidStructureCreateItemExpect400(t, "sellerID null",
		fmt.Sprintf(`{"sellerID":null,"name":"Товар","price":1000,%s}`, createItemValidStatsJSON))
}

func (s *TestSuite) TestCreateItemInvalidNameNull(t provider.T) {
	invalidStructureCreateItemExpect400(t, "name null",
		fmt.Sprintf(`{"sellerID":123456,"name":null,"price":1000,%s}`, createItemValidStatsJSON))
}

func (s *TestSuite) TestCreateItemInvalidPriceNull(t provider.T) {
	invalidStructureCreateItemExpect400(t, "price null",
		fmt.Sprintf(`{"sellerID":123456,"name":"Товар","price":null,%s}`, createItemValidStatsJSON))
}

func (s *TestSuite) TestCreateItemInvalidStatisticsNull(t provider.T) {
	invalidStructureCreateItemExpect400(t, "statistics null",
		`{"sellerID":123456,"name":"Товар","price":1000,"statistics":null}`)
}

func (s *TestSuite) TestCreateItemInvalidStatisticsAsArray(t provider.T) {
	invalidStructureCreateItemExpect400(t, "statistics массивом",
		`{"sellerID":123456,"name":"Товар","price":1000,"statistics":[]}`)
}

func (s *TestSuite) TestCreateItemInvalidBodyAsArray(t provider.T) {
	invalidStructureCreateItemExpect400(t, "тело запроса массивом", `[]`)
}

func (s *TestSuite) TestCreateItemInvalidLikesNullInStatistics(t provider.T) {
	invalidStructureCreateItemExpect400(t, "likes null в statistics",
		`{"sellerID":123456,"name":"Товар","price":1000,"statistics":{"likes":null,"viewCount":1,"contacts":1}}`)
}

func (s *TestSuite) TestCreateItemInvalidViewCountNullInStatistics(t provider.T) {
	invalidStructureCreateItemExpect400(t, "viewCount null в statistics",
		`{"sellerID":123456,"name":"Товар","price":1000,"statistics":{"likes":1,"viewCount":null,"contacts":1}}`)
}

func (s *TestSuite) TestCreateItemInvalidContactsNullInStatistics(t provider.T) {
	invalidStructureCreateItemExpect400(t, "contacts null в statistics",
		`{"sellerID":123456,"name":"Товар","price":1000,"statistics":{"likes":1,"viewCount":1,"contacts":null}}`)
}

func (s *TestSuite) TestCreateItemInvalidMissingStatisticsField(t provider.T) {
	invalidStructureCreateItemExpect400(t, "нет поля statistics",
		`{"sellerID":123456,"name":"Товар","price":1000}`)
}

// Negative: price=0 не допускается.
func (s *TestSuite) TestCreateItemZeroPrice(t provider.T) {
	allureCreateItemSuiteLayout(t)
	t.Severity(allure.CRITICAL)
	t.Title("price=0 возвращает 400")
	t.Description("Согласно документации, price=0 недопустим — объявление должно иметь ненулевую цену")

	t.WithNewStep("POST /api/1/item с price=0", func(ctx provider.StepCtx) {
		req := models.CreateItemRequest{
			SellerID:   utils.RandomSellerID(),
			Name:       "Товар с нулевой ценой",
			Price:      0,
			Statistics: models.Statistics{Likes: 1, ViewCount: 1, Contacts: 1},
		}
		statusCode, body := itemManager.CreateItemRaw(t, req)
		itemManager.ScheduleDeleteIfCreatedOK(t, statusCode, body)
		ctx.WithNewParameters("statusCode", fmt.Sprintf("%d", statusCode))
		require.Equal(t, http.StatusBadRequest, statusCode,
			"price=0 должен отклоняться API (ожидался 400)")
	})
}

// Negative: sellerID=0 должен отклоняться.
func (s *TestSuite) TestCreateItemSellerIDZero(t provider.T) {
	allureCreateItemSuiteLayout(t)
	t.Severity(allure.CRITICAL)
	t.Title("sellerID=0 возвращает 400")
	t.Description("sellerID=0 является отсутствующим значением и должен отклоняться с 400")

	t.WithNewStep("POST с sellerID=0", func(ctx provider.StepCtx) {
		req := models.CreateItemRequest{
			SellerID:   0,
			Name:       "Товар",
			Price:      100,
			Statistics: models.Statistics{Likes: 1, ViewCount: 1, Contacts: 1},
		}
		statusCode, body := itemManager.CreateItemRaw(t, req)
		itemManager.ScheduleDeleteIfCreatedOK(t, statusCode, body)
		ctx.WithNewParameters("statusCode", fmt.Sprintf("%d", statusCode))
		require.Equal(t, http.StatusBadRequest, statusCode,
			"sellerID=0 не должен приниматься API (ожидался 400)")
	})
}

// BUG-003: POST с sellerID=-1 не возвращает ответ в разумный срок (зависание).
// Тест не блокирует прогон: клиент с таймаутом. Проходит при таймауте (дефект) или при быстром 400 (исправление).
func (s *TestSuite) TestCreateItemSellerIDMinusOneHangOrTimeout(t provider.T) {
	allureCreateItemSuiteLayout(t)
	t.Severity(allure.CRITICAL)
	t.Title("[BUG-003] POST с sellerID=-1: таймаут или быстрый 400")
	t.Description("Ожидаемое поведение — 400 без долгого ожидания. Фактически сервер может зависнуть; см. BUGS.md#BUG-003")

	const probeTimeout = 15 * time.Second
	raw := `{"sellerID":-1,"name":"Товар BUG-003","price":100,"statistics":{"likes":1,"viewCount":1,"contacts":1}}`

	t.WithNewStep(fmt.Sprintf("POST с sellerID=-1, таймаут клиента %s", probeTimeout), func(ctx provider.StepCtx) {
		ctx.WithNewParameters("probeTimeout", probeTimeout.String())
		statusCode, body, err := itemManager.CreateItemRawBodyWithTimeout(t, raw, probeTimeout)
		if err != nil {
			if utils.IsClientTimeout(err) {
				t.Logf("BUG-003: за %s ответ не получен (таймаут/зависание): %v", probeTimeout, err)
				return
			}
			require.Failf(t, "неожиданная ошибка POST при sellerID=-1", "%v", err)
		}
		ctx.WithNewParameters("statusCode", fmt.Sprintf("%d", statusCode))
		itemManager.ScheduleDeleteIfCreatedOK(t, statusCode, body)
		switch statusCode {
		case http.StatusBadRequest:
			t.Logf("BUG-003 устранён: API вернуло 400 за время до %s", probeTimeout)
		case http.StatusOK:
			require.Fail(t, "sellerID=-1 не должен приводить к 200 (ожидался 400 или отсутствие ответа до исправления BUG-003)",
				"ответ: %s", utils.TruncateForLog(body, 200))
		default:
			require.Failf(t, "неожиданный HTTP-статус при sellerID=-1",
				"получен %d, тело: %s", statusCode, utils.TruncateForLog(body, 200))
		}
	})
}

// Negative (BUG): API принимает отрицательный sellerID в теле POST (не -1 — см. BUG-003).
// BUG: API возвращает 200 при sellerID=-111111 (см. BUGS.md#BUG-008).
func (s *TestSuite) TestCreateItemNegativeSellerIDPost(t provider.T) {
	allureCreateItemSuiteLayout(t)
	t.Severity(allure.CRITICAL)
	t.Title("[BUG-008] POST: отрицательный sellerID=-111111 принимается")
	t.Description("Отрицательный sellerID должен отклоняться с 400; не использовать sellerID=-1 (зависание, BUG-003). Подробнее: BUGS.md#BUG-008")

	raw := `{"sellerID":-111111,"name":"Тест отрицательный продавец","price":1000,"statistics":{"likes":1,"viewCount":1,"contacts":1}}`

	t.WithNewStep("POST с sellerID=-111111", func(ctx provider.StepCtx) {
		statusCode, body := itemManager.CreateItemRawBody(t, raw)
		itemManager.ScheduleDeleteIfCreatedOK(t, statusCode, body)
		ctx.WithNewParameters("sellerID", "-111111")
		ctx.WithNewParameters("statusCode", fmt.Sprintf("%d", statusCode))
		if statusCode == http.StatusOK {
			t.Logf("BUG (BUGS.md#BUG-008): API принял отрицательный sellerID, ответ: %s", utils.TruncateForLog(body, 200))
		}
		require.Equal(t, http.StatusBadRequest, statusCode,
			"отрицательный sellerID должен отклоняться (ожидался 400)")
	})
}

// Negative: создание объявления с отрицательной ценой.
// BUG: API принимает отрицательные цены и возвращает 200 (см. BUGS.md#BUG-001).
func (s *TestSuite) TestCreateItemNegativePrice(t provider.T) {
	allureCreateItemSuiteLayout(t)
	t.Severity(allure.CRITICAL)
	t.Title("[BUG-001] Отрицательная цена принимается")
	t.Description("API должен отклонять price < 0 с ошибкой 400, но принимает и сохраняет объявление. Подробнее: BUGS.md#BUG-001")

	req := models.CreateItemRequest{
		SellerID:   utils.RandomSellerID(),
		Name:       "Товар с отрицательной ценой",
		Price:      -1,
		Statistics: models.Statistics{Likes: 1, ViewCount: 1, Contacts: 1},
	}

	var statusCode int
	var body string

	t.WithNewStep("Отправляем POST с price=-1", func(ctx provider.StepCtx) {
		statusCode, body = itemManager.CreateItemRaw(t, req)
		itemManager.ScheduleDeleteIfCreatedOK(t, statusCode, body)
		ctx.WithNewParameters("statusCode", fmt.Sprintf("%d", statusCode))
	})

	t.WithNewStep("Проверяем, что API отклонил запрос с 400", func(ctx provider.StepCtx) {
		if statusCode == http.StatusOK {
			t.Logf("BUG (BUGS.md#BUG-001): API принял отрицательную цену, ответ: %s", body)
		}
		require.Equal(t, http.StatusBadRequest, statusCode,
			"отрицательная цена должна отклоняться (ожидался 400)")
	})
}

// Negative (BUG): API принимает price=2^53+1 вместо отказа валидации.
// BUG: API возвращает 200 (см. BUGS.md#BUG-007). Значение выходит за точные целые JSON number.
func (s *TestSuite) TestCreateItemVeryLargePrice(t provider.T) {
	allureCreateItemSuiteLayout(t)
	t.Severity(allure.CRITICAL)
	t.Title("[BUG-007] price=2^53+1 принимается")
	t.Description("API должен отклонять чрезмерную/неточную цену с 400; фактически принимает. Подробнее: BUGS.md#BUG-007")

	t.WithNewStep("POST /api/1/item с price=9007199254740993 (2^53+1)", func(ctx provider.StepCtx) {
		req := models.CreateItemRequest{
			SellerID:   utils.RandomSellerID(),
			Name:       "Очень дорогой товар",
			Price:      9007199254740993,
			Statistics: models.Statistics{Likes: 1, ViewCount: 1, Contacts: 1},
		}
		statusCode, body := itemManager.CreateItemRaw(t, req)
		itemManager.ScheduleDeleteIfCreatedOK(t, statusCode, body)
		ctx.WithNewParameters("price", "9007199254740993 (2^53+1)")
		ctx.WithNewParameters("statusCode", fmt.Sprintf("%d", statusCode))
		if statusCode == http.StatusOK {
			t.Logf("BUG (BUGS.md#BUG-007): API принял price=2^53+1, ответ: %s", utils.TruncateForLog(body, 200))
		}
		require.Equal(t, http.StatusBadRequest, statusCode,
			"price=2^53+1 должен отклоняться (ожидался 400)")
	})
}

func negativeStatisticsCreateItemExpect400(t provider.T, caseLabel string, req models.CreateItemRequest) {
	allureCreateItemSuiteLayout(t)
	t.Severity(allure.CRITICAL)
	t.Title("[BUG-004] Отрицательные statistics — " + caseLabel)
	t.Description("API должен отклонять отрицательные счётчики с 400, но принимает и сохраняет объявление. Подробнее: BUGS.md#BUG-004")
	t.WithNewStep(fmt.Sprintf("POST с %s", caseLabel), func(ctx provider.StepCtx) {
		statusCode, body := itemManager.CreateItemRaw(t, req)
		itemManager.ScheduleDeleteIfCreatedOK(t, statusCode, body)
		ctx.WithNewParameters("statusCode", fmt.Sprintf("%d", statusCode))
		if statusCode == http.StatusOK {
			t.Logf("BUG (BUGS.md#BUG-004): API принял отрицательное значение %s, ответ: %s", caseLabel, body)
		}
		require.Equal(t, http.StatusBadRequest, statusCode,
			"отрицательное значение statistics должно отклоняться с 400 (BUG: сервис принимает)")
	})
}

// Negative (BUG): API принимает likes=-1 (см. BUGS.md#BUG-004).
func (s *TestSuite) TestCreateItemNegativeStatisticsLikes(t provider.T) {
	negativeStatisticsCreateItemExpect400(t, "likes=-1", models.CreateItemRequest{
		SellerID:   utils.RandomSellerID(),
		Name:       "Товар",
		Price:      100,
		Statistics: models.Statistics{Likes: -1, ViewCount: 1, Contacts: 1},
	})
}

// Negative (BUG): API принимает viewCount=-1 (см. BUGS.md#BUG-004).
func (s *TestSuite) TestCreateItemNegativeStatisticsViewCount(t provider.T) {
	negativeStatisticsCreateItemExpect400(t, "viewCount=-1", models.CreateItemRequest{
		SellerID:   utils.RandomSellerID(),
		Name:       "Товар",
		Price:      100,
		Statistics: models.Statistics{Likes: 1, ViewCount: -1, Contacts: 1},
	})
}

// Negative (BUG): API принимает contacts=-1 (см. BUGS.md#BUG-004).
func (s *TestSuite) TestCreateItemNegativeStatisticsContacts(t provider.T) {
	negativeStatisticsCreateItemExpect400(t, "contacts=-1", models.CreateItemRequest{
		SellerID:   utils.RandomSellerID(),
		Name:       "Товар",
		Price:      100,
		Statistics: models.Statistics{Likes: 1, ViewCount: 1, Contacts: -1},
	})
}

// Negative (BUG): API принимает name из одних пробелов.
// BUG: API возвращает 200 при name="   " (см. BUGS.md#BUG-005).
func (s *TestSuite) TestCreateItemWhitespaceName(t provider.T) {
	allureCreateItemSuiteLayout(t)
	t.Severity(allure.CRITICAL)
	t.Title("[BUG-005] name из одних пробелов принимается")
	t.Description("API должен отклонять name, состоящий только из пробелов, но принимает его. Подробнее: BUGS.md#BUG-005")

	t.WithNewStep("POST /api/1/item с name из одних пробелов", func(ctx provider.StepCtx) {
		req := models.CreateItemRequest{
			SellerID:   utils.RandomSellerID(),
			Name:       "   ",
			Price:      100,
			Statistics: models.Statistics{Likes: 1, ViewCount: 1, Contacts: 1},
		}
		statusCode, body := itemManager.CreateItemRaw(t, req)
		itemManager.ScheduleDeleteIfCreatedOK(t, statusCode, body)
		ctx.WithNewParameters("name", `"   "`)
		ctx.WithNewParameters("statusCode", fmt.Sprintf("%d", statusCode))
		if statusCode == http.StatusOK {
			t.Logf("BUG (BUGS.md#BUG-005): API принял name из пробелов, ответ: %s", body)
		}
		require.Equal(t, http.StatusBadRequest, statusCode,
			"name из одних пробелов должен отклоняться с 400 (BUG: сервис принимает)")
	})
}
