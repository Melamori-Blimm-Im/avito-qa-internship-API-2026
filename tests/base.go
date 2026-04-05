package tests

import (
	"log"

	"avito-qa-internship/internal/utils"
)

// SetupSuite выполняет инициализацию перед запуском набора тестов.
func SetupSuite() {
	log.Println("Инициализация переменных окружения")
	utils.LoadEnv()
}

// TearDownSuite выполняет завершение после набора тестов.
func TearDownSuite() {
	log.Println("Завершение набора тестов")
}

// Precondition выводит предусловие теста в лог.
func Precondition(text string) {
	utils.LogWithLabelAndTimestamp("Precondition", text)
}
