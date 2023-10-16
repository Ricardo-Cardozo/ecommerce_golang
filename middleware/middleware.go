package middleware

import (
	"net/http"

	token "github.com/Ricardo-Cardozo/ecommerce_golang/tokens"
	"github.com/gin-gonic/gin"
)

// func Authentication() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		ClientToken := c.Request.Header.Get("token")
// 		if ClientToken == "" {
// 			c.JSON(http.StatusInternalServerError, gin.H{
// 				"error": "No authorization header provided",
// 			})
// 			return
// 		}

// 		claims, err := token.ValidateToken(ClientToken)

// 		if err != "" {
// 			c.JSON(http.StatusInternalServerError, gin.H{
// 				"error": err,
// 			})
// 			return
// 		}

// 		c.Set("email", claims.Email)
// 		c.Set("uid", claims.Uid)
// 		c.Next()
// 	}

// }
func Authentication() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "No Authorization header provided",
			})
			return
		}

		// Verificar se o prefixo "Bearer" está presente
		if len(authHeader) > 7 && authHeader[0:7] == "Bearer " {
			ClientToken := authHeader[7:]

			claims, err := token.ValidateToken(ClientToken)

			if err != "" {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": err,
				})
				return
			}

			c.Set("email", claims.Email)
			c.Set("uid", claims.Uid)
			c.Next()

		} else {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid Authorization header format",
			})
			return
		}
	}
}
