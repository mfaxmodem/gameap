package userservice

import (
	"fmt"

	"github.com/mfaxmodem/gameap/entity"
	"github.com/mfaxmodem/gameap/pkg/phonenumber"
)

type Repository interface {
	IsPhoneNumberUnique(phoneNumber string) (bool, error)
	Register(u entity.User) (entity.User, error)
}

type Service struct {
	repo Repository
}
type RegisterRequest struct {
	Name        string
	PhoneNumber string
}
type RegisterResponse struct {
	User entity.User
}

func New(repo Repository) Service {
	return Service{repo: repo}
}

func (s Service) Register(reg RegisterRequest) (RegisterResponse, error) {

	// TODO: verify and sanitize phone number by verification code
	//validate phone number format
	if !phonenumber.IsValid(reg.PhoneNumber) {
		return RegisterResponse{}, fmt.Errorf("invalid phone number format")
	}

	//check uniqueness of phone number
	if isUnique, err := s.repo.IsPhoneNumberUnique(reg.PhoneNumber); err != nil || !isUnique {

		if err != nil {
			return RegisterResponse{}, fmt.Errorf("failed to check uniqueness: %v", err)
		}
		if !isUnique {
			return RegisterResponse{}, fmt.Errorf("phone number is already in use")
		}

	}

	// validate name format
	if len(reg.Name) < 3 || len(reg.Name) > 50 {
		return RegisterResponse{}, fmt.Errorf("invalid name format")
	}

	// create new user in storage
	user := entity.User{
		ID:          0,
		PhoneNumber: reg.PhoneNumber,
		Name:        reg.Name,
	}

	createUser, err := s.repo.Register(user)
	if err != nil {
		return RegisterResponse{}, fmt.Errorf("failed to create user")
	}
	return RegisterResponse{User: createUser}, nil

	// return created user

}
