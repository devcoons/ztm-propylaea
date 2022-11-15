package routes

import (
	"io/ioutil"
	"net/http"
	"strconv"

	ztm "github.com/devcoons/go-ztm"

	"github.com/gin-gonic/gin"
)

func RoutePOSTRegister(c *gin.Context) {

	srv, ok := c.MustGet("service").(*ztm.Service)

	if !ok || srv == nil {
		c.IndentedJSON(http.StatusExpectationFailed, nil)
		return
	}

	if srv.Config.PathRegister.Host == "" || srv.Config.PathRegister.URL == "" {
		c.AbortWithStatus(400)
		return
	}

	url := srv.Config.PathRegister.Host + ":" + strconv.Itoa(srv.Config.PathRegister.Port) + srv.Config.PathRegister.URL
	res, errn := srv.RequestWithClaims(url, "POST", nil, c.Request.Body, ztm.SJWTClaims{Auth: false, Hop: 2, Role: -1, Service: srv.Config.Ims.Abbeviation, UserId: -1})

	if errn == nil {
		body, _ := ioutil.ReadAll(res.Body)
		c.Data(res.StatusCode, res.Header.Get("Content-Type"), body)
	} else {
		c.Data(503, "application/json", nil)
	}

}
