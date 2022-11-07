package routes

import (
	middleware "api-gateway/middleware"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RouteUserLogout(c *gin.Context) {

	srv, ok := c.MustGet("service").(*middleware.Service)
	if !ok || srv == nil {
		c.IndentedJSON(http.StatusExpectationFailed, nil)
		return
	}

	if srv.Config.PathNonce.Host == "" || srv.Config.PathNonce.URL == "" {
		c.AbortWithStatus(400)
		return
	}

	claims := srv.ValidateUserJWT(c.Request)
	if claims == nil || claims.UserId == -1 {
		c.AbortWithStatus(401)
		return
	}

	if srv.RefreshUserNonceFromAll(claims.UserId) {
		c.Data(200, "application/json", nil)
		return
	}
	c.AbortWithStatus(401)
}
