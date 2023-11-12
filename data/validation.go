package data

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Checks that the organization ID supplied in the request params matches the organization ID that was retrieved when validating the supplied API key
func ValidateOrganizationID(ctx *gin.Context) (id int64, err error) {
	// get authorized org id from request context
	authorizedOrgIdString, exists := ctx.Get("authorizedOrgID")
	if !exists {
		// ctx.AbortWithStatusJSON(500, gin.H{"error": "not authorized to any organization"})
		return -1, errors.New("not authorized to any organization")
	}
	authorizedOrgID := authorizedOrgIdString.(int64)

	//compare authorized org id to org id in request
	requestedOrgID, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		// ctx.AbortWithError(500, err)
		return -1, errors.New("invalid org_id")
	}

	if requestedOrgID != authorizedOrgID {
		// ctx.AbortWithStatusJSON(401, gin.H{"error": "supplied API key is not authorized for organization"})
		return -1, errors.New("supplied API key is not authorized for this organization")
	}

	return authorizedOrgID, nil
}
