package service

import (
	"websckt/internal/repository"
	"websckt/models"
)

type Authorization interface {
	CreateUser(user models.User) (string, error)
	GenerateToken(username, password, email string) (string, string, error)
	ParseRefToken(tokenR string) (string, error)
	ParseToken(token string) (string, error)
	GenerateAccToken(userId string) (string, error)
}
type Service struct {
	Authorization
}

func NewService(rep *repository.Repository) *Service {
	return &Service{
		Authorization: NewAuthService(rep.Authorization),
	}
}
