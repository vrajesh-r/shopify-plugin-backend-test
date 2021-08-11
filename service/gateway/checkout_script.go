package gateway

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"net/http"
	"text/template"
	"time"

	"github.com/getbread/breadkit/desmond"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func (h *Handlers) ShopifyPlusCheckoutJS(c *gin.Context, dc desmond.Context) {

	gatewayKey := c.Param("gatewayKey")
	account, err := findGatewayAccountByGatewayKey(gatewayKey, h)
	if err != nil {
		log.WithFields(log.Fields{
			"gatewayKey": gatewayKey,
			"error":      err.Error(),
		}).Error("(ShopifyPlusCheckoutJS) query for shop produced error")
		c.String(400, err.Error())
		return
	}

	tmp, err := template.ParseFiles("build/templates/checkout.js")
	if err != nil {
		log.WithFields(log.Fields{
			"account": account,
			"error":   err.Error(),
		}).Error("(ShopifyPlusCheckoutJS) parsing checkout.js produced error")
		c.String(400, err.Error())
		return
	}

	breadJS := account.CheckoutHost() + "/bread.js"
	apiKey, _ := account.GetAPIKeys()

	var buf bytes.Buffer
	if err = tmp.ExecuteTemplate(&buf, "checkout.js", gin.H{
		"ApiKey":                     apiKey,
		"BreadJS":                    breadJS,
		"EnableCheckout":             account.PlusEmbeddedCheckout,
		"HealthcareMode":             account.HealthcareMode,
		"MiltonHost":                 gatewayConfig.HostConfig.MiltonHost,
		"TargetedFinancing":          account.TargetedFinancing,
		"TargetedFinancingThreshold": account.TargetedFinancingThreshold * 100,
		"TargetedFinancingProgramID": account.TargetedFinancingID,
		"DatadogToken":               gatewayConfig.DataDogToken.Unmask(),
		"DatadogSite":                gatewayConfig.DatadogSite,
	}); err != nil {
		log.WithFields(log.Fields{
			"account": account,
			"error":   err.Error(),
		}).Error("(ShopifyPlusCheckoutJS) executing checkout.js produced error")
		c.String(400, err.Error())
		return
	}
	serveInMemoryContent(c.Writer, c.Request, "checkout.js", buf.Bytes())
}

func serveInMemoryContent(w http.ResponseWriter, r *http.Request, name string, content []byte) {
	h := sha1.New()
	h.Write(content)
	eTag := fmt.Sprintf(`W/"%x"`, h.Sum(nil))
	w.Header().Set("ETag", eTag)
	http.ServeContent(w, r, name, time.Now(), bytes.NewReader(content))
}
