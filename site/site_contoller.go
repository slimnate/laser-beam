package site

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/slimnate/laser-beam/crypto"
	"github.com/slimnate/laser-beam/data/event"
	"github.com/slimnate/laser-beam/data/organization"
	"github.com/slimnate/laser-beam/data/session"
	"github.com/slimnate/laser-beam/data/user"
	"github.com/slimnate/laser-beam/middleware"
	"github.com/slimnate/laser-beam/validation"
	"github.com/thanhpk/randstr"
)

type SiteController struct {
	orgRepo     *organization.SQLiteRepository
	eventRepo   *event.SQLiteRepository
	userRepo    *user.SQLiteRepository
	sessionRepo *session.SQLiteRepository
}

func NewSiteController(orgRepo *organization.SQLiteRepository, eventRepo *event.SQLiteRepository, userRepo *user.SQLiteRepository, sessionRepo *session.SQLiteRepository) *SiteController {
	return &SiteController{
		orgRepo:     orgRepo,
		eventRepo:   eventRepo,
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
	}
}

func (s *SiteController) GetUserOrg(ctx *gin.Context) (u *user.User, o *organization.Organization, err error) {
	userAny, exists := ctx.Get("user")
	if !exists {
		return nil, nil, errors.New("no user found on request")
	}
	user := userAny.(*user.User)

	org, err := s.orgRepo.GetByID(user.OrganizationID)
	if err != nil {
		return nil, nil, errors.New("unable to get org")
	}

	return user, org, nil
}

func (s *SiteController) Index(ctx *gin.Context) {
	user, org, err := s.GetUserOrg(ctx)
	if err != nil {
		ctx.AbortWithStatus(500)
		return
	}

	events, err := s.eventRepo.AllForOrganization(org.ID)
	if err != nil {
		ctx.AbortWithStatus(500)
		return
	}

	hx := middleware.GetHxHeaders(ctx)

	if hx.Request {
		ctx.HTML(200, "dashboard.html", gin.H{"User": user, "Organization": org, "Events": events})
	} else {
		ctx.HTML(200, "index.html", gin.H{"User": user, "Organization": org, "Events": events, "Route": "/"})
	}
}

func (s *SiteController) RenderLogin(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "login.html", nil)
}

func (s *SiteController) ProcessLogin(ctx *gin.Context) {
	hx := middleware.GetHxHeaders(ctx)

	username := ctx.PostForm("username")
	password := ctx.PostForm("password")

	user, err := s.userRepo.GetByUsername(username)
	if err != nil {
		log.Println("Invalid user")
		log.Println(err.Error())
		if hx.Request {
			ctx.HTML(200, "login_form.html", gin.H{"Error": "Invalid username or password", "Username": username})
		} else {
			ctx.HTML(401, "login.html", gin.H{"Error": "Invalid username or password", "Username": username})
		}
		return
	}

	if !crypto.TestMatch(password, user.Password) {
		log.Println("invalid pass")
		if hx.Request {
			ctx.HTML(200, "login_form.html", gin.H{"Error": "Invalid username or password", "Username": username})
		} else {
			ctx.HTML(401, "login.html", gin.H{"Error": "Invalid username or password", "Username": username})
		}
		return
	}

	session_key := randstr.String(64)
	session, err := s.sessionRepo.Create(session_key, user.ID)
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"Error": "unable to create user session"})
		return
	}

	cookie := &http.Cookie{
		Name:   "session_key",
		Value:  session.Key,
		MaxAge: 0,
	}
	ctx.SetCookie(cookie.Name, cookie.Value, cookie.MaxAge, cookie.Path, cookie.Domain, cookie.Secure, cookie.HttpOnly)

	if hx.Request {
		ctx.Header("HX-Redirect", "/")
		ctx.AbortWithStatus(200)
	} else {
		ctx.Redirect(302, "/")
	}
}

func (s *SiteController) Logout(ctx *gin.Context) {
	sessionCookie, err := ctx.Request.Cookie("session_key")
	if err != nil {
		log.Println("No session cookie found when trying to log out")
		ctx.Redirect(302, "/")
		return
	}

	if err := s.sessionRepo.DeleteByKey(sessionCookie.Value); err != nil {
		log.Println("Unable to delete session entry from db: " + err.Error())
		ctx.Redirect(302, "/")
		return
	}

	ctx.SetCookie(sessionCookie.Name, "", -1, sessionCookie.Path, sessionCookie.Domain, sessionCookie.Secure, sessionCookie.HttpOnly)
	ctx.Redirect(302, "/")
}

