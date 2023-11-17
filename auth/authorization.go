package auth

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Get the organization ID path param, while validating that the organization associated with the supplied API key is either authorized to access data for the requested organization, or is globally authorized. Returns -1 for `id` if the requested org_id is not authorized
func GetAndAuthorizeOrgIDParam(ctx *gin.Context) (id int64, err error) {
	//Get requested org_id
	requestedOrgID, err := strconv.ParseInt(ctx.Param("org_id"), 10, 64)
	if err != nil {
		// ctx.AbortWithError(500, err)
		return -1, errors.New("invalid org_id")
	}

	if !IsAuthorizedForOrgID(ctx, requestedOrgID) {
		return -1, errors.New("not authorized for requested org_id")
	}

	return requestedOrgID, nil
}

func IsAuthorizedForOrgID(ctx *gin.Context, id int64) bool {
	// return false if no "authorizedOrgID" is set on the request context
	authorizedOrgID, exists := ctx.Get("authorizedOrgID")
	if !exists {
		return false
	}

	// return true if the supplied `id` matches or request is global authorized
	if authorizedOrgID.(int64) == id || IsAuthorizedForGlobal(ctx) {
		return true
	}

	return false
}

func IsAuthorizedForGlobal(ctx *gin.Context) bool {
	authorizedGlobal, exists := ctx.Get("authorizedGlobal")
	if exists && authorizedGlobal.(bool) {
		return true
	}
	return false
}
