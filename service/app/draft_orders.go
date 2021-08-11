package app

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"sync"

	"github.com/getbread/breadkit/desmond"
	"github.com/getbread/breadkit/zeus/searcher"
	zeus "github.com/getbread/breadkit/zeus/types"
	"github.com/getbread/shopify_plugin_backend/service/bread"
	"github.com/getbread/shopify_plugin_backend/service/search"
	"github.com/getbread/shopify_plugin_backend/service/shopify"
	"github.com/getbread/shopify_plugin_backend/service/types"
	"github.com/getbread/shopify_plugin_backend/service/update"
	"github.com/gin-gonic/gin"
	"github.com/pborman/uuid"
	"github.com/sirupsen/logrus"
)

const (
	itemsPerPage        = 10
	shopifyMaxItemLimit = 250
)

func (h *Handlers) validateAppSession(c *gin.Context) (types.Session, types.Shop, error) {
	// Find valid session
	var sessionID string
	// Optimistic error handling here since an empty session ID gets caught
	// further along in execution
	if cookie, err := c.Request.Cookie(ADMIN_COOKIE_NAME); err == nil {
		sessionID = cookie.Value
	}

	// Short circuit error logging for empty session id
	if sessionID == "" {
		return types.Session{}, types.Shop{}, fmt.Errorf("empty session id")
	}
	session, err := findValidSessionById(sessionID, h)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"sessionId": sessionID,
		}).Error("(GetDraftOrders) query for session produced error")
		return types.Session{}, types.Shop{}, err
	}

	// Find shop
	shop, err := findShopById(session.ShopId, h)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"session": fmt.Sprintf("%+v", session),
		}).Error("(GetDraftOrders) query for shop produced error")
		return types.Session{}, types.Shop{}, err
	}

	return session, shop, nil
}

// The Shopify Draft Orders are returned from the Shopify API in chronological order,
// but we want to display them to the user in reverse chronological order.  In addition,
// Shopify only returns the set of orders you ask for (up to a limit of 250), and a pair
// of links to fetch either the previous batch or the next batch.  As a result, this
// function needs to determine:
//   * How many Draft Orders to fetch from Shopify in a single batch
//   * How many batches to fetch
//
// The range of draft orders to display to the user is defined as:
//     [numDraftOrders - (page * itemsPerPage), numDraftOrders - ((page - 1) * itemsPerPage)]
//
// As a result, this will return:
//    (numDraftOrders - ((page - 1) * itemsPerPage)) % 250 as the limit
//    ((numDraftOrders - ((page - 1) * itemsPerPage)) / 250) + 1 as the number of batches to fetch
//
func getDraftOrderLimitSizeAndNumberOfFetches(numDraftOrders, page, itemsPerPage int) (int, int) {
	limit := (numDraftOrders - ((page - 1) * itemsPerPage)) % 250
	numRequests := ((numDraftOrders - ((page - 1) * itemsPerPage)) / 250) + 1

	return limit, numRequests
}

func getPageInfoFromNextShopifyLinkUrl(links shopify.Links) (string, error) {
	if links.Next == nil {
		return "", fmt.Errorf("No next URL is available in the Shopify links header.")
	}

	theUrl, err := url.Parse(links.Next.Url)
	if err != nil {
		return "", fmt.Errorf("Unable to parse the next URL %s due to error: %+v", links.Next.Url, err)
	}

	theQuery := theUrl.Query()
	pageInfos, hasPageInfos := theQuery["page_info"]
	if !hasPageInfos || len(pageInfos) != 1 {
		return "", fmt.Errorf("Expected exactly one page_info parameter, but found %d in URL %s", len(pageInfos), links.Next.Url)
	}

	return pageInfos[0], nil
}

