package storage

import (
	"encoding/json"
	"errors"
	"os"
	"sync"

	"whats-app-lite/internal/models"

	"golang.org/x/crypto/bcrypt"
)

type UserManager struct {
	mu       sync.RWMutex
	filePath string
	users    map[string]string
}

func NewUserManager(filePath string) *UserManager {
	um := &UserManager{
		filePath: filePath,
		users:    make(map[string]string),
	}
	um.loadUsers()
	return um
}

func (um *UserManager) loadUsers() {
	um.mu.Lock()
	defer um.mu.Unlock()

	data, err := os.ReadFile(um.filePath)
	if err != nil {
		if !os.IsNotExist(err) {
			return
		}
		return
	}

	var users []models.User
	if err := json.Unmarshal(data, &users); err != nil {
		return
	}

	for _, user := range users {
		um.users[user.Username] = user.Password
	}
}

func (um *UserManager) saveUsers() error {
	um.mu.RLock()
	defer um.mu.RUnlock()

	users := make([]models.User, 0, len(um.users))
	for username, hashedPassword := range um.users {
		users = append(users, models.User{
			Username: username,
			Password: hashedPassword,
		})
	}

	data, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(um.filePath, data, 0644)
}

func (um *UserManager) Register(username, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	um.mu.Lock()
	if _, exists := um.users[username]; exists {
		um.mu.Unlock()
		return errors.New("username already exists")
	}

	um.users[username] = string(hashedPassword)
	um.mu.Unlock()

	return um.saveUsers()
}

func (um *UserManager) Authenticate(username, password string) error {
	um.mu.RLock()
	hashedPassword, exists := um.users[username]
	um.mu.RUnlock()

	if !exists {
		return errors.New("user not found")
	}

	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

func (um *UserManager) UserExists(username string) bool {
	um.mu.RLock()
	defer um.mu.RUnlock()
	_, exists := um.users[username]
	return exists
}
