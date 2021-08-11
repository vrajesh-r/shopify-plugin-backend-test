package bread

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	log "github.com/sirupsen/logrus"
)

type HTTPClient interface {
	ApiRequest(method, url string, body []byte, response interface{}, headers map[string]string) (interface{}, error)
}

type TrxHTTPClient struct {
	*http.Client
}

func NewHTTPClient() HTTPClient {
	return &TrxHTTPClient{&http.Client{}}
}

type TrxProcessor interface {
	AuthorizeTransaction(host, trxID string, trxReq *TrxRequest, trxRes *TrxResponse, headers map[string]string) (*TrxResponse, error)
	SettleTransaction(host, trxID string, trxReq *TrxRequest, trxRes *TrxResponse, headers map[string]string) (*TrxResponse, error)
	RefundTransaction(host, trxID string, trxReq *TrxRequest, trxRes *TrxResponse, headers map[string]string) (*TrxResponse, error)
	CancelTransaction(host, trxID string, trxReq *TrxRequest, trxRes *TrxResponse, headers map[string]string) (*TrxResponse, error)
	GetTransactionAuthToken(host string, req *TrxAuthTokenRequest, trxRes *TrxAuthTokenResponse) (*TrxAuthTokenResponse, error)
	GetTransactionFromApplication(host, applicationID string, headers map[string]string, res *TrxResponse) (*TrxResponse, error)
	GetTransaction(host, transactionID string, res *TrxResponse, headers map[string]string) (*TrxResponse, error)
}

type PlatformTrxProcessor struct {
	httpClient HTTPClient
}

func NewTrxProcessor() TrxProcessor {
	return &PlatformTrxProcessor{httpClient: NewHTTPClient()}
}

type HttpError struct {
	StatusCode int
	Err        error
}

func (httpError HttpError) Error() string {
	return httpError.Err.Error()
}

func NewHttpError(code int, err error) HttpError {
	return HttpError{StatusCode: code, Err: err}
}

func (tc *TrxHTTPClient) ApiRequest(
	method, url string,
	body []byte,
	r interface{},
	headers map[string]string) (interface{}, error) {
	// create request
	var req *http.Request
	var err error
	if method == "GET" {
		req, err = http.NewRequest(method, url, nil)
	} else {

		req, err = http.NewRequest(method, url, bytes.NewBuffer(body))
	}
	if err != nil {
		return nil, err
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	// execute request
	res, err := tc.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	// parse response
	contents, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode >= 400 {
		log.WithFields(log.Fields{
			"error":  string(contents),
			"code":   res.StatusCode,
			"status": res.Status,
			"url":    url,
			"method": method,
		}).Errorf("(TrxApiRequest) Request returned a response with status code %d", res.StatusCode)

		errResponse := &TrxErrorResponse{}
		if err = json.Unmarshal(contents, errResponse); err != nil {
			return nil, NewHttpError(res.StatusCode, fmt.Errorf(string(contents)))
		}

		return nil, NewHttpError(res.StatusCode, errors.New(errResponse.Reason))
	}

	if len(string(contents)) > 0 {
		if err = json.Unmarshal(contents, r); err != nil {
			return nil, err
		}
	}

	return r, nil
}

func assertTrxResponse(interfaceType interface{}) (*TrxResponse, error) {
	trxResponse, ok := interfaceType.(*TrxResponse)
	if !ok {
		return nil, errors.New("asserting interface is of type *bread.TrxResponse failed")
	}

	return trxResponse, nil
}

func (tp *PlatformTrxProcessor) AuthorizeTransaction(
	host, trxID string,
	trxReq *TrxRequest,
	trxRes *TrxResponse,
	headers map[string]string) (*TrxResponse, error) {

	url := AuthorizeTransactionURL(host, trxID)
	body, err := json.Marshal(trxReq)
	if err != nil {
		return nil, err
	}

	resInterface, err := tp.httpClient.ApiRequest("POST", url, body, trxRes, headers)
	if err != nil {
		return nil, err
	}

	return assertTrxResponse(resInterface)
}

func (tp *PlatformTrxProcessor) CancelTransaction(
	host, trxID string,
	trxReq *TrxRequest,
	trxRes *TrxResponse,
	headers map[string]string) (*TrxResponse, error) {

	url := CancelTransactionURL(host, trxID)
	body, err := json.Marshal(trxReq)
	if err != nil {
		return nil, err
	}

	resInterface, err := tp.httpClient.ApiRequest("POST", url, body, trxRes, headers)
	if err != nil {
		return nil, err
	}

	return assertTrxResponse(resInterface)
}

func (tp *PlatformTrxProcessor) RefundTransaction(
	host, trxID string,
	trxReq *TrxRequest,
	trxRes *TrxResponse,
	headers map[string]string) (*TrxResponse, error) {

	url := RefundTransactionURL(host, trxID)
	body, err := json.Marshal(trxReq)
	if err != nil {
		return nil, err
	}

	resInterface, err := tp.httpClient.ApiRequest("POST", url, body, trxRes, headers)
	if err != nil {
		return nil, err
	}

	return assertTrxResponse(resInterface)
}

func (tp *PlatformTrxProcessor) SettleTransaction(
	host, trxID string,
	trxReq *TrxRequest,
	trxRes *TrxResponse,
	headers map[string]string) (*TrxResponse, error) {

	url := SettleTransactionURL(host, trxID)
	body, err := json.Marshal(trxReq)
	if err != nil {
		return nil, err
	}

	resInterface, err := tp.httpClient.ApiRequest("POST", url, body, trxRes, headers)
	if err != nil {
		return nil, err
	}

	return assertTrxResponse(resInterface)
}

func assertTrxAuthTokenResponse(interfaceType interface{}) (*TrxAuthTokenResponse, error) {
	response, ok := interfaceType.(*TrxAuthTokenResponse)
	if !ok {
		return nil, errors.New("asserting interface is of type *bread.TrxResponse failed")
	}

	return response, nil
}

func (tp *PlatformTrxProcessor) GetTransactionAuthToken(
	host string,
	req *TrxAuthTokenRequest,
	trxRes *TrxAuthTokenResponse) (*TrxAuthTokenResponse, error) {

	url := TransactionAuthTokenURL(host)
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	headers := map[string]string{"Content-Type": "application/json"}
	resInterface, err := tp.httpClient.ApiRequest("POST", url, body, trxRes, headers)
	if err != nil {
		return nil, err
	}

	return assertTrxAuthTokenResponse(resInterface)
}

func (tp *PlatformTrxProcessor) GetTransactionFromApplication(
	host, applicationID string,
	headers map[string]string,
	res *TrxResponse) (*TrxResponse, error) {

	url := GetTransactionFromApplicationURL(host, applicationID)
	resInterface, err := tp.httpClient.ApiRequest("GET", url, nil, res, headers)
	if err != nil {
		return nil, err
	}

	return assertTrxResponse(resInterface)

}

func (tp *PlatformTrxProcessor) GetTransaction(
	host, transactionID string,
	res *TrxResponse,
	headers map[string]string) (*TrxResponse, error) {

	url := GetTransactionURL(host, transactionID)
	resInterface, err := tp.httpClient.ApiRequest("GET", url, nil, res, headers)
	if err != nil {
		return nil, err
	}

	return assertTrxResponse(resInterface)
}