func (h *Handlers) GetDraftOrders(c *gin.Context, dc desmond.Context) {
	// Query open draft orders from the Shopify OMS API
	// @param page number
	page := c.Query("page")
	pageInt := func() int {
		if page == "" {
			return 1
		}
		pi, err := strconv.Atoi(page)
		if err != nil {
			logrus.WithError(err).WithField("page", page).
				Errorf("(GetDraftOrders) page from query in url is not a valid number")
			return 1
		}
		return pi
	}()

	session, shop, err := h.validateAppSession(c)
	if err != nil {
		if err.Error() == "empty session id" {
			c.HTML(400, "app_error.html", gin.H{
				"messagePrimary":   "Session expired",
				"messageSecondary": "Please restart your session by selecting Apps and then Bread",
			})
			return
		}
		logrus.WithError(err).WithFields(logrus.Fields{
			"session": fmt.Sprintf("%+v", session),
		}).Error("(GetDraftOrders) validating session app failed")
		c.String(400, err.Error())
		return
	}

	if shop.ActiveVersion == BreadPlatform {
		logrus.WithFields(logrus.Fields{
			"breadVersion": BreadPlatform,
		}).Error("(GetDraftOrders) Draft order not supported on bread platform")

		c.HTML(400, "app_error.html", gin.H{
			"messagePrimary":   "Invalid request",
			"messageSecondary": "Draft order is not supported on bread platform",
		})
		return
	}

	shopifycli := shopify.NewClient(shop.Shop, shop.AccessToken)

	// Query draft orders count
	docQuery := url.Values{}
	docQuery.Add("status", "open")
	doCount := shopify.GetDraftOrdersCountResponse{}
	if err := shopifycli.GetDraftOrdersCount(docQuery, &doCount); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"session": fmt.Sprintf("%+v", session),
		}).Error("(GetDraftOrders) query for draft orders count produced error")
		c.String(400, err.Error())
		return
	}

	limit, numRequests := getDraftOrderLimitSizeAndNumberOfFetches(doCount.Count, pageInt, itemsPerPage)

	formattedDraftOrders := make([]shopify.DraftOrder, 0, (numRequests-1)*250+limit)
	var pageInfo string

	for requestCount := 1; requestCount <= numRequests; requestCount++ {
		requestLimit := limit
		if requestCount < numRequests {
			requestLimit = shopifyMaxItemLimit // Need to page through to get all of the requests.

		} else if limit == 0 {
			// This happens when requesting the first page when there are exactly 250 draft orders.
			// We already fetched the 250 orders in the previous loop iteration, so we are done.
			break
		}

		query := url.Values{}
		query.Add("status", "open")
		query.Add("limit", strconv.Itoa(requestLimit))

		if len(pageInfo) > 0 {
			query.Add("page_info", pageInfo)
		}

		var dos shopify.GetDraftOrdersResponse
		var links shopify.Links

		if err := shopifycli.GetDraftOrders(query, &dos, &links); err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"session": fmt.Sprintf("%+v", session),
			}).Error("(GetDraftOrders) query for draft orders produced error")
			c.String(400, err.Error())
			return
		}

		formattedDraftOrders = append(formattedDraftOrders, dos.DraftOrders...)

		if (requestCount < numRequests) && (links.Next != nil) {
			// Prime the next request with the page_info header from the current response.
			pageInfo, err = getPageInfoFromNextShopifyLinkUrl(links)
			if err != nil {
				logrus.WithError(err).WithFields(logrus.Fields{
					"session": fmt.Sprintf("%+v", session),
					"nextUrl": links.Next.Url,
				}).Error("(GetDraftOrders) query for draft orders produced error")
				c.String(500, err.Error())
				return
			}
		}
	}

	if len(formattedDraftOrders) > 10 {
		formattedDraftOrders = formattedDraftOrders[len(formattedDraftOrders)-10 : len(formattedDraftOrders)]
	}

	// Reverse formatted draft orders
	for i := 0; i < len(formattedDraftOrders)/2; i++ {
		j := len(formattedDraftOrders) - i - 1
		formattedDraftOrders[i], formattedDraftOrders[j] = formattedDraftOrders[j], formattedDraftOrders[i]
	}

	c.HTML(200, "draft_orders.html", gin.H{
		"apiKey":             appConfig.ShopifyConfig.ShopifyApiKey.Unmask(),
		"shopName":           shop.Shop,
		"draftOrders":        formattedDraftOrders,
		"previousPage":       strconv.Itoa(pageInt - 1),
		"previousPageExists": pageInt != 1,
		"nextPage":           strconv.Itoa(pageInt + 1),
		"nextPageExists":     (numRequests > 1) || (numRequests == 1 && limit > 10),
	})
}

func (h *Handlers) ViewDraftOrder(c *gin.Context, dc desmond.Context) {
	idString := c.Param("id")
	id, err := strconv.Atoi(idString)
	if err != nil {
		logrus.WithError(err).WithField("idString", idString).Error("(CreateDraftOrderCart) draft order cart id expected to be an integer")
		c.String(400, "Invalid input.")
		return
	}

	session, shop, err := h.validateAppSession(c)
	if err != nil {
		if err.Error() == "empty session id" {
			c.HTML(400, "app_error.html", gin.H{
				"messagePrimary":   "Session expired",
				"messageSecondary": "Please restart your session by selecting Apps and then Bread",
			})
			return
		}
		logrus.WithError(err).WithFields(logrus.Fields{
			"session": fmt.Sprintf("%+v", session),
		}).Error("(ViewDraftOrder) validating session app failed")
		c.String(400, err.Error())
		return
	}

	// Query the draft_order from Shopify
	shopifycli := shopify.NewClient(shop.Shop, shop.AccessToken)
	dores := shopify.GetDraftOrderResponse{}
	if err := shopifycli.GetDraftOrder(strconv.Itoa(id), &dores); err != nil {
		logrus.WithError(err).Error("(ViewDraftOrder) query for draft order from Shopify failed")
		c.String(400, err.Error())
		return
	}
	draftOrder := dores.DraftOrder

	// Query the draft_order_cart from postgres
	docsr := search.DraftOrderCartSearchRequest{}
	docsr.AddFilter(search.DraftOrderCartSearch_ShopID, shop.Id, searcher.Operator_EQ, searcher.Condition_AND)
	docsr.AddFilter(search.DraftOrderCartSearch_IsProduction, shop.Production, searcher.Operator_EQ, searcher.Condition_AND)
	docsr.AddFilter(search.DraftOrderCartSearch_DraftOrderID, id, searcher.Operator_EQ, searcher.Condition_AND)
	docsr.AddFilter(search.DraftOrderCartSearch_IsDeleted, false, searcher.Operator_EQ, searcher.Condition_AND)
	docsr.Limit = 1
	docs, err := h.DraftOrderCartSearcher.Search(docsr)
	if err != nil {
		logrus.WithError(err).Error("(CreateDraftOrderCart) query for draft order cart failed")
		c.String(400, err.Error())
		return
	}
	var draftOrderCart *types.DraftOrderCart
	if len(docs) == 1 {
		draftOrderCart = &docs[0]
	}

	// Render the template
	c.HTML(200, "draft_order.html", gin.H{
		"apiKey":            appConfig.ShopifyConfig.ShopifyApiKey.Unmask(),
		"shopName":          shop.Shop,
		"draftOrder":        draftOrder,
		"hasDraftOrderCart": len(docs) == 1,
		"draftOrderCart":    draftOrderCart,
	})
}

