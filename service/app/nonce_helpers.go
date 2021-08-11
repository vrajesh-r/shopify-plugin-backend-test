package app

import (
	"fmt"

	"github.com/getbread/breadkit/zeus/searcher"
	ztypes "github.com/getbread/breadkit/zeus/types"
	"github.com/getbread/shopify_plugin_backend/service/search"
	"github.com/getbread/shopify_plugin_backend/service/types"
)

func createNonceByShopId(shopId ztypes.Uuid, h *Handlers) (types.Nonce, error) {
	nonce := types.Nonce{
		ShopId: shopId,
		Nonce:  types.GenerateNonceValue(),
	}
	nonceId, err := h.NonceCreator.Create(nonce)
	if err != nil {
		return nonce, err
	}
	nonce.Id = nonceId
	return nonce, err
}

func useNonceByShopId(test string, shopId ztypes.Uuid, h *Handlers) error {
	nsr := search.NonceSearchRequest{}
	nsr.AddFilter(search.NonceSearch_ShopId, shopId, searcher.Operator_EQ, searcher.Condition_AND)
	nsr.AddFilter(search.NonceSearch_Nonce, test, searcher.Operator_EQ, searcher.Condition_AND)
	nsr.Limit = 1
	nonces, err := h.NonceSearcher.Search(nsr)
	if err != nil {
		return err
	}
	if len(nonces) == 0 {
		return fmt.Errorf("nonce not found")
	}
	go deleteNonceById(nonces[0].Id, h)
	return nil
}

func deleteNonceById(nonceId ztypes.Uuid, h *Handlers) {
	_ = h.NonceUpdater.DeleteById(nonceId)
}

func clearNoncesByShopId(shopId ztypes.Uuid, h *Handlers) {
	nsr := search.NonceSearchRequest{}
	nsr.AddFilter(
		search.NonceSearch_ShopId,
		shopId,
		searcher.Operator_EQ,
		searcher.Condition_AND)
	nonces, err := h.NonceSearcher.Search(nsr)
	if err != nil {
		return
	}
	for _, n := range nonces {
		go deleteNonceById(n.Id, h)
	}
	return
}
