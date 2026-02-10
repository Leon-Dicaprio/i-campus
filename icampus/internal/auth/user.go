package auth

import (
	"encoding/json"
	"errors"
	"os"
	"sync"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Username string `json:"username"`
	Password string `json:"password"` // Hashed password
}

type UserManager struct {
	users    map[string]User
	filePath string
	mu       sync.RWMutex
}

func NewUserManager(filePath string) (*UserManager, error) {
	um := &UserManager{
		users:    make(map[string]User),
		filePath: filePath,
	}
	if err := um.load(); err != nil {
		// If file doesn't exist, that's fine, we'll create it on save
		if !os.IsNotExist(err) {
			return nil, err
		}
	}
	return um, nil
}

func (um *UserManager) load() error {
	um.mu.Lock()
	defer um.mu.Unlock()

	file, err := os.ReadFile(um.filePath)
	if err != nil {
		return err
	}

	if len(file) == 0 {
		return nil
	}

	return json.Unmarshal(file, &um.users)
}

func (um *UserManager) save() error {
	um.mu.RLock()
	defer um.mu.RUnlock()

	data, err := json.MarshalIndent(um.users, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(um.filePath, data, 0644)
}

func (um *UserManager) Register(username, password string) error {
	um.mu.Lock()
	defer um.mu.Unlock()

	if _, exists := um.users[username]; exists {
		return errors.New("user already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	um.users[username] = User{
		Username: username,
		Password: string(hashedPassword),
	}

	// Save immediately for persistence
	// Note: We need to release the lock before calling save if save also locks.
	// But here save() uses RLock, and we have Lock. RLock within Lock is fine? 
	// Actually, RLock blocks if Lock is held by *another* goroutine. 
	// But if the same goroutine holds Lock, RLock might deadlock or be allowed depending on implementation.
	// Standard Go sync.RWMutex: "If a goroutine holds a RWMutex for reading and another goroutine might call Lock, no goroutine should expect to be able to acquire a read lock until the initial read lock is released. In particular, this prohibits recursive read locking."
	// And "If a goroutine holds a RWMutex for writing, it is not allowed to acquire a read lock." -> deadlock!
	
	// So I should NOT call save() (which RLocks) inside Register() (which Locks).
	// I should make an internal save function that doesn't lock, or just write the file here.
	
	// Refactoring to use an internal save helper without locking.
	return um.saveWithoutLock()
}

func (um *UserManager) saveWithoutLock() error {
	data, err := json.MarshalIndent(um.users, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(um.filePath, data, 0644)
}

func (um *UserManager) Login(username, password string) error {
	um.mu.RLock()
	user, exists := um.users[username]
	um.mu.RUnlock()

	if !exists {
		return errors.New("invalid username or password")
	}

	return bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
}
