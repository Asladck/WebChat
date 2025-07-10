package repository

import (
	"github.com/jmoiron/sqlx"
	"websckt/models"
)

type Authorization interface {
	CreateUser(user models.User) (string, error)
	GetUser(username, password, email string) (models.User, error)
	GetUserByID(userId string) (models.User, error)
}
type Repository struct {
	Authorization
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{
		Authorization: NewAuthPostgres(db),
	}
}
