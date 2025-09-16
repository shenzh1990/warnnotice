package jwtutil

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/request"
	"github.com/gin-gonic/gin"
	"net/http"
	"reflect"
	"time"
	"warnnotice/util"
)

const (
	TOKEN_VALUE = "token_value"
)

func AuthorizedMiddelware(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := request.ParseFromRequest(c.Request, request.AuthorizationHeaderExtractor,
			func(token *jwt.Token) (interface{}, error) {
				return []byte(secret), nil
			})
		if err == nil {
			if token.Valid {
				claims := token.Claims
				c.Set(TOKEN_VALUE, GetMapFromClaims(claims)[util.Login_User])
				c.Next()
			} else {
				c.String(http.StatusUnauthorized, util.JsonResponse(-1, "token不合法", nil))
				c.Abort()
				return
			}
		} else {
			c.String(http.StatusUnauthorized, util.JsonResponse(-1, "token不合法", nil))
			c.Abort()
			return
		}
	}
}
func SetToken(maps map[string]interface{}, secret string) string {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := make(jwt.MapClaims)
	//有效期
	claims["claim_exp"] = time.Now().Add(time.Hour * time.Duration(1)).Unix()

	//生成时间
	claims["claim_iat"] = time.Now().Unix()
	for k, v := range maps {
		claims[k] = v
	}
	token.Claims = claims
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return ""
	}
	return tokenString
}
func GetMapFromClaims(claims jwt.Claims) map[string]interface{} {
	v := reflect.ValueOf(claims)
	maps := make(map[string]interface{})
	if v.Kind() == reflect.Map {
		for _, k := range v.MapKeys() {
			value := v.MapIndex(k)
			maps[fmt.Sprintf("%s", k.Interface())] = fmt.Sprintf("%v", value.Interface())
		}
		return maps
	}
	return nil
}
