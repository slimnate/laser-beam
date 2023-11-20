package user

import (
	"errors"
	"log"

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

func (c *UserController) RenderUserForm(ctx *gin.Context) {
	user, err := GetUser(ctx)
	if err != nil {
		ctx.AbortWithStatus(500)
		return
	}

	ctx.HTML(200, "user_form.html", user)
}

func (c *UserController) UpdateUser(ctx *gin.Context) {
	newUsername := ctx.PostForm("username")
	newFirstName := ctx.PostForm("first_name")
	newLastName := ctx.PostForm("last_name")

	user, err := GetUser(ctx)
	if err != nil {
		ctx.AbortWithStatus(500)
		return
	}

	user.Username = newUsername
	user.FirstName = newFirstName
	user.LastName = newLastName

	newUser, err := c.repo.UpdateUserInfo(user.ID, *user)
	if err != nil {
		log.Println(err.Error())
		ctx.AbortWithStatus(500)
		return
	}

	ctx.HTML(200, "user_display.html", newUser)
}
