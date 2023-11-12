package user

import (
	"github.com/gin-gonic/gin"
	"github.com/slimnate/laser-beam/data"
)

type UserController struct {
	repo *SQLiteRepository
}

func NewUserController(repo *SQLiteRepository) *UserController {
	return &UserController{
		repo: repo,
	}
}

func (c *UserController) List(ctx *gin.Context) {
	orgID, err := data.ValidateOrganizationID(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(401, gin.H{"error": err.Error()})
		return
	}

	users, err := c.repo.AllForOrganization(orgID)
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(200, users)
}
