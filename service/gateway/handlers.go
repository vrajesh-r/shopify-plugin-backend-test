package gateway

import (
	"github.com/getbread/shopify_plugin_backend/service/dbhandlers"
)

type Handlers struct {
	*dbhandlers.Handlers
}

func NewHandlers(h *dbhandlers.Handlers) *Handlers {
	return &Handlers{h}
}
