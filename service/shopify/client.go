package shopify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type Client struct {
	*http.Client
	ShopName    string
	AccessToken string
}

type Link struct {
	Url       string
	Direction string
}

type Links struct {
	Previous *Link
	Next     *Link
}

func NewClient(shopName, accessToken string) *Client {
	return &Client{&http.Client{}, shopName, accessToken}
}

func (c *Client) ApiRequest(method, url string, payload []byte, body interface{}) error {
	return c.ApiRequestWithLinkHeader(method, url, payload, body, nil)
}

func (c *Client) ApiRequestWithLinkHeader(method, url string, payload []byte, body interface{}, links *Links) error {
	var req *http.Request
	var err error
	if method == "GET" {
		req, err = http.NewRequest(method, url, nil)
	} else {

		req, err = http.NewRequest(method, url, bytes.NewBuffer(payload))
	}
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("X-Shopify-Access-Token", c.AccessToken)
	res, err := c.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode >= 400 {
		contents, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf(string(contents))
	}
	err = json.NewDecoder(res.Body).Decode(body)
	if err != nil {
		return err
	}
	if links != nil {
		*links, err = parseLinkHeader(res.Header.Get("Link"))
	}
	return err
}

func parseLinkHeader(linkHeader string) (Links, error) {
	var links Links

	if linkHeader == "" {
		return links, nil
	}

	unquoted, err := strconv.Unquote(linkHeader)
	if err != nil {
		return links, fmt.Errorf("Unable to unquote the Link header due to error: %+v\nLink Header: %s", err, linkHeader)
	}

	split := strings.Split(unquoted, ",")
	if len(split) > 2 {
		return links, fmt.Errorf("Expected at most 2 links relative to this one; found %d.\n Link Header: %s", len(split), linkHeader)
	}

	for _, link := range split {
		info := strings.Split(strings.TrimSpace(link), ";")
		if len(info) != 2 {
			return links, fmt.Errorf("Expected link %s to be split by a single semi-colon; instead found %d pieces.\n Link Header: %s", link, len(info), linkHeader)
		}

		url := strings.Trim(strings.TrimSpace(info[0]), "<>")

		relInfo := strings.Split(strings.TrimSpace(info[1]), "=")
		if len(relInfo) != 2 || relInfo[0] != "rel" {
			return links, fmt.Errorf("Expected second half of link (%s) to be a relation split by an equality sign; instead found %d pieces.\nLink Header: %s", info[1], len(relInfo), linkHeader)
		}

		parsedLink := &Link{
			Url:       url,
			Direction: relInfo[1],
		}

		if relInfo[1] == "previous" {
			links.Previous = parsedLink
		} else if relInfo[1] == "next" {
			links.Next = parsedLink
		} else {
			return links, fmt.Errorf("Invalid relation direction %s; expected \"previous\" or \"next\".\nLinkHeader: %s", relInfo[1], linkHeader)
		}
	}

	return links, nil
}

func (c *Client) Request(method, url string, payload []byte) (*http.Response, error) {
	var req *http.Request
	var err error
	if method == "GET" {
		req, err = http.NewRequest(method, url, nil)
	} else {
		req, err = http.NewRequest(method, url, bytes.NewBuffer(payload))
	}
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.Header.Set("X-Shopify-Access-Token", c.AccessToken)
	return c.Do(req)
}

func (c *Client) CreateCustomer(req *CreateCustomerRequest, res *CreateCustomerResponse) error {
	url := CreateCustomerUrl(c.ShopName)
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}
	return c.ApiRequest("POST", url, body, res)
}

func (c *Client) CreateOrder(req *CreateOrderRequest, res *CreateOrderResponse) error {
	url := CreateOrderUrl(c.ShopName)
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}
	return c.ApiRequest("POST", url, body, res)
}

func (c *Client) CancelOrder(orderId int, res *CreateOrderResponse) error {
	url := CancelOrderUrl(c.ShopName, strconv.Itoa(orderId))
	return c.ApiRequest("POST", url, []byte{}, res)
}

func (c *Client) UpdateOrder(orderId int, req *UpdateOrderRequest, res *CreateOrderResponse) error {
	url := UpdateOrderUrl(c.ShopName, strconv.Itoa(orderId))
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}
	return c.ApiRequest("PUT", url, body, res)
}

func (c *Client) CreateTransaction(orderId int, req *CreateTransactionRequest, res *CreateTransactionResponse) error {
	url := CreateTransactionUrl(c.ShopName, strconv.Itoa(orderId))
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}
	return c.ApiRequest("POST", url, body, res)
}

func (c *Client) DeleteWebhook(id int) error {
	var resp DeleteWebhookResponse
	url := SingleWebhookUrl(c.ShopName, id)
	return c.ApiRequest("DELETE", url, []byte{}, &resp)
}

func (c *Client) EmbedScript(req *EmbedScriptRequest, res *EmbedScriptResponse) error {
	url := EmbedScriptUrl(c.ShopName)
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}
	return c.ApiRequest("POST", url, body, res)
}

func (c *Client) ExchangeOAuthCode(req *OAuthExchangeRequest, res *OAuthExchangeResponse) error {
	url := OAuthExchangeUrl(c.ShopName)
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}
	return c.ApiRequest("POST", url, body, res)
}

