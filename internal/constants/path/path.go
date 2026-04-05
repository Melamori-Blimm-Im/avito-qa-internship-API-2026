package path

import (
	"fmt"
	"net/url"
)

const (
	// CreateItemPath POST /api/1/item — создать объявление.
	CreateItemPath = "/api/1/item"

	// GetItemPath GET /api/1/item/:id — получить объявление по ID
	GetItemPath = "/api/1/item"

	// DeleteItemPath DELETE /api/2/item/:id
	DeleteItemPath = "/api/2/item"

	// GetStatisticPath GET /api/1/statistic/:id — получить статистику.
	GetStatisticPath = "/api/1/statistic"
)

// GetSellerItemsPath возвращает путь для получения объявлений продавца.
// GET /api/1/:sellerID/item
func GetSellerItemsPath(sellerID int) string {
	return fmt.Sprintf("/api/1/%d/item", sellerID)
}

// GetItemByIDPath возвращает путь для получения объявления по UUID.
func GetItemByIDPath(id string) string {
	return fmt.Sprintf("%s/%s", GetItemPath, url.PathEscape(id))
}

// GetDeleteItemByIDPath возвращает путь для DELETE объявления по UUID.
func GetDeleteItemByIDPath(id string) string {
	return fmt.Sprintf("%s/%s", DeleteItemPath, url.PathEscape(id))
}

// GetStatisticByIDPath возвращает путь для получения статистики по UUID.
func GetStatisticByIDPath(id string) string {
	return fmt.Sprintf("/api/1/statistic/%s", url.PathEscape(id))
}
