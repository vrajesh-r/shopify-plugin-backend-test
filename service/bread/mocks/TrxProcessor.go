// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package mocks

import (
	bread "github.com/getbread/shopify_plugin_backend/service/bread"
	mock "github.com/stretchr/testify/mock"
)

// TrxProcessor is an autogenerated mock type for the TrxProcessor type
type TrxProcessor struct {
	mock.Mock
}

// AuthorizeTransaction provides a mock function with given fields: host, trxID, trxReq, trxRes, headers
func (_m *TrxProcessor) AuthorizeTransaction(host string, trxID string, trxReq *bread.TrxRequest, trxRes *bread.TrxResponse, headers map[string]string) (*bread.TrxResponse, error) {
	ret := _m.Called(host, trxID, trxReq, trxRes, headers)

	var r0 *bread.TrxResponse
	if rf, ok := ret.Get(0).(func(string, string, *bread.TrxRequest, *bread.TrxResponse, map[string]string) *bread.TrxResponse); ok {
		r0 = rf(host, trxID, trxReq, trxRes, headers)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*bread.TrxResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, *bread.TrxRequest, *bread.TrxResponse, map[string]string) error); ok {
		r1 = rf(host, trxID, trxReq, trxRes, headers)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CancelTransaction provides a mock function with given fields: host, trxID, trxReq, trxRes, headers
func (_m *TrxProcessor) CancelTransaction(host string, trxID string, trxReq *bread.TrxRequest, trxRes *bread.TrxResponse, headers map[string]string) (*bread.TrxResponse, error) {
	ret := _m.Called(host, trxID, trxReq, trxRes, headers)

	var r0 *bread.TrxResponse
	if rf, ok := ret.Get(0).(func(string, string, *bread.TrxRequest, *bread.TrxResponse, map[string]string) *bread.TrxResponse); ok {
		r0 = rf(host, trxID, trxReq, trxRes, headers)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*bread.TrxResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, *bread.TrxRequest, *bread.TrxResponse, map[string]string) error); ok {
		r1 = rf(host, trxID, trxReq, trxRes, headers)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetTransaction provides a mock function with given fields: host, transactionID, res, headers
func (_m *TrxProcessor) GetTransaction(host string, transactionID string, res *bread.TrxResponse, headers map[string]string) (*bread.TrxResponse, error) {
	ret := _m.Called(host, transactionID, res, headers)

	var r0 *bread.TrxResponse
	if rf, ok := ret.Get(0).(func(string, string, *bread.TrxResponse, map[string]string) *bread.TrxResponse); ok {
		r0 = rf(host, transactionID, res, headers)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*bread.TrxResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, *bread.TrxResponse, map[string]string) error); ok {
		r1 = rf(host, transactionID, res, headers)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetTransactionAuthToken provides a mock function with given fields: host, req, trxRes
func (_m *TrxProcessor) GetTransactionAuthToken(host string, req *bread.TrxAuthTokenRequest, trxRes *bread.TrxAuthTokenResponse) (*bread.TrxAuthTokenResponse, error) {
	ret := _m.Called(host, req, trxRes)

	var r0 *bread.TrxAuthTokenResponse
	if rf, ok := ret.Get(0).(func(string, *bread.TrxAuthTokenRequest, *bread.TrxAuthTokenResponse) *bread.TrxAuthTokenResponse); ok {
		r0 = rf(host, req, trxRes)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*bread.TrxAuthTokenResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, *bread.TrxAuthTokenRequest, *bread.TrxAuthTokenResponse) error); ok {
		r1 = rf(host, req, trxRes)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetTransactionFromApplication provides a mock function with given fields: host, applicationID, headers, res
func (_m *TrxProcessor) GetTransactionFromApplication(host string, applicationID string, headers map[string]string, res *bread.TrxResponse) (*bread.TrxResponse, error) {
	ret := _m.Called(host, applicationID, headers, res)

	var r0 *bread.TrxResponse
	if rf, ok := ret.Get(0).(func(string, string, map[string]string, *bread.TrxResponse) *bread.TrxResponse); ok {
		r0 = rf(host, applicationID, headers, res)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*bread.TrxResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, map[string]string, *bread.TrxResponse) error); ok {
		r1 = rf(host, applicationID, headers, res)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RefundTransaction provides a mock function with given fields: host, trxID, trxReq, trxRes, headers
func (_m *TrxProcessor) RefundTransaction(host string, trxID string, trxReq *bread.TrxRequest, trxRes *bread.TrxResponse, headers map[string]string) (*bread.TrxResponse, error) {
	ret := _m.Called(host, trxID, trxReq, trxRes, headers)

	var r0 *bread.TrxResponse
	if rf, ok := ret.Get(0).(func(string, string, *bread.TrxRequest, *bread.TrxResponse, map[string]string) *bread.TrxResponse); ok {
		r0 = rf(host, trxID, trxReq, trxRes, headers)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*bread.TrxResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, *bread.TrxRequest, *bread.TrxResponse, map[string]string) error); ok {
		r1 = rf(host, trxID, trxReq, trxRes, headers)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SettleTransaction provides a mock function with given fields: host, trxID, trxReq, trxRes, headers
func (_m *TrxProcessor) SettleTransaction(host string, trxID string, trxReq *bread.TrxRequest, trxRes *bread.TrxResponse, headers map[string]string) (*bread.TrxResponse, error) {
	ret := _m.Called(host, trxID, trxReq, trxRes, headers)

	var r0 *bread.TrxResponse
	if rf, ok := ret.Get(0).(func(string, string, *bread.TrxRequest, *bread.TrxResponse, map[string]string) *bread.TrxResponse); ok {
		r0 = rf(host, trxID, trxReq, trxRes, headers)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*bread.TrxResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, *bread.TrxRequest, *bread.TrxResponse, map[string]string) error); ok {
		r1 = rf(host, trxID, trxReq, trxRes, headers)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