func (s *SiteController) RenderAccount(ctx *gin.Context) {
	user, org, err := s.GetUserOrg(ctx)
	if err != nil {
		ctx.AbortWithStatus(500)
		return
	}

	hx := middleware.GetHxHeaders(ctx)
	if hx.Request {
		ctx.HTML(200, "user_display.html", user)
	} else {
		ctx.HTML(200, "index.html", gin.H{"User": user, "Organization": org, "Route": "/account"})
	}
}

func (s *SiteController) RenderUserForm(ctx *gin.Context) {
	user, org, err := s.GetUserOrg(ctx)
	if err != nil {
		ctx.AbortWithStatus(500)
		return
	}

	hx := middleware.GetHxHeaders(ctx)
	if hx.Request {
		ctx.HTML(200, "user_form.html", user)
	} else {
		ctx.HTML(200, "index.html", gin.H{"User": user, "Organization": org, "Route": "/account/edit"})
	}
}

func (s *SiteController) UpdateUser(ctx *gin.Context) {
	hx := middleware.GetHxHeaders(ctx)
	newFirstName := ctx.PostForm("first_name")
	newLastName := ctx.PostForm("last_name")

	user, org, err := s.GetUserOrg(ctx)
	if err != nil {
		ctx.AbortWithStatus(500)
		return
	}

	valid, e := validation.ValidateUserUpdate(newFirstName, newLastName)
	if !valid {
		if hx.Request {
			ctx.HTML(200, "user_form.html", gin.H{"User": user, "Errors": e})
		} else {
			ctx.HTML(200, "index.html", gin.H{"User": user, "Organization": org, "Route": "/account/edit", "Errors": e})
		}
		return
	}

	user.FirstName = newFirstName
	user.LastName = newLastName

	newUser, err := s.userRepo.UpdateUserInfo(user.ID, *user)
	if err != nil {
		log.Println(err.Error())
		ctx.AbortWithStatus(500)
		return
	}

	if hx.Request {
		ctx.HTML(200, "user_display.html", newUser)
	} else {
		ctx.HTML(200, "index.html", gin.H{"User": newUser, "Organization": org, "Route": "/account"})
	}
}

func (s *SiteController) RenderPasswordForm(ctx *gin.Context) {
	user, org, err := s.GetUserOrg(ctx)
	if err != nil {
		ctx.AbortWithStatus(500)
		return
	}

	hx := middleware.GetHxHeaders(ctx)
	if hx.Request {
		ctx.HTML(200, "user_password.html", user)
	} else {
		ctx.HTML(200, "index.html", gin.H{"User": user, "Organization": org, "Route": "/account/password"})
	}
}

func (s *SiteController) UpdatePassword(ctx *gin.Context) {
	hx := middleware.GetHxHeaders(ctx)
	newPassword := ctx.PostForm("password")
	confirmPassword := ctx.PostForm("confirm_password")

	u, org, err := s.GetUserOrg(ctx)
	if err != nil {
		ctx.AbortWithStatus(500)
		return
	}

	valid, e := validation.ValidatePasswordUpdate(newPassword, confirmPassword)
	if !valid {
		if hx.Request {
			ctx.HTML(200, "user_password.html", gin.H{"User": u, "Errors": e})
		} else {
			ctx.HTML(200, "index.html", gin.H{"User": u, "Organization": org, "Route": "/account/password", "Errors": e})
		}
		return
	}

	p, err := crypto.HashPassword(newPassword)
	if err != nil {
		ctx.AbortWithStatus(500)
		return
	}

	userSecret := &user.UserSecret{
		User:     *u,
		Password: p,
	}

	newUser, err := s.userRepo.UpdateLoginInfo(u.ID, *userSecret)
	if err != nil {
		log.Println(err.Error())
		ctx.AbortWithStatus(500)
		return
	}

	if hx.Request {
		ctx.HTML(200, "user_display.html", newUser)
	} else {
		ctx.HTML(200, "index.html", gin.H{"User": newUser, "Organization": org, "Route": "/account"})
	}
}
