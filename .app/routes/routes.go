package routes

import (
	middleware "api-gateway/middleware"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
)

func UnmashalBody(body io.ReadCloser) map[string]interface{} {
	var values map[string]interface{}

	bbody, err := ioutil.ReadAll(body)

	if err != nil {
		return nil
	}

	json.Unmarshal([]byte(bbody), &values)
	return values
}

func InitServiceSJWT(c *gin.Context) (*middleware.SJWTClaims, *middleware.Service, bool) {

	srv, ok := c.MustGet("service").(*middleware.Service)

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
