package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RouteDELETENonce(c *gin.Context) {

	claims, srv, ok := InitServiceSJWT(c)

	if !ok || srv == nil {
		c.IndentedJSON(http.StatusExpectationFailed, nil)
		return
	}
	if claims.UserId == -1 || claims.Service != "api-gateway" {
		c.IndentedJSON(418, nil)
		return
	}
	srv.DeleteUserNonceFromDB(claims.UserId)
	c.IndentedJSON(http.StatusAccepted, nil)
}