func (h *Handlers) UpdateDraftOrderCart(c *gin.Context, dc desmond.Context) {
	req := struct {
		UseDraftOrderAsOrder bool `json:"useDraftOrderAsOrder"`
	}{}
	if err := c.BindJSON(&req); err != nil {
		logrus.WithError(err).Error("(UpdateDraftOrderCart) deserializing request body failed")
		c.String(400, err.Error())
		return
	}

	draftOrderCartID := c.Param("id")
	if len(draftOrderCartID) == 0 {
		logrus.Error("(UpdateDraftOrderCart) request should have a draft order cart ID in the path")
		c.String(http.StatusBadRequest, "Draft order cart ID required for this operation.")
		return
	}

	session, shop, err := h.validateAppSession(c)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"session": fmt.Sprintf("%+v", session),
		}).Error("(UpdateDraftOrderCart) validating app session failed")
		c.String(400, err.Error())
		return
	}

	docsr := search.DraftOrderCartSearchRequest{}
	docsr.AddFilter(search.DraftOrderCartSearch_ShopID, shop.Id, searcher.Operator_EQ, searcher.Condition_AND)
	docsr.AddFilter(search.DraftOrderCartSearch_ID, zeus.Uuid(draftOrderCartID), searcher.Operator_EQ, searcher.Condition_AND)
	docsr.AddFilter(search.DraftOrderCartSearch_IsDeleted, false, searcher.Operator_EQ, searcher.Condition_AND)
	docsr.AddFilter(search.DraftOrderCartSearch_IsProduction, shop.Production, searcher.Operator_EQ, searcher.Condition_AND)
	docsr.Limit = 1
	docs, err := h.DraftOrderCartSearcher.Search(docsr)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"draftOrderCartId": draftOrderCartID,
			"shopId":           shop.Id,
		}).Error("(CreateDraftOrderCart) query for draft order cart failed")
		c.String(http.StatusBadRequest, err.Error())
	}
	if len(docs) != 1 {
		c.String(http.StatusBadRequest, "Draft order cart not found.")
		return
	}

	docur := update.DraftOrderCartUpdateRequest{}
	docur.Id = zeus.Uuid(draftOrderCartID)
	docur.AddUpdate(update.DraftOrderCartUpdate_UseDraftOrderAsOrder, req.UseDraftOrderAsOrder)
	if err := h.DraftOrderCartUpdater.Update(docur); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"draftOrderCartId": draftOrderCartID,
			"shopId":           shop.Id,
		}).Error("(CreateDraftOrderCart) update to draft order cart failed")
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.String(http.StatusNoContent, "")
}

