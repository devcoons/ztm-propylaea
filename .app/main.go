package main

import (
	models "api-gateway/models"
	routes "api-gateway/routes"
	"flag"
	"fmt"
	"os"
	"runtime"

	ztm "github.com/devcoons/go-ztm"

	c "github.com/devcoons/go-fmt-colors"
	"github.com/gin-gonic/gin"
)

var APIService ztm.Service

func main() {
	runtime.GOMAXPROCS(8)
	fmt.Println(c.FmtFgBgWhiteLBlue+"[ IMS ]"+c.FmtReset, c.FmtFgBgWhiteBlue+" INFO "+c.FmtReset, c.FmtFgBgWhiteBlack+"Initializing microservice."+c.FmtReset)

	cfgfile, present := os.LookupEnv("IMSCFGFILE")

	if !present {
		wordPtr := flag.String("cfg-file", "", "Service Configuration file")
		flag.Parse()
		if wordPtr == nil || *wordPtr == "" {
			fmt.Println(c.FmtFgBgWhiteLBlue+"[ IMS ]"+c.FmtReset, c.FmtFgBgWhiteRed+" ERRN "+c.FmtReset, c.FmtFgBgWhiteBlack+"Configuration file env.variable `IMSCFGFILE` does not exist"+c.FmtReset)
			return
		}
		cfgfile = *wordPtr
	}

	if !APIService.Initialize(cfgfile) {
		fmt.Println(c.FmtFgBgWhiteLBlue+"[ IMS ]"+c.FmtReset, c.FmtFgBgWhiteRed+" ERRN "+c.FmtReset, c.FmtFgBgWhiteBlack+"Initialization failed. Exiting application.."+c.FmtReset)
		return
	}

	fmt.Println(c.FmtFgBgWhiteLBlue+"[ IMS ]"+c.FmtReset, c.FmtFgBgWhiteBlue+" INFO "+c.FmtReset, c.FmtFgBgWhiteBlack+"Models Database auto-migration"+c.FmtReset)
	models.AutoMigrate(APIService.Database)

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.Use(gin.Recovery())
	router.Use(ztm.AddUSEService(&APIService))

	router.POST("/register", routes.RoutePOSTRegister)
	router.POST("/login", routes.RoutePOSTLogin)
	router.POST("/logout", routes.RouteUserLogout)
	router.GET("/logout", routes.RouteUserLogout)
	router.DELETE("/nonce", routes.RouteDELETENonce)

	APIService.Start(router)

	fmt.Println("[GIN] Starting service at [0.0.0.0:8080]")
	router.Run("0.0.0.0:8080")
}
