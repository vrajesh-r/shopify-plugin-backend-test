package app

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"hash"
	"io/ioutil"
	"net/url"
	"sort"
	"strings"

	"github.com/getbread/breadkit/desmond"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

var shopifySharedSecret string

// steps to authentication
// create test string
// compare test string to control string
func (h *Handlers) ProxyAuthentication(c *gin.Context, dc desmond.Context) {
	// construct inputs for test
	test := constructProxyTest(c.Request.URL.RawQuery)
	control := pullControlFromQuery(c.Request.URL)

	// test
	if testMatchesControl(test, control) {
		c.Next()
		return
	}
	log.WithFields(log.Fields{
		"queryString": c.Request.URL.RawQuery,
		"test":        test,
		"control":     control,
	}).Error("(ProxyAuth) failed")
	c.AbortWithStatus(401)
}

func (h *Handlers) HttpAuthentication(c *gin.Context, dc desmond.Context) {
	// contruct inputs for test
	var (
		test    string
		control string
		err     error
		bb      []byte
	)
	if c.Request.Method == "GET" {
		test = constructHttpGetTest(c.Request.URL.RawQuery)
		control = pullControlFromQuery(c.Request.URL)
	} else {
		bb, err = ioutil.ReadAll(c.Request.Body)
		if err != nil {
			// log error
			log.WithFields(log.Fields{
				"httpMethod":  c.Request.Method,
				"queryString": c.Request.URL.RawQuery,
				"requestBody": string(bb),
				"error":       err.Error(),
				"control":     c.Request.Header.Get("X-Shopify-Hmac-SHA256"),
			}).Error("(HttpAuth) reading body failed")

			// short circuit
			c.AbortWithStatus(401)
			return
		}
		test, err = constructHttpPostTest(bb)
		if err != nil {
			// log error
			log.WithFields(log.Fields{
				"httpMethod":  c.Request.Method,
				"queryString": c.Request.URL.RawQuery,
				"requestBody": string(bb),
				"error":       err.Error(),
				"control":     c.Request.Header.Get("X-Shopify-Hmac-SHA256"),
			}).Error("(HttpAuth) constructing test failed")

			// short circuit
			c.AbortWithStatus(401)
			return
		}
		control = c.Request.Header.Get("X-Shopify-Hmac-SHA256")

		// reset body for use in controller code
		c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bb))
	}

	// test
	if testMatchesControl(test, control) {
		c.Next()
		return
	}

	log.WithFields(log.Fields{
		"httpMethod":  c.Request.Method,
		"queryString": c.Request.URL.RawQuery,
		"requestBody": string(bb),
		"control":     control,
		"test":        test,
	}).Error("(HttpAuth) auth test failed")
	c.AbortWithStatus(401)
}

func pullControlFromQuery(url *url.URL) string {
	v := url.Query()
	var control string
	if control = v.Get("hmac"); control == "" {
		control = v.Get("signature")
	}
	return control
}

func constructProxyTest(rawQuery string) string {
	q, err := url.QueryUnescape(rawQuery)
	if err != nil {
		// log it
		log.WithField("queryString", rawQuery).Error("(HttpAuth) url.QueryUnescape failed in HTTP auth contructProxyTest")
		return ""
	}
	pairs := strings.Split(q, "&")
	var sortPairs sort.StringSlice
	for _, pair := range pairs {
		if strings.HasPrefix(pair, "signature") {
			continue
		}
		// TODO run join on values in key/value pairs which are arrays, with ","
		sortPairs = append(sortPairs, pair)
	}
	sortPairs.Sort()
	mac := newHmacSHA256()
	mac.Write([]byte(strings.Join(sortPairs, "")))
	return hex.EncodeToString(mac.Sum(nil))
}

func constructHttpGetTest(rawQuery string) string {
	q, err := url.QueryUnescape(rawQuery)
	if err != nil {
		// log it
		log.WithField("queryString", rawQuery).Error("(HttpAuth) url.QueryUnescape failed in HTTP auth constructHttpGetTest")
		return ""
	}
	pairs := strings.Split(q, "&")
	var sortPairs sort.StringSlice
	for _, pair := range pairs {
		pieces := strings.Split(pair, "=")
		if pieces[0] == "signature" || pieces[0] == "hmac" { // omit
			continue
		}
		for i, _ := range pieces {
			pieces[i] = strings.Replace(pieces[i], "%", "%25", -1)
			pieces[i] = strings.Replace(pieces[i], "&", "%26", -1)
			if i == 0 {
				pieces[i] = strings.Replace(pieces[i], "=", "%3D", -1)
			}
		}
		sortPairs = append(sortPairs, strings.Join(pieces, "="))
	}
	sortPairs.Sort()
	mac := newHmacSHA256()
	mac.Write([]byte(strings.Join(sortPairs, "&")))
	return hex.EncodeToString(mac.Sum(nil))
}

func constructHttpPostTest(contents []byte) (string, error) {
	mac := newHmacSHA256()
	mac.Write(contents)
	return base64.StdEncoding.EncodeToString(mac.Sum(nil)), nil
}

func newHmacSHA256() hash.Hash {
	return hmac.New(sha256.New, []byte(appConfig.ShopifyConfig.ShopifySharedSecret.Unmask()))
}

func testMatchesControl(test, control string) bool {
	return test == control
}