func (h *Handlers) CreateDraftOrderCart(c *gin.Context, dc desmond.Context) {
	req := struct {
		DraftOrderID int `json:"draftOrderId"`
	}{}
	if err := c.BindJSON(&req); err != nil {
		logrus.WithError(err).Error("(CreateDraftOrderCart) deserializing request body failed")
		c.String(400, err.Error())
		return
	}

	session, shop, err := h.validateAppSession(c)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"session": fmt.Sprintf("%+v", session),
		}).Error("(CreateDraftOrderCart) validating app session failed")
		c.String(400, err.Error())
		return
	}

	// Query to see if there are draft_order_carts in Postgres for this Shopify
	// draft_order. Mark the draft_order_cart as deleted if so.
	docsr := search.DraftOrderCartSearchRequest{}
	docsr.AddFilter(search.DraftOrderCartSearch_ShopID, shop.Id, searcher.Operator_EQ, searcher.Condition_AND)
	docsr.AddFilter(search.DraftOrderCartSearch_IsProduction, shop.Production, searcher.Operator_EQ, searcher.Condition_AND)
	docsr.AddFilter(search.DraftOrderCartSearch_DraftOrderID, req.DraftOrderID, searcher.Operator_EQ, searcher.Condition_AND)
	docsr.AddFilter(search.DraftOrderCartSearch_IsDeleted, false, searcher.Operator_EQ, searcher.Condition_AND)
	docs, err := h.DraftOrderCartSearcher.Search(docsr)
	if err != nil {
		logrus.WithError(err).Error("(CreateDraftOrderCart) query for draft order cart failed")
		c.String(400, err.Error())
		return
	}
	if len(docs) >= 1 {
		for i, _ := range docs {
			docur := update.DraftOrderCartUpdateRequest{}
			docur.Id = docs[i].ID
			docur.AddUpdate(update.DraftOrderCartUpdate_IsDeleted, true)
			if err := h.DraftOrderCartUpdater.Update(docur); err != nil {
				logrus.WithError(err).Error("(CreateDraftOrderCart) deleting existing draft_order_cart failed")
				c.String(400, err.Error())
				return
			}
		}
	}

	// Query draft order from Shopify
	logrus.Infof("Querying Shopify for draft_order with ID: %d", req.DraftOrderID)
	shopifycli := shopify.NewClient(shop.Shop, shop.AccessToken)
	dores := shopify.GetDraftOrderResponse{}
	if err := shopifycli.GetDraftOrder(strconv.Itoa(req.DraftOrderID), &dores); err != nil {
		logrus.WithError(err).WithField("draftOrderId", req.DraftOrderID).Error("(CreateDraftOrderCart) querying the draft order from Shopify failed")
		c.String(400, err.Error())
		return
	}
	draftOrder := dores.DraftOrder
	if draftOrder.Customer.FirstName == "" || draftOrder.Customer.LastName == "" {
		c.JSON(400, gin.H{
			"error": "You must add a customer name to the draft order.",
		})
		return
	}

	// Convert applied_discount to Bread discount
	var discount bread.OptsDiscount
	if len(draftOrder.AppliedDiscount.Amount) > 0 {
		discountAmtMillicents, err := types.USDToMillicents(draftOrder.AppliedDiscount.Amount)
		if err != nil {
			logrus.WithError(err).WithField("AppliedDiscount.Amount", draftOrder.AppliedDiscount.Amount).Error("(CreateDraftOrderCart) parsing discount amount into millicents failed")
		} else {
			discount.Amount = discountAmtMillicents.ToCents()
			discount.Description = draftOrder.AppliedDiscount.Description
		}
	}

	// Calculate total_tax based on draftOrder.TaxLines
	totalTax := 0
	for _, ti := range draftOrder.TaxLines {
		tiPriceMillicents, err := types.USDToMillicents(ti.Price)
		if err != nil {
			logrus.WithError(err).WithField("draftOrderId", req.DraftOrderID).Error("(CreateDraftOrderCart) parsing the tax lines prices into millicents failed")
		}
		totalTax += int(tiPriceMillicents.ToCents())
	}
	totalTaxCents := types.Cents(totalTax)

	// Pre assign the draft_order_cart ID
	draftOrderCartID := uuid.New()

	// Create a new cart with the Bread system
	totalPriceMillicents, err := types.USDToMillicents(draftOrder.TotalPrice)
	if err != nil {
		logrus.WithError(err).WithField("draftOrderId", req.DraftOrderID).Error("(CreateDraftOrderCart) parsing the total price into millicents failed")
		c.String(400, err.Error())
		return
	}
	blis, err := getBreadLineItemsFromDraftOrder(draftOrder, shop)
	if err != nil {
		logrus.WithError(err).WithField("draftOrderID", draftOrder.ID).Error("(CreateDraftOrderCart) mapping Bread line items from draft order produced error")
		c.String(400, err.Error())
		return
	}

	// Prepare bread shippingOption for cartCreateRequest, correctly handle blank shipping lines
	shippingOptions := func() []bread.OptsShippingOption {
		var shippingOpt = bread.OptsShippingOption{}
		if draftOrder.ShippingLine.Title == "" {
			shippingOpt.Type = "No shipping information provided"
		} else {
			shippingOpt.Type = draftOrder.ShippingLine.Title
		}
		if draftOrder.ShippingLine.Code != "" {
			shippingOpt.TypeID = draftOrder.ShippingLine.Code
		} else {
			shippingOpt.TypeID = shippingOpt.Type
		}
		if draftOrder.ShippingLine.Price == "" {
			shippingOpt.Cost = types.Millicents(0).ToCents()
		} else {
			shippingPriceMillicents, err := types.USDToMillicents(draftOrder.ShippingLine.Price)
			if err != nil {
				logrus.WithError(err).WithField("draftOrderId", req.DraftOrderID).Error("(CreateDraftOrderCart) parsing the shipping line price into millicents failed")
				shippingOpt.Cost = types.Millicents(0).ToCents()
			} else {
				shippingOpt.Cost = shippingPriceMillicents.ToCents()
			}
		}
		return []bread.OptsShippingOption{shippingOpt}
	}()
	cartCreateRequest := &bread.Cart{
		Options: bread.CartOptions{
			OrderRef:        string(draftOrderCartID),
			CallbackUrl:     appConfig.MiltonPort + "/portal/draftorder/cart/callback",
			CompleteUrl:     appConfig.MiltonPort + "/portal/draftorder/cart/complete",
			ErrorUrl:        "https://" + shop.Shop + ".myshopify.com", // TODO: Build an actual error view on Milton
			ShippingOptions: shippingOptions,
			ShippingContact: bread.OptsContact{
				FullName: fmt.Sprintf("%s %s", draftOrder.Customer.FirstName, draftOrder.Customer.LastName),
				Address:  draftOrder.ShippingAddress.Address1,
				Address2: draftOrder.ShippingAddress.Address2,
				City:     draftOrder.ShippingAddress.City,
				State:    draftOrder.ShippingAddress.ProvinceCode,
				Zip:      draftOrder.ShippingAddress.Zip,
				Phone:    draftOrder.ShippingAddress.Phone,
			},
			BillingContact: bread.OptsContact{
				FullName: fmt.Sprintf("%s %s", draftOrder.Customer.FirstName, draftOrder.Customer.LastName),
				Address:  draftOrder.BillingAddress.Address1,
				Address2: draftOrder.BillingAddress.Address2,
				City:     draftOrder.BillingAddress.City,
				State:    draftOrder.BillingAddress.ProvinceCode,
				Zip:      draftOrder.BillingAddress.Zip,
				Phone:    draftOrder.BillingAddress.Phone,
				Email:    draftOrder.Email,
			},
			Items:       *blis,
			Tax:         totalTaxCents,
			CustomTotal: totalPriceMillicents.ToCents(),
		},
		CartOrigin: "shopify_carts",
	}

	if discount.Amount > 0 {
		cartCreateRequest.Options.Discounts = []bread.OptsDiscount{discount}
	}

	breadcli := bread.NewClient(shop.GetAPIKeys())
	bhost := appConfig.HostConfig.BreadHost
	if !shop.Production {
		bhost = appConfig.HostConfig.BreadHostDevelopment
	}
	savedCart, err := breadcli.SaveCart(bhost, cartCreateRequest)
	if err != nil {
		logrus.WithError(err).Error("(CreateDraftOrderCart) creating cart with Bread failed")
		c.String(400, err.Error())
		return
	}

	draftOrderCart := types.DraftOrderCart{
		ID:           zeus.Uuid(draftOrderCartID),
		ShopID:       shop.Id,
		DraftOrderID: draftOrder.ID,
		CartID:       zeus.Uuid(savedCart.Id),
		CartURL:      savedCart.Url,
		IsProduction: shop.Production,
		IsDeleted:    false,
	}
	if _, err := h.DraftOrderCartCreator.Create(draftOrderCart); err != nil {
		logrus.WithError(err).Error("(CreateDraftOrderCart) saving the draft order cart in to Postgres failed")
		c.String(400, err.Error())
		return
	}

	// Return everything to render the draft_order detail page
	c.JSON(200, gin.H{
		"draftOrderId": req.DraftOrderID,
	})
}
func queryShopifyProductById(productId string, shop types.Shop) (*shopify.Product, error) {
	sc := shopify.NewClient(shop.Shop, shop.AccessToken)
	var res shopify.SearchProductByIdResponse
	err := sc.QueryProduct(productId, &res)
	if err != nil {
		return nil, err
	}
	return &res.Product, nil

}
func getBreadLineItemsFromDraftOrder(draftOrder shopify.DraftOrder, shop types.Shop) (*[]bread.OptsItem, error) {
	blis := make([]bread.OptsItem, len(draftOrder.LineItems))
	errChan := make(chan error, len(blis)*2) // Error channel buffered for 2x number of line items
	var wg sync.WaitGroup
	for i, doItem := range draftOrder.LineItems {
		wg.Add(1)
		go func(doItem shopify.LineItem, count int, wg *sync.WaitGroup) {
			productIdString := strconv.Itoa(doItem.ProductID)
			product, err := queryShopifyProductById(productIdString, shop)
			errChan <- err

			itemPriceMillicents, err := types.USDToMillicents(doItem.Price)
			errChan <- err
			if product != nil {
				blis[count] = bread.OptsItem{
					Name:      doItem.Name,
					Price:     itemPriceMillicents.ToCents(),
					Sku:       productIdString + ";::;" + doItem.Sku,
					Quantity:  uint32(doItem.Quantity),
					DetailUrl: "https://" + shop.Shop + ".myshopify.com/products/" + product.Handle,
				}
				if len(product.Image) > 0 {
					blis[count].ImageUrl = product.Image[0].Src
				}
			}
			wg.Done()
		}(doItem, i, &wg)
	}
	wg.Wait()
	close(errChan)

	var errs []error
	for err := range errChan {
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error":                err.Error(),
				"draftOrder.LineItems": draftOrder.LineItems,
			}).Error("(CreateDraftOrderCart) error querying Shopify Line Item")
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return nil, errs[0]
	}
	// Return line items formatted for Bread Tx
	return &blis, nil
}

