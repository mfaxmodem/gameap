package main

import (
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/mfaxmodem/gameap/repository/mysql"
	"github.com/mfaxmodem/gameap/service/userservice"
)

var (
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
)

func init() {
	var err error
	privateKey, err = userservice.LoadPrivateKey("keys/private-key.pem")
	if err != nil {
		log.Printf("Error loading private key: %v", err)
		privateKey = nil
	}

	publicKey, err = userservice.LoadPublicKey("keys/public-key.pem")
	if err != nil {
		log.Printf("Error loading public key: %v", err)
		publicKey = nil
	}
}

func main() {

	mux := http.NewServeMux()
	mux.HandleFunc("/health-check", healthCheckHandler)
	mux.HandleFunc("/users/register", userRegisterHandler)
	mux.HandleFunc("/users/login", userLoginHandler)
	mux.HandleFunc("/users/profile", userProfileHandler)

	log.Printf("Server is listening on :8080...")
	server := http.Server{Addr: ":8080", Handler: mux}
	log.Fatal(server.ListenAndServe())
}

func healthCheckHandler(writer http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(writer, `{"message": "everything is OK!"}`)
}

func userRegisterHandler(writer http.ResponseWriter, req *http.Request) {

	data, err := io.ReadAll(req.Body)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(fmt.Sprintf(`{"error":"%s"}`, err.Error())))
		return
	}

	var uReq userservice.RegisterRequest
	err = json.Unmarshal(data, &uReq)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write([]byte(fmt.Sprintf(`{"error":"%s"}`, err.Error())))
		return
	}

	mysqlRepo := mysql.New()
	userSvc := userservice.New(mysqlRepo, privateKey)

	_, err = userSvc.Register(uReq)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(fmt.Sprintf(`{"error":"%s"}`, err.Error())))
		return
	}

	writer.WriteHeader(http.StatusCreated)
	writer.Write([]byte(`{"message": "User registered successfully"}`))
}

func userLoginHandler(writer http.ResponseWriter, req *http.Request) {
	var lReq userservice.LoginRequest
	if err := json.NewDecoder(req.Body).Decode(&lReq); err != nil {
		http.Error(writer, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusBadRequest)
		return
	}

	mysqlRepo := mysql.New()
	userSvc := userservice.New(mysqlRepo, privateKey)

	resp, err := userSvc.Login(lReq)
	if err != nil {
		http.Error(writer, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusUnauthorized)
		return
	}

	if err := json.NewEncoder(writer).Encode(resp); err != nil {
		http.Error(writer, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
	}
}

func userProfileHandler(writer http.ResponseWriter, req *http.Request) {
	auth := req.Header.Get("Authorization")
	claims, err := ParsJWT(auth)
	if err != nil {
		writer.WriteHeader(http.StatusUnauthorized)
		writer.Write([]byte(fmt.Sprintf(`{"error":"Invalid token: %v"}`, err)))
		return
	}

	mysqlRepo := mysql.New()

	userSvc := userservice.New(mysqlRepo, privateKey)
	resp, err := userSvc.Profile(userservice.ProfileRequest{UserID: claims.UserID})
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(fmt.Sprintf(`{"error":"%s"}`, err.Error())))
		return
	}

	data, err := json.Marshal(resp)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(fmt.Sprintf(`{"error":"%s"}`, err.Error())))
		return
	}

	writer.WriteHeader(http.StatusOK)
	writer.Write(data)
}

func ParsJWT(tokenStr string) (*userservice.Claims, error) {
	tokenStr = strings.TrimSpace(strings.Replace(tokenStr, "Bearer ", "", 1))
	token, err := jwt.ParseWithClaims(tokenStr, &userservice.Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return publicKey, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %v", err)
	}

	claims, ok := token.Claims.(*userservice.Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}
