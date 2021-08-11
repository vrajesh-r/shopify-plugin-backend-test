package admin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/getbread/breadkit/zeus/searcher"
	"github.com/getbread/shopify_plugin_backend/service/search"
	"github.com/getbread/shopify_plugin_backend/service/types"
	"github.com/getbread/shopify_plugin_backend/service/update"
)

const GOOGLE_DISCOVERY_DOCUMENT_ENDPOINT = "https://accounts.google.com/.well-known/openid-configuration"

func findAllShops(h *Handlers) (shops []types.Shop, err error) {
	ssr := search.ShopSearchRequest{}
	shops, err = h.ShopSearcher.Search(ssr)
	if err != nil {
		return
	}
	if len(shops) == 0 {
		err = fmt.Errorf("shop not found")
		return
	}
	return
}

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

func saveShopSettings(req shopSettingRequest, h *Handlers) error {
	shop, err := findShopByName(req.ShopName, h)
	if err != nil {
		return err
	}

	updateRequest := update.ShopUpdateRequest{
		Id: shop.Id,
		Updates: map[update.ShopUpdateField]interface{}{
			update.ShopUpdate_AcceleratedCheckoutPermitted: req.EnableAcceleratedCheckout,
			update.ShopUpdate_POSAccess:                    req.POSAccess,
		},
	}

	err = h.ShopUpdater.Update(updateRequest)
	if err != nil {
		return err
	}
	return nil
}

type googleDiscoveryDocument struct {
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
}

func getGoogleDiscoveryDocument() (*googleDiscoveryDocument, error) {
	response, err := http.Get(GOOGLE_DISCOVERY_DOCUMENT_ENDPOINT)
	if err != nil {
		return nil, err
	}

	var res googleDiscoveryDocument
	err = json.NewDecoder(response.Body).Decode(&res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

type tokenRequest struct {
	Code         string `json:"code"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RedirectURI  string `json:"redirect_uri"`
	GrantType    string `json:"grant_type"`
}

type tokenResponse struct {
	AccessToken   string `json:"access_token"`
	ExpiresIn     int    `json:"expires_in"`
	Scope         string `json:"scope"`
	TokenType     string `json:"token_type"`
	IdentityToken string `json:"id_token"`
}

func exchangeAuthCodeForAccessToken(code string) (*tokenResponse, error) {
	// Google endpoints can change so we use the discovery document
	dd, err := getGoogleDiscoveryDocument()
	if err != nil {
		return nil, err
	}

	// Request body
	req := tokenRequest{
		Code:         code,
		ClientID:     oauthClientID,
		ClientSecret: oauthClientSecret,
		RedirectURI:  getOAuthRedirectURI(),
		GrantType:    "authorization_code",
	}

	bb, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	response, err := http.Post(dd.TokenEndpoint, "application/json", bytes.NewBuffer(bb))
	if err != nil {
		return nil, err
	}

	var res tokenResponse
	err = json.NewDecoder(response.Body).Decode(&res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}
