package gateway

import (
	"fmt"

	"github.com/getbread/breadkit/zeus/searcher"
	zeus "github.com/getbread/breadkit/zeus/types"
	"github.com/getbread/shopify_plugin_backend/service/search"
	"github.com/getbread/shopify_plugin_backend/service/types"
)

func findGatewayAccountByGatewayKey(gatewayKey string, h *Handlers) (account types.GatewayAccount, err error) {
	gasr := search.GatewayAccountSearchRequest{}
	gasr.AddFilter(search.GatewayAccountSearch_GatewayKey, gatewayKey, searcher.Operator_EQ, searcher.Condition_AND)
	gasr.Limit = 1
	accounts, err := h.GatewayAccountSearcher.Search(gasr)
	if err != nil {
		return
	}
	if len(accounts) == 0 {
		err = fmt.Errorf("gateway account not found")
		return
	}
	account = accounts[0]
	return
}

func findGatewayAccountById(accountID zeus.Uuid, h *Handlers) (account types.GatewayAccount, err error) {
	gasr := search.GatewayAccountSearchRequest{}
	gasr.AddFilter(search.GatewayAccountSearch_Id, accountID, searcher.Operator_EQ, searcher.Condition_AND)
	gasr.Limit = 1
	accounts, err := h.GatewayAccountSearcher.Search(gasr)
	if err != nil {
		return
	}
	if len(accounts) == 0 {
		err = fmt.Errorf("gateway account not found")
		return
	}
	account = accounts[0]
	return
}

func findGatewayAccountByEmail(email string, h *Handlers) (account types.GatewayAccount, err error) {
	gasr := search.GatewayAccountSearchRequest{}
	gasr.AddFilter(search.GatewayAccountSearch_Email, email, searcher.Operator_EQ, searcher.Condition_AND)
	gasr.Limit = 1
	accounts, err := h.GatewayAccountSearcher.Search(gasr)
	if err != nil {
		return
	}
	if len(accounts) == 0 {
		err = fmt.Errorf("gateway account not found")
		return
	}
	account = accounts[0]
	return
}
