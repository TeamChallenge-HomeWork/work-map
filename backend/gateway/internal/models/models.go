package models

import (
	"errors"
	"net/mail"
)

type Validator interface {
	Validate() error
}

type User struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=4,max=16"`
}

func (u *User) Validate() error {
	//validate := validator.New()
	//
	//if err := validate.Struct(u); err != nil {
	//	return err
	//}

	if u.Email == "" || u.Password == "" {
		return errors.New("invalid request")
	}

	_, err := mail.ParseAddress(u.Email)
	if err != nil {
		return errors.New("invalid email")
	}

	return nil
}
