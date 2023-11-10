package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"github.com/slimnate/laser-beam/data/event"
	"github.com/slimnate/laser-beam/data/organization"
)

const dbFile = "data.db"

func AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		isAuthorized := false
		apiKey := ""

		// check auth header
		apiKeyHeader := ctx.GetHeader("X-API-Key")
		if apiKeyHeader == "valid" {
			apiKey = apiKeyHeader
			isAuthorized = true
		}

		// check query params
		apiKeyQuery, exists := ctx.GetQuery("key")
		if exists && apiKeyQuery == "valid" {
			apiKey = apiKeyQuery
			isAuthorized = true
		}

		if !isAuthorized {
			ctx.AbortWithStatusJSON(401, gin.H{"error": "unauthorized"})
			return
		}

		// set the api key in the query context
		ctx.Set("apiKey", apiKey)

		ctx.Next()
	}
}

// Middleware to check for a valid auth key, and add the corresponding org id to the request context
func OrgAuthMiddleware(orgRepo *organization.SQLiteRepository) gin.HandlerFunc {
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

		ctx.Set("apiKey", key)
		ctx.Set("authorizedOrgID", org.ID)
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
				Name: "Organization 1",
			},
			Key: "secret1",
		},
		{
			Organization: organization.Organization{
				Name: "Organization 2",
			},
			Key: "secret2",
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
			OrganizationID: 1,
		},
		{
			Name:           "Error 2",
			Type:           "error",
			Message:        "Some error message",
			OrganizationID: 1,
		},
		{
			Name:           "Info 1",
			Type:           "info",
			Message:        "Some info message",
			OrganizationID: 1,
		},
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
	}

	for _, event := range events {
		created, err := repo.Create(event, event.OrganizationID)
		if err != nil {
			log.Println("asdf")
			log.Fatal(err)
		}
		fmt.Printf("Created event - id = %d | name = %s | type = %s | message = %s | organization_id = %d | time = %s \n", created.ID, created.Name, created.Type, created.Message, created.OrganizationID, created.Time.Format("20060102150405"))
	}

	return controller, repo
}

func main() {
	// init database and controllers
	db := InitDB()
	orgController, orgRepo := InitOrganization(db)
	eventController, _ := InitEvent(db)

	// init router
	router := gin.Default()

	router.LoadHTMLGlob("templates/**/*.html")
	router.Static("/static", "./static")

	router.GET("/", func(ctx *gin.Context) {
		orgIDCookie, err := ctx.Cookie("organization_id")
		if err != nil {
			ctx.Redirect(http.StatusFound, "/login")
			return
		}
		orgID, err := strconv.ParseInt(orgIDCookie, 10, 64)
		if err != nil {
			ctx.Redirect(http.StatusFound, "/login")
			return
		}

		ctx.HTML(200, "index.html", gin.H{"org_id": orgID})
	})

	router.GET("/login", func(ctx *gin.Context) {
		ctx.HTML(http.StatusFound, "login.html", nil)
	})

	router.POST("/login", func(ctx *gin.Context) {
		org_id, err := strconv.ParseInt(ctx.PostForm("organization_id"), 10, 64)
		if err != nil {
			ctx.AbortWithStatusJSON(500, gin.H{"error": "invalid organization_id"})
		}
		log.Printf("Incoming login request: %d", org_id)
		cookie := &http.Cookie{
			Name:   "organization_id",
			Value:  strconv.FormatInt(org_id, 10),
			MaxAge: 0,
		}
		ctx.SetCookie(cookie.Name, cookie.Value, cookie.MaxAge, cookie.Path, cookie.Domain, cookie.Secure, cookie.HttpOnly)
		ctx.Redirect(302, "/")
	})

	router.GET("/logout", func(ctx *gin.Context) {
		orgIDCookie, err := ctx.Request.Cookie("organization_id")
		if err != nil {
			ctx.Redirect(302, "/")
		}

		ctx.SetCookie(orgIDCookie.Name, orgIDCookie.Value, -1, orgIDCookie.Path, orgIDCookie.Domain, orgIDCookie.Secure, orgIDCookie.HttpOnly)
		ctx.Redirect(302, "/")
	})

	router.GET("/org", orgController.List)

	authGroup := router.Group("/api")
	authGroup.Use(AuthMiddleware())
	{
		authGroup.GET("/data", func(ctx *gin.Context) {
			key, _ := ctx.Get("apiKey")
			ctx.JSON(200, gin.H{"message": "Authenticated", "key": key})
		})
	}

	orgAuthGroup := router.Group("/org/:id")
	orgAuthGroup.Use(OrgAuthMiddleware(orgRepo))
	{
		orgAuthGroup.GET("/", orgController.Details)
		orgAuthGroup.GET("/events", eventController.List)
		orgAuthGroup.GET("/events/:event_id", eventController.Details)
		orgAuthGroup.POST("/events", eventController.Create)
		orgAuthGroup.PUT("/events/:event_id", eventController.Update)
	}

	router.Run(":8080")
}
