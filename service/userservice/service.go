package userservice

import (
	"fmt"

	"github.com/mfaxmodem/gameap/entity"
	"github.com/mfaxmodem/gameap/pkg/phonenumber"
	"golang.org/x/crypto/bcrypt"
)

type Repository interface {
	IsPhoneNumberUnique(phoneNumber string) (bool, error)
	Register(u entity.User) (entity.User, error)
	GetUserByPhoneNumber(phoneNumber string) (entity.User, bool, error)
}

type Service struct {
	repo Repository
}
type RegisterRequest struct {
	Name        string `json:"name"`
	PhoneNumber string `json:"phone_number"`
	Password    string `json:"password"`
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
	//TODO: check uniqueness REGEX
	//validate password
	if len(reg.Password) < 8 {
		return RegisterResponse{}, fmt.Errorf("length password should be greater than 8")
	}

	//TODO - Change password to bycrypt
	pass, err := bcrypt.GenerateFromPassword([]byte(reg.Password), bcrypt.DefaultCost)
	if err != nil {
		return RegisterResponse{}, fmt.Errorf("failed to hash password")
	}

	// create new user in storage

	// create new user in storage
	user := entity.User{
		ID:          0,
		PhoneNumber: reg.PhoneNumber,
		Name:        reg.Name,
		Password:    string(pass),
	}

	createUser, err := s.repo.Register(user)
	if err != nil {
		return RegisterResponse{}, fmt.Errorf("failed to create user")
	}
	return RegisterResponse{User: createUser}, nil

	// return created user

}

type LoginRequest struct {
	PhoneNumber string `json:"phone_number"`
	Password    string `json:"password"`
}

type LoginResponse struct {
}

func (s Service) Login(reg LoginRequest) (LoginResponse, error) {
	// get user from storage
	user, isExists, err := s.repo.GetUserByPhoneNumber(reg.PhoneNumber)
	if err != nil {
		return LoginResponse{}, fmt.Errorf("failed to get user")
	}
	if !isExists {
		return LoginResponse{}, fmt.Errorf("user not found")
	}

	// compare password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(reg.Password))
	if err != nil {
		return LoginResponse{}, fmt.Errorf("invalid password")
	}

	// return success
	return LoginResponse{}, nil
}
