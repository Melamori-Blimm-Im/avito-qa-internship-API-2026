package utils

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/joho/godotenv"
)

// LoadEnv загружает .env и .env.override (override не обязателен).
func LoadEnv() {
	loadSpecificEnvFile(".env")
	loadSpecificEnvFile(".env.override")
}

func loadSpecificEnvFile(envFile string) {
	goModDir := findGoModDir(callerFile())
	if goModDir != "" {
		envFile = filepath.Join(goModDir, envFile)
	}

	err := godotenv.Overload(envFile)
	if errors.Is(err, os.ErrNotExist) {
		return
	}
	if err != nil {
		panic(fmt.Sprintf("не удалось загрузить env-файл %s: %s", envFile, err.Error()))
	}
}

func callerFile() string {
	_, file, _, _ := runtime.Caller(1)
	current := file
	for i := 2; file == current; i++ {
		_, file, _, _ = runtime.Caller(i)
	}
	return file
}

func findGoModDir(from string) string {
	dir := filepath.Dir(from)
	gopath := filepath.Clean(os.Getenv("GOPATH"))
	for dir != "/" && dir != gopath {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		dir = filepath.Dir(dir)
	}
	return ""
}