func (h *Handlers) SendDraftOrderCartEmail(c *gin.Context, dc desmond.Context) {
	// Use the Bread client to send the email
	var req struct {
		CartID string `json:"cartId"`
		Email  string `json:"email"`
		Name   string `json:"name"`
	}
	if err := c.BindJSON(&req); err != nil {
		logrus.WithError(err).Error("(SendDraftOrderCartEmail) deserializing request failed")
		c.String(400, "Invalid request, please refresh and try again.")
		return
	}

	session, shop, err := h.validateAppSession(c)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"session": fmt.Sprintf("%+v", session),
		}).Error("(SendDraftOrderCartEmail) validating app session failed")
		c.String(400, err.Error())
		return
	}

	// Find draft_order_cart
	docsr := search.DraftOrderCartSearchRequest{}
	docsr.AddFilter(search.DraftOrderCartSearch_CartID, req.CartID, searcher.Operator_EQ, searcher.Condition_AND)
	docsr.AddFilter(search.DraftOrderCartSearch_ShopID, shop.Id, searcher.Operator_EQ, searcher.Condition_AND)
	docsr.AddFilter(search.DraftOrderCartSearch_IsDeleted, false, searcher.Operator_EQ, searcher.Condition_AND)
	docsr.AddFilter(search.DraftOrderCartSearch_IsProduction, shop.Production, searcher.Operator_EQ, searcher.Condition_AND)
	docsr.Limit = 1
	docs, err := h.DraftOrderCartSearcher.Search(docsr)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"sessionId": session.Id,
			"request":   fmt.Sprintf("%#v", req),
		}).Error("(SendDraftOrderCartEmail) searching for the draft order cart failed")
		c.String(http.StatusBadRequest, "Could not find the specified cart.")
		return
	}
	if len(docs) != 1 {
		logrus.WithFields(logrus.Fields{
			"sessionId": session.Id,
			"request":   fmt.Sprintf("%#v", req),
		}).Error("(SendDraftOrderCartEmail) could not find the specified cart")
		c.String(http.StatusBadRequest, "Could not find the specified cart.")
		return
	}
	draftOrderCart := docs[0]

	bc := bread.NewClient(shop.GetAPIKeys())

	breadHost := appConfig.HostConfig.BreadHost
	if !draftOrderCart.IsProduction {
		breadHost = appConfig.HostConfig.BreadHostDevelopment
	}
	if err := bc.SendCartEmail(breadHost, req.CartID, bread.SendCartEmailRequest{
		Email: req.Email,
		Name:  req.Name,
	}); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"sessionId": session.Id,
			"request":   fmt.Sprintf("%#v", req),
		}).Error("(SendDraftOrderCartEmail) rpc to send the email failed")
		c.String(http.StatusBadRequest, "Email failed to send.")
		return
	}

	c.String(http.StatusNoContent, "")
}

