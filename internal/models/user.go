package models

import "time"

type User struct {
	ID          int       `json:"id"`
	Login       string    `json:"login"`
	GameSurname string    `json:"gameSurname"`
	Email       string    `json:"email"`
	Password    string    `json:"-"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type UserResponse struct {
	ID          int    `json:"id"`
	Login       string `json:"login"`
	GameSurname string `json:"gameSurname"`
	Email       string `json:"email"`
}

// ToResponse преобразует User в UserResponse (без пароля)
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:          u.ID,
		Login:       u.Login,
		GameSurname: u.GameSurname,
		Email:       u.Email,
	}
}