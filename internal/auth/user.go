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
	Password string `json:"password,omitempty"` // Hashed password, omitted in JSON by default
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	Avatar   string `json:"avatar"`
	Nickname string `json:"nickname"`
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
		password VARCHAR(255) NOT NULL,
		email VARCHAR(255) DEFAULT '',
		phone VARCHAR(20) DEFAULT '',
		avatar VARCHAR(255) DEFAULT '',
		nickname VARCHAR(255) DEFAULT ''
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
		_, err := um.db.Exec("INSERT IGNORE INTO users (username, password, email, phone, avatar, nickname) VALUES (?, ?, ?, ?, ?, ?)",
			user.Username, user.Password, user.Email, user.Phone, user.Avatar, user.Nickname)
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

	_, err = um.db.Exec("INSERT INTO users (username, password, nickname) VALUES (?, ?, ?)", username, string(hashedPassword), username)
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

func (um *UserManager) GetUser(username string) (*User, error) {
	var user User
	err := um.db.QueryRow("SELECT username, email, phone, avatar, nickname FROM users WHERE username = ?", username).
		Scan(&user.Username, &user.Email, &user.Phone, &user.Avatar, &user.Nickname)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

func (um *UserManager) UpdateUser(username string, email, phone, avatar, nickname string) error {
	_, err := um.db.Exec("UPDATE users SET email = ?, phone = ?, avatar = ?, nickname = ? WHERE username = ?",
		email, phone, avatar, nickname, username)
	return err
}

func (um *UserManager) UpdateEmail(username, email string) error {
	_, err := um.db.Exec("UPDATE users SET email = ? WHERE username = ?", email, username)
	return err
}

func (um *UserManager) UpdatePhone(username, phone string) error {
	_, err := um.db.Exec("UPDATE users SET phone = ? WHERE username = ?", phone, username)
	return err
}

func (um *UserManager) UpdateAvatar(username, avatar string) error {
	_, err := um.db.Exec("UPDATE users SET avatar = ? WHERE username = ?", avatar, username)
	return err
}

func (um *UserManager) UpdateNickname(username, nickname string) error {
	_, err := um.db.Exec("UPDATE users SET nickname = ? WHERE username = ?", nickname, username)
	return err
}

func (um *UserManager) Close() error {
	if um.db != nil {
		return um.db.Close()
	}
	return nil
}
