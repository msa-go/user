package models

type RegisterParameter struct {
	Name            string `json:"name"`
	Email           string `json:"email" binding:"required,email"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
}

type User struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role" gorm:"default:user"`
}

type LoginParameter struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}
