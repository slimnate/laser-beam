package organization

import (
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
)

type OrganizationController struct {
	repo *OrganizationRepository
}

func NewOrganizationController(repo *OrganizationRepository) *OrganizationController {
	return &OrganizationController{
		repo: repo,
	}
}

func (c *OrganizationController) List(ctx *gin.Context) {
	authorizedGlobal, exists := ctx.Get("authorizedGlobal")

	if !exists || !authorizedGlobal.(bool) {
		ctx.AbortWithStatus(401)
		return
	}

	orgs, err := c.repo.All()
	if err != nil {
		ctx.AbortWithError(500, err)
		return
	}

	ctx.JSON(200, orgs)
}

func (c *OrganizationController) Details(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.AbortWithError(500, err)
		return
	}

	// validate

	org, err := c.repo.GetByID(id)
	log.Println("Org:" + org.Name)
	if err != nil {
		ctx.AbortWithStatusJSON(404, gin.H{"error": err})
		return
	}

	ctx.JSON(200, org)
}
