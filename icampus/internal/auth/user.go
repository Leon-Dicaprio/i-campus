package auth

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Username string `json:"username"`
	Password string `json:"password"` // Hashed password
}

type UserManager struct {
	db *sql.DB
}

func NewUserManager(dsn string) (*UserManager, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	// Ping to check connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	um := &UserManager{
		db: db,
	}

	if err := um.initDB(); err != nil {
		return nil, fmt.Errorf("failed to initialize database schema: %v", err)
	}

	return um, nil
}

func (um *UserManager) initDB() error {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		username VARCHAR(255) PRIMARY KEY,
		password VARCHAR(255) NOT NULL
	);`
	_, err := um.db.Exec(query)
	return err
}

// MigrateFromJson loads users from a JSON file and inserts them into the database
func (um *UserManager) MigrateFromJson(filePath string) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil // No file to migrate
	}

	file, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read migration file: %v", err)
	}

	if len(file) == 0 {
		return nil
	}

	var users map[string]User
	if err := json.Unmarshal(file, &users); err != nil {
		return fmt.Errorf("failed to unmarshal migration data: %v", err)
	}

	for _, user := range users {
		// Use INSERT IGNORE to skip existing users
		_, err := um.db.Exec("INSERT IGNORE INTO users (username, password) VALUES (?, ?)", user.Username, user.Password)
		if err != nil {
			log.Printf("Failed to migrate user %s: %v", user.Username, err)
		}
	}

	log.Printf("Migration completed from %s", filePath)
	return nil
}

func (um *UserManager) Register(username, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = um.db.Exec("INSERT INTO users (username, password) VALUES (?, ?)", username, string(hashedPassword))
	if err != nil {
		return fmt.Errorf("user already exists or registration failed: %v", err)
	}

	return nil
}

func (um *UserManager) Login(username, password string) error {
	var hashedPassword string
	err := um.db.QueryRow("SELECT password FROM users WHERE username = ?", username).Scan(&hashedPassword)
	if err != nil {
		if err == sql.ErrNoRows {
			return errors.New("invalid username or password")
		}
		return err
	}

	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

func (um *UserManager) Close() error {
	if um.db != nil {
		return um.db.Close()
	}
	return nil
}