func (h *Handlers) SendDraftOrderCartText(c *gin.Context, dc desmond.Context) {
	// Use the Bread client to send the text
	var req struct {
		CartID string `json:"cartId"`
		Phone  string `json:"phone"`
	}
	if err := c.BindJSON(&req); err != nil {
		logrus.WithError(err).Error("(SendDraftOrderCartText) deserializing request failed")
		c.String(400, "Invalid request, please refresh and try again.")
		return
	}

	session, shop, err := h.validateAppSession(c)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"session": fmt.Sprintf("%+v", session),
		}).Error("(SendDraftOrderCartText) validating app session failed")
		c.String(400, err.Error())
		return
	}

	// Find draft_order_cart
	docsr := search.DraftOrderCartSearchRequest{}
	docsr.AddFilter(search.DraftOrderCartSearch_CartID, req.CartID, searcher.Operator_EQ, searcher.Condition_AND)
	docsr.AddFilter(search.DraftOrderCartSearch_ShopID, shop.Id, searcher.Operator_EQ, searcher.Condition_AND)
	docsr.AddFilter(search.DraftOrderCartSearch_IsDeleted, false, searcher.Operator_EQ, searcher.Condition_AND)
	docsr.AddFilter(search.DraftOrderCartSearch_IsProduction, shop.Production, searcher.Operator_EQ, searcher.Condition_AND)
	docsr.Limit = 1
	docs, err := h.DraftOrderCartSearcher.Search(docsr)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"sessionId": session.Id,
			"request":   fmt.Sprintf("%#v", req),
		}).Error("(SendDraftOrderCartText) searching for the draft order cart failed")
		c.String(http.StatusBadRequest, "Could not find the specified cart.")
		return
	}
	if len(docs) != 1 {
		logrus.WithFields(logrus.Fields{
			"sessionId": session.Id,
			"request":   fmt.Sprintf("%#v", req),
		}).Error("(SendDraftOrderCartText) could not find the specified cart")
		c.String(http.StatusBadRequest, "Could not find the specified cart.")
		return
	}
	draftOrderCart := docs[0]

	bc := bread.NewClient(shop.GetAPIKeys())

	breadHost := appConfig.HostConfig.BreadHost
	if !draftOrderCart.IsProduction {
		breadHost = appConfig.HostConfig.BreadHostDevelopment
	}
	if err := bc.SendCartText(breadHost, req.CartID, bread.SendCartTextRequest{
		Phone: req.Phone,
	}); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"sessionId": session.Id,
			"request":   fmt.Sprintf("%#v", req),
		}).Error("(SendDraftOrderCartText) rpc to send the text failed")
		c.String(http.StatusBadRequest, "Text message failed to send.")
		return
	}

	c.String(http.StatusNoContent, "")
}

func (h *Handlers) checkExistingDraftOrderCartCheckout(txID, draftOrderCartID zeus.Uuid) (bool, error) {
	doccsr := search.DraftOrderCartCheckoutSearchRequest{}
	doccsr.AddFilter(search.DraftOrderCartCheckoutSearch_TxID, txID, searcher.Operator_EQ, searcher.Condition_AND)
	doccsr.AddFilter(search.DraftOrderCartCheckoutSearch_DraftOrderCartID, draftOrderCartID, searcher.Operator_EQ, searcher.Condition_AND)
	doccs, err := h.DraftOrderCartCheckoutSearcher.Search(doccsr)
	if err != nil {
		return false, err
	}
	return (len(doccs) == 1), nil
}

