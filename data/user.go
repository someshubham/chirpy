package data

import "github.com/someshubham/chirpy/internal/database"

type User struct {
	ID        string `json:"id"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	Email     string `json:"email"`
}

func NewUserFromDB(usr database.User) User {
	return User{
		ID:        usr.ID.String(),
		CreatedAt: usr.CreatedAt.String(),
		UpdatedAt: usr.UpdatedAt.String(),
		Email:     usr.Email,
	}
}
