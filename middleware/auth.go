package middleware

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/slimnate/laser-beam/data/session"
	"github.com/slimnate/laser-beam/data/user"
)

func AuthMiddleware(sessionRepo *session.SessionRepository, userRepo *user.UserRepository) gin.HandlerFunc {
	autoLoginUser := os.Getenv("AUTO_LOGIN_USER")

	// if auto-login is enabled, we skip checking for any session keys
	// and approve the request as if the `autoLoginUser` is already logged in
	if autoLoginUser != "" {
		return func(ctx *gin.Context) {
			user, err := userRepo.GetByUsername(autoLoginUser)
			if err != nil {
				msg := fmt.Sprintf("Error on auto-login, user '%s' not found", autoLoginUser)
				ctx.AbortWithStatusJSON(500, gin.H{"error": msg})
				return
			}

			ctx.Set("user", &user.User)
		}
	}

	return func(ctx *gin.Context) {
		sessionKey, err := ctx.Cookie("session_key")
		if err != nil {
			ctx.Redirect(302, "/login")
			return
		}

		session, err := sessionRepo.GetByKey(sessionKey)
		if err != nil {
			ctx.Redirect(302, "/login")
			return
		}

		user, err := userRepo.GetByID(session.UserID)
		if err != nil {
			ctx.AbortWithStatus(401)
			return
		}

		// set the userID and orgID on the query context
		ctx.Set("user", user)

		ctx.Next()
	}
}
