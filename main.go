package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
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

func InitDB() *sql.DB {
	// remove existing db
	os.Remove(dbFile)

	// open db
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		log.Fatal(err)
	}

	//return db
	return db
}

func InitOrganization(db *sql.DB) *organization.OrganizationController {
	repo := organization.NewSQLiteRepository(db)
	controller := organization.NewOrganizationController(repo)
	// migrate
	if err := repo.Migrate(); err != nil {
		log.Fatal("Migration error: ", err)
	}

	//set up dummy data
	orgs := []organization.Organization{
		{
			Name: "Organization 1",
			Key:  "secret1",
		},
		{
			Name: "Organization 2",
			Key:  "secret2",
		},
	}
	for _, org := range orgs {
		createdOrg, err := repo.Create(org)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Created org: %s with ID: %d \n", createdOrg.Name, createdOrg.ID)
	}

	return controller
}

func main() {
	// init database and controllers
	db := InitDB()
	orgController := InitOrganization(db)

	// init router
	router := gin.Default()

	router.GET("/", func(ctx *gin.Context) {
		ctx.String(200, "Hello!")
	})

	router.GET("/org", orgController.List)

	router.GET("org/:id", orgController.Details)

	authGroup := router.Group("/api")
	authGroup.Use(AuthMiddleware())
	{
		authGroup.GET("/data", func(ctx *gin.Context) {
			key, _ := ctx.Get("apiKey")
			ctx.JSON(200, gin.H{"message": "Authenticated", "key": key})
		})
	}

	router.Run(":8080")
}
