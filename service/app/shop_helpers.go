package app

import (
	"fmt"
	"strings"

	"github.com/getbread/breadkit/zeus/searcher"
	ztypes "github.com/getbread/breadkit/zeus/types"
	"github.com/getbread/shopify_plugin_backend/service/search"
	"github.com/getbread/shopify_plugin_backend/service/types"
)

func findShopByName(shopName string, h *Handlers) (shop types.Shop, err error) {
	shopName = strings.ToLower(shopName)
	ssr := search.ShopSearchRequest{}
	ssr.AddFilter(search.ShopSearch_Shop, shopName, searcher.Operator_EQ, searcher.Condition_AND)
	ssr.Limit = 1
	shops, err := h.ShopSearcher.Search(ssr)
	if err != nil {
		return
	}
	if len(shops) == 0 {
		err = fmt.Errorf("shop not found")
		return
	}
	shop = shops[0]
	return
}

func findShopByBreadApiKey(breadApiKey string, h *Handlers) (shop types.Shop, err error) {
	ssr := search.ShopSearchRequest{}
	ssr.AddFilter(search.ShopSearch_BreadApiKey, breadApiKey, searcher.Operator_EQ, searcher.Condition_AND)
	ssr.Limit = 1
	shops, err := h.ShopSearcher.Search(ssr)
	if err != nil {
		return
	}
	if len(shops) == 0 {
		err = fmt.Errorf("shop not found")
		return
	}
	shop = shops[0]
	return
}

func findShopById(shopId ztypes.Uuid, h *Handlers) (shop types.Shop, err error) {
	ssr := search.ShopSearchRequest{}
	ssr.AddFilter(search.ShopSearch_Id, shopId, searcher.Operator_EQ, searcher.Condition_AND)
	ssr.Limit = 1
	shops, err := h.ShopSearcher.Search(ssr)
	if err != nil {
		return
	}
	if len(shops) == 0 {
		err = fmt.Errorf("shop not found")
		return
	}
	shop = shops[0]
	return
}

func createShopByName(shopName string, h *Handlers) (shop types.Shop, err error) {
	shopName = strings.ToLower(shopName)
	shop = types.Shop{
		Shop: shopName,
	}
	shopId, err := h.ShopCreator.Create(shop)
	if err != nil {
		return
	}
	shop.Id = shopId
	return
}
