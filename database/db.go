package database

import (
	"fmt"
	"log"
	"sync"
	"time"

	"gorm.io/driver/mysql"
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
		// refer https://github.com/go-sql-driver/mysql#dsn-data-source-name for details
		user := vars.Get("MYSQL_USER")
		pass := vars.Get("MYSQL_PASS")
		host := vars.Get("MYSQL_HOST")
		port := vars.Get("MYSQL_PORT")
		dbname := vars.Get("MYSQL_DB")
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
