package app

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/getbread/shopify_plugin_backend/service/bread"
	"github.com/getbread/shopify_plugin_backend/service/shopify"
	"github.com/getbread/shopify_plugin_backend/service/types"
	log "github.com/sirupsen/logrus"
)

func createShopifyOrder(transaction *bread.TransactionResponse, customer *shopify.Customer, shop types.Shop) (*shopify.Order, error) {
	// map Ostia transaction into Shopify order
	orderReq, err := constructShopifyOrderRequest(transaction, customer, shop)
	if err != nil {
		return nil, err
	}

	// create new order on Shopify associated with customer if possible
	so, err := saveShopifyOrder(orderReq, shop)
	if err != nil {
		log.WithFields(log.Fields{
			"order request": orderReq,
			"Shopify order": so,
			"error":         err,
		}).Error("(createShopifyOrder) produced an error")
		return nil, err
	}
	return so, nil
}

func saveShopifyOrder(orderReq *shopify.CreateOrderRequest, shop types.Shop) (so *shopify.Order, err error) {
	sc := shopify.NewClient(shop.Shop, shop.AccessToken)
	var res shopify.CreateOrderResponse
	if err = sc.CreateOrder(orderReq, &res); err != nil {
		return
	}
	so = &res.Order
	return
}

func constructShopifyOrderRequest(transaction *bread.TransactionResponse, customer *shopify.Customer, shop types.Shop) (*shopify.CreateOrderRequest, error) {
	// pull line items from Shopify
	slis, err := getShopifyLineItemsFromTransaction(transaction, shop)
	if err != nil {
		return nil, err
	}
	// create order request
	cor := &shopify.CreateOrderRequest{
		Order: shopify.RequestOrder{
			BillingAddress:         *convertContactToShopifyAddress(&transaction.BillingContact),
			ShippingAddress:        *convertContactToShopifyAddress(&transaction.ShippingContact),
			Email:                  transaction.BillingContact.Email,
			TotalTax:               types.Cents(transaction.TotalTax).ToString(),
			Currency:               "USD",
			FinancialStatus:        "authorized",
			SendWebhooks:           true,
			SendReceipt:            true,
			SendFulfillmentReceipt: false,
			LineItems:              *slis,
			TaxesIncluded:          "false",
			TotalPrice:             strconv.FormatFloat(float64(transaction.AdjustedTotal)/100, 'f', 2, 64),
			TaxLines:               []shopify.TaxLine{},
			InventoryBehaviour:     "decrement_obeying_policy",
		},
	}
	//Add Discounts if there are any
	if len(transaction.Discounts) > 0 {
		var discounts []shopify.DiscountCode
		var total types.Cents
		for _, disc := range transaction.Discounts {
			//transforming bread discount to shopify discount
			el := shopify.DiscountCode{
				Amount: disc.Amount.ToString(),
				Code:   disc.Description,
				Type:   "fixed_amount",
			}
			if el.Code == "" {
				el.Code = "No code or description"
			}
			discounts = append(discounts, el)
			total += disc.Amount
		}
		cor.Order.DiscountCodes = discounts
		cor.Order.TotalDiscounts = total.ToString()
	}

	// add shipping lines, transaction, customer to order request
	cor.Order.ShippingLines = []shopify.ShippingLine{
		shopify.ShippingLine{
			Price: types.Cents(transaction.ShippingCost).ToString(),
			Code:  transaction.ShippingMethodCode,
			Title: func() string {
				if transaction.ShippingMethodName != "" {
					return transaction.ShippingMethodName
				}
				return transaction.ShippingMethodCode
			}(),
		},
	}
	cor.Order.Transactions = []shopify.Transaction{
		shopify.Transaction{
			Kind:     "authorization",
			Status:   "success",
			Amount:   strconv.FormatFloat(float64(transaction.AdjustedTotal)/100, 'f', 2, 64),
			Currency: "USD",
			Gateway:  "Bread Shopify Payments",
			Test:     !shop.Production,
		},
	}
	if customer.ID > 0 {
		cor.Order.Customer = shopify.Customer{
			ID: customer.ID,
		}
	} else {
		firstName, lastName := splitContactFullName(&transaction.ShippingContact)
		cor.Order.Customer = shopify.Customer{
			Email:     transaction.ShippingContact.Email,
			FirstName: firstName,
			LastName:  lastName,
		}
	}

	// respond
	return cor, nil
}

func getShopifyLineItemsFromTransaction(transaction *bread.TransactionResponse, shop types.Shop) (*[]shopify.LineItem, error) {
	// search for product variants (line items) concurrently
	slis := make([]shopify.LineItem, len(transaction.LineItems))
	errChan := make(chan error, len(slis))
	for i, tli := range transaction.LineItems {
		go func(tli bread.LineItem, i int) {
			pv, err := queryShopifyProductVariant(tli.Product.Sku, shop)
			errChan <- err
			if pv != nil {
				slis[i] = shopify.LineItem{
					Taxable:   true,
					Quantity:  int(tli.Quantity),
					VariantID: pv.ID,
				}
			}

		}(tli, i)
	}

	// collect errors
	var errs []error
	for i := 0; i < cap(errChan); i++ {
		err := <-errChan
		if err != nil {
			// logging these here because mulitple errors
			// will get lost with single error return value
			log.WithFields(log.Fields{
				"error":                 err.Error(),
				"transaction.LineItems": transaction.LineItems,
			}).Error("error querying Shopify Line Item ")
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return nil, errs[0]
	}

	// return shopify line items
	return &slis, nil
}

func queryShopifyProductVariant(specialSku string, shop types.Shop) (spv *shopify.ProductVariant, err error) {
	// prepare for request
	sc := shopify.NewClient(shop.Shop, shop.AccessToken)
	productId, sku := splitSpecialSku(specialSku)

	// query product
	var res shopify.SearchProductByIdResponse
	err = sc.QueryProduct(productId, &res)
	if err != nil {
		return nil, err
	}

	//If there is only one variant we use that instead
	if len(res.Product.Variants) == 1 && sku == "" {
		spv = &(res.Product.Variants[0])
		return spv, nil
	}
	for _, variant := range res.Product.Variants {
		if variant.Sku == sku {
			spv = &variant
			return
		}
	}
	return nil, fmt.Errorf("[shop => %s] product (productId => %s) does not contain variant with sku (sku => %s)", shop.Shop, productId, sku)
}

func subTotalStringFromTransaction(transaction *bread.TransactionResponse) string {
	var totalCents types.Cents
	for _, li := range transaction.LineItems {
		totalCents += types.Cents(li.Price)
	}
	return totalCents.ToString()
}

func calcDiscountsFromTransaction(transaction *bread.TransactionResponse) string {
	discounts := types.Cents(transaction.Total) - types.Cents(transaction.AdjustedTotal)
	return discounts.ToString()
}

// special sku is a concatenation: <productId> + ";::;" + <sku>
func splitSpecialSku(specialSku string) (productId, sku string) {
	pieces := strings.Split(specialSku, ";::;")
	if len(pieces) > 1 {
		return pieces[0], pieces[1]
	}
	return pieces[0], ""
}
