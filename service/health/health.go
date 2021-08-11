package health

import (
	"net/http"

	"github.com/getbread/breadkit/desmond"
	"github.com/gin-gonic/gin"
)

func Live(c *gin.Context, dc desmond.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "live"})
}
func Ready(c *gin.Context, dc desmond.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ready"})
}
func Health(c *gin.Context, dc desmond.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "healthy"})
}
