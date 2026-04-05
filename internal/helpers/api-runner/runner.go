package api_runner

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"

	"github.com/steinfletcher/apitest"
)

type APIRunner struct {
	host string
}

func New() *APIRunner {
	return &APIRunner{
		host: os.Getenv("API_URL"),
	}
}

func GetRunner() *APIRunner {
	return New()
}

// Create создаёт новый apitest.APITest с базовыми настройками (debug, перехват URL).
func (r *APIRunner) Create() *apitest.APITest {
	apitestNew := apitest.New().EnableNetworking()
	if os.Getenv("DEBUG") == "true" {
		apitestNew = apitestNew.Debug()
	}
	host := r.host
	return apitestNew.
		Intercept(func(req *http.Request) {
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Accept", "application/json")
			_ = mergeServiceURLs(host, req.URL)
		})
}

func mergeServiceURLs(host string, r *url.URL) error {
	parsed, err := url.Parse(host)
	if err != nil {
		return fmt.Errorf("host cannot be parsed: %w", err)
	}
	if parsed.Path != "" {
		r.Path = path.Join(parsed.Path, r.Path)
	}
	r.Scheme = parsed.Scheme
	r.Host = parsed.Host
	return nil
}
