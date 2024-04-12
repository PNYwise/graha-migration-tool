package internal

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Dbinstance struct {
	Db *gorm.DB
}

var DB Dbinstance

// connectDb
func ConnectDb() {

	conf := New()
	var (
		dbHost     = conf.GetString("database.host")
		dbPort     = conf.GetInt("database.port")
		dbName     = conf.GetString("database.name")
		dbUsername = conf.GetString("database.username")
		dbPassword = conf.GetString("database.password")
	)

	dsn := fmt.Sprintf(
		"host=%s user=%s password='%s' dbname=%s port=%d sslmode=disable TimeZone=UTC",
		dbHost, dbUsername, dbPassword, dbName, dbPort,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	})

	if err != nil {
		log.Fatal("Failed to connect to database. \n", err)
	}

	log.Println("connected")
	db.Logger = logger.Default.LogMode(logger.Info)

	DB = Dbinstance{
		Db: db,
	}
}

func CloseDb() error {
	db, err := DB.Db.DB()
	if err != nil {
		return err
	}
	return db.Close()
}

func Ping() error {
	db, err := DB.Db.DB()
	if err != nil {
		return err
	}
	log.Println("pong")
	return db.Ping()
}
