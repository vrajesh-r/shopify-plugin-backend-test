package app

import (
	"testing"

	"github.com/getbread/shopify_plugin_backend/service/shopify"

	"github.com/stretchr/testify/assert"
)

func TestGetDraftOrderLimitSizeAndNumberOfFetches(t *testing.T) {
	type testConfig struct {
		Name                string
		TotalCount          int
		PageToFetch         int
		ResultsPerPage      int
		ExpectedLimit       int
		ExpectedNumRequests int
	}

	testTable := []testConfig{
		testConfig{
			Name:                "12ItemsPage1",
			TotalCount:          12,
			PageToFetch:         1,
			ResultsPerPage:      10,
			ExpectedLimit:       12,
			ExpectedNumRequests: 1,
		},
		testConfig{
			Name:                "12ItemsPage2",
			TotalCount:          12,
			PageToFetch:         2,
			ResultsPerPage:      10,
			ExpectedLimit:       2,
			ExpectedNumRequests: 1,
		},
		testConfig{
			Name:                "256ItemsPage1",
			TotalCount:          256,
			PageToFetch:         1,
			ResultsPerPage:      10,
			ExpectedLimit:       6,
			ExpectedNumRequests: 2,
		},
		testConfig{
			Name:                "256ItemsPage2",
			TotalCount:          256,
			PageToFetch:         2,
			ResultsPerPage:      10,
			ExpectedLimit:       246,
			ExpectedNumRequests: 1,
		},
		testConfig{
			Name:                "256ItemsPage3",
			TotalCount:          256,
			PageToFetch:         3,
			ResultsPerPage:      10,
			ExpectedLimit:       236,
			ExpectedNumRequests: 1,
		},
		testConfig{
			Name:                "250ItemsPage1",
			TotalCount:          250,
			PageToFetch:         1,
			ResultsPerPage:      10,
			ExpectedLimit:       0,
			ExpectedNumRequests: 2,
		},
	}

	for _, test := range testTable {
		t.Run(test.Name, func(t *testing.T) {
			limit, numRequests := getDraftOrderLimitSizeAndNumberOfFetches(test.TotalCount, test.PageToFetch, test.ResultsPerPage)

			assert.Equal(t, test.ExpectedLimit, limit)
			assert.Equal(t, test.ExpectedNumRequests, numRequests)
		})
	}
}

func TestGetPageInfoFromNextShopifyLinkUrl(t *testing.T) {
	links := shopify.Links{
		Next: &shopify.Link{
			Url:       "https://mikes-markers.myshopify.com/admin/api/2019-07/products.json?page_info=hijgklmn&limit=6",
			Direction: "next",
		},
	}

	pageInfo, err := getPageInfoFromNextShopifyLinkUrl(links)

	if !assert.NoError(t, err, "No error should have occurred") {
		return
	}

	assert.Equal(t, "hijgklmn", pageInfo)
}
