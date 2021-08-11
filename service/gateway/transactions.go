package gateway

import (
	"errors"
	"fmt"

	"github.com/getbread/shopify_plugin_backend/service/bread"
	"github.com/getbread/shopify_plugin_backend/service/cache"
	"github.com/getbread/shopify_plugin_backend/service/types"
	log "github.com/sirupsen/logrus"
)

const TrxAuthTokenKey = "transaction-service-auth-token"

var ErrByteArrayAssertion = errors.New("failed asserting value is []byte")

var authTokenToCache = func(cache cache.Cache, key, value string) (interface{}, error) {
	return cache.Write(key, value)
}

var authTokenFromCache = func(cache cache.Cache, key string) (string, error) {
	rawToken, err := cache.Get(key)
	if err != nil {
		return "", err
	}

	byteToken, ok := rawToken.([]byte)
	if !ok {
		return "", ErrByteArrayAssertion
	}

	return string(byteToken), nil
}

var authTokenFromRemote = func(host string, tp bread.TrxProcessor, req *bread.TrxAuthTokenRequest) (string, error) {

	res, err := tp.GetTransactionAuthToken(host, req, &bread.TrxAuthTokenResponse{})
	if err != nil {
		return "", err
	}

	return res.Token, nil
}

func authTokenRemoteRequestAndCache(host string, tp bread.TrxProcessor, cache cache.Cache, req *bread.TrxAuthTokenRequest) (string, error) {

	token, err := authTokenFromRemote(host, tp, req)
	if err != nil {
		return "", err
	}

	go func() {
		if _, err := authTokenToCache(cache, TrxAuthTokenKey, token); err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Warn("(authTokenRequestAndCache) Could not save token to cache")
		}
	}()

	return token, nil
}

func platformAuthorizeTransaction(host, apiKey, apiSecret, trxID string, trxReq bread.TrxRequest, cache cache.Cache) *HttpError {

	res := new(bread.TrxResponse)
	tp := bread.NewTrxProcessor()
	authTokenReq := &bread.TrxAuthTokenRequest{
		ApiKey: apiKey,
		Secret: apiSecret,
	}

	// Attempt to get auth token from cache
	token, err := authTokenFromCache(cache, TrxAuthTokenKey)
	if err != nil {
		log.WithFields(log.Fields{
			"error":        err.Error(),
			"authTokenKey": TrxAuthTokenKey,
		}).Error("(platformAuthorizeTransaction) Could not retrieve auth token from cache")

		log.Infof("(platformAuthorizeTransaction) Requesting auth token from transaction service ...")

		token, err = authTokenRemoteRequestAndCache(host, tp, cache, authTokenReq)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("(platformAuthorizeTransaction) Could not retrieve auth token from transaction service")

			return NewHttpError("An error occurred while processing transaction", 500)
		}

		log.Info("(platformAuthorizeTransaction) Auth token retrieved from transaction service")
	}

	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", token),
		"Content-Type":  "application/json",
	}

	if _, err = tp.AuthorizeTransaction(host, trxID, &trxReq, res, headers); err != nil {
		httpError, isHttpError := err.(bread.HttpError)

		if isHttpError {
			if httpError.StatusCode == 401 {
				// Token has most likely expired
				//Attempt to retrive new token from remote and cache new token
				log.Info("(platformAuthorizeTransaction) Requesting new token from transaction service ...")

				token, err = authTokenRemoteRequestAndCache(host, tp, cache, authTokenReq)
				if err != nil {
					log.WithFields(log.Fields{
						"error": err.Error(),
					}).Error("(platformAuthorizeTransaction) Could not retrieve auth token from transaction service")

					return NewHttpError("An error occurred while processing transaction", 500)
				}

				log.Info("(platformAuthorizeTransaction) New token retrieved from transaction service")

				//Update headers with new token
				headers["Authorization"] = fmt.Sprintf("Bearer %s", token)

				//Resend transaction processing request
				if _, err = tp.AuthorizeTransaction(host, trxID, &trxReq, res, headers); err != nil {
					log.WithFields(log.Fields{
						"error":         err.Error(),
						"transactionID": trxID,
						"trxReq":        trxReq,
					}).Error("(platformAuthorizeTransaction) Authorize transaction retry failed")

					return NewHttpError("An error occurred while processing transaction", 400)

				}
			} else { // Not a 401 bread.HttpError
				log.WithFields(log.Fields{
					"error":         err.Error(),
					"transactionID": trxID,
					"trxReq":        trxReq,
				}).Error("(platformAuthorizeTransaction) Initial request to authorize transaction failed with bread.HttpError")

				return NewHttpError("An error occurred while processing transaction", 400)
			}
		} else { // Not a bread.HTTPError
			log.WithFields(log.Fields{
				"error":         err.Error(),
				"transactionID": trxID,
				"trxReq":        trxReq,
			}).Error("(platformAuthorizeTransaction) Initial request to authorize transaction failed")

			return NewHttpError("An error occurred while processing transaction", 400)
		}
	}

	log.WithFields(log.Fields{
		"response": res,
	}).Info("(platformAuthorizeTransaction) Transaction authorization success")

	return nil
}

