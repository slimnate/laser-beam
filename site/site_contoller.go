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

// Send a different response depending whether we are responding to an HTMX request or not
func HxRespond(status int, ctx *gin.Context, htmxTemplate string, defaultTemplate string, data any) {
	hx := middleware.GetHxHeaders(ctx)
	if hx.Request {
		ctx.HTML(200, htmxTemplate, data)
	} else {
		ctx.HTML(status, defaultTemplate, data)
	}
}

// Redirect the page, using different methods for HTMX and non-htmx requests
func HxRedirect(ctx *gin.Context, path string) {
	hx := middleware.GetHxHeaders(ctx)
	if hx.Request {
		ctx.Header("HX-Redirect", path)
		ctx.AbortWithStatus(200)
	} else {
		ctx.Redirect(302, path)
	}
}

// Extract the user and organization data from the request context
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

// GET /
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

	data := PageData{
		User:         user,
		Organization: org,
		Events:       events,
		Route:        "/",
	}

	HxRespond(200, ctx, "dashboard.html", "index.html", data)
}

// GET /login
func (s *SiteController) RenderLogin(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "login.html", nil)
}

// POST /login
func (s *SiteController) ProcessLogin(ctx *gin.Context) {
	username := ctx.PostForm("username")
	password := ctx.PostForm("password")

	data := gin.H{"Error": "Invalid username or password", "Username": username}

	user, err := s.userRepo.GetByUsername(username)
	if err != nil {
		log.Println("Invalid user")
		log.Println(err.Error())
		HxRespond(401, ctx, "login_form.html", "index.html", data)
		return
	}

	if !crypto.TestMatch(password, user.Password) {
		log.Println("invalid pass")
		HxRespond(401, ctx, "login_form.html", "index.html", data)
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

	HxRedirect(ctx, "/")
}

// GET /logout
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

// GET /account
func (s *SiteController) RenderAccount(ctx *gin.Context) {
	user, org, err := s.GetUserOrg(ctx)
	if err != nil {
		ctx.AbortWithStatus(500)
		return
	}

	data := PageData{
		User:         user,
		Organization: org,
		Route:        "/account",
	}

	HxRespond(200, ctx, "user_display.html", "index.html", data)
}

// GET /account/edit
func (s *SiteController) RenderUserForm(ctx *gin.Context) {
	user, org, err := s.GetUserOrg(ctx)
	if err != nil {
		ctx.AbortWithStatus(500)
		return
	}

	data := PageData{
		User:         user,
		Organization: org,
		Route:        "/account/edit",
	}

	HxRespond(200, ctx, "user_form.html", "index.html", data)
}

// POST /account/edit
func (s *SiteController) UpdateUser(ctx *gin.Context) {
	newFirstName := ctx.PostForm("first_name")
	newLastName := ctx.PostForm("last_name")
	newEmail := ctx.PostForm("email")
	newPhone := ctx.PostForm("phone")

	user, org, err := s.GetUserOrg(ctx)
	if err != nil {
		ctx.AbortWithStatus(500)
		return
	}

	user.FirstName = newFirstName
	user.LastName = newLastName
	user.Email = newEmail
	user.Phone = newPhone

	data := PageData{
		User:         user,
		Organization: org,
		Route:        "/account/edit",
	}

	valid, e := validation.ValidateUserUpdate(user)
	if !valid {
		data.Errors = e
		HxRespond(200, ctx, "user_form.html", "index.html", data)
		return
	}

	newUser, err := s.userRepo.UpdateUserInfo(user.ID, *user)
	if err != nil {
		log.Println(err.Error())
		ctx.AbortWithStatus(500)
		return
	}

	data.User = newUser
	data.Route = "/account"
	data.AddToast("Successfully updated user account!")

	HxRespond(200, ctx, "user_display.html", "index.html", data)
}

// GET /account/password
func (s *SiteController) RenderPasswordForm(ctx *gin.Context) {
	user, org, err := s.GetUserOrg(ctx)
	if err != nil {
		ctx.AbortWithStatus(500)
		return
	}

	data := PageData{
		User:         user,
		Organization: org,
		Route:        "/account/password",
	}

	HxRespond(200, ctx, "user_password.html", "index.html", data)
}

// POST /account/password
func (s *SiteController) UpdatePassword(ctx *gin.Context) {
	newPassword := ctx.PostForm("password")
	confirmPassword := ctx.PostForm("confirm_password")

	u, org, err := s.GetUserOrg(ctx)
	if err != nil {
		ctx.AbortWithStatus(500)
		return
	}

	data := PageData{
		User:         u,
		Organization: org,
		Route:        "/account/password",
	}

	valid, e := validation.ValidatePasswordUpdate(newPassword, confirmPassword)
	if !valid {
		data.Errors = e
		HxRespond(200, ctx, "user_password.html", "index.html", data)
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

	data.User = newUser
	data.Route = "/account"
	data.AddToast("Successfully updated password!")

	HxRespond(200, ctx, "user_display.html", "index.html", data)
}

// GET /events
func (s *SiteController) RenderEvents(ctx *gin.Context) {
	u, o, err := s.GetUserOrg(ctx)
	if err != nil {
		log.Println(err.Error())
		ctx.AbortWithStatus(500)
		return
	}

	e, err := s.eventRepo.AllForOrganization(o.ID)
	if err != nil {
		log.Println(err.Error())
		ctx.AbortWithStatus(500)
		return
	}

	data := PageData{
		User:         u,
		Organization: o,
		Events:       e,
		Route:        "/events",
	}

	HxRespond(200, ctx, "events.html", "index.html", data)
}
