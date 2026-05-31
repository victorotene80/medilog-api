package bootstrap

import (
	"net/http"

	svchttp "github.com/victorotene80/medilog-api/internal/infrastructure/services/http"
	"github.com/victorotene80/medilog-api/internal/shared/config"
)

func initializeHTTPClient(cfg config.HTTPConfig) svchttp.HTTPService {
	return svchttp.NewDefaultHTTPService(&http.Client{
		Timeout: cfg.Timeout,
	})
}
