package tests

// Константы для вкладки Behaviors в Allure: верхний уровень — Epic, вложенный — Feature.
const (
	AllureEpic           = "Сервис объявлений"
	AllureEpicAds        = "Объявления"
	AllureEpicStatistics = "Статистика"

	AllureFeatureAPIContract = "API-контракт"
	AllureFeatureSellerList  = "Получение объявлений продавца"
	AllureFeatureGetItem     = "Получение объявления"
	AllureFeatureCreateItem  = "Создание объявления"
	AllureFeatureDeleteItem  = "Удаление объявления"

	AllureFeatureItems = "items"

	// AllureStoryCreateItem — sub_suite и базовая история «создание объявления» (аналог AllureStory.CREATE_ITEM).
	AllureStoryCreateItem = "Создание объявления"

	// AllureFeatureStatistics — уровень suite для сценариев статистики (отдельный epic).
	AllureFeatureStatistics = "statistics"

	// AllureFeatureNonFunctional — нефункциональные проверки (Content-Type, время ответа).
	AllureFeatureNonFunctional = "Нефункциональные проверки"
)
