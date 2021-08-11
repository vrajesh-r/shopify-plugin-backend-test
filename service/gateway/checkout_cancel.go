package gateway

import (
	"github.com/getbread/breadkit/desmond"
	zeus "github.com/getbread/breadkit/zeus/types"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (h *Handlers) GatewayCheckoutCancel(c *gin.Context, dc desmond.Context) {
	// Consider removing the need for this route handler altogether by using x_url_callback for offsite checkout error url
	orderRef := c.Query("orderRef")

	checkout, err := findGatewayCheckoutById(zeus.Uuid(orderRef), h)
	if err != nil {
		logrus.WithError(err).WithField("orderRef", orderRef).Error("(GatewayCheckoutCancel) search for offsite checkout produced error")
		c.String(400, err.Error())
		return
	}

	c.Redirect(302, checkout.CancelUrl)
}
