package gateway

import (
	"testing"

	"github.com/getbread/shopify_plugin_backend/service/bread"
	bmocks "github.com/getbread/shopify_plugin_backend/service/bread/mocks"
	"github.com/getbread/shopify_plugin_backend/service/bread/samples"
	cmocks "github.com/getbread/shopify_plugin_backend/service/cache/mocks"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAuthTokenFromRemote(t *testing.T) {

	//t.Parallel()
	req := samples.NewTrxAuthTokenRequest()
	zeroAuthTokenRes := &bread.TrxAuthTokenResponse{}
	tpMock := new(bmocks.TrxProcessor)

	t.Run("transaction service auth token request failed", func(tt *testing.T) {
		tpMock.On(
			"GetTransactionAuthToken",
			"http://host",
			req,
			zeroAuthTokenRes,
		).Return(nil, samples.Err).Once()

		token, err := authTokenFromRemote("http://host", tpMock, req)

		assert.NotNil(tt, err)
		assert.Equal(tt, "", token)
		tpMock.AssertExpectations(tt)
	})

	t.Run("transaction service auth token request success", func(tt *testing.T) {
		loadedAuthTokenRes := samples.NewAuthTokenResponse()
		tpMock.On("GetTransactionAuthToken", "http://host", req, zeroAuthTokenRes).Return(loadedAuthTokenRes, nil).Once()

		token, err := authTokenFromRemote("http://host", tpMock, req)

		assert.Nil(tt, err)
		assert.Equal(tt, token, loadedAuthTokenRes.Token)
		tpMock.AssertExpectations(tt)
	})
}

func TestAuthTokenToCache(t *testing.T) {
	//t.Parallel()

	t.Run("calls method to write to cache", func(tt *testing.T) {
		key, value := "k-e-y", "v-a-l-u-e"
		cacheMock := new(cmocks.Cache)
		cacheMock.On("Write", key, value).Return(mock.Anything, nil).Once()

		authTokenToCache(cacheMock, key, value)

		cacheMock.AssertExpectations(tt)
	})
}

func TestAuthTokenFromCache(t *testing.T) {
	//t.Parallel()

	t.Run("returns a token string on success", func(tt *testing.T) {
		key, value := "k-e-y", "v-a-l-u-e"
		cacheMock := new(cmocks.Cache)
		cacheMock.On("Get", key).Return([]byte(value), nil).Once()

		token, err := authTokenFromCache(cacheMock, key)

		assert.Nil(tt, err)
		assert.Equal(tt, value, token)
		cacheMock.AssertExpectations(tt)
	})

	t.Run("returns empty token and error if []byte type assertion fails", func(tt *testing.T) {
		key := "k-e-y"
		cacheMock := new(cmocks.Cache)
		cacheMock.On("Get", "k-e-y").Return(nil, nil).Once()

		token, err := authTokenFromCache(cacheMock, key)

		assert.Equal(tt, "", token)
		assert.EqualError(tt, err, ErrByteArrayAssertion.Error())
		cacheMock.AssertExpectations(tt)
	})

	t.Run("fails when reading from cache results in an error", func(tt *testing.T) {
		key := "k-e-y"
		cacheMock := new(cmocks.Cache)
		cacheMock.On("Get", key).Return(nil, samples.Err).Once()

		token, err := authTokenFromCache(cacheMock, key)

		assert.EqualError(tt, err, samples.Err.Error())
		assert.Equal(tt, "", token)
		cacheMock.AssertExpectations(tt)
	})

}

func TestAuthTokenRemoteRequestAndCache(t *testing.T) {

	origAuthTokenFromRemote := authTokenFromRemote
	cacheMock := new(cmocks.Cache)
	tpMock := new(bmocks.TrxProcessor)
	tokenReq := samples.NewTrxAuthTokenRequest()

	t.Run("Fails if remote request for auth token returns error", func(tt *testing.T) {
		// Stub authTokenFromRemote
		authTokenFromRemote = func(host string, tp bread.TrxProcessor, req *bread.TrxAuthTokenRequest) (string, error) {
			return "", samples.Err
		}

		token, err := authTokenRemoteRequestAndCache("http://host", tpMock, cacheMock, tokenReq)

		assert.Equal(tt, "", token)
		assert.Error(tt, err)
	})

	t.Run("Success if remote request for auth token suceeds", func(tt *testing.T) {
		token := uuid.NewRandom().String()
		authTokenFromRemote = func(host string, tp bread.TrxProcessor, req *bread.TrxAuthTokenRequest) (string, error) {
			return token, nil
		}

		resToken, err := authTokenRemoteRequestAndCache("http://host", tpMock, cacheMock, tokenReq)

		assert.Nil(tt, err)
		assert.Equal(tt, token, resToken)
	})

	authTokenFromRemote = origAuthTokenFromRemote
}