func platformSettleTransaction(host, apiKey, apiSecret, trxID string, trxReq bread.TrxRequest, cache cache.Cache) *HttpError {

	res := new(bread.TrxResponse)
	tp := bread.NewTrxProcessor()
	authTokenReq := &bread.TrxAuthTokenRequest{
		ApiKey: apiKey,
		Secret: apiSecret,
	}

	// Attempt to get auth token from cache
	token, err := authTokenFromCache(cache, TrxAuthTokenKey)
	if err != nil {
		log.WithFields(log.Fields{
			"error":        err.Error(),
			"authTokenKey": TrxAuthTokenKey,
		}).Error("(platformSettleTransaction) Could not retrieve auth token from cache")

		log.Infof("(platformSettleTransaction) Requesting auth token from transaction service ...")

		token, err = authTokenRemoteRequestAndCache(host, tp, cache, authTokenReq)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("(platformSettleTransaction) Could not retrieve auth token from transaction service")

			return NewHttpError("An error occurred while processing transaction", 500)
		}

		log.Info("(platformSettleTransaction) Auth token retrieved from transaction service")
	}

	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", token),
		"Content-Type":  "application/json",
	}

	if _, err = tp.SettleTransaction(host, trxID, &trxReq, res, headers); err != nil {
		httpError, isHttpError := err.(bread.HttpError)

		if isHttpError {
			if httpError.StatusCode == 401 {
				// Token has most likely expired
				//Attempt to retrive new token from remote and cache new token
				log.Info("(platformSettleTransaction) Requesting new token from transaction service ...")

				token, err = authTokenRemoteRequestAndCache(host, tp, cache, authTokenReq)
				if err != nil {
					log.WithFields(log.Fields{
						"error": err.Error(),
					}).Error("(platformSettleTransaction) Could not retrieve auth token from transaction service")

					return NewHttpError("An error occurred while processing transaction", 500)
				}

				log.Info("(platformSettleTransaction) New token retrieved from transaction service")

				//Update headers with new token
				headers["Authorization"] = fmt.Sprintf("Bearer %s", token)

				//Resend transaction processing request
				if _, err = tp.SettleTransaction(host, trxID, &trxReq, res, headers); err != nil {
					log.WithFields(log.Fields{
						"error":         err.Error(),
						"transactionID": trxID,
					}).Error("(platformSettleTransaction) Settle transaction retry failed")

					return NewHttpError("An error occurred while processing transaction", 400)

				}
			} else { // Not a 401 bread.HttpError
				log.WithFields(log.Fields{
					"error":         err.Error(),
					"transactionID": trxID,
				}).Error("(platformSettleTransaction) Initial request to settle transaction failed with bread.HttpError")

				return NewHttpError("An error occurred while processing transaction", 400)
			}
		} else { // Not a bread.HTTPError
			log.WithFields(log.Fields{
				"error":         err.Error(),
				"transactionID": trxID,
			}).Error("(platformSettleTransaction) Initial request to settle transaction failed")

			return NewHttpError("An error occurred while processing transaction", 400)
		}
	}

	log.WithFields(log.Fields{
		"response": res,
	}).Info("(platformSettleTransaction) Transaction settling success")

	return nil
}

