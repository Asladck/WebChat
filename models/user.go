package models

type User struct {
	Id       string `json:"-" db:"id"`
	Email    string `json:"email" binding:"required"`
	Name     string `json:"name" binding:"required"`
	Username string `json:"username" binding:"required"`
	Password string `json:"password_hash" binding:"required"`
}
