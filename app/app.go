package app

import (
	"insider-case/app/database"
	"insider-case/config"
	"log"

	"insider-case/app/routes"
	"net/http"

	"gorm.io/gorm"

	"github.com/gorilla/mux"
)

// App struct represents the application
type App struct {
	Database *gorm.DB
	Router   *mux.Router
}

// NewApp initializes a new App instance
func (a *App) Initialize(c *config.Config) *App {
	database.Connect(c)
	database.MigrateAll()
	db := database.GetDB()

	if db == nil {
		panic("Database connection is not initialized")
	}
	return &App{
		Database: db,
	}
}

func (a *App) Run() {

	Router := mux.NewRouter()
	routes.RegisterRoutes(Router)

	log.Println("Server starting at :8081...")
	if err := http.ListenAndServe(":8081", Router); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
