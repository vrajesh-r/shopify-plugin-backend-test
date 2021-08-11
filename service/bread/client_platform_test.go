package bread

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAssertTrxResponse(t *testing.T) {
	t.Run("fails if value to assert is not of type *bread.TrxResponse", func(tt *testing.T) {
		res, err := assertTrxResponse(make(map[string]string))

		assert.Nil(tt, res)
		assert.Error(tt, err)
	})

	t.Run("success if value to assert is of type *bread.TrxResponse", func(tt *testing.T) {
		res, err := assertTrxResponse(new(TrxResponse))

		assert.NotNil(tt, res)
		assert.Nil(tt, err)
	})
}

func TestAssertTrxAuthTokenResponse(t *testing.T) {
	t.Run("fails if *bread.TrxAuthTokenResponse type assertion fails", func(tt *testing.T) {
		res, err := assertTrxAuthTokenResponse(make(map[string]string))

		assert.Nil(tt, res)
		assert.Error(tt, err)
	})

	t.Run("success if *bread.TrxAuthTokenResponse type assertion succeeds", func(tt *testing.T) {
		res, err := assertTrxAuthTokenResponse(new(TrxAuthTokenResponse))

		assert.NotNil(tt, res)
		assert.Nil(tt, err)
	})
}
