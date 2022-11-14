package routes

import (
	"net/http"

	ztm "github.com/devcoons/go-ztm"

	"github.com/gin-gonic/gin"
)

func InitServiceSJWT(c *gin.Context) (*ztm.SJWTClaims, *ztm.Service, bool) {

	srv, ok := c.MustGet("service").(*ztm.Service)

	if !ok {
		c.IndentedJSON(http.StatusExpectationFailed, nil)
		return nil, nil, false
	}

	claims := srv.ValidateServiceJWT(c.Request)

	if claims == nil {
		return nil, nil, false
	}

	return claims, srv, true
}
