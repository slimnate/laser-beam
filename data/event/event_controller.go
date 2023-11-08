package event

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
)

type EventController struct {
	repo *SQLiteRepository
}

func NewEventController(repo *SQLiteRepository) *EventController {
	return &EventController{
		repo: repo,
	}
}

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

// Handler for /org/:id/events
func (c *EventController) List(ctx *gin.Context) {
	orgID, err := ValidateOrganizationID(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(401, gin.H{"error": err.Error()})
		return
	}

	// Request events list from repo
	orgs, err := c.repo.AllForOrganization(orgID)
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(200, orgs)
}

func (c *EventController) Details(ctx *gin.Context) {
	orgID, err := ValidateOrganizationID(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(401, gin.H{"error": err.Error()})
		return
	}

	id, err := strconv.ParseInt(ctx.Param("event_id"), 10, 64)
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}

	org, err := c.repo.GetByID(id, orgID)
	if err != nil {
		ctx.AbortWithStatusJSON(404, gin.H{"error": "event not found"})
		return
	}

	ctx.JSON(200, org)
}

func (c *EventController) Create(ctx *gin.Context) {
	orgID, err := ValidateOrganizationID(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(401, gin.H{"error": err.Error()})
		return
	}

	var e Event
	if err := ctx.ShouldBindJSON(&e); err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}

	created, err := c.repo.Create(e, orgID)
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(200, created)
}
