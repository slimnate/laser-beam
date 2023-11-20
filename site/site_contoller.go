package site

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/slimnate/laser-beam/data/event"
	"github.com/slimnate/laser-beam/data/organization"
	"github.com/slimnate/laser-beam/data/session"
	"github.com/slimnate/laser-beam/data/user"
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

	ctx.HTML(200, "index.html", gin.H{"User": user, "Organization": org, "Events": events})
}

func (s *SiteController) RenderLogin(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "login.html", nil)
}

func (s *SiteController) ProcessLogin(ctx *gin.Context) {
	username := ctx.PostForm("username")
	password := ctx.PostForm("password")

	log.Println("Username: " + username)
	log.Println("Password: " + password)

	user, err := s.userRepo.GetByUsername(username)
	if err != nil {
		log.Println("Invalid user")
		log.Println(err.Error())
		ctx.HTML(401, "login.html", gin.H{"Error": "Invalid username or password"})
		return
	}

	if user.Password != password {
		log.Println("invalid pass")
		ctx.HTML(401, "login.html", gin.H{"Error": "Invalid username or password"})
		return
	}

	session_key := randstr.String(64)
	session, err := s.sessionRepo.Create(session_key, user.ID)
	if err != nil {
		ctx.AbortWithStatusJSON(500, gin.H{"error": "unable to create user session"})
	}

	cookie := &http.Cookie{
		Name:   "session_key",
		Value:  session.Key,
		MaxAge: 0,
	}
	ctx.SetCookie(cookie.Name, cookie.Value, cookie.MaxAge, cookie.Path, cookie.Domain, cookie.Secure, cookie.HttpOnly)

	ctx.Redirect(302, "/")
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

	ctx.HTML(200, "user.html", gin.H{"User": user, "Organization": org})
}