func (h *Handlers) DraftOrderCartCallback(c *gin.Context, dc desmond.Context) {
	var req struct {
		DraftOrderCartID   zeus.Uuid `json:"orderRef"`
		BreadTransactionID zeus.Uuid `json:"transactionId"`
	}
	if err := c.Bind(&req); err != nil {
		logrus.WithError(err).WithField("request", req).Error("(DraftOrderCartCallback) deserializing request failed")
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	// Ensure this draft order cart X bread transaction combo has not already been recorded as a checkout
	exists, err := h.checkExistingDraftOrderCartCheckout(req.BreadTransactionID, req.DraftOrderCartID)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"draftOrderCartId": req.DraftOrderCartID,
			"transactionId":    req.BreadTransactionID,
		}).Error("(DraftOrderCartCallback) precondition check for draft order cart checkouts failed")
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	if exists {
		logrus.WithFields(logrus.Fields{
			"draftOrderCartId": req.DraftOrderCartID,
			"transactionId":    req.BreadTransactionID,
		}).Infof("(DraftOrderCartCallback) skipped processing callback because draft_order_cart_checkout already exists")
		c.String(http.StatusNoContent, "")
		return
	}

	logrus.WithFields(logrus.Fields{
		"draftOrderCartId": req.DraftOrderCartID,
		"transactionId":    req.BreadTransactionID,
	}).Infof("(DraftOrderCartCallback) processing offsite checkout callback")

	doc, err := h.DraftOrderCartSearcher.ById(req.DraftOrderCartID)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"draftOrderCartId": req.DraftOrderCartID,
			"transactionId":    req.BreadTransactionID,
		}).Warn("(DraftOrderCartCallbak) failed to find draft order cart.")
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	shop, err := h.ShopSearcher.ById(doc.ShopID)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"draftOrderCartId": req.DraftOrderCartID,
			"transactionId":    req.BreadTransactionID,
			"shopId":           doc.ShopID,
		}).Warn("(DraftOrderCartCallback) failed to find the shopify shop")
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	bc := bread.NewClient(shop.GetAPIKeys())
	bt, err := bc.QueryTransaction(string(req.BreadTransactionID), shop.BreadHost())
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error":            err.Error(),
			"draftOrderCartId": req.DraftOrderCartID,
			"transactionId":    req.BreadTransactionID,
			"shopId":           doc.ShopID,
		}).Error("(DraftOrderCartCallback) failed to query the Bread transaction")
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	if _, err := bc.AuthorizeTransaction(bt.BreadTransactionId, shop.BreadHost(), &bread.TransactionActionRequest{
		Type: "authorize",
	}); err != nil {
		logrus.WithFields(logrus.Fields{
			"error":            err.Error(),
			"draftOrderCartId": req.DraftOrderCartID,
			"transactionId":    req.BreadTransactionID,
			"shopId":           doc.ShopID,
		}).Error("(DraftOrderCartCallback) failed to authorize the Bread transaction")
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	// create customer on shopify backend
	customer, err := getShopifyCustomer(&bt.BillingContact, &bt.ShippingContact, shop)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error":         err.Error(),
			"request":       req,
			"shop":          shop,
			"transactionID": bt.BreadTransactionId,
		}).Error("(DraftOrderCartCallback) creating Shopify customer produced error")
		c.String(400, err.Error())
		return
	}

	// re-create the cart transaction via Order API
	so, err := createShopifyOrder(bt, customer, shop)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error":         err.Error(),
			"request":       req,
			"shop":          shop,
			"transactionID": bt.BreadTransactionId,
		}).Error("(DraftOrderCartCallback) creating Shopify order produced error")
		c.String(400, err.Error())
		return
	}

	// Update transaction with Shopify order number
	updateRequest := &bread.TransactionActionRequest{
		MerchantOrderId: strconv.Itoa(so.OrderNumber),
	}
	_, err = bc.UpdateTransaction(bt.BreadTransactionId, shop.BreadHost(), updateRequest)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error":              err.Error(),
			"request":            req,
			"shop":               shop,
			"transactionID":      bt.BreadTransactionId,
			"shopifyOrderNumber": so.OrderNumber,
			"shopifyOrderID":     so.ID,
		}).Error("(DraftOrderCartCallback) updating transaction with Shopify order ID produced error")
	}

	// add order to Milton order -> transaction lookup
	_, err = createOrder(shop, bt.BreadTransactionId, so.ID, h)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error":           err.Error(),
			"request":         req,
			"shop":            shop,
			"transactionID":   bt.BreadTransactionId,
			"shopifyCustomer": customer,
			"shopifyOrderID":  so.ID,
		}).Error("(DraftOrderCartCallback) creating Milton order produced error")
		c.String(400, err.Error())
		return
	}

	// Query the draft_order from Shopify
	shopifycli := shopify.NewClient(shop.Shop, shop.AccessToken)
	dores := shopify.DeleteDraftOrderResponse{}
	if err := shopifycli.DeleteDraftOrder(strconv.Itoa(doc.DraftOrderID), &dores); err != nil {
		logrus.WithError(err).Info("(DraftOrderCartCallback) deleting draft order from Shopify failed")
		return
	}

	// Expire the Bread cart
	if err := bc.ExpireCart(shop.BreadHost(), string(doc.CartID)); err != nil {
		logrus.WithError(err).Info("(DraftOrderCartCallback) expiring Bread cart produced error")
		return
	}

	if shop.AutoSettle {
		// Settle the transaction
		captureRequest := &shopify.CreateTransactionRequest{
			Transaction: shopify.Transaction{
				Kind:     "capture",
				Status:   "success",
				Amount:   types.Cents(bt.AdjustedTotal).ToString(),
				Currency: "USD",
				Gateway:  "Bread Shopify Payments",
				Test:     !shop.Production,
			},
		}

		settled := true
		var captureRes shopify.CreateTransactionResponse
		if err := shopifycli.CreateTransaction(so.ID, captureRequest, &captureRes); err != nil {
			// log and continue with response
			logrus.WithFields(logrus.Fields{
				"error":          err.Error(),
				"request":        req,
				"shop":           shop,
				"transactionID":  bt.BreadTransactionId,
				"shopifyOrderID": so.ID,
			}).Error("(CopyOrder) creating Shopify capture transaction produced error")

			settled = false
		}

		if settled && !shop.EnableOrderWebhooks {
			if _, err := bc.SettleTransaction(bt.BreadTransactionId, shop.BreadHost(), &bread.TransactionActionRequest{
				Type: "settle",
			}); err != nil {
				logrus.WithFields(logrus.Fields{
					"draftOrderCartId": req.DraftOrderCartID,
					"transactionId":    req.BreadTransactionID,
					"shopId":           doc.ShopID,
				}).Error("(DraftOrderCartCallback) failed to settle the Bread transaction")
				c.String(http.StatusInternalServerError, err.Error())
				return
			}
		}

	}

	// Save a draft order cart checkout record
	docc := types.DraftOrderCartCheckout{
		ShopID:           shop.Id,
		TxID:             req.BreadTransactionID,
		DraftOrderCartID: req.DraftOrderCartID,
		OrderID:          so.ID,
		IsProduction:     doc.IsProduction,
		Completed:        true,
		Errored:          false,
	}
	if _, err := h.DraftOrderCartCheckoutCreator.Create(docc); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"draftOrderCartId": req.DraftOrderCartID,
			"transactionId":    req.BreadTransactionID,
			"shopId":           doc.ShopID,
		}).Error("(DraftOrderCartCallback) failed to create the draft order cart checkout")
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.String(http.StatusNoContent, "")
}

