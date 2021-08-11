package gateway

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/getbread/breadkit/desmond"
	"github.com/getbread/breadkit/zeus/searcher"
	"github.com/getbread/shopify_plugin_backend/service/search"
	"github.com/getbread/shopify_plugin_backend/service/types"
)

type ImageResponse struct {
	Images []Image `json:"images"`
}

type Image struct {
	ProductId int    `json:"product_id"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Alt       string `json:"alt"`
	Src       string `json:"src"`
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

func (h *Handlers) GetProductImage(c *gin.Context, dc desmond.Context) {
	var shopName string = c.Params.ByName("ShopName")
	var productId string = c.Params.ByName("ProductId")

	shop, err := findShopByName(shopName, h)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error":     err.Error(),
			"shopName":  shopName,
			"productId": productId,
		}).Error("(GetProductImage) query for shop produced error")

		productIdInt, _ := strconv.Atoi(productId)
		noImage := Image{
			ProductId: productIdInt,
			Width:     160,
			Height:    160,
			Alt:       "No image",
			Src:       fmt.Sprintf("%s/assets/no-image.gif", gatewayConfig.HostConfig.MiltonHost),
		}

		imageResponse := &ImageResponse{
			Images: []Image{noImage},
		}

		c.JSON(200, imageResponse)
		return
	}

	client := &http.Client{}

	req, err := http.NewRequest("GET", fmt.Sprintf("https://%s.myshopify.com/admin/products/%s/images.json", shopName, productId), nil)

	if err != nil {
		logrus.Error(fmt.Sprintf("(GetProductImage) failed to create request: %s", err))
		c.JSON(500, gin.H{})
		return
	}

	logrus.Info(fmt.Sprintf("(GetProductImage) attaching token of ... %s", shop.AccessToken))

	req.Header.Add("X-Shopify-Access-Token", shop.AccessToken)
	resp, err := client.Do(req)

	if err != nil {
		logrus.Error(fmt.Sprintf("(GetProductImage) failed to execute product request: %s", err))
		c.JSON(500, gin.H{})
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logrus.Error(fmt.Sprintf("(GetProductImage) failed to read in response: %s", err))
		c.JSON(500, gin.H{})
		return
	}

	logrus.Info(fmt.Sprintf("(GetProductImage) Got response: %v", string(body)))

	images := &ImageResponse{}
	if err := json.Unmarshal(body, &images); err != nil {
		logrus.Error(fmt.Sprintf("(GetProductImage) failed to unmarshal response: %s", err))
		c.JSON(500, gin.H{})
		return
	}

	if images == nil {
		logrus.Error(fmt.Sprintf("(GetProductImage) response is invalid: %s", string(body)))
		c.JSON(500, gin.H{})
		return
	}

	c.JSON(200, images)
}
