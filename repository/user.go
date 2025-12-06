package repository

import (
	"database/sql"

	"github.com/Kamekure-Maisuke/maiyumi/model"
)

type UserRepository interface {
	Create(username, password string) error
	FindByUsername(username string) (*model.User, error)
	GetID(username string) (int, error)
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(username, password string) error {
	_, err := r.db.Exec("INSERT INTO users (username, password) VALUES (?, ?)", username, password)
	return err
}

func (r *userRepository) FindByUsername(username string) (*model.User, error) {
	var user model.User
	err := r.db.QueryRow("SELECT id, username, password, created_at FROM users WHERE username = ?", username).
		Scan(&user.ID, &user.Username, &user.Password, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetID(username string) (int, error) {
	var id int
	err := r.db.QueryRow("SELECT id FROM users WHERE username = ?", username).Scan(&id)
	return id, err
}
