package shopify

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseShopifyLinksResponsHeader(t *testing.T) {
	t.Run("NextOnly", func(t *testing.T) {
		header := "\"<https://{shop}.myshopify.com/admin/api/2019-07/products.json?page_info=hijgklmn&limit=3>; rel=next\""
		links, err := parseLinkHeader(header)

		if !assert.NoError(t, err, "No error should have occurred.") {
			return
		}

		assert.Nil(t, links.Previous, "This header did not have a PREVIOUS portion.")

		if assert.NotNil(t, links.Next, "This header had a NEXT portion.") {
			assert.Equal(t, "https://{shop}.myshopify.com/admin/api/2019-07/products.json?page_info=hijgklmn&limit=3", links.Next.Url)
			assert.Equal(t, "next", links.Next.Direction)
		}
	})

	t.Run("PreviousOnly", func(t *testing.T) {
		header := "\"<https://{shop}.myshopify.com/admin/api/2019-07/products.json?page_info=hijgklmn&limit=3>; rel=previous\""
		links, err := parseLinkHeader(header)

		if !assert.NoError(t, err, "No error should have occurred.") {
			return
		}

		if assert.NotNil(t, links.Previous, "This header had a PREVIOUS portion.") {
			assert.Equal(t, "https://{shop}.myshopify.com/admin/api/2019-07/products.json?page_info=hijgklmn&limit=3", links.Previous.Url)
			assert.Equal(t, "previous", links.Previous.Direction)
		}

		assert.Nil(t, links.Next, "This header did not have a NEXT portion.")
	})

	t.Run("PreviousAndNext", func(t *testing.T) {
		header := "\"<https://{shop}.myshopify.com/admin/api/{version}/products.json?page_info=next&limit={limit}>; rel=next, <https://{shop}.myshopify.com/admin/api/{version}/products.json?page_info=previous&limit={limit}>; rel=previous\""
		links, err := parseLinkHeader(header)

		if !assert.NoError(t, err, "No error should have occurred.") {
			return
		}

		if assert.NotNil(t, links.Previous, "This header had a PREVIOUS portion.") {
			assert.Equal(t, "https://{shop}.myshopify.com/admin/api/{version}/products.json?page_info=previous&limit={limit}", links.Previous.Url)
			assert.Equal(t, "previous", links.Previous.Direction)
		}

		if assert.NotNil(t, links.Next, "This header had a NEXT portion.") {
			assert.Equal(t, "https://{shop}.myshopify.com/admin/api/{version}/products.json?page_info=next&limit={limit}", links.Next.Url)
			assert.Equal(t, "next", links.Next.Direction)
		}
	})

	t.Run("None", func(t *testing.T) {
		links, err := parseLinkHeader("")

		if !assert.NoError(t, err, "No error should have occurred") {
			return
		}

		assert.Nil(t, links.Previous, "This header did not have a PREVIOUS portion.")
		assert.Nil(t, links.Next, "This header did not have a NEXT portion.")
	})
}
