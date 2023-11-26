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

	codes := []int{
		1001,
		1002,
		1003,
		1004,
		1005,
	}

	messages := []string{
		"Error 1001: Database connection failed. Please check your database credentials.",
		"Error 1002: Database query failed. Please check your SQL syntax.",
		"Error 1003: Database write failed. Please check your database permissions.",
		"Error 1004: File upload failed. Please check your file size and format.",
		"Error 1005: Email delivery failed. Please check your email server settings.",
	}

	// loop over org ids - start at org 2, since 1 is the global org and doesn't need events
	for orgID := 2; orgID <= 3; orgID++ {
		// loop over events
		for eventNum := 1; eventNum <= 15; eventNum++ {
			var (
				name    string
				eType   string
				message string
				app     string
			)

			// Half of events should be error, other half info
			typeCode := eventNum % 2
			// Share between three different app names
			appCode := eventNum % 3
			messageCode := eventNum % 5

			if typeCode == 0 {
				eType = "error"
				name = fmt.Sprintf("Error %d", codes[messageCode])
				message = messages[messageCode]
			} else {
				eType = "info"
				name = fmt.Sprintf("Info %d", codes[messageCode])
				message = messages[messageCode]
			}

			if appCode == 0 {
				app = "TechNexus"
			} else if appCode == 1 {
				app = "InnovateX"
			} else {
				app = "CodeWave"
			}

			e := event.Event{
				Name:           name,
				Application:    app,
				Type:           eType,
				Message:        message,
				OrganizationID: int64(orgID),
			}

			created, err := repo.Create(e, e.OrganizationID)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("Created event - id = %d | type = %s | app = %s | name = %s |  message = %s | organization_id = %d | time = %s \n", created.ID, created.Type, created.Application, created.Name, created.Message, created.OrganizationID, created.Time.Format("20060102150405"))
		}
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
				Email:          "admin1@globalorg.com",
				Phone:          "1234567890",
				AdminStatus:    2,
				OrganizationID: 1,
			},
			Password: "$2a$15$dRgGBE56DiFg/I2sarfnKOYk6GMHSo/A5U38OIDpjKeePBGlLFqKe",
			// Password: "password",
		},
		{
			User: user.User{
				Username:       "admin2",
				FirstName:      "Admin",
				LastName:       "OrgTwo",
				Email:          "admin2@org2.com",
				Phone:          "1234567890",
				AdminStatus:    1,
				OrganizationID: 2,
			},
			Password: "$2a$15$221/N0pnu5epRsGzs39JCucTXzNMYh22YHFu5oIW36lJ3bYKghz3K",
			// Password: "password",
		},
		{
			User: user.User{
				Username:       "user2",
				FirstName:      "User",
				LastName:       "OrgTwo",
				Email:          "user2@org2.com",
				Phone:          "1234567890",
				AdminStatus:    0,
				OrganizationID: 2,
			},
			Password: "$2a$15$TIeBxsBMN94IxawycrT4Ce1HcomMwBoJHt3wsEX5rE56XCV3slN7e",
			// Password: "password",
		},
	}

	for _, user := range users {
		//Hash user password before storing
		// hashed, err := crypto.HashPassword(user.Password)
		// if err != nil {
		// 	log.Fatal("Unable to hash password: " + err.Error())
		// }
		// user.Password = hashed

		created, err := repo.Create(user)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Create user - id: %d | u: %s | name: %s | e: %s | p: %s | admin: %d, org_id: %d \n", created.ID, created.Username, created.FullName(), created.Email, created.Phone, created.AdminStatus, created.OrganizationID)
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
	authGroup.Use(middleware.AuthMiddleware(sessionRepo, userRepo), middleware.HTMXMiddleware())
	{
		authGroup.GET("/", siteController.Index)
		authGroup.GET("/account", siteController.RenderAccount)
		authGroup.PUT("/account", siteController.UpdateUser)
		authGroup.POST("/account", siteController.UpdateUser)
		authGroup.GET("/account/edit", siteController.RenderUserForm)
		authGroup.GET("/account/password", siteController.RenderPasswordForm)
		authGroup.PUT("/account/password", siteController.UpdatePassword)
		authGroup.POST("/account/password", siteController.UpdatePassword)
		authGroup.GET("/events", siteController.RenderEvents)
	}

	router.GET("/login", siteController.RenderLogin)
	router.POST("/login", siteController.ProcessLogin)
	router.GET("/logout", siteController.Logout)

	// API routes
	apiAuthGroup := router.Group("/api")
	apiAuthGroup.Use(middleware.ApiAuthMiddleware(orgRepo))
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
