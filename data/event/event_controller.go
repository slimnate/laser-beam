package event

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/slimnate/laser-beam/data"
)

type EventController struct {
	repo *SQLiteRepository
}

func NewEventController(repo *SQLiteRepository) *EventController {
	return &EventController{
		repo: repo,
	}
}

// Handler for /org/:id/events
func (c *EventController) List(ctx *gin.Context) {
	orgID, err := data.ValidateOrganizationID(ctx)
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
	orgID, err := data.ValidateOrganizationID(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(401, gin.H{"error": err.Error()})
		return
	}

	id, err := strconv.ParseInt(ctx.Param("event_id"), 10, 64)
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}

	org, err := c.repo.GetByIDAndOrg(id, orgID)
	if err != nil {
		ctx.AbortWithStatusJSON(404, gin.H{"error": "event not found"})
		return
	}

	ctx.JSON(200, org)
}

func (c *EventController) Create(ctx *gin.Context) {
	orgID, err := data.ValidateOrganizationID(ctx)
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

func (c *EventController) Update(ctx *gin.Context) {
	// validate org ID, but we don't need it for the request
	_, err := data.ValidateOrganizationID(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(401, gin.H{"error": err.Error()})
		return
	}

	// get event id from the query params
	id, err := strconv.ParseInt(ctx.Param("event_id"), 10, 64)
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}

	// read updated event from body
	var e Event
	if err := ctx.ShouldBindJSON(&e); err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}

	updated, err := c.repo.Update(id, e)
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(200, updated)
}
