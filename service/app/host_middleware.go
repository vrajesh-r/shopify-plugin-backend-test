package app

import (
	"github.com/getbread/breadkit/desmond"
	"github.com/gin-gonic/gin"
)

func (h *Handlers) VerifyHost(c *gin.Context, dc desmond.Context) {
	// how do you parse the hostname of the requester?
	c.Next()
}
