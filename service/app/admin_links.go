package app

import (
	"net/http"
	"strings"
	"time"

	"github.com/getbread/breadkit/desmond"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func (h *Handlers) DraftOrderDetails(c *gin.Context, dc desmond.Context) {
	_, _, err := h.validateAppSession(c)
	if err != nil {
		c.Next()
		return
	}
	c.Redirect(302, "/portal/draftorder/"+c.Query("id"))
	return
}

func (h *Handlers) AuthorizeDraftOrderView(c *gin.Context, dc desmond.Context) {
	shopUrl := c.Query("shop")
	shopName := strings.Split(shopUrl, ".")[0]

	shop, err := findShopByName(shopName, h)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error":       err.Error(),
			"queryString": c.Request.URL.RawQuery,
		}).Error("(AuthorizeDraftOrderView) query for shop produced error")
		c.AbortWithStatus(401)
		return
	}

	// generate new session
	session, err := createSessionByShopId(shop.Id, h)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error":       err.Error(),
			"queryString": c.Request.URL.RawQuery,
			"shop":        shop,
		}).Error("(AuthorizeDraftOrderView) creating session produced error")
		c.AbortWithStatus(401)
		return
	}

	// add cookie
	expire := time.Unix(session.Expiration, 0)
	cookie := &http.Cookie{
		Name:       ADMIN_COOKIE_NAME,
		Value:      string(session.Id),
		Expires:    expire,
		RawExpires: expire.Format(time.UnixDate),
		Secure:     true,
		Path:       "/portal",
		HttpOnly:   true,
		SameSite:   http.SameSiteNoneMode,
	}
	http.SetCookie(c.Writer, cookie)

	c.Redirect(302, "/portal/draftorder/"+c.Query("id"))
}
