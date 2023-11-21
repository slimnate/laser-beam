package user

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/slimnate/laser-beam/auth"
)

type UserController struct {
	repo *SQLiteRepository
}

func NewUserController(repo *SQLiteRepository) *UserController {
	return &UserController{
		repo: repo,
	}
}

func GetUser(ctx *gin.Context) (*User, error) {
	userAny, exists := ctx.Get("user")
	if !exists {
		return nil, errors.New("no user found on request")
	}
	user := userAny.(*User)

	return user, nil
}

func (c *UserController) List(ctx *gin.Context) {
	orgID, err := auth.GetAndAuthorizeOrgIDParam(ctx)
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

func (c *UserController) RenderUser(ctx *gin.Context) {
	user, err := GetUser(ctx)
	if err != nil {
		ctx.AbortWithStatus(500)
		return
	}

	ctx.HTML(200, "user_display.html", user)
}
