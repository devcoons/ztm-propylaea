package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	aJWT "github.com/devcoons/go-auth-jwt"
	c "github.com/devcoons/go-fmt-colors"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/logrusorgru/aurora"
	"github.com/mitchellh/mapstructure"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Service struct {
	UJwt     *aJWT.AuthJWT
	SJwt     *aJWT.AuthJWT
	Config   *ServiceConfiguration
	Database *gorm.DB
	Rdb      *redis.Client
}

func (u *Service) Initialize(cfgpath string) bool {
	var err error

	u.Config = &ServiceConfiguration{}
	r := u.Config.Load(cfgpath)
	fmt.Println(c.FmtFgBgWhiteLBlue+"[ IMS ]"+c.FmtReset, c.FmtFgBgWhiteBlue+" INFO "+c.FmtReset, c.FmtFgBgWhiteBlack+" Loading service configuration for: "+u.Config.Ims.Title+"."+c.FmtReset)

	if !r {
		return false
	}

	if u.Config.RedisDB.Host != "" {
		fmt.Println(c.FmtFgBgWhiteLBlue+"[ IMS ]"+c.FmtReset, c.FmtFgBgWhiteBlue+" INFO "+c.FmtReset, c.FmtFgBgWhiteBlack+"Redis Instance will be available"+c.FmtReset)

		u.Rdb = redis.NewClient(&redis.Options{
			Addr:     u.Config.RedisDB.Host + ":" + strconv.Itoa(u.Config.RedisDB.Port),
			Username: u.Config.RedisDB.Username,
			Password: u.Config.RedisDB.Password,
			DB:       u.Config.RedisDB.DB,
		})
		var ctx = context.Background()
		u.Rdb.FlushDB(ctx)

	} else {
		fmt.Println(aurora.BgBrightYellow("[ IMS ] Redis Instance will NOT be available.."))
	}

	if u.Config.Secrets != nil {
		for _, s := range u.Config.Secrets {
			if strings.ToLower(s.Name) == "sjwt" {
				fmt.Println(c.FmtFgBgWhiteLBlue+"[ IMS ]"+c.FmtReset, c.FmtFgBgWhiteBlue+" INFO "+c.FmtReset, c.FmtFgBgWhiteBlack+"SJWT Token will be available"+c.FmtReset)
				u.SJwt = &aJWT.AuthJWT{}
				u.SJwt.SecretKey = s.Secret
				u.SJwt.TokenDuration = time.Duration(s.Duration) * time.Second
				u.SJwt.AuthType = s.AuthType
			}
			if strings.ToLower(s.Name) == "ujwt" {
				fmt.Println(c.FmtFgBgWhiteLBlue+"[ IMS ]"+c.FmtReset, c.FmtFgBgWhiteBlue+" INFO "+c.FmtReset, c.FmtFgBgWhiteBlack+"UJWT Token will be available"+c.FmtReset)
				u.UJwt = &aJWT.AuthJWT{}
				u.UJwt.SecretKey = s.Secret
				u.UJwt.TokenDuration = time.Duration(s.Duration) * time.Second
				u.UJwt.AuthType = s.AuthType
			}
		}
	} else {
		fmt.Println(aurora.BgRed("[ IMS ] Microservice cannot work without Secrets"))
		return false
	}

	if u.Config.Database.Host != "" {
		u.Database = &gorm.DB{}

		dsn := u.Config.Database.Username + ":" + u.Config.Database.Password + "@tcp("
		dsn += u.Config.Database.Host + ":" + strconv.Itoa(u.Config.Database.Port) + ")/"
		dsn += u.Config.Database.DbName + "?parseTime=true"

		for i := 1; i <= 5; i++ {
			fmt.Println(c.FmtFgBgWhiteLBlue+"[ IMS ]"+c.FmtReset, c.FmtFgBgWhiteBlue+" INFO "+c.FmtReset, c.FmtFgBgWhiteBlack+"Connecting SQL database: "+u.Config.Database.Host+":"+strconv.Itoa(u.Config.Database.Port)+c.FmtReset)

			u.Database, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
			if err != nil {
				fmt.Println(aurora.BgBrightYellow("[ IMS ] Connection failed. Retring in 7 seconds.."))
				time.Sleep(7 * time.Second)
			} else {
				fmt.Println(c.FmtFgBgWhiteLBlue+"[ IMS ]"+c.FmtReset, c.FmtFgBgWhiteBlue+" INFO "+c.FmtReset, c.FmtFgBgWhiteBlack+"Connection succesfully completed"+c.FmtReset)
				break
			}
		}
	} else {
		fmt.Println(aurora.BgBrightYellow("[ IMS ] Sql Database will NOT be available.."))
	}

	return err == nil
}

