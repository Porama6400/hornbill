package auth

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func (s *Service) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := s.GetUser(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, NewResultErrorMessage(err))
			return
		}

		c.Set("user", user)
	}
}
