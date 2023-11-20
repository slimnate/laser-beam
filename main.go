package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"github.com/slimnate/laser-beam/data/event"
	"github.com/slimnate/laser-beam/data/organization"
	"github.com/slimnate/laser-beam/data/session"
	"github.com/slimnate/laser-beam/data/user"
	"github.com/slimnate/laser-beam/middleware"
	"github.com/slimnate/laser-beam/site"
)

const dbFile = "data.db"
const autoLogin = true
const autoLoginUser = "admin2"

func AuthMiddleware(sessionRepo *session.SQLiteRepository, userRepo *user.SQLiteRepository) gin.HandlerFunc {
	// if auto-login is enabled, we skip checking for any session keys
	// and approve the request as if the `autoLoginUser` is already logged in
	if autoLogin {
		return func(ctx *gin.Context) {
			user, err := userRepo.GetByUsername(autoLoginUser)
			if err != nil {
				ctx.AbortWithStatusJSON(500, gin.H{"error": "Error on auto-login, user not found"})
				return
			}

			ctx.Set("user", &user.User)
		}
	}

	return func(ctx *gin.Context) {
		sessionKey, err := ctx.Cookie("session_key")
		if err != nil {
			ctx.Redirect(302, "/login")
			return
		}

		session, err := sessionRepo.GetByKey(sessionKey)
		if err != nil {
			ctx.Redirect(302, "/login")
			return
		}

		user, err := userRepo.GetByID(session.UserID)
		if err != nil {
			ctx.AbortWithStatus(401)
			return
		}

		// set the userID and orgID on the query context
		ctx.Set("user", user)

		ctx.Next()
	}
}

// Middleware to check for a valid auth key, and add the corresponding org id to the request context
func ApiAuthMiddleware(orgRepo *organization.SQLiteRepository) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		key, exists := ctx.GetQuery("key")
		if !exists {
			ctx.AbortWithStatusJSON(401, gin.H{"error": "no api key supplied"})
			return
		}

		org, err := orgRepo.GetByKey(key)
		if err != nil {
			ctx.AbortWithStatusJSON(401, gin.H{"error": "invalid api key"})
			return
		}

		// Set flag for global authorization if org ID matches the global org
		if org.ID == 1 {
			ctx.Set("authorizedGlobal", true)
		}

		ctx.Set("apiKey", key)
		ctx.Set("authorizedOrgID", org.ID)

		ctx.Next()
	}
}

func InitDB() *sql.DB {
	// remove existing db
	os.Remove(dbFile)

	// open db
	db, err := sql.Open("sqlite3", dbFile+"?_fk=true")
	if err != nil {
		log.Fatal(err)
	}

	//return db
	return db
}

func InitOrganization(db *sql.DB) (*organization.OrganizationController, *organization.SQLiteRepository) {
	repo := organization.NewSQLiteRepository(db)
	controller := organization.NewOrganizationController(repo)
	// migrate
	if err := repo.Migrate(); err != nil {
		log.Fatal("Migration error: ", err)
	}

	//set up dummy data
	orgs := []organization.OrganizationSecret{
		{
			Organization: organization.Organization{
				Name: "Global Org",
			},
			Key: "secret1",
		},
		{
			Organization: organization.Organization{
				Name: "Organization 2",
			},
			Key: "secret2",
		},
		{
			Organization: organization.Organization{
				Name: "Organization 3",
			},
			Key: "secret3",
		},
	}
	for _, org := range orgs {
		createdOrg, err := repo.Create(org.Organization, org.Key)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Created org: %s with ID: %d \n", createdOrg.Name, createdOrg.ID)
	}

	return controller, repo
}

