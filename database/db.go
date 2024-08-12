package database

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"techybat.org/go-vpn/models"
)

var (
	db     *gorm.DB
	dbOnce sync.Once
)

func Setup() {
	dbInstance := GetDB()
	MigrateAll(dbInstance)
}

func GetDB() *gorm.DB {
	dbOnce.Do(func() {
		// refer https://github.com/go-sql-driver/mysql#dsn-data-source-name for details
		user := os.Getenv("MYSQL_USER")
		pass := os.Getenv("MYSQL_PASS")
		host := os.Getenv("MYSQL_HOST")
		port := os.Getenv("MYSQL_PORT")
		dbname := os.Getenv("MYSQL_DB")
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", user, pass, host, port, dbname)

		dbInstance, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err != nil {
			log.Fatalln("Cannot connect to database. Err: ", err)
		}

		// Configure the database connection pool
		sqlDB, err := dbInstance.DB()
		if err != nil {
			log.Fatal("Failed to get sql.DB:", err)
		}

		sqlDB.SetMaxOpenConns(100)
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetConnMaxLifetime(time.Hour)

		db = dbInstance
	})

	return db
}

func MigrateAll(db *gorm.DB) {
	migrations := []models.Model{
		&models.Category{},
		&models.Pack{},
		&models.Card{},
		&models.User{},
		&models.Order{},
		&models.Receipt{},
	}

	for _, migration := range migrations {
		migration.Migrate(db)
	}
}
