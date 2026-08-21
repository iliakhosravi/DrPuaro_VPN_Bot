package database

import (
	"fmt"
	"log"
	"sync"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"techybat.org/go-vpn/models"
	"techybat.org/go-vpn/vars"
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
		// refer https://pkg.go.dev/gorm.io/driver/postgres for DSN details
		user := vars.Get("POSTGRES_USER")
		pass := vars.Get("POSTGRES_PASS")
		host := vars.Get("POSTGRES_HOST")
		port := vars.Get("POSTGRES_PORT")
		dbname := vars.Get("POSTGRES_DB")
		sslmode := vars.Get("POSTGRES_SSLMODE")
		if sslmode == "" {
			sslmode = "disable"
		}
		dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=UTC",
			host, port, user, pass, dbname, sslmode)

		dbInstance, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
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

		// db.Logger = logger.Default.LogMode(logger.Info)

		// db.Debug()
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
		&models.Config{},
		&models.Guide{},
		&models.ChargeOrder{},
		&models.ChargeReceipt{},
		&models.InlineKeyboard{},
	}

	for _, migration := range migrations {
		migration.Migrate(db)
	}
}
