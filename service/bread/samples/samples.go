package samples

import (
	"errors"

	"github.com/getbread/shopify_plugin_backend/service/bread"
	"github.com/pborman/uuid"
)

var Err = errors.New("an error occurred")

func NewTrxAuthTokenRequest() *bread.TrxAuthTokenRequest {
	return &bread.TrxAuthTokenRequest{
		ApiKey: uuid.NewRandom().String(),
		Secret: uuid.NewRandom().String(),
	}
}

func NewAuthTokenResponse() *bread.TrxAuthTokenResponse {
	return &bread.TrxAuthTokenResponse{Token: uuid.NewRandom().String()}
}
