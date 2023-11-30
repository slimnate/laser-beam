package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/slimnate/laser-beam/data/organization"
)

// Middleware to check for a valid auth key, and add the corresponding org id to the request context
func ApiAuthMiddleware(orgRepo *organization.OrganizationRepository) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		key, exists := ctx.GetQuery("key")
		if !exists {
			ctx.AbortWithStatusJSON(401, gin.H{"error": "no api key supplied"})
			return
		}

		org, err := orgRepo.GetByKey(key)
		if err != nil {
			ctx.AbortWithStatusJSON(401, gin.H{"error": "invalid api key"})
			return
		}

		// Set flag for global authorization if org ID matches the global org
		if org.ID == 1 {
			ctx.Set("authorizedGlobal", true)
		}

		ctx.Set("apiKey", key)
		ctx.Set("authorizedOrgID", org.ID)

		ctx.Next()
	}
}
