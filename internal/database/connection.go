package database

import (
	"crypto/tls"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

// DB holds the database connection
var DB *sql.DB

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Username string
	Password string
	Host     string
	Port     string
	Database string
}

// LoadDatabaseConfig loads database configuration from environment variables
func LoadDatabaseConfig() (*DatabaseConfig, error) {
	godotenv.Load()

	config := &DatabaseConfig{
		Username: os.Getenv("TIDB_USERNAME"),
		Password: os.Getenv("TIDB_PASSWORD"),
		Host:     os.Getenv("TIDB_HOST"),
		Port:     os.Getenv("TIDB_PORT"),
		Database: os.Getenv("TIDB_DATABASE"),
	}

	if config.Port == "" {
		config.Port = "4000"
	}
	if config.Database == "" {
		config.Database = "test"
	}

	if config.Username == "" || config.Password == "" || config.Host == "" {
		return nil, fmt.Errorf("missing required environment variables: username, password, and host must be set")
	}

	return config, nil
}

// ConnectDatabase establishes a connection to TiDB Cloud
func ConnectDatabase() (*sql.DB, error) {
	config, err := LoadDatabaseConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %v", err)
	}

	mysql.RegisterTLSConfig("tidb", &tls.Config{
		MinVersion: tls.VersionTLS12,
		ServerName: config.Host,
	})

	cfg := mysql.Config{
		User:                 config.Username,
		Passwd:               config.Password,
		Net:                  "tcp",
		Addr:                 fmt.Sprintf("%s:%s", config.Host, config.Port),
		DBName:               config.Database,
		TLSConfig:           "tidb",
		AllowNativePasswords: true,
		CheckConnLiveness:    true,
		MaxAllowedPacket:     4 << 20,
		Timeout:              30 * time.Second,
		ReadTimeout:          30 * time.Second,
		WriteTimeout:         30 * time.Second,
		ParseTime:            true,
		Loc:                  time.UTC,
	}

	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		return nil, fmt.Errorf("error opening database: %v", err)
	}

	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(1 * time.Minute)

	err = db.Ping()
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("error connecting to database: %v", err)
	}

	return db, nil
}

// Initialize connects to the database and stores it in the global DB variable
func Initialize() error {
	var err error
	DB, err = ConnectDatabase()
	return err
}

// Close closes the database connection
func Close() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

// EnsureConnection checks if the database connection is alive and reconnects if needed
func EnsureConnection() error {
	if DB == nil {
		return Initialize()
	}

	if err := DB.Ping(); err != nil {
		DB.Close()
		return Initialize()
	}

	return nil
}
