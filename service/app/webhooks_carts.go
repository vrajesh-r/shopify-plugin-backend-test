package app

import (
	"encoding/json"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/getbread/breadkit/desmond"
	"github.com/getbread/shopify_plugin_backend/service/types"
)

func (h *Handlers) PersistCheckout(c *gin.Context, dc desmond.Context) {
	c.String(200, "success")

	var req types.CreateCheckoutRequest
	if err := c.BindJSON(&req); err != nil {
		logrus.WithFields(logrus.Fields{
			"error":       err.Error(),
			"queryString": c.Request.URL.RawQuery,
		}).Error("(PersistCheckout) Binding request to model produced error")
		return
	}

	// Serialize request to store into redis
	serializedRequest, err := json.Marshal(req)
	if err != nil {
		logrus.Error(fmt.Sprintf("(PersistCheckout) Failed to serialize CreateCheckoutRequest into JSON: %s", err.Error()))
		return
	}

	// Store req in redis
	var key string = fmt.Sprintf("checkout-%d", req.Id)

	conn := h.RedisPool.Get()
	defer conn.Close()

	set, err := conn.Do("SET", key, serializedRequest)

	if err != nil {
		logrus.Error(fmt.Sprintf("(PersistCheckout) Failed to set stuff in redis: %s", err))
	} else {
		logrus.Info(fmt.Sprintf("(PersistCheckout) Stored %s, result: %v, value: %s", key, set, serializedRequest))
	}

	expire, err := conn.Do("EXPIRE", key, "172800") // Expires after 48 hours

	if err != nil {
		logrus.Warn(fmt.Sprintf("(PersistCheckout) Failed to set expire time for key: %s", key))
	} else {
		logrus.Info(fmt.Sprintf("(PersistCheckout) Set expire on key: %s, expire result: %s", key, expire))
	}

}
