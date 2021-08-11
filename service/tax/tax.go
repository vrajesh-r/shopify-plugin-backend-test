package tax

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
)

const avalaraUrl = "https://taxrates.api.avalara.com:443/"

var avalaraKey string

func TaxInitConfig(key string) {
	avalaraKey = key
}

type TaxRateResponse struct {
	TotalRate float64   `json:"totalRate"`
	Rates     []TaxRate `json:"rates"`
}

type TaxRate struct {
	Rate float64 `json:"rate"`
	Name string  `json:"name"`
	Type string  `json:"type"`
}

func TaxRateByZip(zip int, body *TaxRateResponse) error {
	// construct request
	path := "/postal?"
	v := url.Values{}
	v.Add("country", "usa")
	v.Add("postal", strconv.Itoa(zip))
	u := avalaraUrl + path + v.Encode()

	// make request
	c := &http.Client{}
	req, err := http.NewRequest("GET", u, bytes.NewBuffer([]byte{}))
	req.Header.Set("Authorization", "AvalaraApiKey "+avalaraKey)
	res, err := c.Do(req)
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