func platformCancelTransaction(host, apiKey, apiSecret, trxID string, trxReq bread.TrxRequest, cache cache.Cache) *HttpError {

	res := new(bread.TrxResponse)
	tp := bread.NewTrxProcessor()
	authTokenReq := &bread.TrxAuthTokenRequest{
		ApiKey: apiKey,
		Secret: apiSecret,
	}

	// Attempt to get auth token from cache
	token, err := authTokenFromCache(cache, TrxAuthTokenKey)
	if err != nil {
		log.WithFields(log.Fields{
			"error":        err.Error(),
			"authTokenKey": TrxAuthTokenKey,
		}).Error("(platformCancelTransaction) Could not retrieve auth token from cache")

		log.Infof("(platformCancelTransaction) Requesting auth token from transaction service ...")

		token, err = authTokenRemoteRequestAndCache(host, tp, cache, authTokenReq)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("(platformCancelTransaction) Could not retrieve auth token from transaction service")

			return NewHttpError("An error occurred while processing transaction", 500)
		}

		log.Info("(platformCancelTransaction) Auth token retrieved from transaction service")
	}

	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", token),
		"Content-Type":  "application/json",
	}

	if _, err = tp.CancelTransaction(host, trxID, &trxReq, res, headers); err != nil {
		httpError, isHttpError := err.(bread.HttpError)

		if isHttpError {
			if httpError.StatusCode == 401 {
				// Token has most likely expired
				//Attempt to retrive new token from remote and cache new token
				log.Info("(platformCancelTransaction) Requesting new token from transaction service ...")

				token, err = authTokenRemoteRequestAndCache(host, tp, cache, authTokenReq)
				if err != nil {
					log.WithFields(log.Fields{
						"error": err.Error(),
					}).Error("(platformCancelTransaction) Could not retrieve auth token from transaction service")

					return NewHttpError("An error occurred while processing transaction", 500)
				}

				log.Info("(platformCancelTransaction) New token retrieved from transaction service")

				//Update headers with new token
				headers["Authorization"] = fmt.Sprintf("Bearer %s", token)

				//Resend transaction processing request
				if _, err = tp.CancelTransaction(host, trxID, &trxReq, res, headers); err != nil {
					log.WithFields(log.Fields{
						"error":         err.Error(),
						"transactionID": trxID,
					}).Error("(platformCancelTransaction) Cancel transaction retry failed")

					return NewHttpError("An error occurred while processing transaction", 400)

				}
			} else { // Not a 401 bread.HttpError
				log.WithFields(log.Fields{
					"error":         err.Error(),
					"transactionID": trxID,
				}).Error("(platformCancelTransaction) Initial request to cancel transaction failed with bread.HttpError")

				return NewHttpError("An error occurred while processing transaction", 400)
			}
		} else { // Not a bread.HTTPError
			log.WithFields(log.Fields{
				"error":         err.Error(),
				"transactionID": trxID,
			}).Error("(platformCancelTransaction) Initial request to cancel transaction failed")

			return NewHttpError("An error occurred while processing transaction", 400)
		}
	}

	log.WithFields(log.Fields{
		"response": res,
	}).Info("(platformCancelTransaction) Transaction cancelling success")

	return nil
}

func platformRefundTransaction(host, apiKey, apiSecret, trxID string, trxReq bread.TrxRequest, cache cache.Cache) *HttpError {

	res := new(bread.TrxResponse)
	tp := bread.NewTrxProcessor()
	authTokenReq := &bread.TrxAuthTokenRequest{
		ApiKey: apiKey,
		Secret: apiSecret,
	}

	// Attempt to get auth token from cache
	token, err := authTokenFromCache(cache, TrxAuthTokenKey)
	if err != nil {
		log.WithFields(log.Fields{
			"error":        err.Error(),
			"authTokenKey": TrxAuthTokenKey,
		}).Error("(platformRefundTransaction) Could not retrieve auth token from cache")

		log.Infof("(platformRefundTransaction) Requesting auth token from transaction service ...")

		token, err = authTokenRemoteRequestAndCache(host, tp, cache, authTokenReq)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("(platformRefundTransaction) Could not retrieve auth token from transaction service")

			return NewHttpError("An error occurred while processing transaction", 500)
		}

		log.Info("(platformRefundTransaction) Auth token retrieved from transaction service")
	}

	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", token),
		"Content-Type":  "application/json",
	}

	if _, err = tp.RefundTransaction(host, trxID, &trxReq, res, headers); err != nil {
		httpError, isHttpError := err.(bread.HttpError)

		if isHttpError {
			if httpError.StatusCode == 401 {
				// Token has most likely expired
				//Attempt to retrive new token from remote and cache new token
				log.Info("(platformRefundTransaction) Requesting new token from transaction service ...")

				token, err = authTokenRemoteRequestAndCache(host, tp, cache, authTokenReq)
				if err != nil {
					log.WithFields(log.Fields{
						"error": err.Error(),
					}).Error("(platformRefundTransaction) Could not retrieve auth token from transaction service")

					return NewHttpError("An error occurred while processing transaction", 500)
				}

				log.Info("(platformRefundTransaction) New token retrieved from transaction service")

				//Update headers with new token
				headers["Authorization"] = fmt.Sprintf("Bearer %s", token)

				//Resend transaction processing request
				if _, err = tp.RefundTransaction(host, trxID, &trxReq, res, headers); err != nil {
					log.WithFields(log.Fields{
						"error":         err.Error(),
						"transactionID": trxID,
					}).Error("(platformRefundTransaction) Refund transaction retry failed")

					return NewHttpError("An error occurred while processing transaction", 400)

				}
			} else { // Not a 401 bread.HttpError
				log.WithFields(log.Fields{
					"error":         err.Error(),
					"transactionID": trxID,
				}).Error("(platformRefundTransaction) Initial request to refund transaction failed with bread.HttpError")

				return NewHttpError("An error occurred while processing transaction", 400)
			}
		} else { // Not a bread.HTTPError
			log.WithFields(log.Fields{
				"error":         err.Error(),
				"transactionID": trxID,
			}).Error("(platformRefundTransaction) Initial request to refund transaction failed")

			return NewHttpError("An error occurred while processing transaction", 400)
		}
	}

	log.WithFields(log.Fields{
		"response": res,
	}).Info("(platformRefundTransaction) Transaction refund success")

	return nil
}

