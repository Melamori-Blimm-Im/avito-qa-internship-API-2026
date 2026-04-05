package models

// CreateItemStatusResponse — ответ на успешное создание объявления.
// Формат: {"status": "Сохранили объявление - <UUID>"}
type CreateItemStatusResponse struct {
	Status string `json:"status"`
}

// ItemResponse — ответ на получение объявления (GET /api/1/item/:id или GET /api/1/:sellerID/item).
// Эндпоинты возвращают массив: []ItemResponse.
type ItemResponse struct {
	ID         string     `json:"id"`
	SellerID   int        `json:"sellerId"`
	Name       string     `json:"name"`
	Price      int        `json:"price"`
	Statistics Statistics `json:"statistics"`
	CreatedAt  string     `json:"createdAt"`
}

// StatisticResponse — один элемент ответа GET /api/1/statistic/:id.
// Эндпоинт возвращает массив: []StatisticResponse.
type StatisticResponse struct {
	Likes     int `json:"likes"`
	ViewCount int `json:"viewCount"`
	Contacts  int `json:"contacts"`
}

// ErrorResponse — тело ответа при ошибке (4xx, 5xx).
type ErrorResponse struct {
	Result ErrorResult `json:"result"`
	Status string      `json:"status"`
}

// ErrorResult — вложенное поле result в теле ошибки.
type ErrorResult struct {
	Message  string            `json:"message"`
	Messages map[string]string `json:"messages"`
}
