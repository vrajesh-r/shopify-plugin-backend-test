package app

import (
	"fmt"

	"github.com/getbread/breadkit/zeus/searcher"
	ztypes "github.com/getbread/breadkit/zeus/types"
	"github.com/getbread/shopify_plugin_backend/service/search"
	"github.com/getbread/shopify_plugin_backend/service/types"
)

func createOrder(shop types.Shop, txId string, orderId int, h *Handlers) (order types.Order, err error) {
	order = types.Order{
		ShopId:     shop.Id,
		TxId:       ztypes.Uuid(txId),
		OrderId:    orderId,
		Production: shop.Production,
	}
	oid, err := h.OrderCreator.Create(order)
	if err != nil {
		return
	}
	order.Id = oid
	return
}

func findOrderByOrderId(orderId int, h *Handlers) (order types.Order, err error) {
	osr := search.OrderSearchRequest{}
	osr.AddFilter(search.OrderSearch_OrderId, orderId, searcher.Operator_EQ, searcher.Condition_AND)
	osr.Limit = 1
	orders, err := h.OrderSearcher.Search(osr)
	if err != nil {
		return
	}
	if len(orders) == 0 {
		err = fmt.Errorf("order not found (order_id => %d)", orderId)
		return
	}
	order = orders[0]
	return
}

func findOrderByTransactionId(transactionId string, h *Handlers) (order *types.Order, err error) {
	osr := search.OrderSearchRequest{}
	osr.AddFilter(search.OrderSearch_TxId, ztypes.Uuid(transactionId), searcher.Operator_EQ, searcher.Condition_AND)
	osr.Limit = 1
	orders, err := h.OrderSearcher.Search(osr)
	if err != nil {
		return
	}
	if len(orders) == 0 {
		err = fmt.Errorf("order not found (transaction_id => %s)", transactionId)
		return
	}
	order = &orders[0]
	return
}
