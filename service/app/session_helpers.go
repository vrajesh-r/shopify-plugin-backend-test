package app

import (
	"fmt"
	"time"

	"github.com/getbread/breadkit/zeus/searcher"
	ztypes "github.com/getbread/breadkit/zeus/types"
	"github.com/getbread/shopify_plugin_backend/service/search"
	"github.com/getbread/shopify_plugin_backend/service/types"
)

func createSessionByShopId(shopId ztypes.Uuid, h *Handlers) (types.Session, error) {
	session := types.Session{
		ShopId:     shopId,
		Expiration: types.GenerateSessionExpiration(),
	}
	sessionId, err := h.SessionCreator.Create(session)
	if err != nil {
		return session, err
	}
	session.Id = sessionId
	return session, err
}

func findValidSessionById(inputId interface{}, h *Handlers) (session types.Session, err error) {
	var sessionId ztypes.Uuid
	switch inputId.(type) {
	case string:
		sessionId = ztypes.Uuid(inputId.(string))
	case ztypes.Uuid:
		sessionId = inputId.(ztypes.Uuid)
	}
	ssr := search.SessionSearchRequest{}
	ssr.AddFilter(search.SessionSearch_Id, sessionId, searcher.Operator_EQ, searcher.Condition_AND)
	ssr.AddFilter(search.SessionSearch_Expiration, time.Now().Unix(), searcher.Operator_GT, searcher.Condition_AND)
	ssr.Limit = 1
	sessions, err := h.SessionSearcher.Search(ssr)
	if err != nil {
		return
	}
	if len(sessions) == 0 {
		err = fmt.Errorf("session not found")
		return
	}
	session = sessions[0]
	return
}
