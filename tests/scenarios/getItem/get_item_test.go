package getItem

import (
	"fmt"
	"net/http"
	"regexp"
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

// createdAtDuplicateOffsetPattern — признак BUG-010 (дубль смещения в конце createdAt). См. BUGS.md#BUG-010.
var createdAtDuplicateOffsetPattern = regexp.MustCompile(`(\s[+-]\d{4})\s[+-]\d{4}$`)

func createdAtParsesWithKnownLayouts(createdAt string) bool {
	layouts := []string{
		time.RFC3339,
		"2006-01-02T15:04:05-07:00",
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02T15:04:05-0700",
		"2006-01-02 15:04:05.999999999 -0700",
		"2006-01-02 15:04:05.999999 -0700",
		"2006-01-02T15:04:05Z",
	}
	for _, layout := range layouts {
		if _, err := time.Parse(layout, createdAt); err == nil {
			return true
		}
	}
	return false
}

type TestSuite struct {
	suite.Suite
}

// allureGetItemSuiteLayout — epic/parent/suite/feature и sub_suite/story для GET /api/1/item/:id.
func allureGetItemSuiteLayout(t provider.T) {
	t.Epic(base.AllureEpic)
	t.AddParentSuite(base.AllureEpic)
	t.AddSubSuite(base.AllureFeatureGetItem)
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

// Positive: E2E — создать объявление, получить по ID, сверить все поля.
func (s *TestSuite) TestGetItemByIDPositive(t provider.T) {
	allureGetItemSuiteLayout(t)
	t.Severity(allure.BLOCKER)
	t.Title("E2E: создать объявление и получить по ID")
	t.Description("Создаём объявление через POST, затем получаем по UUID через GET и проверяем все поля")

	sellerID := utils.RandomSellerID()
	req := models.CreateItemRequest{
		SellerID: sellerID,
		Name:     "E2E получение " + utils.RandomString(6),
		Price:    750,
		Statistics: models.Statistics{
			Likes:     3,
			ViewCount: 15,
			Contacts:  1,
		},
	}

	var createdID string

	t.WithNewStep("POST /api/1/item — создаём объявление", func(ctx provider.StepCtx) {
		createdID = itemManager.CreateItem(t, req)
		ctx.WithNewParameters("createdID", createdID)
		ctx.WithNewParameters("sellerID", fmt.Sprintf("%d", sellerID))
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

// Negative (BUG): в ответе GET поле createdAt может содержать дубль смещения (+0300 +0300) и не соответствовать RFC3339.
// BUG: фактический формат даты; ожидается одна зона в конце строки. Подробнее: BUGS.md#BUG-010.
func (s *TestSuite) TestGetItemCreatedAtFormat(t provider.T) {
	allureGetItemSuiteLayout(t)
	t.Severity(allure.CRITICAL)
	t.Title("[BUG-010] createdAt должен парситься как ISO 8601 / RFC3339 без предобработки")
	t.Description("API должен отдавать дату в стандартном виде; при дубле смещения строка не разбирается time.Parse. Подробнее: BUGS.md#BUG-010")

	var createdID string
	var createdAt string

	t.WithNewStep("POST /api/1/item — создаём объявление", func(ctx provider.StepCtx) {
		createdID = itemManager.CreateItem(t, models.CreateItemRequest{
			SellerID:   utils.RandomSellerID(),
			Name:       "Тест формата даты",
			Price:      200,
			Statistics: models.Statistics{Likes: 1, ViewCount: 1, Contacts: 1},
		})
		ctx.WithNewParameters("createdID", createdID)
	})

	t.WithNewStep("GET /api/1/item/:id — проверяем формат createdAt", func(ctx provider.StepCtx) {
		item := itemManager.GetItem(t, createdID)
		createdAt = item.CreatedAt
		ctx.WithNewParameters("createdAt", createdAt)

		parsed := createdAtParsesWithKnownLayouts(createdAt)

		if !parsed {
			if createdAtDuplicateOffsetPattern.MatchString(createdAt) {
				t.Logf("BUG (BUGS.md#BUG-010): createdAt не соответствует ожидаемому формату (дубль смещения): %q", createdAt)
			} else {
				t.Logf("createdAt не соответствует ни одному из ожидаемых layout'ов: %q", createdAt)
			}
		}

		require.True(t, parsed,
			"createdAt %q должен соответствовать ISO 8601 / RFC3339 без предобработки (BUG: некорректный формат / дубль смещения, BUGS.md#BUG-010)", createdAt)
	})
}

// Positive: два подряд GET с одним и тем же id возвращают идентичное тело ответа.
func (s *TestSuite) TestGetItemDoubleGETSameBody(t provider.T) {
	allureGetItemSuiteLayout(t)
	t.Severity(allure.CRITICAL)
	t.Title("Повторный GET возвращает то же тело ответа")
	t.Description("Два последовательных GET /api/1/item/:id должны дать одинаковый JSON")

	var createdID string
	req := models.CreateItemRequest{
		SellerID:   utils.RandomSellerID(),
		Name:       "Идемпотентность GET " + utils.RandomString(4),
		Price:      400,
		Statistics: models.Statistics{Likes: 2, ViewCount: 5, Contacts: 1},
	}

	t.WithNewStep("Создаём объявление", func(ctx provider.StepCtx) {
		createdID = itemManager.CreateItem(t, req)
		ctx.WithNewParameters("createdID", createdID)
	})

	t.WithNewStep("Два GET подряд и сравнение тел", func(ctx provider.StepCtx) {
		_, body1 := itemManager.GetItemRaw(t, createdID, http.StatusOK)
		_, body2 := itemManager.GetItemRaw(t, createdID, http.StatusOK)
		require.Equal(t, body1, body2, "тела двух GET должны совпадать байт-в-байт")
	})
}

// --- Негативные сценарии ---

// Negative: получение по несуществующему UUID возвращает 404.
func (s *TestSuite) TestGetItemNonExistentID(t provider.T) {
	allureGetItemSuiteLayout(t)
	t.Severity(allure.CRITICAL)
	t.Title("Несуществующий UUID возвращает 404")
	t.Description("GET /api/1/item/:id с UUID, которого нет в базе, должен возвращать 404 Not Found")

	nonExistentID := "00000000-0000-0000-0000-000000000000"

	t.WithNewStep("GET /api/1/item/:id с несуществующим UUID", func(ctx provider.StepCtx) {
		ctx.WithNewParameters("id", nonExistentID)
		statusCode, _ := itemManager.GetItemRaw(t, nonExistentID, http.StatusNotFound)
		require.Equal(t, http.StatusNotFound, statusCode,
			"несуществующий UUID должен возвращать 404")
	})
}

func getItemInvalidIDExpect400(t provider.T, displayName, id string) {
	allureGetItemSuiteLayout(t)
	t.Severity(allure.CRITICAL)
	t.Title("Negative: невалидный формат ID — " + displayName)
	t.Description("GET /api/1/item/:id с не-UUID значением должен возвращать 400 Bad Request")
	t.WithNewStep(fmt.Sprintf("GET с невалидным ID: %s", id), func(ctx provider.StepCtx) {
		ctx.WithNewParameters("id", id)
		ctx.WithNewParameters("описание", displayName)
		statusCode, _ := itemManager.GetItemRaw(t, id, http.StatusBadRequest)
		require.Equal(t, http.StatusBadRequest, statusCode,
			"невалидный ID %q должен возвращать 400", id)
	})
}

// Negative: строка вместо UUID в path — 400.
func (s *TestSuite) TestGetItemInvalidIDNotUUID(t provider.T) {
	getItemInvalidIDExpect400(t, "строка вместо UUID", "not-a-uuid")
}

// Negative: числовой id в path — 400.
func (s *TestSuite) TestGetItemInvalidIDNumeric(t provider.T) {
	getItemInvalidIDExpect400(t, "числовой ID", "12345")
}

// Negative: UUID неправильной длины — 400.
func (s *TestSuite) TestGetItemInvalidIDWrongLengthUUID(t provider.T) {
	getItemInvalidIDExpect400(t, "UUID неправильной длины", "123e4567-e89b-12d3-a456-4266141400000")
}

// Negative: SQL-подобная строка в path-параметре id — ожидается 400, не 500.
func (s *TestSuite) TestGetItemSQLInjectionLikeID(t provider.T) {
	allureGetItemSuiteLayout(t)
	t.Severity(allure.CRITICAL)
	t.Title("Path с SQL-подобной строкой — 400, не 500")
	t.Description("GET /api/1/item/{id} с id вида SQL-инъекции не должен приводить к 500")

	sqlLikeID := "'; DROP TABLE items; --"

	t.WithNewStep("GET с SQL-подобным id в пути", func(ctx provider.StepCtx) {
		ctx.WithNewParameters("id", sqlLikeID)
		statusCode, _ := itemManager.GetItemRaw(t, sqlLikeID, http.StatusBadRequest)
		ctx.WithNewParameters("statusCode", fmt.Sprintf("%d", statusCode))
		require.NotEqual(t, http.StatusInternalServerError, statusCode,
			"не ожидается 500 Internal Server Error")
		require.Equal(t, http.StatusBadRequest, statusCode,
			"невалидный id должен отклоняться с 400")
	})
}

// Negative: получение с пустым ID.
func (s *TestSuite) TestGetItemEmptyID(t provider.T) {
	allureGetItemSuiteLayout(t)
	t.Severity(allure.CRITICAL)
	t.Title("Пустой ID возвращает 400")
	t.Description("GET /api/1/item/:id с пробелом вместо ID должен возвращать 400 Bad Request")

	t.WithNewStep("GET /api/1/item/ с пустым ID", func(ctx provider.StepCtx) {
		statusCode, _ := itemManager.GetItemRaw(t, " ", http.StatusBadRequest)
		require.Equal(t, http.StatusBadRequest, statusCode,
			"пустой ID должен возвращать 400")
	})
}
