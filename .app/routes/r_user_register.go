package routes

import (
	middleware "api-gateway/middleware"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func RoutePOSTRegister(c *gin.Context) {

	srv, ok := c.MustGet("service").(*middleware.Service)

	if !ok || srv == nil {
		c.IndentedJSON(http.StatusExpectationFailed, nil)
		return
	}

	if srv.Config.PathRegister.Host == "" || srv.Config.PathRegister.URL == "" {
		c.AbortWithStatus(400)
		return
	}

	var sclaims middleware.SJWTClaims

	sclaims.Auth = false
	sclaims.Hop = 2
	sclaims.Role = 0
	sclaims.Service = "api-gateway"
	sclaims.UserId = -1

	token := srv.SJwt.GenerateJWT(sclaims)

	client := &http.Client{}
	req, _ := http.NewRequest(c.Request.Method, srv.Config.PathRegister.Host+":"+strconv.Itoa(srv.Config.PathRegister.Port)+srv.Config.PathRegister.URL, nil)
	req.Header = c.Request.Header
	req.Header.Del("Authorization")
	req.Header.Add("Authorization", srv.SJwt.AuthType+" "+token)
	req.Body = c.Request.Body
	res, errn := client.Do(req)
	if errn == nil {
		body, _ := ioutil.ReadAll(res.Body)
		c.Data(res.StatusCode, res.Header.Get("Content-Type"), body)
	} else {
		c.Data(503, "application/json", nil)
	}

}
