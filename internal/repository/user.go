package repository

import (
	"github.com/jmoiron/sqlx"
	"websckt/models"
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}
func (r *UserRepository) GetUserByID(userId string) (models.User, error) {
	var user models.User
	query := `SELECT email, name, username FROM users WHERE id = $1`
	err := r.db.Get(&user, query, userId)
	if err != nil {
		return user, err
	}
	return user, nil
}