func (h *Handlers) completeShopifyDraftOrder(shop types.Shop, draftOrderID int) (shopify.DraftOrder, error) {
	shopifycli := shopify.NewClient(shop.Shop, shop.AccessToken)
	query := url.Values{}
	query.Add("payment_pending", "false")
	dos := shopify.GetDraftOrderResponse{}
	if err := shopifycli.CompleteDraftOrder(strconv.Itoa(draftOrderID), &dos); err != nil {
		return shopify.DraftOrder{}, err
	}
	return dos.DraftOrder, nil
}

func (h *Handlers) DraftOrderCartComplete(c *gin.Context, dc desmond.Context) {
	// Copy the bread transaction into the merchants OMS
	draftOrderCartID := zeus.Uuid(c.Query("orderRef"))
	breadTransactionID := zeus.Uuid(c.Query("transactionId"))

	logrus.WithFields(logrus.Fields{
		"draftOrderCartId": draftOrderCartID,
		"transactionId":    breadTransactionID,
	}).Infof("(DraftOrderCartComplete) received offsite checkout complete")

	// Query draft order cart
	// Ensure this draft order cart X bread transaction combo has not already been recorded
	_, err := h.checkExistingDraftOrderCartCheckout(breadTransactionID, draftOrderCartID)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"draftOrderCartId": draftOrderCartID,
			"transactionId":    breadTransactionID,
		}).Error("(DraftOrderCartComplete) precondition check for draft order cart checkouts failed")
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	doc, err := h.DraftOrderCartSearcher.ById(draftOrderCartID)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"draftOrderCartId": draftOrderCartID,
			"transactionId":    breadTransactionID,
		}).Warn("(DraftOrderCartComplete) failed to find draft order cart.")
		c.String(http.StatusBadRequest, err.Error())
		return
	}

	shop, err := h.ShopSearcher.ById(doc.ShopID)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"draftOrderCartId": draftOrderCartID,
			"transactionId":    breadTransactionID,
			"shopId":           doc.ShopID,
		}).Warn("(DraftOrderCartComplete) failed to find the shopify shop")
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	// TODO: Figure out of this failsafe is necessary
	// Attempt to complete checkout, then save draft order cart checkout record
	// if !exists {
	// }

	doccsr := search.DraftOrderCartCheckoutSearchRequest{}
	doccsr.AddFilter(search.DraftOrderCartCheckoutSearch_DraftOrderCartID, draftOrderCartID, searcher.Operator_EQ, searcher.Condition_AND)
	doccsr.AddFilter(search.DraftOrderCartCheckoutSearch_TxID, breadTransactionID, searcher.Operator_EQ, searcher.Condition_AND)
	doccsr.AddFilter(search.DraftOrderCartCheckoutSearch_ShopID, shop.Id, searcher.Operator_EQ, searcher.Condition_AND)
	doccsr.AddFilter(search.DraftOrderCartCheckoutSearch_IsProduction, doc.IsProduction, searcher.Operator_EQ, searcher.Condition_AND)
	doccsr.Limit = 1
	doccs, err := h.DraftOrderCartCheckoutSearcher.Search(doccsr)
	if err != nil || len(doccs) == 0 {
		if err == nil {
			err = fmt.Errorf("Query to to retrieve draft_order_cart_checkout resource returned 0 results")
		}
		logrus.WithError(err).WithFields(logrus.Fields{
			"transactionId":    breadTransactionID,
			"draftOrderCartId": draftOrderCartID,
			"shopId":           doc.ShopID,
		}).Errorf("(DraftOrderCartComplete) failed to retrieve the draft_order_cart_checkout resource")
		c.String(400, err.Error())
		return
	}
	docc := doccs[0]

	// Get the order
	shopifycli := shopify.NewClient(shop.Shop, shop.AccessToken)
	var res shopify.SearchOrderResponse
	if err := shopifycli.QueryOrder(strconv.Itoa(docc.OrderID), &res); err != nil {
		logrus.WithFields(logrus.Fields{
			"error":                    err.Error(),
			"transactionId":            breadTransactionID,
			"draftOrderCartId":         draftOrderCartID,
			"draftOrderCartCheckoutId": docc.ID,
			"shopId":                   shop.Id,
		}).Error("(DraftOrderCartComplete) query for Shopify order produced error")
		c.String(400, err.Error())
		return
	}

	c.Redirect(302, fmt.Sprintf("https://%s.myshopify.com/apps/bread/orders/confirmation/%d", shop.Shop, res.Order.ID))
}

func (h *Handlers) DraftOrderCartError(c *gin.Context, dc desmond.Context) {
	draftOrderCartID := c.Query("orderRef")

	logrus.WithFields(logrus.Fields{
		"draftOrderCartId": draftOrderCartID,
	}).Infof("(DraftOrderCartError) received offsite checkout cancel")

}