func InitEvent(db *sql.DB) (*event.EventController, *event.SQLiteRepository) {
	repo := event.NewSQLiteRepository(db)
	controller := event.NewEventController(repo)

	if err := repo.Migrate(); err != nil {
		log.Fatal("Migration error", err)
	}

	events := []event.Event{
		{
			Name:           "Error 1",
			Type:           "error",
			Message:        "Some error message",
			OrganizationID: 2,
		},
		{
			Name:           "Error 2",
			Type:           "error",
			Message:        "Some error message",
			OrganizationID: 2,
		},
		{
			Name:           "Info 1",
			Type:           "info",
			Message:        "Some info message",
			OrganizationID: 2,
		},
		{
			Name:           "Error 1",
			Type:           "error",
			Message:        "Some error message",
			OrganizationID: 3,
		},
		{
			Name:           "Error 2",
			Type:           "error",
			Message:        "Some error message",
			OrganizationID: 3,
		},
		{
			Name:           "Info 1",
			Type:           "info",
			Message:        "Some info message",
			OrganizationID: 3,
		},
	}

	for _, event := range events {
		created, err := repo.Create(event, event.OrganizationID)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Created event - id = %d | name = %s | type = %s | message = %s | organization_id = %d | time = %s \n", created.ID, created.Name, created.Type, created.Message, created.OrganizationID, created.Time.Format("20060102150405"))
	}

	return controller, repo
}

func InitUser(db *sql.DB) (*user.UserController, *user.SQLiteRepository) {
	repo := user.NewSQLiteRepository(db)
	controller := user.NewUserController(repo)

	if err := repo.Migrate(); err != nil {
		log.Fatal("Migration error", err)
	}

	users := []user.UserSecret{
		{
			User: user.User{
				Username:       "admin1",
				FirstName:      "Admin",
				LastName:       "Global",
				AdminStatus:    2,
				OrganizationID: 1,
			},
			Password: "password",
		},
		{
			User: user.User{
				Username:       "admin2",
				FirstName:      "Admin",
				LastName:       "One",
				AdminStatus:    1,
				OrganizationID: 2,
			},
			Password: "password",
		},
		{
			User: user.User{
				Username:       "user2",
				FirstName:      "User",
				LastName:       "One",
				AdminStatus:    0,
				OrganizationID: 2,
			},
			Password: "password",
		},
	}

	for _, user := range users {
		created, err := repo.Create(user)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Create user - id: %d | username: %s | first: %s | last: %s | admin: %d, org_id: %d \n", created.ID, created.Username, created.FirstName, created.LastName, created.AdminStatus, created.OrganizationID)
	}

	return controller, repo
}

func InitSession(db *sql.DB) *session.SQLiteRepository {
	repo := session.NewSQLiteRepository(db)

	if err := repo.Migrate(); err != nil {
		log.Fatal(err)
	}

	return repo
}

func main() {
	// init database and controllers
	db := InitDB()
	orgController, orgRepo := InitOrganization(db)
	eventController, eventRepo := InitEvent(db)
	userController, userRepo := InitUser(db)
	sessionRepo := InitSession(db)
	siteController := site.NewSiteController(orgRepo, eventRepo, userRepo, sessionRepo)

	// init router
	router := gin.Default()

	// Load templates and static files
	router.LoadHTMLGlob("templates/**/*.html")
	router.Static("/static", "./static")

	// Website routes
	authGroup := router.Group("")
	authGroup.Use(AuthMiddleware(sessionRepo, userRepo), middleware.HTMXMiddleware())
	{
		authGroup.GET("/", siteController.Index)
		authGroup.GET("/account", siteController.RenderAccount)
		authGroup.GET("/user", userController.RenderUser)
		authGroup.GET("/user/edit", userController.RenderUserForm)
		authGroup.PUT("/user", userController.UpdateUser)
	}

	router.GET("/login", siteController.RenderLogin)
	router.POST("/login", siteController.ProcessLogin)
	router.GET("/logout", siteController.Logout)

	// API routes
	apiAuthGroup := router.Group("/api")
	apiAuthGroup.Use(ApiAuthMiddleware(orgRepo))
	{
		// Global auth only routes
		apiAuthGroup.GET("/org", orgController.List)
		apiAuthGroup.GET("/events", eventController.ListGlobal)

		// org specific routes
		orgGroup := apiAuthGroup.Group("/org/:org_id")
		{
			orgGroup.GET("/", orgController.Details)

			// event specific routes
			eventGroup := orgGroup.Group("/events")
			{
				eventGroup.GET("/", eventController.List)
				eventGroup.GET("/:event_id", eventController.Details)
				eventGroup.POST("/", eventController.Create)
				eventGroup.PUT("/:event_id", eventController.Update)
			}
		}

		apiAuthGroup.GET("/users", userController.List)
	}

	router.Run(":8080")
}
