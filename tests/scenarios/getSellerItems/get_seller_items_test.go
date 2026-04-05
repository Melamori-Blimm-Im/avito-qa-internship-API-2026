package getSellerItems

import (
	"encoding/json"
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

// allureGetSellerItemsSuiteLayout — epic/parent/suite/feature и sub_suite/story для GET /api/1/:sellerID/item.
func allureGetSellerItemsSuiteLayout(t provider.T) {
	t.Epic(base.AllureEpic)
	t.AddParentSuite(base.AllureEpic)
	t.AddSubSuite(base.AllureFeatureSellerList)
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

// Positive: создать несколько объявлений одного продавца и получить их список.
func (s *TestSuite) TestGetSellerItemsPositive(t provider.T) {
	allureGetSellerItemsSuiteLayout(t)
	t.Severity(allure.BLOCKER)
	t.Title("Получение списка объявлений продавца")
	t.Description("Создаём два объявления одного продавца и проверяем, что оба присутствуют в ответе GET /api/1/:sellerID/item")

	sellerID := utils.RandomSellerID()
	var id1, id2 string

	t.WithNewStep("Создаём два объявления одного продавца", func(ctx provider.StepCtx) {
		ctx.WithNewParameters("sellerID", fmt.Sprintf("%d", sellerID))

		req1 := models.CreateItemRequest{
			SellerID:   sellerID,
			Name:       "Товар первый " + utils.RandomString(4),
			Price:      300,
			Statistics: models.Statistics{Likes: 1, ViewCount: 5, Contacts: 2},
		}
		req2 := models.CreateItemRequest{
			SellerID:   sellerID,
			Name:       "Товар второй " + utils.RandomString(4),
			Price:      600,
			Statistics: models.Statistics{Likes: 2, ViewCount: 10, Contacts: 1},
		}

		id1 = itemManager.CreateItem(t, req1)
		id2 = itemManager.CreateItem(t, req2)
		ctx.WithNewParameters("id1", id1)
		ctx.WithNewParameters("id2", id2)
	})

	t.WithNewStep("GET /api/1/:sellerID/item — проверяем наличие обоих объявлений", func(ctx provider.StepCtx) {
		items := itemManager.GetSellerItems(t, sellerID)
		require.NotEmpty(t, items, "список объявлений продавца не должен быть пустым")

		foundIDs := make(map[string]bool, len(items))
		for _, item := range items {
			foundIDs[item.ID] = true
		}

		require.True(t, foundIDs[id1], "первое созданное объявление должно быть в списке")
		require.True(t, foundIDs[id2], "второе созданное объявление должно быть в списке")
	})
}

// Positive: все объявления в ответе принадлежат запрошенному sellerId.
func (s *TestSuite) TestGetSellerItemsAllBelongToSeller(t provider.T) {
	allureGetSellerItemsSuiteLayout(t)
	t.Severity(allure.CRITICAL)
	t.Title("Все объявления в ответе принадлежат нужному sellerId")
	t.Description("Создаём несколько объявлений одного продавца и проверяем, что все возвращённые объявления имеют корректный sellerId")

	sellerID := utils.RandomSellerID()

	t.WithNewStep("Создаём 3 объявления одного продавца", func(ctx provider.StepCtx) {
		ctx.WithNewParameters("sellerID", fmt.Sprintf("%d", sellerID))
		for i := 1; i <= 3; i++ {
			itemManager.CreateItem(t, models.CreateItemRequest{
				SellerID:   sellerID,
				Name:       "Товар для проверки " + utils.RandomString(4),
				Price:      100 + i*50,
				Statistics: models.Statistics{Likes: i, ViewCount: i * 2, Contacts: i},
			})
		}
	})

	t.WithNewStep("GET /api/1/:sellerID/item — проверяем sellerId каждого объявления", func(ctx provider.StepCtx) {
		items := itemManager.GetSellerItems(t, sellerID)
		require.NotEmpty(t, items)

		for _, item := range items {
			require.Equal(t, sellerID, item.SellerID,
				"объявление %s должно принадлежать продавцу %d", item.ID, sellerID)
		}
	})
}

// Positive: объявления другого продавца не попадают в выдачу при запросе по sellerID первого.
func (s *TestSuite) TestGetSellerItemsExcludesOtherSellers(t provider.T) {
	allureGetSellerItemsSuiteLayout(t)
	t.Severity(allure.CRITICAL)
	t.Title("Список продавца A не содержит объявлений продавца B")
	t.Description("Создаём объявления у двух разных продавцов; GET по sellerID первого должен вернуть только его объявления")

	sellerA := utils.RandomSellerID()
	sellerB := utils.RandomSellerID()
	for sellerB == sellerA {
		sellerB = utils.RandomSellerID()
	}

	var idA, idB string

	t.WithNewStep("Создаём по одному объявлению у продавца A и у продавца B", func(ctx provider.StepCtx) {
		ctx.WithNewParameters("sellerA", fmt.Sprintf("%d", sellerA))
		ctx.WithNewParameters("sellerB", fmt.Sprintf("%d", sellerB))

		idA = itemManager.CreateItem(t, models.CreateItemRequest{
			SellerID:   sellerA,
			Name:       "Товар продавца A " + utils.RandomString(4),
			Price:      100,
			Statistics: models.Statistics{Likes: 1, ViewCount: 1, Contacts: 1},
		})
		idB = itemManager.CreateItem(t, models.CreateItemRequest{
			SellerID:   sellerB,
			Name:       "Товар продавца B " + utils.RandomString(4),
			Price:      200,
			Statistics: models.Statistics{Likes: 2, ViewCount: 2, Contacts: 2},
		})
		ctx.WithNewParameters("idA", idA)
		ctx.WithNewParameters("idB", idB)
	})

	t.WithNewStep("GET списка продавца A — своё объявление есть, чужого нет", func(ctx provider.StepCtx) {
		items := itemManager.GetSellerItems(t, sellerA)

		ids := make(map[string]struct{}, len(items))
		for _, it := range items {
			ids[it.ID] = struct{}{}
			require.Equal(t, sellerA, it.SellerID,
				"ожидался sellerId=%d, получено %d (id=%s)", sellerA, it.SellerID, it.ID)
		}

		_, hasA := ids[idA]
		require.True(t, hasA, "объявление продавца A должно быть в списке")
		_, hasB := ids[idB]
		require.False(t, hasB, "объявление продавца B не должно попадать в список продавца A")
	})
}

// Positive: 50 объявлений одного продавца — все присутствуют в ответе без усечения.
func (s *TestSuite) TestGetSellerItemsFiftyItemsAllReturned(t provider.T) {
	allureGetSellerItemsSuiteLayout(t)
	t.Severity(allure.NORMAL)
	t.Title("Список из 50 объявлений продавца без пагинационного обрезания")
	t.Description("Создаём 50 объявлений с одним sellerID и проверяем, что GET возвращает все 50")

	sellerID := utils.RandomSellerID()

	t.WithNewStep("Создаём 50 объявлений", func(ctx provider.StepCtx) {
		ctx.WithNewParameters("sellerID", fmt.Sprintf("%d", sellerID))
		for i := 0; i < 50; i++ {
			itemManager.CreateItem(t, models.CreateItemRequest{
				SellerID:   sellerID,
				Name:       fmt.Sprintf("Массовый товар %02d %s", i, utils.RandomString(4)),
				Price:      100 + i,
				Statistics: models.Statistics{Likes: 1, ViewCount: 1, Contacts: 1},
			})
		}
	})

	t.WithNewStep("GET списка и подсчёт элементов", func(ctx provider.StepCtx) {
		items := itemManager.GetSellerItems(t, sellerID)
		ids := make(map[string]struct{}, len(items))
		for _, it := range items {
			if it.SellerID == sellerID {
				ids[it.ID] = struct{}{}
			}
		}
		ctx.WithNewParameters("countForSeller", fmt.Sprintf("%d", len(ids)))
		require.GreaterOrEqual(t, len(ids), 50,
			"в ответе должно быть не менее 50 объявлений данного продавца")
	})
}

// Corner: продавец без объявлений — 200 и пустой JSON-массив
func (s *TestSuite) TestGetSellerItemsEmptySeller(t provider.T) {
	allureGetSellerItemsSuiteLayout(t)
	t.Severity(allure.NORMAL)
	t.Title("Продавец без объявлений — 200 и пустой массив")
	t.Description("GET /api/1/:sellerID/item для продавца без объявлений возвращает 200 и тело []")

	// Диапазон вне типичных 111111–999999 из RandomSellerID — меньше шанс пересечься с данными на общем стенде.
	sellerID := 1_000_000_000 + utils.RandomIntN(900_000_000)

	t.WithNewStep(fmt.Sprintf("GET /api/1/%d/item — sellerID без объявлений", sellerID), func(ctx provider.StepCtx) {
		ctx.WithNewParameters("sellerID", fmt.Sprintf("%d", sellerID))
		statusCode, body := itemManager.GetSellerItemsRaw(t, sellerID, http.StatusOK)
		ctx.WithNewParameters("statusCode", fmt.Sprintf("%d", statusCode))

		require.Equal(t, http.StatusOK, statusCode, "ожидался 200 с пустым списком объявлений")

		var items []models.ItemResponse
		require.NoError(t, json.Unmarshal([]byte(body), &items), "тело ответа должно быть JSON-массивом")
		require.Empty(t, items, "для продавца без объявлений ожидается []")
	})
}

// Boundary: очень большой sellerID (2^53+1). Ограничение не задокументировано — фиксируем поведение.
func (s *TestSuite) TestGetSellerItemsVeryLargeSellerID(t provider.T) {
	allureGetSellerItemsSuiteLayout(t)
	t.Severity(allure.NORMAL)
	t.Title("sellerID=2^53+1 — очень большой sellerID")
	t.Description("Граничное значение: sellerID=9007199254740993 (2^53+1). Ограничение сверху не задокументировано — фиксируем поведение")

	t.WithNewStep("GET /api/1/9007199254740993/item", func(ctx provider.StepCtx) {
		ctx.WithNewParameters("sellerID", "9007199254740993 (2^53+1)")
		statusCode, body := itemManager.GetSellerItemsByRawID(t, "9007199254740993")
		ctx.WithNewParameters("statusCode", fmt.Sprintf("%d", statusCode))
		t.Logf("Поведение API при sellerID=2^53+1: statusCode=%d, body=%s", statusCode, body)
		require.Equal(t, http.StatusOK, statusCode,
			"текущее поведение API: sellerID=2^53+1 в пути возвращает 200")
	})
}

// Negative (BUG): отрицательный sellerID в URL должен вернуть 400, но API возвращает 200.
// BUG: API принимает отрицательный sellerID и возвращает пустой список (см. BUGS.md#BUG-006).
func (s *TestSuite) TestGetSellerItemsNegativeSellerID(t provider.T) {
	allureGetSellerItemsSuiteLayout(t)
	t.Severity(allure.CRITICAL)
	t.Title("[BUG-006] Отрицательный sellerID в пути — фиксация ответа 200")
	t.Description("Ожидаемое поведение — 400; фактически 200 (BUG-006). Тест фиксирует текущий ответ.")

	t.WithNewStep("GET /api/1/-111111/item с отрицательным sellerID", func(ctx provider.StepCtx) {
		ctx.WithNewParameters("sellerID", "-111111")
		statusCode, body := itemManager.GetSellerItemsByRawID(t, "-111111")
		ctx.WithNewParameters("statusCode", fmt.Sprintf("%d", statusCode))
		if statusCode == http.StatusOK {
			t.Logf("BUG (BUGS.md#BUG-006): API принял отрицательный sellerID, ответ: %s", body)
		}
		require.Equal(t, http.StatusOK, statusCode,
			"текущее поведение API: отрицательный sellerID в пути возвращает 200 (дефект — см. BUGS.md#BUG-006)")
	})
}

// --- Негативные сценарии ---

// Negative: дробный sellerID в пути — ожидается 400.
func (s *TestSuite) TestGetSellerItemsFractionalSellerID(t provider.T) {
	allureGetSellerItemsSuiteLayout(t)
	t.Severity(allure.CRITICAL)
	t.Title("Дробный sellerID в URL возвращает 400")
	t.Description("GET /api/1/111111.5/item — идентификатор должен быть целым числом")

	t.WithNewStep("GET /api/1/111111.5/item", func(ctx provider.StepCtx) {
		statusCode, body := itemManager.GetSellerItemsByRawID(t, "111111.5")
		ctx.WithNewParameters("statusCode", fmt.Sprintf("%d", statusCode))
		t.Logf("ответ: %s", body)
		require.Equal(t, http.StatusBadRequest, statusCode,
			"дробный sellerID должен отклоняться с 400")
	})
}

func getSellerItemsInvalidRawIDExpect400(t provider.T, displayName, rawID string) {
	allureGetSellerItemsSuiteLayout(t)
	t.Severity(allure.CRITICAL)
	t.Title("Negative: нечисловой sellerID — " + displayName)
	t.Description("Строковые значения в пути /api/1/:sellerID/item должны возвращать 400 Bad Request")
	t.WithNewStep(fmt.Sprintf("GET /api/1/%s/item", rawID), func(ctx provider.StepCtx) {
		ctx.WithNewParameters("rawID", rawID)
		statusCode, _ := itemManager.GetSellerItemsByRawID(t, rawID)
		ctx.WithNewParameters("statusCode", fmt.Sprintf("%d", statusCode))
		require.Equal(t, http.StatusBadRequest, statusCode,
			"нечисловой sellerID %q должен возвращать 400", rawID)
	})
}

// Negative: латинские буквы в sellerID в пути — 400.
func (s *TestSuite) TestGetSellerItemsInvalidSellerIDLatin(t provider.T) {
	getSellerItemsInvalidRawIDExpect400(t, "строка латиница", "abc")
}

// Negative: кириллица в sellerID в пути — 400.
func (s *TestSuite) TestGetSellerItemsInvalidSellerIDCyrillic(t provider.T) {
	getSellerItemsInvalidRawIDExpect400(t, "строка кириллица", "продавец")
}
