package repository

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"websckt/models"
)

type AuthRepository struct {
	db *sqlx.DB
}

func NewAuthRepository(db *sqlx.DB) *AuthRepository {
	return &AuthRepository{db: db}
}
func (r *AuthRepository) CreateUser(user models.User) (string, error) {
	var id string

	query := fmt.Sprintf("INSERT INTO %s (name,username,email,password_hash) values ($1,$2,$3,$4) RETURNING id", users)
	row := r.db.QueryRow(query, user.Name, user.Username, user.Email, user.Password)
	if err := row.Scan(&id); err != nil {
		return "", err
	}
	return id, nil
}
func (r *AuthRepository) GetUser(username, password, email string) (models.User, error) {
	var user models.User
	fmt.Println(username, " ", password, " ", email)
	query := fmt.Sprintf("SELECT id, username, email FROM %s WHERE username=$1 AND password_hash=$2 AND email=$3", users)
	err := r.db.Get(&user, query, username, password, email)
	if err != nil {
		logrus.Printf("user: %s Username or Password is incorrect", username)
		return user, err

	}
	logrus.Println(user.Username)
	return user, err
}
func (r *AuthRepository) GetUserByID(id string) (models.User, error) {
	var user models.User
	query := `SELECT id, username, email FROM users WHERE id = $1`
	err := r.db.Get(&user, query, id)
	return user, err
}
