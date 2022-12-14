package routes

import (
	"encoding/json"
	"net/http"
	"strconv"

	ztm "github.com/devcoons/go-ztm"

	"github.com/gin-gonic/gin"
)

func RoutePOSTLogin(c *gin.Context) {

	srv, ok := c.MustGet("service").(*ztm.Service)

	if !ok || srv == nil {
		c.IndentedJSON(http.StatusExpectationFailed, nil)
		return
	}

	if srv.Config.PathAuth.Host == "" || srv.Config.PathAuth.URL == "" {
		c.AbortWithStatus(400)
		return
	}

	url := srv.Config.PathAuth.Host + ":" + strconv.Itoa(srv.Config.PathAuth.Port) + srv.Config.PathAuth.URL
	res, errn := srv.RequestWithClaims(url, "POST", nil, c.Request.Body, ztm.SJWTClaims{Auth: false, Hop: 2, Role: -1, Service: srv.Config.Ims.Abbeviation, UserId: -1})

	if errn == nil && res != nil && res.StatusCode == 200 {

		values := ztm.UnmashalBody(res.Body)

		if values["id"] == "" {
			c.IndentedJSON(http.StatusNotAcceptable, nil)
			return
		}

		id := ztm.ConvertToInt(values["id"], -1)
		role := ztm.ConvertToInt(values["role"], -1)
		nonce := values["nonce"].(string)

		srv.ClearUserNonceFromAll(id)

		if id == -1 || role == -1 {
			c.IndentedJSON(http.StatusExpectationFailed, nil)
			return
		}

		isAdmin := srv.IsUserAdmin(&ztm.UJWTClaims{UserId: id, Role: role, Nonce: nonce, Auth: true})
		token, ok := srv.IssueNewUserJWT(ztm.UJWTClaims{UserId: id, Role: role, Nonce: nonce, Auth: true, SysAdm: isAdmin})

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
