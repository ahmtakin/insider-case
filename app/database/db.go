package database

import (
	"fmt"
	"insider-case/config"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect(c *config.Config) {

	dsn := "host=" + c.DB.Host +
		" user=" + c.DB.User +
		" password=" + c.DB.Password +
		" dbname=" + c.DB.Name +
		" port=" + c.DB.Port +
		" sslmode=" + c.DB.SSLMode

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to DB: ", err)
	} else {
		log.Println("Connected to DB successfully")
	}
	DB = db
}
func Close() {
	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatal("Failed to get DB instance: ", err)
	}
	if err := sqlDB.Close(); err != nil {
		log.Fatal("Failed to close DB connection: ", err)
	} else {
		log.Println("DB connection closed successfully")
	}
}

func GetDB() *gorm.DB {
	if DB == nil {
		log.Fatal("Database connection is not initialized")
	}
	return DB
}

func ExecuteSQLFile(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("error reading SQL file: %v", err)
	}

	if err := DB.Exec(string(content)).Error; err != nil {
		return fmt.Errorf("error executing SQL file: %v", err)
	}

	log.Println("SQL migrations executed successfully")
	return nil
}

func MigrateAll() {
	if err := ExecuteSQLFile("app/database/migrations/001_create_tables.sql"); err != nil {
		log.Fatal("Failed to execute SQL migrations: ", err)
	}
}
