package getStatistic

import (
	"fmt"
	"net/http"
	"testing"

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

// allureGetStatisticSuiteLayout — epic/parent/suite/feature и sub_suite/story для GET /api/1/statistic/:id.
func allureGetStatisticSuiteLayout(t provider.T) {
	t.Epic(base.AllureEpic)
	t.AddParentSuite(base.AllureEpic)
}

func TestSuiteRun(t *testing.T) {
	suite.RunNamedSuite(t, base.AllureEpicStatistics, new(TestSuite))
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

// Positive: E2E — создать объявление, получить статистику, проверить наличие полей.
func (s *TestSuite) TestGetStatisticPositive(t provider.T) {
	allureGetStatisticSuiteLayout(t)
	t.Severity(allure.BLOCKER)
	t.Title("Получение статистики по созданному объявлению")
	t.Description("Создаём объявление, получаем статистику по UUID и проверяем структуру ответа")

	req := models.CreateItemRequest{
		SellerID:   utils.RandomSellerID(),
		Name:       "Статистический товар " + utils.RandomString(5),
		Price:      999,
		Statistics: models.Statistics{Likes: 7, ViewCount: 42, Contacts: 3},
	}

	var createdID string

	t.WithNewStep("POST /api/1/item — создаём объявление", func(ctx provider.StepCtx) {
		createdID = itemManager.CreateItem(t, req)
		ctx.WithNewParameters("createdID", createdID)
	})

	t.WithNewStep("GET /api/1/statistic/:id — получаем статистику", func(ctx provider.StepCtx) {
		stats := itemManager.GetStatistic(t, createdID)
		require.NotEmpty(t, stats, "список статистик не должен быть пустым")
		ctx.WithNewParameters("statsCount", fmt.Sprintf("%d", len(stats)))
	})

	t.WithNewStep("Проверяем наличие числовых полей в статистике", func(ctx provider.StepCtx) {
		stats := itemManager.GetStatistic(t, createdID)
		require.NotEmpty(t, stats)
		stat := stats[0]
		require.GreaterOrEqual(t, stat.Likes, 0, "likes должен быть >= 0")
		require.GreaterOrEqual(t, stat.ViewCount, 0, "viewCount должен быть >= 0")
		require.GreaterOrEqual(t, stat.Contacts, 0, "contacts должен быть >= 0")
	})
}

// Positive: E2E полный сценарий — создать → получить по ID → получить статистику.
func (s *TestSuite) TestGetStatisticFullE2E(t provider.T) {
	allureGetStatisticSuiteLayout(t)
	t.Severity(allure.CRITICAL)
	t.Title("Полный E2E: создать → получить объявление → получить статистику")
	t.Description("Полная цепочка: POST /api/1/item → GET /api/1/item/:id → GET /api/1/statistic/:id")

	sellerID := utils.RandomSellerID()
	req := models.CreateItemRequest{
		SellerID:   sellerID,
		Name:       "E2E статистика " + utils.RandomString(5),
		Price:      450,
		Statistics: models.Statistics{Likes: 1, ViewCount: 10, Contacts: 2},
	}

	var createdID string

	t.WithNewStep("POST /api/1/item — создаём объявление", func(ctx provider.StepCtx) {
		createdID = itemManager.CreateItem(t, req)
		require.NotEmpty(t, createdID)
		ctx.WithNewParameters("createdID", createdID)
		ctx.WithNewParameters("sellerID", fmt.Sprintf("%d", sellerID))
	})

	t.WithNewStep("GET /api/1/item/:id — проверяем созданное объявление", func(ctx provider.StepCtx) {
		item := itemManager.GetItem(t, createdID)
		require.Equal(t, createdID, item.ID)
		require.Equal(t, req.Name, item.Name)
		require.Equal(t, sellerID, item.SellerID)
	})

	t.WithNewStep("GET /api/1/statistic/:id — проверяем структуру ответа статистики", func(ctx provider.StepCtx) {
		stats := itemManager.GetStatistic(t, createdID)
		require.NotEmpty(t, stats, "статистика не должна быть пустой")
		require.GreaterOrEqual(t, stats[0].Likes, 0)
		require.GreaterOrEqual(t, stats[0].ViewCount, 0)
		require.GreaterOrEqual(t, stats[0].Contacts, 0)
	})
}

// Positive: повторный GET статистики не меняет viewCount (чтение идемпотентно).
func (s *TestSuite) TestGetStatisticRepeatedGETViewCountStable(t provider.T) {
	allureGetStatisticSuiteLayout(t)
	t.Severity(allure.CRITICAL)
	t.Title("Повторный GET /statistic не изменяет viewCount")
	t.Description("Два запроса статистики подряд возвращают одинаковый viewCount")

	req := models.CreateItemRequest{
		SellerID:   utils.RandomSellerID(),
		Name:       "Стабильность viewCount " + utils.RandomString(4),
		Price:      300,
		Statistics: models.Statistics{Likes: 1, ViewCount: 10, Contacts: 1},
	}

	var createdID string

	t.WithNewStep("Создаём объявление с viewCount=10", func(ctx provider.StepCtx) {
		createdID = itemManager.CreateItem(t, req)
		ctx.WithNewParameters("createdID", createdID)
	})

	t.WithNewStep("Два GET statistic подряд — viewCount совпадает", func(ctx provider.StepCtx) {
		s1 := itemManager.GetStatistic(t, createdID)
		require.NotEmpty(t, s1)
		v1 := s1[0].ViewCount
		s2 := itemManager.GetStatistic(t, createdID)
		require.NotEmpty(t, s2)
		v2 := s2[0].ViewCount
		ctx.WithNewParameters("viewCount1", fmt.Sprintf("%d", v1))
		ctx.WithNewParameters("viewCount2", fmt.Sprintf("%d", v2))
		require.Equal(t, v1, v2, "повторный GET статистики не должен менять viewCount")
	})
}

// --- Негативные сценарии ---

// Negative: статистика по несуществующему UUID возвращает 404.
func (s *TestSuite) TestGetStatisticNonExistentID(t provider.T) {
	allureGetStatisticSuiteLayout(t)
	t.Severity(allure.CRITICAL)
	t.Title("Несуществующий UUID возвращает 404")
	t.Description("GET /api/1/statistic/:id с несуществующим UUID должен возвращать 404 Not Found")

	nonExistentID := "00000000-0000-0000-0000-000000000000"

	t.WithNewStep("GET /api/1/statistic/:id с нулевым UUID", func(ctx provider.StepCtx) {
		ctx.WithNewParameters("id", nonExistentID)
		statusCode, _ := itemManager.GetStatisticRaw(t, nonExistentID, http.StatusNotFound)
		require.Equal(t, http.StatusNotFound, statusCode,
			"несуществующий UUID должен возвращать 404")
	})
}

// Negative: статистика по пустому ID возвращает 400.
func (s *TestSuite) TestGetStatisticEmptyID(t provider.T) {
	allureGetStatisticSuiteLayout(t)
	t.Severity(allure.CRITICAL)
	t.Title("Пустой ID возвращает 400")
	t.Description("GET /api/1/statistic/:id с пробелом вместо ID должен возвращать 400 Bad Request")

	t.WithNewStep("GET /api/1/statistic/ с пустым ID", func(ctx provider.StepCtx) {
		statusCode, _ := itemManager.GetStatisticRaw(t, " ", http.StatusBadRequest)
		require.Equal(t, http.StatusBadRequest, statusCode,
			"пустой ID должен возвращать 400")
	})
}

func getStatisticInvalidIDExpect400(t provider.T, displayName, id string) {
	allureGetStatisticSuiteLayout(t)
	t.Severity(allure.CRITICAL)
	t.Title("Negative: невалидный формат ID статистики — " + displayName)
	t.Description("GET /api/1/statistic/:id со строкой не в формате UUID должен возвращать 400 Bad Request")
	t.WithNewStep(fmt.Sprintf("GET с невалидным ID: %s", id), func(ctx provider.StepCtx) {
		ctx.WithNewParameters("id", id)
		ctx.WithNewParameters("описание", displayName)
		statusCode, _ := itemManager.GetStatisticRaw(t, id, http.StatusBadRequest)
		require.Equal(t, http.StatusBadRequest, statusCode,
			"невалидный ID %q должен возвращать 400", id)
	})
}

// Negative: статистика по строке вместо UUID — 400.
func (s *TestSuite) TestGetStatisticInvalidIDNotUUID(t provider.T) {
	getStatisticInvalidIDExpect400(t, "строка вместо UUID", "not-a-uuid")
}

// Negative: статистика по числовому id — 400.
func (s *TestSuite) TestGetStatisticInvalidIDNumeric(t provider.T) {
	getStatisticInvalidIDExpect400(t, "числовой ID", "99999")
}
