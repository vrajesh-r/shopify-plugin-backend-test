package shopify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/google/go-querystring/query"
)

type Session struct {
	Shop              string
	Jar               *cookiejar.Jar
	Url               *url.URL
	CheckoutUrl       *url.URL
	AuthenticityToken string
}

func NewSession(shopName string) (*Session, error) {
	// create
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	s := &Session{
		Shop: shopName,
		Jar:  jar,
	}

	// set url
	u, err := url.Parse(s.ShopUrl())
	if err != nil {
		return nil, err
	}
	s.Url = u

	// initialze Shopify cookies
	if err := s.initializeShopCookies(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Session) Clear() error {
	j, err := cookiejar.New(nil)
	if err != nil {
		return err
	}
	s.Jar = j
	s.AuthenticityToken = ""
	s.CheckoutUrl = nil
	return nil
}

func (s *Session) ShopUrl() string {
	return fmt.Sprintf("https://%s.myshopify.com", s.Shop)
}

func (s *Session) initializeShopCookies() error {
	// create request
	req, err := http.NewRequest("GET", s.ShopUrl(), bytes.NewBuffer([]byte{}))
	if err != nil {
		return err
	}

	// make request
	client := &http.Client{
		Jar: s.Jar,
	}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode >= 400 {
		contents, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf(string(contents))
	}
	return nil
}

func (s *Session) JsonRequest(method, url string, payload []byte, body interface{}) error {
	// create request
	req, err := http.NewRequest(method, url, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	// make request
	client := &http.Client{
		Jar: s.Jar,
	}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode >= 400 {
		contents, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf(string(contents))
	}
	return json.NewDecoder(res.Body).Decode(body)
}

func (s *Session) Request(method, url string, payload []byte) (*http.Response, error) {
	// create request
	req, err := http.NewRequest(method, url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}

	// make request
	client := &http.Client{
		Jar: s.Jar,
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode >= 400 {
		contents, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf(string(contents))
	}
	return res, nil
}

func (s *Session) GetCart(res *CartResponse) error {
	url := CartUrl(s.Shop)
	return s.JsonRequest("GET", url, []byte{}, res)
}

func (s *Session) AddToCart(req *AddToCartRequest, res *AddToCartResponse) error {
	url := AddToCartUrl(s.Shop)
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}
	return s.JsonRequest("POST", url, body, res)
}

func (s *Session) ClearCart(res *CartResponse) error {
	url := ClearCartUrl(s.Shop)
	return s.JsonRequest("POST", url, []byte{}, res)
}

func (s *Session) GetCartShippingRates(zip, country, province string, res *ShippingRatesResponse) error {
	url := ShippingRatesUrl(s.Shop, zip, country, province)
	return s.JsonRequest("GET", url, []byte{}, res)
}

func (s *Session) CreateCheckout() error {
	// make request
	url := CartCheckoutUrl(s.Shop)
	res, err := s.Request("POST", url, []byte{})
	if err != nil {
		return err
	}

	// set session CheckoutUrl
	s.CheckoutUrl = res.Request.URL

	// parse Authenticity Token from HTML response
	doc, err := goquery.NewDocumentFromResponse(res)
	if err != nil {
		return err
	}
	tokenNodes := doc.Find("input[name=authenticity_token]")
	if tn := tokenNodes.Get(0); tn != nil {
		for _, a := range tn.Attr {
			if a.Key == "value" {
				s.AuthenticityToken = a.Val
				return nil
			}
		}
	}
	return fmt.Errorf("checkout request failed, no authenticity token parsed")
}

func (s *Session) CartTaxCheck(req *CartTaxCheckRequest) (string, error) {
	// tax string
	var tax string

	// construct url
	req.Method = "patch"
	req.PreviousStep = "contact_information"
	req.Step = "contact_information"
	u := s.CheckoutUrl.String()
	v, _ := query.Values(req)
	u += "?" + v.Encode()

	// send request
	res, err := s.Request("GET", u, []byte{})
	if err != nil {
		return "", err
	}

	// parse tax from HTML response
	doc, err := goquery.NewDocumentFromResponse(res)
	if err != nil {
		return "", err
	}

	taxNodes := doc.Find("tr.total-line--taxes")
	data := strings.Split(taxNodes.Text(), " ")
	for _, d := range data {
		d = strings.TrimLeft(strings.TrimSpace(d), "$")
		t, err := strconv.ParseFloat(d, 64)
		if err != nil {
			continue
		}
		tax = strconv.Itoa(int(t * 100.00))
		return tax, err
	}
	err = fmt.Errorf("tax check failed, unable to parse tax node")
	return "", err
}
