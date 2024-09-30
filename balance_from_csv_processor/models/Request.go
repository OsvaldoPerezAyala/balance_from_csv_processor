package models

type Request struct {
	Email string `json:"email" validate:"required,email"`
}
