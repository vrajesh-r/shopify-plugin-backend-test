package app

import (
	"github.com/getbread/breadkit/desmond"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type errorLogRequest struct {
	Error            string `json:"error"`
	Message          string `json:"message"`
	StackTrace       string `json:"stackTrace"`
	UserAgent        string `json:"userAgent"`
	Referrer         string `json:"referrer"`
	PageType         string `json:"pageType"`
	APIKey           string `json:"apiKey"`
	GatewayReference string `json:"gatewayReference"`
	TransactionID    string `json:"transactionID"`
}

// Endpoint for forwarding front end errors to Sentry
func (h *Handlers) LogFrontEndError(c *gin.Context, dc desmond.Context) {
	var r errorLogRequest
	err := c.Bind(&r)
	if err != nil {
		log.WithFields(log.Fields{
			"error":       err.Error(),
			"queryString": c.Request.URL.RawQuery,
		}).Error("(FrontendError Internal) binding request produced an error")

		c.JSON(400, gin.H{
			"status":  "400",
			"message": "Bad Request",
		})
		return
	}

	shopDomain := c.Query("shop")

	ipAddress := c.Request.Header.Get("X-Forwarded-For")

	log.WithFields(log.Fields{
		"error":            r.Error,
		"message":          r.Message,
		"stackTrace":       r.StackTrace,
		"userAgent":        r.UserAgent,
		"referrer":         r.Referrer,
		"pageType":         r.PageType,
		"apiKey":           r.APIKey,
		"gatewayReference": r.GatewayReference,
		"transactionID":    r.TransactionID,
		"shop":             shopDomain,
		"ipAddress":        ipAddress,
	}).Error("(FrontendError) logging frontend error")

	c.JSON(200, gin.H{
		"status":  "200",
		"message": "OK",
	})
}
