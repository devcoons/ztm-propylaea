package routes

import (
	middleware "api-gateway/middleware"
	utilities "api-gateway/utilities"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func RoutePOSTLogin(c *gin.Context) {

	srv, ok := c.MustGet("service").(*middleware.Service)

	if !ok || srv == nil {
		c.IndentedJSON(http.StatusExpectationFailed, nil)
		return
	}

	if srv.Config.PathAuth.Host == "" || srv.Config.PathAuth.URL == "" {
		c.AbortWithStatus(400)
		return
	}

	url := srv.Config.PathAuth.Host + ":" + strconv.Itoa(srv.Config.PathAuth.Port) + srv.Config.PathAuth.URL
	res, errn := srv.ZTServiceRequest(url, "POST", c.Request.Header, c.Request.Body, middleware.SJWTClaims{Auth: false, Hop: 2, Role: -1, Service: "api-gateway", UserId: -1})

	if errn == nil {

		if res.StatusCode != 200 {
			c.Data(503, "application/json", nil)
			return
		}

		values := UnmashalBody(res.Body)

		if values["id"] == "" {
			c.IndentedJSON(http.StatusNotAcceptable, nil)
			return
		}

		id := utilities.ConvertToInt(values["id"], -1)
		role := utilities.ConvertToInt(values["role"], -1)
		nonce := values["nonce"].(string)

		srv.ClearUserNonceFromAll(id)

		token, ok := srv.IssueNewUserJWT(middleware.UJWTClaims{UserId: id, Role: role, Nonce: nonce, Auth: true})

		if !ok {
			c.IndentedJSON(http.StatusExpectationFailed, nil)
			return
		}
		r, _ := json.Marshal(struct {
			Id       int    `json:"id"`
			Username string `json:"username"`
			Role     int    `json:"role"`
			Token    string `json:"token"`
		}{id, values["username"].(string), role, token})
		c.Data(http.StatusOK, gin.MIMEJSON, (r))
	} else {
		c.Data(503, "application/json", nil)
	}
}
