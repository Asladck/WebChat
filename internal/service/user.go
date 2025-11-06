package service

import (
	"websckt/internal/repository"
	"websckt/models"
)

type UserService struct {
	repo repository.User
}

func NewUserService(repo repository.User) *UserService {
	return &UserService{repo: repo}
}
func (s *UserService) GetUserByID(userId string) (models.User, error) {
	user, err := s.repo.GetUserByID(userId)
	if err != nil {
		return user, err
	}
	return user, nil
}
