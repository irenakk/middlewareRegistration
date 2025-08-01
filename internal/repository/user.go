package repository

import (
	"simpleRegistration/internal/config"
	"simpleRegistration/internal/models"
)

type InterfaceUserRepository interface {
	Store(user models.UserRegister) (int, error)
	Find(username string) (models.User, error)
	ExistsByUsername(username string) (bool, error)
}

type UserRepository struct {
	db *config.Database
}

func NewUserRepository(db *config.Database) InterfaceUserRepository {
	return &UserRepository{db}
}

// Storing data to database
func (r *UserRepository) Store(user models.UserRegister) (int, error) {
	var id int
	err := r.db.DB.QueryRow(`INSERT INTO users (username, password) VALUES ($1, $2) RETURNING id`,
		user.Username, user.Password).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// Find user from database
func (r *UserRepository) Find(username string) (models.User, error) {
	var user models.User
	err := r.db.DB.QueryRow(`SELECT id, username, password FROM users WHERE username = $1`, username).Scan(&user.ID, &user.Username, &user.Password)
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}

func (r *UserRepository) ExistsByUsername(username string) (bool, error) {
	var exists bool
	err := r.db.DB.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)", username).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}
