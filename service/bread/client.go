package bread

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type BreadClient struct {
	*http.Client
	ApiKey    string
	SecretKey string
}

func NewClient(apiKey, secretKey string) *BreadClient {
	return &BreadClient{&http.Client{}, apiKey, secretKey}
}

func (bc *BreadClient) ApiRequest(method, url string, body []byte, r interface{}) error {
	// create request
	var req *http.Request
	var err error
	if method == "GET" {
		req, err = http.NewRequest(method, url, nil)
	} else {

		req, err = http.NewRequest(method, url, bytes.NewBuffer(body))
	}
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(bc.ApiKey, bc.SecretKey)

	// execute request
	res, err := bc.Do(req)
	if err != nil {
		return err
	}

	// parse request
	contents, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if res.StatusCode >= 400 {
		return fmt.Errorf(string(contents))
	}

	if len(string(contents)) > 0 {
		if err = json.Unmarshal(contents, r); err != nil {
			return err
		}
	}
	return nil
}

func (bc *BreadClient) QueryTransaction(txId, ostiaHost string) (*TransactionResponse, error) {
	url := QueryTransactionUrl(txId, ostiaHost)
	var tr TransactionResponse
	if err := bc.ApiRequest("GET", url, nil, &tr); err != nil {
		return nil, err
	}
	return &tr, nil
}

func (bc *BreadClient) transactionActionPut(txId, ostiaHost string, tar *TransactionActionRequest, tr *TransactionResponse) error {
	url := QueryTransactionUrl(txId, ostiaHost)
	body, err := json.Marshal(tar)
	if err != nil {
		return err
	}
	return bc.ApiRequest("PUT", url, body, tr)
}

func (bc *BreadClient) transactionActionPost(txId, ostiaHost string, tar *TransactionActionRequest, tr *TransactionResponse) error {
	url := TransactionActionUrl(txId, ostiaHost)
	body, err := json.Marshal(tar)
	if err != nil {
		return err
	}
	return bc.ApiRequest("POST", url, body, tr)
}

func (bc *BreadClient) UpdateTransaction(txId, ostiaHost string, tar *TransactionActionRequest) (*TransactionResponse, error) {
	var tr TransactionResponse
	if err := bc.transactionActionPut(txId, ostiaHost, tar, &tr); err != nil {
		return nil, err
	}
	return &tr, nil
}

func (bc *BreadClient) AuthorizeTransaction(txId, ostiaHost string, tar *TransactionActionRequest) (*TransactionResponse, error) {
	var tr TransactionResponse
	if err := bc.transactionActionPost(txId, ostiaHost, tar, &tr); err != nil {
		return nil, err
	}
	return &tr, nil
}

func (bc *BreadClient) SettleTransaction(txId, ostiaHost string, tar *TransactionActionRequest) (*TransactionResponse, error) {
	var tr TransactionResponse
	if err := bc.transactionActionPost(txId, ostiaHost, tar, &tr); err != nil {
		return nil, err
	}
	return &tr, nil
}

func (bc *BreadClient) CancelTransaction(txId, ostiaHost string, tar *TransactionActionRequest) (*TransactionResponse, error) {
	var tr TransactionResponse
	if err := bc.transactionActionPost(txId, ostiaHost, tar, &tr); err != nil {
		return nil, err
	}
	return &tr, nil
}

func (bc *BreadClient) RefundTransaction(txId, ostiaHost string, tar *TransactionActionRequest) (*TransactionResponse, error) {
	var tr TransactionResponse
	if err := bc.transactionActionPost(txId, ostiaHost, tar, &tr); err != nil {
		return nil, err
	}
	return &tr, nil
}

func (bc *BreadClient) SaveCart(ostiaHost string, cart *Cart) (*Cart, error) {
	url := SaveCartUrl(ostiaHost)
	body, err := json.Marshal(cart)
	if err != nil {
		return nil, err
	}

	var resp Cart
	if err := bc.ApiRequest("POST", url, body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (bc *BreadClient) SendCartEmail(ostiaHost, cartID string, req SendCartEmailRequest) error {
	url := SendCartEmailUrl(ostiaHost, cartID)
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}
	return bc.ApiRequest("POST", url, body, &MessageResponse{})

}

func (bc *BreadClient) SendCartText(ostiaHost, cartID string, req SendCartTextRequest) error {
	url := SendCartTextUrl(ostiaHost, cartID)
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}
	return bc.ApiRequest("POST", url, body, &MessageResponse{})
}

func (bc *BreadClient) ExpireCart(ostiaHost, cartID string) error {
	url := ExpireCartUrl(ostiaHost, cartID)
	return bc.ApiRequest("POST", url, nil, nil)
}

func (bc *BreadClient) SetShippingDetails(transactionID, ostiaHost string, req TransactionShipmentRequest) (*TransactionResponse, error) {
	var tr TransactionResponse
	url := TransactionShipmentURL(transactionID, ostiaHost)
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	err = bc.ApiRequest("POST", url, body, &tr)
	if err != nil {
		return nil, err
	}
	return &tr, nil
}
