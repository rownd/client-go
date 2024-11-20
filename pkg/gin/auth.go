package gin

import (
    "github.com/gin-gonic/gin"
    "github.com/rgthelen/rownd-go-test/pkg/rownd"
    "strings"
)

type AuthOptions struct {
    FetchUserInfo     bool
    ErrOnInvalidToken bool
    ErrOnMissingUser  bool
}

func Authenticate(client *rownd.Client, opts AuthOptions) gin.HandlerFunc {
    return func(c *gin.Context) {
        // Get token from Authorization header
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.AbortWithStatusJSON(401, gin.H{"error": "No authorization header"})
            return
        }

        // Extract token
        token := strings.TrimPrefix(authHeader, "Bearer ")
        
        // Validate token
        tokenInfo, err := client.ValidateToken(token)
        if err != nil {
            if opts.ErrOnInvalidToken {
                c.AbortWithStatusJSON(401, gin.H{"error": "Invalid token"})
                return
            }
            c.Next()
            return
        }

        // Fetch user info if requested
        if opts.FetchUserInfo {
            user, err := client.GetUser(tokenInfo.UserID)
            if err != nil {
                if opts.ErrOnMissingUser {
                    c.AbortWithStatusJSON(404, gin.H{"error": "User not found"})
                    return
                }
            } else {
                c.Set("rownd_user", user)
            }
        }

        c.Set("rownd_token_info", tokenInfo)
        c.Next()
    }
}