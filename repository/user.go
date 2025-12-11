package repository

import (
	"database/sql"

	"github.com/Kamekure-Maisuke/maiyumi/model"
)

type UserRepository interface {
	Create(username, password string) error
	FindByUsername(username string) (*model.User, error)
	GetID(username string) (int, error)
	UpdateUsername(userID int, newUsername string) error
	UpdatePassword(userID int, newPassword string) error
	FindByID(userID int) (*model.User, error)
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

func (r *userRepository) UpdateUsername(userID int, newUsername string) error {
	_, err := r.db.Exec("UPDATE users SET username = ? WHERE id = ?", newUsername, userID)
	return err
}

func (r *userRepository) UpdatePassword(userID int, newPassword string) error {
	_, err := r.db.Exec("UPDATE users SET password = ? WHERE id = ?", newPassword, userID)
	return err
}

func (r *userRepository) FindByID(userID int) (*model.User, error) {
	var user model.User
	err := r.db.QueryRow("SELECT id, username, password, created_at FROM users WHERE id = ?", userID).
		Scan(&user.ID, &user.Username, &user.Password, &user.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