func AddUSEService(u *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("service", u)
		c.Next()
	}
}

func (u *Service) IssueNewUserJWT(claims UJWTClaims) (string, bool) {

	var ctx = context.Background()
	if u.Rdb == nil {
		return "", false
	}
	_, err := u.Rdb.Get(ctx, strconv.Itoa(claims.UserId)).Result()
	if err == nil {
		_, _ = u.Rdb.Del(ctx, strconv.Itoa(claims.UserId)).Result()
	}
	u.Rdb.Set(ctx, strconv.Itoa(claims.UserId), claims.Nonce, 0)
	token := u.UJwt.GenerateJWT(claims)

	return token, true
}

func (u *Service) ValidateUserJWT(r *http.Request) *UJWTClaims {

	if u.Rdb == nil {
		return nil
	}

	iclaims, ok := u.UJwt.IsAuthorized(r)
	if !ok {
		return nil
	}

	var claims UJWTClaims
	var claimsmin UJWTClaimsMinimal
	err := mapstructure.Decode(iclaims, &claimsmin)

	if err != nil {
		return nil
	}
	claims.Auth = claimsmin.A
	claims.Nonce = claimsmin.N
	claims.Role = claimsmin.R
	claims.UserId = claimsmin.U

	res := u.CompareUserNonce(claims.UserId, claims.Nonce)

	if res {
		return &claims
	}
	return nil
}

func (u *Service) ValidateServiceJWT(r *http.Request) *SJWTClaims {

	if u.Rdb == nil {
		return nil
	}

	iclaims, ok := u.SJwt.IsAuthorized(r)

	if !ok {
		return nil
	}

	var claims SJWTClaims
	err := mapstructure.Decode(iclaims, &claims)

	if err != nil {
		return nil
	}

	return &claims
}

func (u *Service) UpdateUserNonce(userId int, userNonce string) bool {

	var ctx = context.Background()
	if u.Rdb == nil {
		return false
	}

	_, err := u.Rdb.Get(ctx, strconv.Itoa(userId)).Result()
	if err == nil {
		_, _ = u.Rdb.Del(ctx, strconv.Itoa(userId)).Result()
	}
	u.Rdb.Set(ctx, strconv.Itoa(userId), userNonce, 0)

	for _, gw := range u.Config.Gateways {

		var sclaims SJWTClaims
		sclaims.Auth = true
		sclaims.UserId = userId
		sclaims.Role = 0
		sclaims.Service = "api-gateway"

		token := u.SJwt.GenerateJWT(sclaims)

		client := &http.Client{}
		req, _ := http.NewRequest("GET", gw.Host+":"+strconv.Itoa(gw.Port)+"/syncunc", nil)
		req.Header.Del("Authorization")
		req.Header.Add("Authorization", "X-Fowarder "+token)
		req.Body = nil
		client.Do(req)
	}
	return true
}

func (u *Service) ReloadUserNonceFromDB(userId int, userNonce string) bool {

	var ctx = context.Background()
	if u.Rdb == nil {
		return false
	}

	_, err := u.Rdb.Get(ctx, strconv.Itoa(userId)).Result()
	if err == nil {
		_, _ = u.Rdb.Del(ctx, strconv.Itoa(userId)).Result()
	}
	u.Rdb.Set(ctx, strconv.Itoa(userId), userNonce, 0)

	return true
}

