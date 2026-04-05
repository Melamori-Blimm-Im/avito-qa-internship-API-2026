package deleteItem

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

// allureDeleteItemSuiteLayout — epic/parent/suite/feature и sub_suite/story для DELETE /api/2/item/:id.
func allureDeleteItemSuiteLayout(t provider.T) {
	t.Epic(base.AllureEpic)
	t.AddParentSuite(base.AllureEpic)
	t.AddSubSuite(base.AllureFeatureDeleteItem)
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

// Positive: создать объявление, удалить DELETE, затем GET возвращает 404.
func (s *TestSuite) TestDeleteItemExisting(t provider.T) {
	allureDeleteItemSuiteLayout(t)
	t.Severity(allure.BLOCKER)
	t.Title("Удаление существующего объявления")
	t.Description("POST /api/1/item создаёт объявление; удаление — DELETE /api/2/item/:id; проверка отсутствия — GET /api/1/item/:id → 404.")

	sellerID := utils.RandomSellerID()
	req := models.CreateItemRequest{
		SellerID:   sellerID,
		Name:       "DELETE E2E " + utils.RandomString(6),
		Price:      500,
		Statistics: models.Statistics{Likes: 1, ViewCount: 2, Contacts: 1},
	}

	var id string
	t.WithNewStep("POST /api/1/item — создаём объявление", func(ctx provider.StepCtx) {
		id = itemManager.CreateItem(t, req)
		ctx.WithNewParameters("id", id)
	})

	t.WithNewStep("DELETE /api/2/item/:id — удаляем", func(ctx provider.StepCtx) {
		code, body := itemManager.DeleteItemRaw(t, id)
		ctx.WithNewParameters("status", fmt.Sprintf("%d", code))
		require.Contains(t, []int{http.StatusOK, http.StatusNoContent}, code,
			"ожидали 200 или 204 на DELETE существующего объявления, получили %d, тело: %s", code, utils.TruncateForLog(body, 500))
	})

	t.WithNewStep("GET /api/1/item/:id — объявление отсутствует", func(ctx provider.StepCtx) {
		itemManager.GetItemRaw(t, id, http.StatusNotFound)
	})
}

// Negative: DELETE несуществующего UUID — 404.
func (s *TestSuite) TestDeleteItemNonExistentID(t provider.T) {
	allureDeleteItemSuiteLayout(t)
	t.Severity(allure.CRITICAL)
	t.Title("Удаление несуществующего UUID")
	t.Description("DELETE для заведомо отсутствующего объявления должен возвращать 404.")

	const missingID = "00000000-0000-0000-0000-000000000000"

	t.WithNewStep("DELETE /api/2/item/:id — несуществующий id", func(ctx provider.StepCtx) {
		code, body := itemManager.DeleteItemRaw(t, missingID)
		ctx.WithNewParameters("status", fmt.Sprintf("%d", code))
		require.Equal(t, http.StatusNotFound, code,
			"ожидали 404 для несуществующего id, получили %d, тело: %s", code, utils.TruncateForLog(body, 500))
	})
}
