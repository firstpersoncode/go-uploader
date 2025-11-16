package repositories

import (
	"fmt"
	"sync"

	"firstpersoncode/go-uploader/domain"
	"firstpersoncode/go-uploader/internal/util"
)

type userRepository struct {
	mu    sync.RWMutex
	users map[string]*domain.User
}

func NewUserRepository() domain.UserRepository {
	return &userRepository{
		users: make(map[string]*domain.User),
	}
}

func (r *userRepository) Save(user *domain.User) (*domain.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if user.Username == "" {
		return nil, fmt.Errorf("user username is required")
	}

	for _, record := range r.users {
		if user.Username == record.Username {
			return nil, fmt.Errorf("username already exists")
		}
	}

	if user.Password == "" {
		return nil, fmt.Errorf("user password is required")
	}

	user.ID = util.GenerateRandomID()

	r.users[user.ID] = user
	return user, nil
}

func (r *userRepository) FindByID(id string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	user, exists := r.users[id]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}

	return user, nil
}

func (r *userRepository) FindByUsername(username string) (*domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, user := range r.users {
		if user.Username == username {
			return user, nil
		}
	}

	return nil, fmt.Errorf("user not found")
}