func (c *Client) QueryCustomer(params map[string]string, res *SearchCustomerResponse) error {
	// get base url
	address := SearchCustomerUrl(c.ShopName)

	// construct url query
	var query string
	for key, value := range params {
		if len(query) > 1 {
			query += " " // add a space for each query param
		}
		query += key + ":" + value
	}
	values := url.Values{}
	values.Add("query", query)
	values.Add("limit", "1")
	address += "?" + values.Encode()

	// request
	return c.ApiRequest("GET", address, []byte{}, res)
}

func (c *Client) QueryOrder(orderId string, res *SearchOrderResponse) error {
	url := SearchOrderUrl(c.ShopName, orderId)
	return c.ApiRequest("GET", url, []byte{}, res)
}

func (c *Client) QueryTransactions(orderId string, res *SearchTransactionsResponse) error {
	url := SearchTransactionsUrl(c.ShopName, orderId)
	return c.ApiRequest("GET", url, []byte{}, res)
}

func (c *Client) QueryTransaction(orderId, transactionId string, res *SearchTransactionResponse) error {
	url := SearchTransactionUrl(c.ShopName, orderId, transactionId)
	return c.ApiRequest("GET", url, []byte{}, res)
}

func (c *Client) QueryProduct(productId string, res *SearchProductByIdResponse) error {
	url := SearchProductByIdUrl(c.ShopName, productId)
	return c.ApiRequest("GET", url, []byte{}, res)
}

func (c *Client) QueryProductVariant(variantId string, res *SearchProductVariantByIdResponse) error {
	url := SearchProductVariantUrl(c.ShopName, variantId)
	return c.ApiRequest("GET", url, []byte{}, res)
}

func (c *Client) QueryWebhook(shopName string, res *QueryWebhooksResponse) error {
	url := WebhookUrl(shopName)
	return c.ApiRequest("GET", url, []byte{}, res)
}

func (c *Client) RegisterWebhook(req *RegisterWebhookRequest, res *RegisterWebhookResponse) error {
	url := WebhookUrl(c.ShopName)
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}
	return c.ApiRequest("POST", url, body, res)
}

func (c *Client) QueryLocations(res *SearchLocationsResponse) error {
	url := LocationUrl(c.ShopName)
	return c.ApiRequest("GET", url, []byte{}, res)
}

func (c *Client) QueryShop(res *SearchShopResponse) error {
	url := ShopUrl(c.ShopName)
	return c.ApiRequest("GET", url, []byte{}, res)
}

func (c *Client) QueryEmbeddedScripts(v url.Values, res *SearchEmbeddedScriptResponse) error {
	url := SearchEmbeddedScriptUrl(c.ShopName)
	url += "?" + v.Encode()
	return c.ApiRequest("GET", url, []byte{}, res)
}

func (c *Client) DeleteEmbeddedScript(scriptID string, res *DeleteEmbeddedScriptResponse) error {
	url := DeleteEmbeddedScriptUrl(c.ShopName, scriptID)
	return c.ApiRequest("DELETE", url, []byte{}, res)
}

func (c *Client) GetDraftOrdersCount(query url.Values, res *GetDraftOrdersCountResponse) error {
	url := GetDraftOrdersCountUrl(c.ShopName, query)
	return c.ApiRequest("GET", url, nil, res)
}

func (c *Client) GetDraftOrders(query url.Values, res *GetDraftOrdersResponse, links *Links) error {
	url := GetDraftOrdersUrl(c.ShopName, query)
	return c.ApiRequestWithLinkHeader("GET", url, nil, res, links)
}

func (c *Client) GetDraftOrder(id string, res *GetDraftOrderResponse) error {
	url := GetDraftOrderUrl(c.ShopName, id)
	return c.ApiRequest("GET", url, nil, res)
}

func (c *Client) GetDraftOrderRequest(id string, res *GetDraftOrderResponse) (int, error) {
	url := GetDraftOrderUrl(c.ShopName, id)
	response, err := c.Request("GET", url, nil)
	if err != nil {
		return 500, err
	}
	defer response.Body.Close()
	err = json.NewDecoder(response.Body).Decode(res)
	return response.StatusCode, err
}

func (c *Client) CreateDraftOrder(req *CreateDraftOrderRequest, res *GetDraftOrderResponse) error {
	url := CreateDraftOrderUrl(c.ShopName)
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}
	return c.ApiRequest("POST", url, body, res)
}

func (c *Client) CreateDraftOrderRequest(req *CreateDraftOrderRequest, res *GetDraftOrderResponse) (int, error) {
	url := CreateDraftOrderUrl(c.ShopName)
	body, err := json.Marshal(req)
	if err != nil {
		return 500, err
	}
	response, err := c.Request("POST", url, body)
	if err != nil {
		return 500, err
	}
	defer response.Body.Close()
	err = json.NewDecoder(response.Body).Decode(res)
	return response.StatusCode, err
}

func (c *Client) CompleteDraftOrder(id string, res *GetDraftOrderResponse) error {
	url := CompleteDraftOrderUrl(c.ShopName, id)
	return c.ApiRequest("PUT", url, nil, res)
}

func (c *Client) DeleteDraftOrder(orderId string, res *DeleteDraftOrderResponse) error {
	url := DeleteDraftOrderUrl(c.ShopName, orderId)
	return c.ApiRequest("DELETE", url, []byte{}, res)
}
