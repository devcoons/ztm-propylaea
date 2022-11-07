package middleware

import (
	"encoding/json"

	"fmt"
	"io/ioutil"
	"os"

	c "github.com/devcoons/go-fmt-colors"
)

type ServicesStatus struct {
	Name    string
	IsAlive bool
}

type imsConfiguration struct {
	Title       string
	Abbeviation string
}

type serviceConfigurationJWT struct {
	Name     string
	Secret   string
	Duration int
	AuthType string
}

type serviceConfigurationDatabase struct {
	Host     string
	Port     int
	Username string
	Password string
	DbName   string
}

type serviceConfigurationService struct {
	Name string
	Host string
	Port int
	URL  string
}

type serviceConfigurationGateway struct {
	Host string
	Port int
}

type serviceConfigurationRedisDB struct {
	Host     string
	Port     int
	Username string
	Password string
	DB       int
}

type serviceConfigurationAuth struct {
	Host string
	Port int
	URL  string
}
type serviceConfigurationRegister struct {
	Host string
	Port int
	URL  string
}

type serviceConfigurationNonce struct {
	Host string
	Port int
	URL  string
}

type ServiceConfiguration struct {
	Ims          imsConfiguration
	Secrets      []serviceConfigurationJWT
	Database     serviceConfigurationDatabase
	RedisDB      serviceConfigurationRedisDB
	Services     []serviceConfigurationService
	Gateways     []serviceConfigurationGateway
	PathAuth     serviceConfigurationAuth
	PathRegister serviceConfigurationRegister
	PathNonce    serviceConfigurationNonce
}

func (u *ServiceConfiguration) Load(dbconfig string) bool {

	fmt.Println(c.FmtFgBgWhiteLBlue+"[ IMS ]"+c.FmtReset, c.FmtFgBgWhiteBlue+" INFO "+c.FmtReset, c.FmtFgBgWhiteBlack+"Loading configuration file ("+dbconfig+")"+c.FmtReset)

	jsonFile, err := os.Open(dbconfig)
	if err != nil {
		fmt.Println(c.FmtFgBgWhiteLBlue+"[ IMS ]"+c.FmtReset, c.FmtFgBgWhiteRed+" ERRN "+c.FmtReset, c.FmtFgBgWhiteBlack+"Cannot open the configuration file"+c.FmtReset)
		return false
	}
	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		fmt.Println(c.FmtFgBgWhiteLBlue+"[ IMS ]"+c.FmtReset, c.FmtFgBgWhiteRed+" ERRN "+c.FmtReset, c.FmtFgBgWhiteBlack+"Cannot read the configuration file"+c.FmtReset)
		return false
	}
	err = json.Unmarshal(byteValue, u)

	if err != nil {
		fmt.Println(c.FmtFgBgWhiteLBlue+"[ IMS ]"+c.FmtReset, c.FmtFgBgWhiteRed+" ERRN "+c.FmtReset, c.FmtFgBgWhiteBlack+"Cannot parse the configuration file"+c.FmtReset)
		return false
	}

	fmt.Println(c.FmtFgBgWhiteLBlue+"[ IMS ]"+c.FmtReset, c.FmtFgBgWhiteBlue+" INFO "+c.FmtReset, c.FmtFgBgWhiteBlack+"Loading configuration file ("+dbconfig+")"+c.FmtReset)

	return true
}
