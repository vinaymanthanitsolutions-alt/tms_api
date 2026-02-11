package config

import (
	"backend/internal/utils"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// Global database instance
var DB *sql.DB

// InitDB initializes the database connection once
func InitDB() error {
	db, err := connectDB()
	if err != nil {
		return err
	}
	DB = db
	return nil
}

func connectDB() (*sql.DB, error) {
	var dbUsername string = utils.GetEnv("DB_USERNAME")
	var dbPassword string = utils.GetEnv("DB_PASSWORD")
	var dbHost string = utils.GetEnv("DB_HOST")
	var dbPort string = utils.GetEnv("DB_PORT")
	var dbName string = utils.GetEnv("DB_NAME")

	// dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbUsername, dbPassword, dbHost, dbPort, dbName)

	// db, err := sql.Open("mysql", dsn)
	// if err != nil {
	// 	return nil, err
	// }

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&loc=Asia%%2FKolkata",
		dbUsername, dbPassword, dbHost, dbPort, dbName)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	// Connection pooling parameters
	maxIdleConns, _ := strconv.Atoi(utils.GetEnv("DB_MAX_IDLE_CONNS"))
	maxOpenConns, _ := strconv.Atoi(utils.GetEnv("DB_MAX_OPEN_CONNS"))
	connMaxLifetime, _ := strconv.Atoi(utils.GetEnv("DB_CONN_MAX_LIFETIME")) // in seconds

	db.SetMaxIdleConns(maxIdleConns)
	db.SetMaxOpenConns(maxOpenConns)
	db.SetConnMaxLifetime(time.Duration(connMaxLifetime) * time.Second)

	if err := db.Ping(); err != nil {
		db.Close() // Clean up the connection if some err has occurred
		return nil, err
	}

	fmt.Println("Connected to the database successfully!")
	return db, nil
}

// GetDB returns the database instance (optional getter)
func GetDB() *sql.DB {
	return DB
}

// CloseDB closes the database connection
func CloseDB() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}