func platformGetTransaction(host, apiKey, apiSecret, trxID string, cache cache.Cache) (*bread.TrxResponse, *HttpError) {

	var trxRes *bread.TrxResponse
	res := new(bread.TrxResponse)
	tp := bread.NewTrxProcessor()
	authTokenReq := &bread.TrxAuthTokenRequest{
		ApiKey: apiKey,
		Secret: apiSecret,
	}

	// Attempt to get auth token from cache
	token, err := authTokenFromCache(cache, TrxAuthTokenKey)
	if err != nil {
		log.WithFields(log.Fields{
			"error":        err.Error(),
			"authTokenKey": TrxAuthTokenKey,
		}).Error("(platformGetTransaction) Could not retrieve auth token from cache")

		log.Infof("(platformGetTransaction) Requesting auth token from transaction service ...")

		token, err = authTokenRemoteRequestAndCache(host, tp, cache, authTokenReq)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Error("(platformGetTransaction) Could not retrieve auth token from transaction service")

			return nil, NewHttpError("An error occurred while retrieving transaction", 500)
		}

		log.Info("(platformGetTransaction) Auth token retrieved from transaction service")
	}

	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", token),
		"Content-Type":  "application/json",
	}

	if trxRes, err = tp.GetTransaction(host, trxID, res, headers); err != nil {
		httpError, isHttpError := err.(bread.HttpError)

		if isHttpError {
			if httpError.StatusCode == 401 {
				// Token has most likely expired
				//Attempt to retrive new token from remote and cache new token
				log.Info("(platformGetTransaction) Requesting new token from transaction service ...")

				token, err = authTokenRemoteRequestAndCache(host, tp, cache, authTokenReq)
				if err != nil {
					log.WithFields(log.Fields{
						"error": err.Error(),
					}).Error("(platformGetTransaction) Could not retrieve auth token from transaction service")

					return nil, NewHttpError("An error occurred while processing transaction", 500)
				}

				log.Info("(platformGetTransaction) New token retrieved from transaction service")

				//Update headers with new token
				headers["Authorization"] = fmt.Sprintf("Bearer %s", token)

				//Resend transaction processing request
				if trxRes, err = tp.GetTransaction(host, trxID, res, headers); err != nil {
					log.WithFields(log.Fields{
						"error":         err.Error(),
						"transactionID": trxID,
					}).Error("(platformGetTransaction) Get transaction retry failed")

					return nil, NewHttpError("An error occurred while processing transaction", 400)

				}
			} else { // Not a 401 bread.HttpError
				log.WithFields(log.Fields{
					"error":         err.Error(),
					"transactionID": trxID,
				}).Error("(platformGetTransaction) Initial request to get transaction failed with bread.HttpError")

				return nil, NewHttpError("An error occurred while processing transaction", 400)
			}
		} else { // Not a bread.HTTPError
			log.WithFields(log.Fields{
				"error":         err.Error(),
				"transactionID": trxID,
			}).Error("(platformGetTransaction) Initial request to get transaction failed")

			return nil, NewHttpError("An error occurred while processing transaction", 400)
		}
	}

	log.WithFields(log.Fields{
		"response": res,
	}).Info("(platformGetTransaction) Transaction retrieved")

	return trxRes, nil
}

func platformApiParams(isTestMode bool, account types.GatewayAccount) (string, string, string) {
	var apiKey string
	var apiSecret string
	var host string
	if isTestMode { // Test transaction
		apiKey = account.PlatformSandboxApiKey
		apiSecret = account.PlatformSandboxSharedSecret
		host = gatewayConfig.TransactionService.TransactionServiceHostDevelopment
	} else {
		apiKey = account.PlatformApiKey
		apiSecret = account.PlatformSharedSecret
		host = gatewayConfig.TransactionService.TransactionServiceHost
	}

	return apiKey, apiSecret, host
}