func (u *Service) CompareUserNonce(userId int, nonce string) bool {

	var ctx = context.Background()
	if u.Rdb == nil {
		return false
	}

	rdb_nonce, err := u.Rdb.Get(ctx, strconv.Itoa(userId)).Result()
	if err == nil && rdb_nonce == nonce {
		return true
	}

	var sclaims SJWTClaims
	sclaims.Auth = true
	sclaims.Role = 0
	sclaims.UserId = userId
	sclaims.Service = "api-gateway"
	sclaims.Hop = 2
	token := u.SJwt.GenerateJWT(sclaims)
	client := &http.Client{}
	req, _ := http.NewRequest("GET", u.Config.PathNonce.Host+":"+strconv.Itoa(u.Config.PathNonce.Port)+u.Config.PathNonce.URL, nil)
	req.Header.Del("Authorization")
	req.Header.Add("Authorization", u.SJwt.AuthType+" "+token)
	req.Body = nil
	res, errn := client.Do(req)

	values := UnmashalBody(res.Body)

	if errn != nil || values == nil {
		return false
	}

	db_nonce := values["nonce"].(string)
	if db_nonce != nonce {
		return false
	}

	u.Rdb.Set(ctx, strconv.Itoa(userId), db_nonce, 0)
	return true
}

func (u *Service) DeleteUserNonceFromDB(userId int) bool {

	var ctx = context.Background()
	if u.Rdb == nil {
		return false
	}
	_, err := u.Rdb.Get(ctx, strconv.Itoa(userId)).Result()
	if err == nil {
		_, _ = u.Rdb.Del(ctx, strconv.Itoa(userId)).Result()
	} else {
	}

	return true
}

func UnmashalBody(body io.ReadCloser) map[string]interface{} {
	var values map[string]interface{}

	bbody, err := ioutil.ReadAll(body)

	if err != nil {
		return nil
	}

	json.Unmarshal([]byte(bbody), &values)
	return values
}

func (u *Service) ZTServiceRequest(url string, method string, header http.Header, body io.ReadCloser, claims SJWTClaims) (*http.Response, error) {

	if url == "" || method == "" {
		return nil, errors.New("Failed")
	}

	token := u.SJwt.GenerateJWT(claims)
	client := &http.Client{}
	req, _ := http.NewRequest(method, url, nil)
	req.Header = header
	req.Header.Del("Authorization")
	req.Header.Add("Authorization", u.SJwt.AuthType+" "+token)
	req.Body = body
	res, errn := client.Do(req)

	if errn != nil {
		return nil, errors.New("Failed")
	}
	return res, nil
}

func (u *Service) ClearUserNonceFromAll(userId int) bool {

	u.DeleteUserNonceFromDB(userId)

	errns := 0
	var sclaims SJWTClaims
	sclaims.Auth = true
	sclaims.Hop = 2
	sclaims.Role = 9
	sclaims.Service = "api-gateway"
	sclaims.UserId = userId

	token := u.SJwt.GenerateJWT(sclaims)

	for _, gateway := range u.Config.Gateways {
		gclient := &http.Client{}
		req1, _ := http.NewRequest("DELETE", gateway.Host+":"+strconv.Itoa(gateway.Port)+"/nonce", nil)
		req1.Header.Del("Authorization")
		req1.Header.Add("Authorization", u.SJwt.AuthType+" "+token)
		req1.Body = nil
		_, errn := gclient.Do(req1)
		if errn != nil {
			errns = errns + 1
		}
	}
	return errns == 0
}

func (u *Service) RefreshUserNonceFromAll(userId int) bool {

	var sclaims SJWTClaims
	sclaims.Auth = true
	sclaims.Hop = 2
	sclaims.Role = 9
	sclaims.Service = "api-gateway"
	sclaims.UserId = userId

	token := u.SJwt.GenerateJWT(sclaims)
	client := &http.Client{}
	req, _ := http.NewRequest("PATCH", u.Config.PathNonce.Host+":"+strconv.Itoa(u.Config.PathNonce.Port)+u.Config.PathNonce.URL, nil)
	req.Header.Del("Authorization")
	req.Header.Add("Authorization", u.SJwt.AuthType+" "+token)
	req.Body = nil
	client.Do(req)
	return u.ClearUserNonceFromAll(userId)
}
