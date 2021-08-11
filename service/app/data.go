package app

import (
	"database/sql"
	"fmt"

	"github.com/getbread/breadkit/zeus/searcher"
	zeus "github.com/getbread/breadkit/zeus/types"
	"github.com/getbread/shopify_plugin_backend/service/search"
	"github.com/getbread/shopify_plugin_backend/service/types"
	"github.com/getbread/shopify_plugin_backend/service/update"
)

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

func findCompletedGatewayCheckoutByReference(reference string, h *Handlers) (checkout types.GatewayCheckout, err error) {
	gcsr := search.GatewayCheckoutSearchRequest{}
	gcsr.AddFilter(search.GatewayCheckoutSearch_Reference, reference, searcher.Operator_EQ, searcher.Condition_AND)
	gcsr.AddFilter(search.GatewayCheckoutSearch_Completed, true, searcher.Operator_EQ, searcher.Condition_AND)
	gcsr.Limit = 1
	checkouts, err := h.GatewayCheckoutSearcher.Search(gcsr)
	if err != nil {
		return
	}
	if len(checkouts) == 0 {
		err = fmt.Errorf("checkout not found")
		return
	}
	checkout = checkouts[0]
	return
}

func redactOrderByOrderID(orderID int, h *Handlers) error {
	searchRequest := search.AnalyticsOrderSearchRequest{}
	searchRequest.AddFilter(search.AnalyticsOrderSearch_OrderID, orderID, searcher.Operator_EQ, searcher.Condition_AND)
	searchRequest.Limit = 1
	orders, err := h.AnalyticsOrderSearcher.Search(searchRequest)

	if err != nil {
		return err
	}

	if len(orders) == 0 {
		err = fmt.Errorf("order not found")
		return err
	}

	updateRequest := update.AnalyticsOrderUpdateRequest{
		Id:      orders[0].ID,
		Updates: map[update.AnalyticsOrderUpdateField]interface{}{},
	}

	ns := zeus.NullString{NullString: sql.NullString{Valid: false}}

	updateRequest.Updates[update.AnalyticsOrderUpdate_CustomerEmail] = ns
	updateRequest.Updates[update.AnalyticsOrderUpdate_TotalPrice] = ns
	updateRequest.Updates[update.AnalyticsOrderUpdate_Gateway] = ns
	updateRequest.Updates[update.AnalyticsOrderUpdate_FinancialStatus] = ns
	updateRequest.Updates[update.AnalyticsOrderUpdate_FulfillmentStatus] = ns
	updateRequest.Updates[update.AnalyticsOrderUpdate_Redacted] = true

	err = h.AnalyticsOrderUpdater.Update(updateRequest)
	return err
}
