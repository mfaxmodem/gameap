package userservice

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/mfaxmodem/gameap/entity"
	"github.com/mfaxmodem/gameap/pkg/phonenumber"
	"golang.org/x/crypto/bcrypt"
)

type Repository interface {
	IsPhoneNumberUnique(phoneNumber string) (bool, error)
	Register(u entity.User) (entity.User, error)
	GetUserByPhoneNumber(phoneNumber string) (entity.User, bool, error)
	GetUserByID(userID uint) (entity.User, error)
}

type Service struct {
	signKey string
	repo    Repository
}

type RegisterRequest struct {
	Name        string `json:"name"`
	PhoneNumber string `json:"phone_number"`
	Password    string `json:"password"`
}

type RegisterResponse struct {
	User entity.User
}

func New(repo Repository, signKey string) Service {
	return Service{repo: repo, signKey: signKey}
}

func (s Service) Register(reg RegisterRequest) (RegisterResponse, error) {
	if !phonenumber.IsValid(reg.PhoneNumber) {
		return RegisterResponse{}, fmt.Errorf("invalid phone number format")
	}

	if isUnique, err := s.repo.IsPhoneNumberUnique(reg.PhoneNumber); err != nil || !isUnique {
		if err != nil {
			return RegisterResponse{}, fmt.Errorf("failed to check uniqueness: %v", err)
		}
		if !isUnique {
			return RegisterResponse{}, fmt.Errorf("phone number is already in use")
		}
	}

	if len(reg.Name) < 3 || len(reg.Name) > 50 {
		return RegisterResponse{}, fmt.Errorf("invalid name format")
	}

	if len(reg.Password) < 8 {
		return RegisterResponse{}, fmt.Errorf("length password should be greater than 8")
	}

	pass, err := bcrypt.GenerateFromPassword([]byte(reg.Password), bcrypt.DefaultCost)
	if err != nil {
		return RegisterResponse{}, fmt.Errorf("failed to hash password")
	}

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
}

type LoginRequest struct {
	PhoneNumber string `json:"phone_number"`
	Password    string `json:"password"`
}

type LoginResponse struct {
	AccessToken string `json:"access_token"`
}

func (s Service) Login(reg LoginRequest) (LoginResponse, error) {
	user, isExists, err := s.repo.GetUserByPhoneNumber(reg.PhoneNumber)
	if err != nil {
		return LoginResponse{}, fmt.Errorf("failed to get user")
	}
	if !isExists {
		return LoginResponse{}, fmt.Errorf("user not found")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(reg.Password))
	if err != nil {
		return LoginResponse{}, fmt.Errorf("invalid password")
	}

	privateKeyPath := "keys/private-key.pem" // مسیر صحیح را مشخص کنید
	privateKey, err := loadPrivateKey(privateKeyPath)
	if err != nil {
		return LoginResponse{}, fmt.Errorf("failed to load private key: %v", err)
	}
	token, err := createToken(user.ID, privateKey)
	if err != nil {
		return LoginResponse{}, fmt.Errorf("failed to create token: %v", err)
	}
	return LoginResponse{AccessToken: token}, nil
}

type ProfileRequest struct {
	UserID uint
}

type ProfileResponse struct {
	Name string `json:"name"`
}

func (s Service) Profile(reg ProfileRequest) (ProfileResponse, error) {
	user, err := s.repo.GetUserByID(reg.UserID)
	if err != nil {
		return ProfileResponse{}, fmt.Errorf("failed to get user")
	}
	return ProfileResponse{Name: user.Name}, nil
}

type Claims struct {
	jwt.RegisteredClaims
	UserID uint `json:"user_id"`
}

func loadPrivateKey(path string) (*rsa.PrivateKey, error) {
	keyData, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(keyData)
	if block == nil || block.Type != "PRIVATE KEY" {
		return nil, fmt.Errorf("failed to decode PEM block containing private key")
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	rsaPrivateKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA private key")
	}

	return rsaPrivateKey, nil
}

func createToken(userID uint, privateKey *rsa.PrivateKey) (string, error) {
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 7)),
		},
		UserID: userID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}
	return tokenString, nil
}
