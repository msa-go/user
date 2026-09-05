package middleware

import (
	// golang package
	"net/http"
	"strings"

	// external package
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		// 유효성 검사 1 - 헤더값 존재 여부
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "authorization header required",
			})
			c.Abort()
			return
		}

		// 유효성 검사 2 - Bearer, 토큰값 존재 여부
		// format token
		// Authorization: Bearer xxx
		// split: [Bearer] [xxx]
		tokenString := strings.Split(authHeader, " ")
		if len(tokenString) != 2 || tokenString[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token",
			})
			c.Abort()
			return
		}

		// 유효성 검사 3 - 토큰 유효성 검사
		//
		// keyfunc 에서 서명 알고리즘을 확인하고 파서에도 HS256 만 허용한다.
		// 이 검사가 없으면 파서가 등록된 모든 알고리즘을 받아들여
		// 알고리즘 스위칭 공격에 노출된다.
		// exp 는 존재할 때만 검증되므로, 만료 없는 토큰을 거부하도록 필수화한다.
		token, err := jwt.Parse(tokenString[1], func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(secret), nil
		},
			jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
			jwt.WithExpirationRequired(),
		)
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid Token",
			})
			c.Abort()
			return
		}

		// 유효성 검사 4 - 클레임 유효성 검사
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token",
			})
			c.Abort()
			return
		}

		// 유효성 검사 5 - user_id 클레임 타입 확인
		// 타입 단언에 실패하면 panic 이 나므로 반드시 확인 후 사용한다.
		userID, ok := claims["user_id"].(float64)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid token",
			})
			c.Abort()
			return
		}

		c.Set("user_id", userID)
		c.Next()
	}
}
