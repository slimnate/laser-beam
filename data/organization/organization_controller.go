package organization

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

type OrganizationController struct {
	repo *SQLiteRepository
}

func NewOrganizationController(repo *SQLiteRepository) *OrganizationController {
	return &OrganizationController{
		repo: repo,
	}
}

func (c *OrganizationController) List(ctx *gin.Context) {
	orgs, err := c.repo.All()
	if err != nil {
		ctx.AbortWithError(500, err)
	}

	ctx.JSON(200, orgs)
}

func (c *OrganizationController) Details(ctx *gin.Context) {
	id, err := strconv.ParseInt(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.AbortWithError(500, err)
	}
	org, err := c.repo.GetByID(id)
	if err != nil {
		ctx.AbortWithStatusJSON(404, gin.H{"error": err})
		return
	}

	ctx.JSON(200, org)
}
