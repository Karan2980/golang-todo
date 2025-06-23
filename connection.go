package main

import (
	"crypto/tls"
	"database/sql"
	"fmt"
	"log"
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
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	config := &DatabaseConfig{
		Username: os.Getenv("TIDB_USERNAME"),
		Password: os.Getenv("TIDB_PASSWORD"),
		Host:     os.Getenv("TIDB_HOST"),
		Port:     os.Getenv("TIDB_PORT"),
		Database: os.Getenv("TIDB_DATABASE"),
	}

	// Set defaults
	if config.Port == "" {
		config.Port = "4000"
	}
	if config.Database == "" {
		config.Database = "test"
	}

	// Validate required fields
	if config.Username == "" || config.Password == "" || config.Host == "" {
		return nil, fmt.Errorf("missing required environment variables: username, password, and host must be set")
	}

	return config, nil
}

// ConnectDatabase establishes a connection to TiDB Cloud
func ConnectDatabase() (*sql.DB, error) {
	// Load configuration
	config, err := LoadDatabaseConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %v", err)
	}

	// Register TLS config for secure connection
	mysql.RegisterTLSConfig("tidb", &tls.Config{
		MinVersion: tls.VersionTLS12,
		ServerName: config.Host,
	})

	// Create MySQL config with improved settings and time parsing
	cfg := mysql.Config{
		User:                 config.Username,
		Passwd:               config.Password,
		Net:                  "tcp",
		Addr:                 fmt.Sprintf("%s:%s", config.Host, config.Port),
		DBName:               config.Database,
		TLSConfig:           "tidb",
		AllowNativePasswords: true,
		CheckConnLiveness:    true,
		MaxAllowedPacket:     4 << 20, // 4 MiB
		Timeout:              30 * time.Second,
		ReadTimeout:          30 * time.Second,
		WriteTimeout:         30 * time.Second,
		ParseTime:            true,  // This is crucial for time parsing
		Loc:                  time.UTC, // Set timezone to UTC
	}

	// Get the formatted DSN
	dsn := cfg.FormatDSN()

	// Open database connection
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %v", err)
	}

	// Set connection pool settings with better values for cloud database
	db.SetMaxOpenConns(5)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(1 * time.Minute)

	// Test the connection
	err = db.Ping()
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("error connecting to database: %v", err)
	}

	return db, nil
}

// InitializeDatabase connects to the database and stores it in the global DB variable
func InitializeDatabase() error {
	var err error
	DB, err = ConnectDatabase()
	if err != nil {
		return err
	}
	
	fmt.Println("Database connected successfully!")
	return nil
}

// CloseDatabase closes the database connection
func CloseDatabase() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}

// GetDatabase returns the database connection
func GetDatabase() *sql.DB {
	return DB
}

// EnsureConnection checks if the database connection is alive and reconnects if needed
func EnsureConnection() error {
	if DB == nil {
		log.Println("Database connection is nil, reconnecting...")
		return InitializeDatabase()
	}

	if err := DB.Ping(); err != nil {
		log.Printf("Database ping failed: %v, reconnecting...", err)
		DB.Close()
		return InitializeDatabase()
	}

	return nil
}

// GetTiDBVersion returns the TiDB version
func GetTiDBVersion() (string, error) {
	if err := EnsureConnection(); err != nil {
		return "", err
	}

	var version string
	err := DB.QueryRow("SELECT VERSION()").Scan(&version)
	if err != nil {
		return "", fmt.Errorf("error getting version: %v", err)
	}

	return version, nil
}

// GetCurrentDatabase returns the current database name
func GetCurrentDatabase() (string, error) {
	if err := EnsureConnection(); err != nil {
		return "", err
	}

	var currentDB string
	err := DB.QueryRow("SELECT DATABASE()").Scan(&currentDB)
	if err != nil {
		return "", fmt.Errorf("error getting current database: %v", err)
	}

	return currentDB, nil
}

// TestConnection tests if the database connection is working
func TestConnection() error {
	return EnsureConnection()
}
