package models

type User struct {
	Id       string `json:"-" db:"id"`
	Email    string `json:"email" binding:"required" db:"email"`
	Name     string `json:"name" binding:"required" db:"name"`
	Username string `json:"username" binding:"required" db:"username"`
	Password string `json:"password" binding:"required" db:"password_hash"`
}
