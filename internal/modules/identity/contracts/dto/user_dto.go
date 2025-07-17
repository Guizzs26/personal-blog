package dto

import (
	"time"

	"github.com/Guizzs26/personal-blog/internal/modules/identity/model"
)

type CreateUserRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=67,alpha"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=2"`
}

func (cur *CreateUserRequest) ToModel() model.User {
	return model.User{
		Name:     cur.Name,
		Email:    cur.Email,
		Password: cur.Password,
	}
}

type UserFullResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func ToUserFullResponse(user *model.User) UserFullResponse {
	return UserFullResponse{
		ID:        user.ID.String(),
		Name:      user.Name,
		Email:     user.Email,
		Active:    user.Active,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}
